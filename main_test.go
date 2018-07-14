package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

var stripSlashesTests = []struct {
	s string // supplied
	e string // expected
}{
	{"Test \\@1\\, both", "Test @1, both"},
	{"Test\\@ at", "Test@ at"},
	{"Test \\,comma", "Test ,comma"},
}

var addSlashesTests = []struct {
	s string // supplied
	e string // expected
}{
	{"Test @1, both", "Test \\@1\\, both"},
	{"Test@ at", "Test\\@ at"},
	{"Test ,comma", "Test \\,comma"},
}

var stripCommasTests = []struct {
	s string // supplied
	e string // expected
}{
	{",beginning,and,end,", "beginning,and,end"},
	{",just,the,start", "just,the,start"},
	{"one,right,", "one,right"},
	{"multi,right,,,", "multi,right"},
	{",,,multi,left", "multi,left"},
}

var generateCommandsTests = []struct {
	s string   // supplied
	e []string // expected
}{
	{"echo \"@hello,world@\"", []string{"echo \"hello\"", "echo \"world\""}},
	{"echo \"@1,2@ @3,4@\"", []string{"echo \"1 3\"", "echo \"1 4\"", "echo \"2 3\"", "echo \"2 4\""}},
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

var runeInTests = []struct {
	r rune // text
	e bool // expected
}{
	{'a', false},
	{'x', false},
	{'?', true},
	{'{', true},
	{'}', true},
	{'!', true},
	{'*', true},
}

var unescapeGlobCharsTests = []struct {
	s string // text
	e string // expected
}{
	{"/tmp/\\*/", "/tmp/*/"},
	{"/tmp/*", "/tmp/*"},
	{"/tmp/a\\?c/", "/tmp/a?c/"},
	{"/tmp/1\\!2/", "/tmp/1!2/"},
	{"/tmp/3\\{4/", "/tmp/3{4/"},
	{"/tmp/5\\}6/", "/tmp/5}6/"},
	{"/t\\*\\\\m\\?\\{\\!\\?p/5\\}6/", "/t*\\\\m?{!?p/5}6/"},
}

var splitByTests = []struct {
	s string   // text
	r rune     // split on
	e []string // expected
}{
	{"hello,world", ',', []string{"hello", "world"}},
	{"this\\,is,a,test", ',', []string{"this,is", "a", "test"}},
}

var getNodesTests = []struct {
	dp string // text
	ep string
	k  string
	e  string // expected
}{
	{"/tmp/luptests/getnodes/*", "", "files", "0,1,2,3,4,5"},
	{"/tmp/luptests/getnodes/*", "", "dirs", "a"},
	{"/tmp/luptests/getnodes/*", "", "all", "0,1,2,3,4,5,a"},
	{"*", "/tmp/luptests/getnodes/", "files", "0,1,2,3,4,5"},
	{"*", "/tmp/luptests/getnodes/", "dirs", "a"},
	{"*", "/tmp/luptests/getnodes/", "all", "0,1,2,3,4,5,a"},
	{"*", "/tmp/luptests/getn?des/", "files", "/tmp/luptests/getnodes/0,/tmp/luptests/getnodes/1,/tmp/luptests/getnodes/2,/tmp/luptests/getnodes/3,/tmp/luptests/getnodes/4,/tmp/luptests/getnodes/5"},
	{"*", "/tmp/luptests/getn?des/", "dirs", "/tmp/luptests/getnodes/a"},
	{"*", "/tmp/luptests/getn?des/", "all", "/tmp/luptests/getnodes/0,/tmp/luptests/getnodes/1,/tmp/luptests/getnodes/2,/tmp/luptests/getnodes/3,/tmp/luptests/getnodes/4,/tmp/luptests/getnodes/5,/tmp/luptests/getnodes/a"},
	{"getnodes/*", "/tmp/luptests/", "all", "getnodes/0,getnodes/1,getnodes/2,getnodes/3,getnodes/4,getnodes/5,getnodes/a"},
}

var backrefTests = []struct {
	s            string // text
	mapName      string
	curTerms     map[string]string
	curMap       int
	numMaps      int
	replacements map[string]string
	e            []string
}{
	{"hello world ##1", "##1", map[string]string{"##0": "blah"}, 1, 2, map[string]string{"##0": "hello,world", "##1": "1"}, []string{"hello world blah"}},
}

var unwrapTests = []struct {
	s string // supplied
	e string // expected
}{
	{"@1..5@", "@1,2,3,4,5@"},
	{"@files:/tmp/luptests/unwrap/*@", "@0,1,2,3,4,5@"},
	{"@dirs:/tmp/luptests/unwrap/*@", "@a@"},
	{"@all:/tmp/luptests/unwrap/*@", "@0,1,2,3,4,5,a@"},
}

func makeNodes(parent string) {
	path := "/tmp/luptests/" + parent
	if _, err := os.Stat(path); err == nil {
		for i := 0; i <= 5; i++ {
			file := path + "/" + strconv.Itoa(i)
			f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0700)
			if err != nil {
				panic(err.Error())
			}
			f.Close()
		}
	} else {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			panic("Couldn't create " + path)
		}
		for i := 0; i <= 5; i++ {
			dir := path + "/" + "a"
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				panic("Couldn't create " + path)
			}
		}
		makeNodes(parent)
	}
}
func TestUnwrap(t *testing.T) {
	makeNodes("unwrap")
	for _, x := range unwrapTests {
		result := unwrap(x.s)
		if x.e != result {
			t.Errorf("Failed unwrap - expected: %s, got %s", x.e, result)
		}
	}
}
func TestBackref(t *testing.T) {
	for _, x := range backrefTests {
		replacements = x.replacements
		result := backref(x.s, x.mapName, x.curTerms, x.curMap, x.numMaps)
		if result[0] != x.e[0] {
			t.Errorf("Backref - expected %s got %s", x.e[0], result[0])
		}
		//t.Errorf("%s", result)
	}
}

func TestGetNodes(t *testing.T) {
	makeNodes("getnodes")
	for _, x := range getNodesTests {
		result := getNodes(x.dp, x.ep, x.k)
		if result != x.e {
			t.Errorf("getNodes failed on '%s', got '%s'", x.dp, result)
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

func TestRuneIn(t *testing.T) {
	for _, x := range runeInTests {
		result := runeIn(globChars, x.r)
		if result != x.e {
			t.Errorf("runeIn failed on '%c'", x.r)
		}
	}
}

func TestUnescapeGlobChars(t *testing.T) {
	for _, x := range unescapeGlobCharsTests {
		result := unescapeGlobChars(x.s)
		if result != x.e {
			t.Errorf("unescapeGlobChars failed - got: %s, expect: %s", x.s, x.e)
		}
	}
}

func TestSplitBy(t *testing.T) {
	//validates getUnescapedIndices and splitByIndices
	for _, x := range splitByTests {
		result := splitBy(x.s, x.r)
		for i, v := range result {
			if v != result[i] {
				t.Errorf("splitBy failed - got: %v |  expect: %v", result[i], x.e[i])
			}
		}
	}
}

func TestStripSlashes(t *testing.T) {
	for _, x := range stripSlashesTests {
		result := stripSlashes(x.s)
		if result != x.e {
			t.Errorf("stripSlashes failed. Got: %s, Want: %s", result, x.e)
		}
	}
}

func TestAddSlashes(t *testing.T) {
	for _, x := range addSlashesTests {
		result := addSlashes(x.s)
		if result != x.e {
			t.Errorf("addSlashes failed. Got: %s, Want: %s", result, x.e)
		}
	}
}

func TestStripCommas(t *testing.T) {
	for _, x := range stripCommasTests {
		result := stripCommas(x.s)
		if result != x.e {
			t.Errorf("stripCommas failed. Got: %s, Want: %s", result, x.e)
		}
	}
}

func TestDetectShell(t *testing.T) {
	//kinda platform specific
	if shell := detectShell(); shell != "bash" {
		t.Errorf("Shell is %s, expected bash(change this in the test for other shells)", shell)
	}
}
