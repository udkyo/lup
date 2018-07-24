package main

import (
	"testing"
)

func TestDetectShell(t *testing.T) {
	if shell := detectShell(); shell != "go" {
		t.Errorf("Shell is %s, expected go (change this in shell_test.go for other shells)", shell)
	}
}

func TestShowHelp(t *testing.T) {
	testrun = true
	showHelp()
}
