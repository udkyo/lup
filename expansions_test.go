package main

import (
	"io/ioutil"
	"testing"
)

var backrefTests = []struct {
	s        string // text
	curTerms []string
	e        []string
}{
	{"1", []string{"blah"}, []string{"blah"}},
	{"3", []string{"blah", "yadda", "hey"}, []string{"hey"}},
}

var expandLinesTests = []struct {
	s string // text
	e []string
}{
	{"lines:/tmp/luptests/names.txt", []string{"f\\@rnsworth", "fry", "lee\\,la", "zoid berg", "bender", "hermes"}},
}

var expandRangesTests = []struct {
	s string // text
	e []string
}{
	{"1..5", []string{"1", "2", "3", "4", "5"}},
	{"5..1", []string{"5", "4", "3", "2", "1"}},
	{"abc", []string{"abc"}},
}

var expandPathsTests = []struct {
	s string // text
	e []string
}{
	{"files:/tmp/luptests/expandpaths/*", []string{`a\ a`, `a""a`, `a"a`, `a'"a`, `a''a`, `a'a`, `a,a`, `a@a`, `a\\ a`, `a\,a`, `a\@a`, `aa`, `ab`, `ac`}},
	{"dirs:/tmp/luptests/expandpaths/*", []string{`0\ 0`, `0""0`, `0"0`, `0'"0`, `0''0`, `0'0`, `0,0`, `00`, `01`, `02`, `0@0`, `0\\ 0`, `0\,0`, `0\@0`}},
}

var getNodesTests = []struct {
	dp string // text
	ep string
	k  string
	e  []string
}{
	{"/tmp/luptests/getnodes/*", "", "files", []string{`a\ a`, `a""a`, `a"a`, `a'"a`, `a''a`, `a'a`, `a,a`, `a@a`, `a\\ a`, `a\,a`, `a\@a`, `aa`, `ab`, `ac`}},
	{"/tmp/luptests/getnodes/*", "", "dirs", []string{`0\ 0`, `0""0`, `0"0`, `0'"0`, `0''0`, `0'0`, `0,0`, `00`, `01`, `02`, `0@0`, `0\\ 0`, `0\,0`, `0\@0`}},
	{"/tmp/luptests/getnodes/*", "", "all", []string{`0\ 0`, `0""0`, `0"0`, `0'"0`, `0''0`, `0'0`, `0,0`, `00`, `01`, `02`, `0@0`, `0\\ 0`, `0\,0`, `0\@0`, `a\ a`, `a""a`, `a"a`, `a'"a`, `a''a`, `a'a`, `a,a`, `a@a`, `a\\ a`, `a\,a`, `a\@a`, `aa`, `ab`, `ac`}},
	{"*", "/tmp/luptests/getnodes/", "files", []string{`a\ a`, `a""a`, `a"a`, `a'"a`, `a''a`, `a'a`, `a,a`, `a@a`, `a\\ a`, `a\,a`, `a\@a`, `aa`, `ab`, `ac`}},
	{"*", "/tmp/luptests/getnodes/", "dirs", []string{`0\ 0`, `0""0`, `0"0`, `0'"0`, `0''0`, `0'0`, `0,0`, `00`, `01`, `02`, `0@0`, `0\\ 0`, `0\,0`, `0\@0`}},
	{"*", "/tmp/luptests/getnodes/", "all", []string{`0\ 0`, `0""0`, `0"0`, `0'"0`, `0''0`, `0'0`, `0,0`, `00`, `01`, `02`, `0@0`, `0\\ 0`, `0\,0`, `0\@0`, `a\ a`, `a""a`, `a"a`, `a'"a`, `a''a`, `a'a`, `a,a`, `a@a`, `a\\ a`, `a\,a`, `a\@a`, `aa`, `ab`, `ac`}},
	{"*", "/tmp/luptests/getn?des/", "files", []string{`/tmp/luptests/getnodes/a\ a`, `/tmp/luptests/getnodes/a""a`, `/tmp/luptests/getnodes/a"a`, `/tmp/luptests/getnodes/a'"a`, `/tmp/luptests/getnodes/a''a`, `/tmp/luptests/getnodes/a'a`, `/tmp/luptests/getnodes/a,a`, `/tmp/luptests/getnodes/a@a`, `/tmp/luptests/getnodes/a\\ a`, `/tmp/luptests/getnodes/a\,a`, `/tmp/luptests/getnodes/a\@a`, `/tmp/luptests/getnodes/aa`, `/tmp/luptests/getnodes/ab`, `/tmp/luptests/getnodes/ac`}},
	{"*", "/tmp/luptests/getn?des/", "dirs", []string{`/tmp/luptests/getnodes/0\ 0`, `/tmp/luptests/getnodes/0""0`, `/tmp/luptests/getnodes/0"0`, `/tmp/luptests/getnodes/0'"0`, `/tmp/luptests/getnodes/0''0`, `/tmp/luptests/getnodes/0'0`, `/tmp/luptests/getnodes/0,0`, `/tmp/luptests/getnodes/00`, `/tmp/luptests/getnodes/01`, `/tmp/luptests/getnodes/02`, `/tmp/luptests/getnodes/0@0`, `/tmp/luptests/getnodes/0\\ 0`, `/tmp/luptests/getnodes/0\,0`, `/tmp/luptests/getnodes/0\@0`}},
	{"*", "/tmp/luptests/getn?des/", "all", []string{`/tmp/luptests/getnodes/0\ 0`, `/tmp/luptests/getnodes/0""0`, `/tmp/luptests/getnodes/0"0`, `/tmp/luptests/getnodes/0'"0`, `/tmp/luptests/getnodes/0''0`, `/tmp/luptests/getnodes/0'0`, `/tmp/luptests/getnodes/0,0`, `/tmp/luptests/getnodes/00`, `/tmp/luptests/getnodes/01`, `/tmp/luptests/getnodes/02`, `/tmp/luptests/getnodes/0@0`, `/tmp/luptests/getnodes/0\\ 0`, `/tmp/luptests/getnodes/0\,0`, `/tmp/luptests/getnodes/0\@0`, `/tmp/luptests/getnodes/a\ a`, `/tmp/luptests/getnodes/a""a`, `/tmp/luptests/getnodes/a"a`, `/tmp/luptests/getnodes/a'"a`, `/tmp/luptests/getnodes/a''a`, `/tmp/luptests/getnodes/a'a`, `/tmp/luptests/getnodes/a,a`, `/tmp/luptests/getnodes/a@a`, `/tmp/luptests/getnodes/a\\ a`, `/tmp/luptests/getnodes/a\,a`, `/tmp/luptests/getnodes/a\@a`, `/tmp/luptests/getnodes/aa`, `/tmp/luptests/getnodes/ab`, `/tmp/luptests/getnodes/ac`}},
	{"getnodes/*", "/tmp/luptests/", "all", []string{`getnodes/0\ 0`, `getnodes/0""0`, `getnodes/0"0`, `getnodes/0'"0`, `getnodes/0''0`, `getnodes/0'0`, `getnodes/0,0`, `getnodes/00`, `getnodes/01`, `getnodes/02`, `getnodes/0@0`, `getnodes/0\\ 0`, `getnodes/0\,0`, `getnodes/0\@0`, `getnodes/a\ a`, `getnodes/a""a`, `getnodes/a"a`, `getnodes/a'"a`, `getnodes/a''a`, `getnodes/a'a`, `getnodes/a,a`, `getnodes/a@a`, `getnodes/a\\ a`, `getnodes/a\,a`, `getnodes/a\@a`, `getnodes/aa`, `getnodes/ab`, `getnodes/ac`}},
}

func TestBackref(t *testing.T) {
	for _, x := range backrefTests {
		result := backref(x.s, x.curTerms)
		if result != x.e[0] {
			t.Errorf("Backref - expected %s got %s", x.e[0], result)
		}
	}
}

func TestExpandLines(t *testing.T) {
	d1 := []byte("f@rnsworth\nfry\nlee,la\nzoid berg\nbender\nhermes")
	err := ioutil.WriteFile("/tmp/luptests/names.txt", d1, 0700)
	if err != nil {
		t.Fatalf("Failed to write names.txt in TestExpandLines")
	}
	for _, x := range expandLinesTests {
		result := expandLines([]string{x.s})
		for i, r := range result {
			if x.e[i] != r {
				t.Errorf("Failed expandLines - expected: %s, got %s", x.e, result)
			}
		}
	}
}

func TestExpandRanges(t *testing.T) {
	for _, x := range expandRangesTests {
		result := expandRanges([]string{x.s})
		for i, r := range result {
			if x.e[i] != r {
				t.Errorf("Failed expandRanges - expected: %s, got %s", x.e, result)
			}
		}
	}
}

func TestExpandPaths(t *testing.T) {
	//todo: various
	makeNodes("expandpaths")
	for _, x := range expandPathsTests {
		result := expandPaths([]string{x.s}, "")
		for i, r := range result {
			if x.e[i] != r {
				t.Errorf("Failed unwrap - expected: %s, got %s", x.e, result)
			}
		}
	}
}

func TestGetNodes(t *testing.T) {
	makeNodes("getnodes")
	for _, x := range getNodesTests {
		result := getNodes(x.dp, x.ep, x.k)
		for i, r := range result {
			if r != x.e[i] {
				t.Errorf("getNodes failed on '%s', expected , got '%s'", x, result)
			}
		}
	}
}
