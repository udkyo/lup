package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

type group struct {
	hidden bool
	terms  []string
	// paths which precede a group externally
	// are used by files/dirs/all directives
	externalPath string
}

type command struct {
	tokens   []string
	original string
	template string
	groups   []group
	commands []string
}

func newCommand(tokens ...string) (c command) {
	c.tokens = tokens
	c.checkFlags()
	c.getGroups()
	c.getCommands(0, "", []string{})
	return
}

func newGroup(s string, externalPath string, inSingles state, inDoubles state) (g group) {
	var terms []string
	var escaping state
	lastComma := -1
	g.hidden, s = isHidden(s)
	for i, char := range s {
		if !escaping.on && char == ',' {
			terms = append(terms, expand(stripSlashes(s[lastComma+1:i]), externalPath, inSingles.on, inDoubles.on)...)
			lastComma = i
		}
		escaping.toggle(char, '\\', true)
	}
	terms = append(terms, expand(stripSlashes(s[lastComma+1:len(s)]), externalPath, inSingles.on, inDoubles.on)...)
	g.externalPath = externalPath
	g.terms = terms
	return
}

func (c *command) getCommands(startGroup int, s string, curTerms []string) {
	var curTerm string
	if s == "" {
		s = c.template
	}
	if len(c.groups) == 0 {
		c.commands = append(c.commands, s)
		return
	}
	for _, t := range c.groups[startGroup].terms {
		if len(c.groups[startGroup].terms) == 1 {
			t = backref(t, curTerms)
		}
		if c.groups[startGroup].hidden {
			curTerm = ""
		} else {
			curTerm = t
		}
		if startGroup < len(c.groups)-1 {
			c.getCommands(startGroup+1, strings.Replace(s, lupGroup(startGroup), curTerm, 1), append(curTerms, t))
		} else {
			newCommand := strings.Replace(s, lupGroup(startGroup), curTerm, 1)
			c.commands = append(c.commands, newCommand)
		}
	}
}

func (c *command) getGroups() {
	var escaping bool
	var path string
	var curGroup int
	inSingles := state{on: false}
	inDoubles := state{on: false}
	pathStart := -1
	groupStart := -1
	for i, char := range c.original {
		if !escaping {
			if groupStart == -1 {
				if pathStart == -1 && char == '/' {
					pathStart = i
				}
				if pathStart > -1 && char == ' ' {
					pathStart = -1
				}
				inSingles.toggle(char, '\'', false)
				inDoubles.toggle(char, '"', false)
			}
			if char == '\\' {
				escaping = true
			} else if char == delimiter {
				if pathStart > -1 {
					path = c.original[pathStart:i]
				}
				pathStart = -1
				if groupStart == -1 {
					groupStart = i
				} else {
					c.groups = append(c.groups, newGroup(c.original[groupStart+1:i], path, inSingles, inDoubles))
					c.template = strings.Replace(c.template, c.original[groupStart-len(path):i+1], lupGroup(curGroup), 1)
					path = ""
					curGroup++
					groupStart = -1
				}
			}
		} else {
			if char != '\\' {
				escaping = false
			}
		}
	}
	c.template = stripSlashes(c.template)
}

func (c *command) checkFlags() {
	if len(c.tokens) > 0 {
		for i := 0; i < len(c.tokens); i++ {
			if !strings.HasPrefix(c.tokens[i], "-") {
				c.tokens = c.tokens[i:]
				c.original = shellquote.Join(c.tokens...)
				c.template = c.original
				break
			}
			switch c.tokens[i] {
			case "-V", "--version":
				fmt.Println(version)
				os.Exit(0)
			case "-h", "--help":
				showHelp()
				os.Exit(0)
			case "-t", "--test":
				dryRun = true
			default:
				fmt.Fprintf(os.Stderr, "Flag not recognised (%s), try using lup -h to see the help\n", c.tokens[i])
				os.Exit(2)
			}
		}
	}
}

func (c *command) run() int {
	var retcode int
	var cmd *exec.Cmd

	for _, command := range c.commands {
		switch shell {
		case "powershell.exe":
			command = "powershell -C " + strings.Replace(command, "\\!", "!", -1)
			if input != "" {
				command = shell + " -c \"Write-Host " + shellquote.Join(input) + " | " + command + "\""
			}
		case "cmd.exe":
			//todo
		default:
			if input != "" {
				command = shell + " -c \"echo " + shellquote.Join(input) + " | " + command + "\""
			}
		}
		if shell == "powershell" {
			if dryRun {
				fmt.Println("powershell -Command", command)
			} else {
				cmd = exec.Command("powershell", "-Command", command)
				cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
				if err := cmd.Run(); err != nil {
					retcode = 1
				}
			}
		} else {
			t, err := shellquote.Split(command)
			if err != nil {
				errOn(err, "Couldn't split command", 5)
			}
			if dryRun {
				fmt.Println(command)
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
				if err := cmd.Run(); err != nil {
					retcode = 1
				}
			}
		}
	}
	return retcode
}
