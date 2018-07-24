package main

import (
	"fmt"
	"testing"
)

var toggleTests = []struct {
	s     rune
	m     rune // matching
	st    state
	start bool
	d     bool // disableOnNoMatch
	e     bool
}{
	{'\\', '\\', state{}, true, true, true},
	{'\\', '\\', state{}, false, false, true},
	{'\\', 'x', state{}, true, true, false},
}

var stripSlashesTests = []struct {
	s string
	e string
}{
	{"Test \\@1\\, both", "Test @1, both"},
	{"Test\\@ at", "Test@ at"},
	{"Test \\,comma", "Test ,comma"},
}

var lupGroupTests = []struct {
	s int
	e string
}{
	{1, fmt.Sprintf("##LUP_GROUP_%d##", 1)},
	{99, fmt.Sprintf("##LUP_GROUP_%d##", 99)},
}

var addSlashesTests = []struct {
	s string
	e string
}{
	{"Test @1, both", "Test \\@1\\, both"},
	{"Test@ at", "Test\\@ at"},
	{"Test ,comma", "Test \\,comma"},
}

var stripCommasTests = []struct {
	s string
	e string
}{
	{",beginning,and,end,", "beginning,and,end"},
	{",just,the,start", "just,the,start"},
	{"one,right,", "one,right"},
	{"multi,right,,,", "multi,right"},
	{",,,multi,left", "multi,left"},
}

var hasGlobsTests = []struct {
	s string
	e bool
}{
	{"/tmp/*/", true},
	{"/tmp/*", true},
	{"/tmp/a?c/", true},
	{"/tmp/1!2/", true},
	{"/tmp/3{4/", true},
	{"/tmp/5}6/", true},
	{"/tmp/506/", false},
}

var isEscapedTests = []struct {
	s string // text
	n int    // position
	e bool
}{
	{"He\\llo", 3, true},
	{"He\\llo", 4, false},
	{"He\\llo", 2, false},
	{"\\Hello", 1, true},
	{"Hell\\o", 5, true},
	{"Hell\\\\o", 6, true},
}

var runeInTests = []struct {
	r rune // text
	e bool
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
	e string
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
	s string // text
	r rune   // split on
	e []string
}{
	{"hello,world", ',', []string{"hello", "world"}},
	{"this\\,is,a,test", ',', []string{"this,is", "a", "test"}},
}

var isHiddenTests = []struct {
	s string
	e bool
	o string
}{
	{"-:Hello", true, "Hello"},
	{"-Goodbye", false, "-Goodbye"},
	{"cat", false, "cat"},
	{":-dog", false, ":-dog"},
}

func TestHasGlobs(t *testing.T) {
	for _, x := range hasGlobsTests {
		result := hasGlobs(x.s)
		if result != x.e {
			t.Errorf("hasGlobs failed on '%s'", x.s)
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

func TestLupGroup(t *testing.T) {
	for _, x := range lupGroupTests {
		result := lupGroup(x.s)
		if result != x.e {
			t.Errorf("lupGroup failed. Got: %s, Want: %s", result, x.e)
		}
	}
}
func TestToggle(t *testing.T) {
	for _, x := range toggleTests {
		x.st.toggle(x.s, x.m, x.d)
		if x.st.on != x.e {
			t.Errorf("lupGroup failed. Got: %t, Want: %t", x.st.on, x.e)
		}
	}
}

func TestIsHidden(t *testing.T) {
	for _, x := range isHiddenTests {
		b, o := isHidden(x.s)
		if b != x.e || o != x.o {
			t.Errorf("isHidden failed on '%s'", x.s)
		}
	}
}
