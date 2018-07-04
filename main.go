//go:generate go build
//go:generate cp ./lup /usr/local/bin

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var version string

func showHelp(force bool) {
	if len(os.Args) == 1 || force {
		fmt.Println(`
Usage: lup [OPTION] COMMANDLINE

Run multiple similar commands expanding ampersand encapsulated, comma-separated lists similarly to nested for loops.

     -h     Show this help message and exit
     -V     Show version information and exit
     -t     Show commands, but do not execute them  
`)
		os.Exit(0)
	}
}

// detectShell tries to detect the current shell, although  I'm fairly certain
// that hobbling through ps output and ham-fistedly checking for shell names
// isn't the best way of doing this
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
func getCommandLine() string {
	cmd := ""
	for i, arg := range os.Args[1:] {
		if strings.Contains(arg, " ") {
			cmd += fmt.Sprintf("\"%s\"", arg)
		} else if len(arg) == 0 {
			cmd += "\"\""
		} else {
			cmd += arg
		}
		if i < len(os.Args) {
			cmd += " "
		}
	}
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

// generateCommands generates the list of commands which will be executed
func generateCommands(str string, rep map[string]string, curTerms map[string]string, curMap int, numMaps int) []string {
	var commas []int
	var words []string
	var output []string
	mapName := fmt.Sprintf("##%d", curMap)
	rep[mapName] = stripCommas(rep[mapName])
	if regexp.MustCompile(`^/[0-9]+/$`).MatchString(rep[mapName]) {
		i, _ := strconv.Atoi(strings.Replace(rep[mapName], "/", "", -1))
		if i > curMap {
			log.Fatal(fmt.Sprintf("Forward references are not currently possible - switch @/%d/@ and the contents of group %d around", i, i))
		}
		refMap := fmt.Sprintf("##%d", i-1)

		if curMap == numMaps-1 {
			output = append(output, strings.Replace(str, mapName, curTerms[refMap], -1))
		} else {
			output = append(output, generateCommands(strings.Replace(str, mapName, curTerms[refMap], -1), rep, curTerms, curMap+1, numMaps)...)
		}
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
			output = append(output, generateCommands(strings.Replace(str, mapName, string(word), -1), rep, curTerms, curMap+1, numMaps)...)
		}
	} else {
		outerstr := str
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
	shell := detectShell()
	for _, command := range commands {
		command = strings.TrimRight(command, " ")
		if input != "" {
			command = "echo " + input + " | " + command
		}
		cmd := exec.Command(shell, "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			retcode = 1
		}
	}
	return retcode
}

func main() {
	version = "v1.0.0"

	replacements := make(map[string]string)
	commandLine := getCommandLine()
	dryRun := false
	switch firstWord := strings.Split(commandLine, " ")[0]; firstWord {
	case "-V":
		fmt.Println(version)
		os.Exit(0)
	case "-h":
		showHelp(true)
		os.Exit(0)
	case "-t":
		dryRun = true
		commandLine = strings.Join(strings.Split(commandLine, " ")[1:], " ")
	default:
		if strings.HasPrefix(firstWord, "-") {
			fmt.Printf("I don't recognise that flag (%s), try using lup -h to see the help\n", firstWord)
			os.Exit(2)
		}
	}
	capturing := -1
	word := ""
	count := 0

	delimiter := '@'
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

	if dryRun {
		for _, cmd := range generateCommands(commandLine, replacements, make(map[string]string), 0, count) {
			fmt.Println(cmd)
		}
	} else {
		if count > 0 {
			os.Exit(runCommands(generateCommands(commandLine, replacements, make(map[string]string), 0, count)))
		} else {
			os.Exit(runCommands([]string{commandLine}))
		}
	}
}
