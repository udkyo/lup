package main

import (
	"strings"
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

var hasGlobsTests = []struct {
	s string // supplied
	e bool   // expected
}{
	{"/tmp/*/", true},
	{"/tmp/*", true},
	{"/tmp/a?c/", true},
	{"/tmp/1!2/", true},
	{"/tmp/3{4/", true},
	{"/tmp/5}6/", true},
}

var isHiddenTests = []struct {
	s string // supplied
	e bool   // expected
}{
	{"-:Hello", true},
	{"-Goodbye", false},
	{"cat", false},
	{":-dog", false},
}

var isEscapedTests = []struct {
	s string // text
	n int    // position
	e bool   // expected result
}{
	{"He\\llo", 3, true},
	{"He\\llo", 4, false},
	{"He\\llo", 2, false},
	{"\\Hello", 1, true},
	{"Hell\\o", 5, true},
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
		//command = x.s
		command, replacements = parseCommand(x.s)
		expected := x.e
		result := generateCommands(command, make(map[string]string), 0, len(replacements))
		for n := range result {
			if result[n] != expected[n] {
				t.Errorf("generateCommands failed - Got: %s, Want: %s", result[n], expected[n])
			}
		}
	}
}

func TestHasGlobs(t *testing.T) {
	for _, x := range hasGlobsTests {
		result := hasGlobs(x.s)
		if result != x.e {
			t.Errorf("hasGlobs failed on '%s'", x.s)
		}
	}
}

func TestIsHidden(t *testing.T) {
	for _, x := range isHiddenTests {
		result := strings.HasPrefix(x.s, hider)
		if result != x.e {
			t.Errorf("isHidden failed on '%s'", x.s)
		}
	}
}

func TestIsEscaped(t *testing.T) {
	for _, x := range isEscapedTests {
		result := isEscaped(x.s, x.n)
		if result != x.e {
			t.Errorf("isEscaped failed on '%s:%d'", x.s, x.n)
		}
	}
}

// func TestGenerateCommands(t *testing.T) {
// 	s, m := parseCommand("")
// }

//func parseCommand(cmd string) (string, map[string]string)
