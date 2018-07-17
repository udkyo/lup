//go:generate go get github.com/udkyo/go-shellquote
//go:generate go test
//go:generate go build main.go expansions.go help.go
//go:generate mv ./main /usr/local/bin/lup

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

var (
	version      = "v0.3.2"
	delimiter    = '@'
	shell        = "sh"
	input        = ""
	command      = ""
	replacements map[string]string
	dryRun       = false
	errors       = false
	hider        = "-:"
	globChars    = []rune{'*', '?', '!', '{', '}'}
)

func errIf(err error) {
	if err != nil {
		errors = true
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func errOn(failed bool, msg string, returnCode int) {
	if failed {
		fmt.Fprintln(os.Stderr, "fatal:", msg)
		os.Exit(returnCode)
	}
}

func hasGlobs(text string) (b bool) {
	for _, c := range text {
		if runeIn(globChars, c) {
			b = true
		}
	}
	return b
}

func isHidden(text string) (b bool) {
	if strings.HasPrefix(text, hider) {
		b = true
	}
	return b
}

func isEscaped(text string, n int) (b bool) {
	if n > 0 && text[n-1] == '\\' {
		b = true
	}
	return b
}

func runeIn(group []rune, r rune) (b bool) {
	for _, x := range group {
		if string(x) == string(r) {
			b = true
		}
	}
	return b
}

func unescapeGlobChars(text string) string {
	for _, char := range globChars {
		text = strings.Replace(text, "\\"+string(char), string(char), -1)
	}
	return text
}

func splitBy(s string, r rune) []string {
	commas := getUnescapedIndices(s, r)
	words := splitByIndices(s, commas)
	return words
}

func getUnescapedIndices(text string, char rune) (commas []int) {
	var prev rune
	for i, c := range text {
		if c == rune(char) && i > 0 && prev != '\\' {
			commas = append(commas, i)
		}
		prev = c
	}
	return commas
}

func splitByIndices(text string, indices []int) (words []string) {
	var start int
	if isHidden(text) {
		start = 2
	}
	prevIndex := -1

	for _, i := range indices {
		words = append(words, text[prevIndex+1+start:i])
		prevIndex = i
		if start > 0 {
			start = 0
		}
	}
	words = append(words, text[prevIndex+1:len(text)])
	return words
}

func addSlashes(word string) string {
	delimiter := '@'
	word = strings.Replace(word, string(delimiter), "\\"+string(delimiter), -1)
	word = strings.Replace(word, ",", "\\,", -1)
	return word
}

func stripSlashes(word string) string {
	delimiter := '@'
	word = strings.Replace(word, "\\"+string(delimiter), string(delimiter), -1)
	word = strings.Replace(word, "\\,", ",", -1)
	return word
}

func stripCommas(word string) string {
	word = strings.TrimLeft(word, ",")
	if len(word) > 2 {
		if word[len(word)-1] == ',' && word[len(word)-2] != '\\' {
			word = strings.TrimRight(word, ",")
		}
	}
	return word
}

func getStdin() string {
	inp := ""
	// https://stackoverflow.com/a/22564526 didn't hurt here
	file := os.Stdin
	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "file.Stat()", err)
	}
	size := fi.Size()
	if size > 0 {
		data, _ := ioutil.ReadAll(os.Stdin)
		inp = string(data)
		if len(inp) > 0 {
			inp = inp[:len(string(data))-1]
		}
	}
	return inp
}

func detectShell() string {
	shells := [6]string{"bash", "tcsh", "csh", "ksh", "zsh", "sh"}
	cmd := exec.Command("ps")
	t, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't execute ps to detect shell - using sh")
		return "sh"
	}
	psOutput := string(t)
	scanner := bufio.NewScanner(strings.NewReader(psOutput))
	for scanner.Scan() {
		for _, w := range strings.Fields(scanner.Text()) {
			for _, s := range shells {
				if w == s || w == "-"+s || strings.HasSuffix(w, s) {
					return strings.Trim(w, "-")
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "Couldn't detect shell - using sh")
	return "sh"
}

func backref(cmdStr string, mapName string, curTerms map[string]string, curMap int, numMaps int) []string {
	var output []string
	if regexp.MustCompile(`^/[0-9]+/$`).MatchString(replacements[mapName]) || regexp.MustCompile(`^[0-9]+$`).MatchString(replacements[mapName]) {
		i, _ := strconv.Atoi(strings.Replace(replacements[mapName], "/", "", -1))
		errOn(i > curMap, fmt.Sprintf("Forward references are not currently possible - switch @%d@ and the contents of group %d around", i, i), 4)
		refMap := fmt.Sprintf("##%d", i-1)
		if curMap == numMaps-1 {
			output = append(output, strings.Replace(cmdStr, mapName, curTerms[refMap], -1))
		} else {
			output = append(output, generateCommands(strings.Replace(cmdStr, mapName, curTerms[refMap], -1), curTerms, curMap+1, numMaps)...)
		}
	}
	return output
}

func unwrap(text string) string {
	return expandLines(expandPaths(expandRanges(text)))
}

// generateCommands generates the list of commands which will be executed
func generateCommands(cmdStr string, curTerms map[string]string, curMap int, numMaps int) []string {
	var words []string
	var output []string
	mapName := fmt.Sprintf("##%d", curMap)
	r := stripCommas(replacements[mapName])
	output = backref(cmdStr, mapName, curTerms, curMap, numMaps)

	if len(output) > 0 {
		return output
	}

	words = splitBy(r, ',')
	if curMap < numMaps-1 {
		for _, word := range words {
			curTerms[mapName] = string(word)
			if strings.HasPrefix(r, "-:") {
				word = ""
			}
			output = append(output, generateCommands(strings.Replace(cmdStr, mapName, string(word), 1),
				curTerms, curMap+1, numMaps)...)
		}
	} else {
		outerstr := cmdStr
		for _, word := range words {
			if strings.HasPrefix(r, "-:") {
				word = ""
			}
			output = append(output, stripSlashes(strings.Replace(outerstr, mapName, string(word), -1)))
		}
	}
	return output
}

func runCommands(commands []string) int {
	var retcode int
	var cmd *exec.Cmd

	for _, command := range commands {
		if input != "" {
			command = shell + " -c \"echo " + shellquote.Join(input) + " | " + command + "\""
		}
		t, err := shellquote.Split(command)
		errIf(err)
		if dryRun {
			//null tokens were presumably once quotes, doubles are just a guess
			for i := range t {
				if t[i] == "" {
					t[i] = "\"\""
				}
			}
			fmt.Println(shellquote.Join(t...))
		} else if len(t) > 0 {
			if len(t) > 1 {
				cmd = exec.Command(t[0], t[1:]...)
			} else if len(t) == 1 {
				cmd = exec.Command(t[0])
			} else {
				showHelp()
				os.Exit(0)
			}
			cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
			err := cmd.Run()
			errIf(err)
		}
	}
	return retcode
}

func checkFlags() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-V", "--version":
			fmt.Println(version)
			os.Exit(0)
		case "-h", "--help":
			showHelp()
			os.Exit(0)
		case "-t", "--test":
			dryRun = true
			command = strings.Join(strings.Split(command, " ")[1:], " ")
		default:
			if strings.HasPrefix(os.Args[1], "-") {
				fmt.Fprintf(os.Stderr, "Flag not recognised (%s), try using lup -h to see the help\n", os.Args[1])
				os.Exit(2)
			}
		}
	}
}

// parseCommand goes through the string finding bits that need replaced
// building a replacements map and adding placeholders
func parseCommand(cmd string) (string, map[string]string) {
	var word string
	capturing := -1
	count := 0
	r := make(map[string]string)
	originalCommand := cmd
	for n, char := range cmd {
		if capturing == -1 && char == delimiter {
			if n == 0 || originalCommand[n-1] != '\\' {
				capturing = n + 1
			}
		} else if char == delimiter && capturing > 0 && originalCommand[n-1] != '\\' {
			ref := fmt.Sprintf("##%d", count)
			cmd = strings.Replace(cmd, string(delimiter)+word+string(delimiter), ref, 1)
			r[ref] = word
			capturing = -1
			word = ""
			count++
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
	}
	return cmd, r
}

func main() {
	input = getStdin()
	shell = detectShell()

	command = unwrap(shellquote.Join(os.Args[1:]...))

	checkFlags()
	command, replacements = parseCommand(command)

	if len(replacements) > 0 {
		triggers := generateCommands(command, make(map[string]string), 0, len(replacements))
		os.Exit(runCommands(triggers))
	} else {
		os.Exit(runCommands([]string{command}))
	}
}
