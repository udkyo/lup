package main

import (
	"os"
	"testing"
)

var newGroupTests = []struct {
	s  string
	ep string
	is state
	id state
	e  group
	et []string
}{
	{
		s:  "hello,\\@well\\, goodbye\\@,farewell",
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: group{
			hidden:       false,
			externalPath: "",
			terms: []string{
				"hello",
				"\\@well\\, goodbye\\@",
				"farewell",
			},
		},
		et: []string{
			"hello",
			"@well, goodbye@",
			"farewell",
		},
	},
}

var getCommandsTests = []struct {
	s  []string
	ep string
	is state
	id state
	e  []group
	et []string
}{
	{
		s:  []string{"echo @hello,goodbye@"},
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: []group{{
			hidden:       false,
			externalPath: "",
			terms: []string{
				"hello",
				"goodbye",
				"farewell",
			}},
		},
		et: []string{
			"'echo hello'",
			"'echo goodbye'",
		},
	},
	{
		s:  []string{"echo @hello,goodbye@ @1@"},
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: []group{{
			hidden:       false,
			externalPath: "",
			terms: []string{
				"hello",
				"goodbye",
				"farewell",
			},
		}, {
			hidden:       false,
			externalPath: "",
			terms: []string{
				"1",
			},
		}},

		et: []string{
			"'echo hello hello'",
			"'echo goodbye goodbye'",
		},
	},
	{
		s:  []string{"echo hello"},
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: []group{{
			hidden:       false,
			externalPath: "",
			terms:        []string{},
		}},
		et: []string{"'echo hello'"},
	},
	{
		s:  []string{"echo @-:hello@"},
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: []group{{
			hidden:       false,
			externalPath: "",
			terms:        []string{""},
		}},
		et: []string{"'echo '"},
	},
	{
		s:  []string{"@hello,bonjour@,\\@well\\, goodbye\\@,farewell /usr/bin/ @files:/dev@files:*@"},
		ep: "",
		is: state{on: false},
		id: state{on: false},
		e: []group{
			{
				hidden:       false,
				externalPath: "",
				terms: []string{
					"",
				},
			},
		},
		et: []string{
			"'hello,@well, goodbye@,farewell /usr/bin/ fd'",
			"'bonjour,@well, goodbye@,farewell /usr/bin/ fd'",
		},
	},
}

func TestGetCommands(t *testing.T) {
	for _, x := range getCommandsTests {
		c := newCommand(x.s...)
		for i, y := range c.commands {
			if y != x.et[i] {
				t.Errorf("Failed TestGetCommands - expected %s, got %s", y, x.et[i])
			}
		}
	}
}

func TestNewGroup(t *testing.T) {
	for _, x := range newGroupTests {
		r := newGroup(x.s, x.ep, x.is, x.id)
		if x.ep == r.externalPath {
			for j, o := range r.terms {
				if o != x.et[j] {
					t.Errorf("Failed TestNewGroup - expected %s, got %s", o, x.et[j])
				}
			}
		}
	}
}

var checkFlagsTests = []struct {
	s  []string
	e  []group
	et []string
}{
	{
		s: []string{"-t", "echo", "hello", "world"},
	},
}

func TestCheckFlags(t *testing.T) {
	dryRun = true
	newCommand(checkFlagsTests[0].s...)
	if dryRun != true {
		t.Errorf("Failed TestCheckFlags")
	}
}

var runTests = []struct {
	s  []string
	e  []group
	et []string
}{
	{
		s: []string{"sh", "-c", "touch /tmp/luptests/runtest"},
	},
}

func TestRun(t *testing.T) {
	dryRun = false
	for _, r := range runTests {
		c := newCommand(r.s...)
		c.run()
		if _, err := os.Stat("/tmp/luptests/runtest"); os.IsNotExist(err) {
			t.Errorf("Failed TestRun: %s", r.s)
		} else {
			os.Remove("/tmp/luptests/runtest")
		}
	}
}
