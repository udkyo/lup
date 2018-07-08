package main

import (
	"testing"
)

var stripSlashesTests = []struct {
	s string // supplied
	e string // expected
}{
	{"\\@\\,\\@\\,", "@,@,"},
}
var generateCommandsTests = []struct {
	s string   // supplied
	e []string // expected
}{
	{"echo \"@hello,world@\"", []string{"echo \"hello\"", "echo \"world\""}},
}

// var Tests = []struct {
// 	s string // supplied
// 	e string // expected
// }{
// 	{"", ""},
// }

//func runCommands(commands []string) int
//func checkFlags()
//func parseCommandLine(commandLine string, replacements map[string]string) (string, map[string]string, int)

func TestStripSlashes(t *testing.T) {
	for _, x := range stripSlashesTests {
		result := stripSlashes(x.s)
		if result != x.e {
			t.Errorf("stripSlashes failed. Got: %s, Want: %s", result, x.e)
		}
	}
}

func TestGenerateCommands(t *testing.T) {
	delimiter = '@'
	for _, x := range generateCommandsTests {
		commandLine := x.s
		commandLine, replacements = parseCommand(commandLine)
		expected := x.e
		result := generateCommands(commandLine, make(map[string]string), 0, len(replacements))
		for n := range result {
			if result[n] != expected[n] {
				t.Errorf("generateCommands failed - Got: %s, Want: %s", result[n], expected[n])
			}
		}
	}
}

// func TestGenerateCommands(t *testing.T) {
// 	s, m := parseCommand("")
// }

//func parseCommand(cmd string) (string, map[string]string)
