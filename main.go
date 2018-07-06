//go:generate go get github.com/udkyo/go-shellquote
//go:generate go test
//go:generate go build main.go
//go:generate cp ./main /usr/local/bin/lup
//in lieu of comprehensive tests, I'm going with this for now!
//go:generate lup echo "@Hello,Bonjour,Yo\\, wud up@ user\\@domain"
//go:generate lup echo '@Hello,Bonjour,Yo\, wud up@ user\@domain'
//go:generate lup echo "@hello,goodbye@ @old,new@ @world,friend@ (@1@ @/3/@)"
//go:generate sh -c "echo 'hello' | lup cat @-n,-e@"
//go:generate lup @-:1..3@ echo "Iteration @1@"
//go:generate lup sh -c "echo @1..3@"
//go:generate lup echo virsh @destroy,start@ @dev,test@_@1..3@
//go:generate lup echo 'this\"works'
//go:generate lup echo "this\"too"
//go:generate lup echo '"hello world"'
//go:generate lup echo "quoted" 'quoted' "\"nested\"" "'nested'" '"nested"'
//go:generate echo "MATCH FILES:"
//go:generate lup echo "  - @match_files:/tmp/*@"
//go:generate echo "MATCH DIRS"
//go:generate lup echo "  - @match_dirs:/tmp/*@"
//go:generate echo "MATCH ALL"
//go:generate lup echo "  - @match_all:/tmp/*@"

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
	"runtime"
	"strconv"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

var (
	version     = "v0.1.4"
	dryRun      = false
	commandLine = ""
	delimiter   = '@'
	shell       = "sh"
)

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

// addSlashes escaped delimiters and commas
func addSlashes(word string) string {
	delimiter := '@'
	word = strings.Replace(word, string(delimiter), "\\"+string(delimiter), -1)
	word = strings.Replace(word, ",", "\\,", -1)
	return word
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

// backref injects the term from a previous group
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

// expand expands numeric ranges and paths
func expand(text string) string {
	expanded := ""
	capturing := -1
	word := ""
	for n, char := range text {
		if capturing == -1 && (char == delimiter || char == ',') {
			if n == 0 || n > 0 && text[n-1] != '\\' {
				capturing = n + 1
				expanded = ""
				word = ""
			}
		} else if (char == delimiter || char == ',') && capturing > -1 && text[n-1] != '\\' {
			// expand files/dirs
			if strings.HasPrefix(word, "match_files:") {
				text = strings.Replace(text, word, getNodes(strings.Replace(word, "match_files:", "", 1), "files"), 1)
			} else if strings.HasPrefix(word, "match_dirs:") {
				text = strings.Replace(text, word, getNodes(strings.Replace(word, "match_dirs:", "", 1), "dirs"), 1)
			} else if strings.HasPrefix(word, "match_all:") {
				text = strings.Replace(text, word, getNodes(strings.Replace(word, "match_all:", "", 1), "all"), 1)
			} else {
				// expand ranges
				re := regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)$`)
				res := re.FindAllStringSubmatch(word, -1)
				if len(res) > 0 {
					first, _ := strconv.Atoi(res[0][1])
					last, _ := strconv.Atoi(res[0][2])
					if last < first {
						for i := first; i >= last; i-- {
							expanded += strconv.Itoa(i)
							if i > last {
								expanded += ","
							}
						}
						text = strings.Replace(text, word, expanded, 1)
					} else if first < last {
						for i := first; i <= last; i++ {
							expanded += strconv.Itoa(i)
							if i < last {
								expanded += ","
							}
						}
						text = strings.Replace(text, word, expanded, 1)
					} else if first == last {
						log.Fatal("Integer range starts and ends on the same number, please check")
					}
				}
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
		if char == ',' && n > 0 && text[n-1] != '\\' {
			capturing = n
			word = ""
			expanded = ""
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
	start := 0
	if rep[mapName][:2] == "-:" {
		start = 2
	}
	for _, i := range commas {
		words = append(words, rep[mapName][previousComma+1+start:i])
		previousComma = i
		if start > 0 {
			start = 0
		}
	}
	words = append(words, rep[mapName][previousComma+1:len(rep[mapName])])

	if curMap < numMaps-1 {
		for _, word := range words {
			curTerms[mapName] = string(word)
			if strings.HasPrefix(rep[mapName], "-:") {
				word = ""
			}
			output = append(output, generateCommands(strings.Replace(cmdStr, mapName, string(word), 1),
				rep, curTerms, curMap+1, numMaps)...)
		}
	} else {
		outerstr := cmdStr
		for _, word := range words {
			if strings.HasPrefix(rep[mapName], "-:") {
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
	retcode := 0
	input := getStdin()
	var winCmd []string
	var cmd *exec.Cmd

	for _, command := range commands {
		winCmd = winCmd[:0]
		if input != "" {
			command = shell + " -c \"echo " + shellquote.Join(input) + " | " + command + "\""
		}
		t, _ := shellquote.Split(command)
		if runtime.GOOS == "windows" {
			winCmd = append(winCmd, "cmd")
			winCmd = append(winCmd, "/C")
			winCmd = append(winCmd, t...)
			t = winCmd
		}
		if dryRun {
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
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				retcode = 1
			}
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
			commandLine = strings.Join(strings.Split(commandLine, " ")[1:], " ")
		default:
			if strings.HasPrefix(os.Args[1], "-") {
				fmt.Printf("Flag not recognised (%s), try using lup -h to see the help\n", os.Args[1])
				os.Exit(2)
			}
		}
	}
}

// parseCommandLine goes through the string finding bits that need replaced
// building a replacements map and adding placeholders
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

func getNodes(path string, kind string) string {
	var files []string
	var globChars = []string{"*", "?", "!", "{", "}"}
	for _, char := range globChars {
		path = strings.Replace(path, "\\"+char, char, -1)
	}
	contents, err := filepath.Glob(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range contents {
		fi, err := os.Stat(f)
		if err != nil {
			log.Fatal(err)
		}
		f = "\"" + f + "\""
		switch mode := fi.Mode(); {
		case mode.IsDir() && (kind == "all" || kind == "dirs"):
			files = append(files, f)
		case mode.IsRegular() && (kind == "all" || kind == "files"):
			files = append(files, f)
		}
	}
	return strings.Join(files, ",")
}

func main() {
	//fmt.Println(getNodes("/tmp/1", "dirs"))
	shell = detectShell()
	replacements := make(map[string]string)

	commandLine = shellquote.Join(os.Args[1:]...)
	commandLine = expand(commandLine)
	checkFlags()

	commandLine, replacements, numMaps := parseCommandLine(commandLine)
	if numMaps > 0 {
		os.Exit(runCommands(generateCommands(commandLine, replacements, make(map[string]string), 0, numMaps)))
	} else {
		os.Exit(runCommands([]string{commandLine}))
	}
}
