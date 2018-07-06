package main

import (
	"testing"
)

//func runCommands(commands []string) int
//func checkFlags()
//func parseCommandLine(commandLine string, replacements map[string]string) (string, map[string]string, int)

func TestStripSlashes(t *testing.T) {
	expected := "@,@,"
	result := stripSlashes("\\@\\,\\@\\,")
	if result != expected {
		t.Errorf("stripSlashes failed. Got: %s, Want: %s", result, expected)
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
