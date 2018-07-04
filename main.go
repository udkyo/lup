//go:generate go test
//go:generate go build main.go
//go:generate cp ./main /usr/local/bin/lup
//in lieu of comprehensive tests, I'm going with this for now!
//go:generate lup echo "@Hello,Bonjour,Yo\\, wud up@ user\\@domain"
//go:generate lup echo @Hello,Bonjour,Yo\\, wud up@ user\\@domain
//go:generate lup echo '@Hello,Bonjour,Yo\, wud up@ user\@domain'
//go:generate lup echo "@hello,goodbye@ @old,new@ @world,friend@ (@1@ @/3/@)"

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"
)

var version string

var dryRun bool
var commandLine string
var delimiter rune

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

// getStdin grabs input so we can pipe it to generated commands if required
func getStdin() string {
	inp := ""
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inp += scanner.Text()
		}
	}
	return inp
}

// getCommandLine returns the current command line sans first word (always lup)
// and encapsulates spaced args with double quotes
func getCommandString(argList []string) string {
	cmd := ""
	for i, arg := range argList[1:] {
		if strings.Contains(arg, " ") {
			cmd += fmt.Sprintf("\"%s\"", arg)
		} else if len(arg) == 0 {
			cmd += "\"\""
		} else {
			cmd += arg
		}
		if i < len(argList)-1 {
			cmd += " "
		}
	}
	cmd = strings.TrimRight(cmd, " ")
	return cmd
}

// stripSlashes removes slashes from delimiters and commas
func stripSlashes(word string) string {
	delimiter := '@'
	word = strings.Replace(word, "\\"+string(delimiter), string(delimiter), -1)
	word = strings.Replace(word, "\\,", ",", -1)
	return word
}

// stripCommas removes commas from terms found at the start and end of a group
func stripCommas(word string) string {
	word = strings.TrimLeft(word, ",")

	if len(word) > 2 {
		if word[len(word)-1] == ',' && word[len(word)-2] != '\\' {
			word = strings.TrimRight(word, ",")
		}
	}
	return word
}

func backref(cmdStr string, rep map[string]string, mapName string, curTerms map[string]string, curMap int, numMaps int) []string {
	var output []string
	if regexp.MustCompile(`^/[0-9]+/$`).MatchString(rep[mapName]) || regexp.MustCompile(`^[0-9]+$`).MatchString(rep[mapName]) {
		i, _ := strconv.Atoi(strings.Replace(rep[mapName], "/", "", -1))
		if i > curMap {
			log.Fatal(fmt.Sprintf("Forward references are not currently possible - switch @/%d/@ and the contents of group %d around", i, i))
		}
		refMap := fmt.Sprintf("##%d", i-1)
		if curMap == numMaps-1 {
			output = append(output, strings.Replace(cmdStr, mapName, curTerms[refMap], -1))
		} else {
			output = append(output, generateCommands(strings.Replace(cmdStr, mapName, curTerms[refMap], -1), rep, curTerms, curMap+1, numMaps)...)
		}
	}
	return output
}

func expand(text string) string {
	expanded := ""
	capturing := -1
	word := ""
	for n, char := range text {
		if capturing == -1 && char == delimiter {
			if n == 0 || n > 0 && text[n-1] != '\\' {
				capturing = n + 1
				expanded = ""
				word = ""
			}
		} else if char == delimiter && capturing > -1 && text[n-1] != '\\' {
			re := regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)$`)
			res := re.FindAllStringSubmatch(word, -1)
			if len(res) > 0 {
				first, _ := strconv.Atoi(res[0][1])
				last, _ := strconv.Atoi(res[0][2])
				for i := first; i <= last; i++ {
					expanded += strconv.Itoa(i)
					if i < last {
						expanded += ","
					}
				}
				text = strings.Replace(text, word, expanded, 1)
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
	}
	return text
}

// generateCommands generates the list of commands which will be executed
func generateCommands(cmdStr string, rep map[string]string, curTerms map[string]string, curMap int, numMaps int) []string {
	var commas []int
	var words []string
	var output []string
	mapName := fmt.Sprintf("##%d", curMap)
	rep[mapName] = stripCommas(rep[mapName])

	output = backref(cmdStr, rep, mapName, curTerms, curMap, numMaps)

	if len(output) > 0 {
		return output
	}

	for i := range rep[mapName] {
		if rep[mapName][i] == ',' && i > 0 && rep[mapName][i-1] != '\\' {
			commas = append(commas, i)
		}
	}
	previousComma := -1
	for _, i := range commas {
		words = append(words, rep[mapName][previousComma+1:i])
		previousComma = i
	}

	words = append(words, rep[mapName][previousComma+1:len(rep[mapName])])
	if curMap < numMaps-1 {
		for _, word := range words {
			curTerms[mapName] = string(word)
			output = append(output, generateCommands(strings.Replace(cmdStr, mapName, string(word), -1), rep, curTerms, curMap+1, numMaps)...)
		}
	} else {
		outerstr := cmdStr
		for _, word := range words {
			output = append(output, stripSlashes(strings.Replace(outerstr, mapName, string(word), -1)))
		}
	}

	return output
}

// runCommands triggers the commands which have been generated
func runCommands(commands []string) int {
	retcode := 0
	input := getStdin()
	var winCmd []string
	for _, command := range commands {
		winCmd = winCmd[:0]
		if input != "" {
			command = "echo " + input + " | " + command
		}
		t, _ := shellquote.Split(command)
		if runtime.GOOS == "windows" {
			winCmd = append(winCmd, "cmd")
			winCmd = append(winCmd, "/C")
			winCmd = append(winCmd, t...)
			t = winCmd
		}
		if len(t) > 1 {
			cmd := exec.Command(t[0], t[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				retcode = 1
			}
		} else {
			showHelp(true)
			os.Exit(0)
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
			showHelp(true)
			os.Exit(0)
		case "-t", "--test":
			dryRun = true
			commandLine = strings.Join(strings.Split(commandLine, " ")[1:], " ")
		default:
			if strings.HasPrefix(os.Args[1], "-") {
				fmt.Printf("Flag not recognised (%s), try using lup -h to see the help\n", os.Args[1])
				os.Exit(2)
			}
		}
	}
}

func parseCommandLine(commandLine string) (string, map[string]string, int) {
	var replacements map[string]string
	capturing := -1
	word := ""
	count := 0
	replacements = make(map[string]string)
	originalCommand := commandLine
	for n, char := range commandLine {
		if capturing == -1 && char == delimiter {
			if n == 0 || n > 0 && originalCommand[n-1] != '\\' {
				capturing = n + 1
			}
		} else if char == delimiter && capturing > -1 && originalCommand[n-1] != '\\' {
			ref := fmt.Sprintf("##%d", count)
			commandLine = strings.Replace(commandLine, string(delimiter)+word+string(delimiter), ref, 1)
			replacements[ref] = word
			capturing = -1
			word = ""
			count++
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
	}
	return commandLine, replacements, count
}

func main() {
	version = "v0.1.2"
	delimiter = '@'
	dryRun = false
	replacements := make(map[string]string)

	commandLine = getCommandString(os.Args)
	commandLine = expand(commandLine)
	checkFlags()

	commandLine, replacements, numMaps := parseCommandLine(commandLine)

	if dryRun {
		for _, cmd := range generateCommands(commandLine, replacements, make(map[string]string), 0, numMaps) {
			fmt.Println(cmd)
		}
	} else {
		if numMaps > 0 {
			os.Exit(runCommands(generateCommands(commandLine, replacements, make(map[string]string), 0, numMaps)))
		} else {
			os.Exit(runCommands([]string{commandLine}))
		}
	}
}
