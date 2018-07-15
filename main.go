//go:generate go get github.com/udkyo/go-shellquote
//go:generate go test
//go:generate go build main.go
//go:generate mv ./main /usr/local/bin/lup

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

var (
	version      = "v0.3.0"
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
		fmt.Println(err)
		os.Exit(1)
	}
}

func errOn(failed bool, msg string, returnCode int) {
	if failed {
		fmt.Println("fatal:", msg)
		os.Exit(returnCode)
	}
}

func hasGlobs(text string) bool {
	var m bool
	for _, c := range text {
		if runeIn(globChars, c) {
			m = true
		}
	}
	return m
}

func isHidden(text string) bool {
	if strings.HasPrefix(text, hider) {
		return true
	}
	return false
}

func isEscaped(text string, n int) bool {
	if n > 0 && text[n-1] == '\\' {
		return true
	}
	return false
}

func runeIn(group []rune, r rune) bool {
	for _, x := range group {
		if string(x) == string(r) {
			return true
		}
	}
	return false
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

func getUnescapedIndices(text string, char rune) []int {
	var commas []int
	var prev rune
	for i, c := range text {
		if c == rune(char) && i > 0 && prev != '\\' {
			commas = append(commas, i)
		}
		prev = c
	}
	return commas
}

func splitByIndices(text string, indices []int) []string {
	var words []string
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

func showHelp(force bool) {
	if len(os.Args) == 1 || force {
		fmt.Println(`Usage: lup [OPTION] COMMANDLINE

Run multiple similar commands expanding ampersand encapsulated, comma-separated lists similarly to nested for loops.

e.g:

  lup @rm,nano@ foo_@1,2@

Expands to and executes:

  rm foo_1
  rm foo_2
  nano foo_1
  nano foo_2

More usage info is available at https://github.com/udkyo/lup

Options:

     -h, --help     Show this help message and exit
     -V, --version  Show version information and exit
     -t, --test     Show commands, but do not execute them`)
		os.Exit(0)
	}
}

func getStdin() string {
	inp := ""
	// https://stackoverflow.com/a/22564526 didn't hurt here
	file := os.Stdin
	fi, err := file.Stat()
	if err != nil {
		fmt.Println("file.Stat()", err)
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
		log.Println("Can't execute ps to detect shell - using sh")
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
	fmt.Println("Couldn't detect shell - using sh")
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
	expanded := ""
	capturing := -1
	word := ""
	newtext := text
	pathStart := -1
	path := ""

	for n, char := range text {
		escaped := isEscaped(text, n)

		if capturing == -1 {
			if pathStart == -1 && char == '/' && !escaped {
				pathStart = n
			}
		}
		if char == delimiter && !escaped {
			if pathStart > -1 {
				path = text[pathStart:n]
			}
			pathStart = -1
		}

		if (capturing == -1 && char == delimiter) && !escaped {
			capturing = n + 1
			expanded = ""
			word = ""
		} else if char == delimiter || char == ',' && capturing > -1 && !escaped {
			if hasGlobs(path) {
				newtext = strings.Replace(newtext, path, "", 1)
			}

			start := 0
			prefix := ""
			if isHidden(word) {
				start = 2
				prefix = "-:"
			}

			re := regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)`)
			res := re.FindAllStringSubmatch(word[start:], -1)

			for m, x := range res {
				if len(x) == 0 {
					break
				}
				first, _ := strconv.Atoi(res[m][1])
				last, _ := strconv.Atoi(res[m][2])

				if last < first {
					for i := first; i >= last; i-- {
						expanded += strconv.Itoa(i)
						if i > last {
							expanded += ","
						}
					}
					newtext = strings.Replace(newtext, word, prefix+expanded, 1)
				} else if first < last {
					for i := first; i <= last; i++ {
						expanded += strconv.Itoa(i)
						if i < last {
							expanded += ","
						}
					}
					newtext = strings.Replace(newtext, word, prefix+expanded, 1)
				} else {
					errOn(first == last, "Integer range starts and ends on the same number, please check", 3)
				}
			}

			for _, marker := range []string{"files", "dirs", "all"} {
				if strings.HasPrefix(word, marker+":") || strings.HasPrefix(word, "-:"+marker+":") {
					replacement := getNodes(strings.Replace(word, marker+":", "", 1), path, marker)
					newtext = strings.Replace(newtext, word, replacement, 1)
				}
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
		if char == ',' && n > 0 && !escaped {
			capturing = n
			word = ""
			expanded = ""
		}
	}
	return newtext
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

// runCommands triggers the commands which have been generated
// and does a bunch of other stuff that really has no business
// being in this function but hey, time is in short supply
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
			//null tokens were presumably once quotes
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
				showHelp(true)
				os.Exit(0)
			}
			cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
			err := cmd.Run()
			errIf(err)
		}
	}
	return retcode
}

// checkFlags checks if the first token is a flag and does the needful
func checkFlags() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-V", "--version":
			fmt.Println(version)
			os.Exit(0)
		case "-h", "--help":
			showHelp(true)
			os.Exit(0)
		case "-t", "--test":
			dryRun = true
			command = strings.Join(strings.Split(command, " ")[1:], " ")
		default:
			if strings.HasPrefix(os.Args[1], "-") {
				fmt.Printf("Flag not recognised (%s), try using lup -h to see the help\n", os.Args[1])
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

func getNodes(directivePath string, externalPath string, kind string) string {
	var nodes []string
	var prefix string
	var tainted bool
	var rel bool

	if isHidden(directivePath) {
		directivePath = directivePath[2:]
		prefix = hider
	}
	if !strings.HasPrefix(directivePath, "/") {
		rel = true
	}
	errOn(!rel && externalPath != "", "fatal: don't mix and match importing paths from outside @@ blocks and absolute paths inside them, no bueno (map: "+directivePath+")", 5)

	if hasGlobs(externalPath) {
		tainted = true
	}
	directivePath = unescapeGlobChars(directivePath)
	fullPath := directivePath
	if externalPath != "" {
		if !strings.HasSuffix(externalPath, "/") || isEscaped(externalPath, len(externalPath)-1) {
			fullPath = externalPath + "/" + directivePath
		} else {
			fullPath = externalPath + directivePath
		}
	}
	contents, err := filepath.Glob(fullPath)
	errIf(err)
	errOn(len(contents) == 0, "No nodes matched ("+kind+": "+directivePath+")", 3)
	for _, f := range contents {
		nodeStat, err := os.Stat(f)
		errIf(err)
		if tainted {
			f = strings.Replace(f, " ", "\\ ", -1)
		} else {
			if filepath.Dir(f) != "." && rel {
				s, err := filepath.Rel(externalPath, f)
				errIf(err)
				f = s
			} else {
				f = strings.Replace(filepath.Base(f), " ", "\\ ", -1)
			}
		}
		switch mode := nodeStat.Mode(); {
		case mode.IsDir() && (kind == "all" || kind == "dirs"):
			nodes = append(nodes, f)
		case mode.IsRegular() && (kind == "all" || kind == "files"):
			nodes = append(nodes, f)
		}
	}
	return prefix + strings.Join(nodes, ",")
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
