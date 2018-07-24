//go:generate go get github.com/kballard/go-shellquote
//go:generate go build main.go expansions.go shell.go command.go strings.go
//go:generate sh -c "GOOS=windows GOARCH=amd64 go build main.go expansions.go shell.go command.go strings.go"
//go:generate mv ./main /usr/local/bin/lup
//go:generate mv ./main.exe lup.exe

package main

import (
	"fmt"
	"os"
)

var (
	version   = "v0.4.0"
	shell     string
	input     string
	dryRun    = false
	delimiter = '@'
	hider     = "-:"
	testrun   = false
)

func main() {
	shell = detectShell()
	input = getStdin()
	c := newCommand(os.Args[1:]...)
	r := c.run()
	if !testrun {
		os.Exit(r)
	}
}

func errOn(e error, s string, r int) {
	fmt.Fprintln(os.Stderr, s)
	fmt.Fprintf(os.Stderr, "  - %s", e)
	os.Exit(r)
}
