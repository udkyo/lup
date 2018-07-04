package main

import (
	"testing"
)

//func runCommands(commands []string) int
//func checkFlags()
//func parseCommandLine(commandLine string, replacements map[string]string) (string, map[string]string, int)

func TestGetCommandString(t *testing.T) {
	expected := "\"test test\" \"\""
	args := []string{"lup", "test test", ""}
	result := getCommandString(args)
	if result != expected {
		t.Errorf("getCommandString failed. Got: %s, Want: %s", result, expected)
	}
}

func TestStripSlashes(t *testing.T) {
	expected := "@,@,"
	result := stripSlashes("\\@\\,\\@\\,")
	if result != expected {
		t.Errorf("stripSlashes failed. Got: %s, Want: %s", result, expected)
	}
}

func TestParseCommandLine(t *testing.T) {
	//parseCommandLine(commandLine string) (string, map[string]string, int)
	//TODO: check replacements
	delimiter = '@'
	commandLine, _, numMaps := parseCommandLine(getCommandString([]string{"lup", "echo", "\"@hello,world@\""}))
	if numMaps != 1 {
		t.Errorf("parseCommandLine failed (numMaps) - Got: %d, Want: %d", numMaps, 1)
	}
	expected := "echo \"##0\""
	if commandLine != expected {
		t.Errorf("parseCommandLine failed (commandLine) - Got: [%s], Want: [%s]", commandLine, expected)
	}
}

func TestGenerateCommands(t *testing.T) {
	delimiter = '@'
	commandLine := "echo \"@hello,world@\""
	var replacements map[string]string
	var numMaps int
	//rep := make(map[string]string)
	commandLine, replacements, numMaps = parseCommandLine(commandLine)
	replacements = replacements
	expected := []string{"echo \"hello\"", "echo \"world\""}
	result := generateCommands(commandLine, replacements, make(map[string]string), 0, numMaps)
	for n, _ := range result {
		if result[n] != expected[n] {
			t.Errorf("generateCommands failed - Got: %s, Want: %s", result[n], expected[n])
		}
	}
}
