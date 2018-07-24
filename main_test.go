package main

import (
	"os"
	"testing"
)

func makeNodes(parent string) {
	dirs := []string{"00", "01", "02", "0'0", "0''0", "0\"\"0", "0\"0", "0'\"0", "0@0", "0\\@0", "0,0", "0\\,0", "0 0", "0\\ 0"}
	files := []string{"aa", "ab", "ac", "a'a", "a''a", "a\"\"a", "a\"a", "a'\"a", "a@a", "a\\@a", "a,a", "a\\,a", "a a", "a\\ a"}

	path := "/tmp/luptests/" + parent
	if _, err := os.Stat(path); err == nil {
		for _, i := range files {
			file := path + "/" + i
			f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0700)
			if err != nil {
				errOn(err, "Couldn't open file", 11)
			}
			f.Close()
		}
	} else {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			errOn(err, "Couldn't create "+path, 12)
		}
		for _, i := range dirs {
			dir := path + "/" + i
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				errOn(err, "Couldn't create "+path, 13)
			}
		}
		makeNodes(parent)
	}
}

func TestMain(t *testing.T) {
	testrun = true
	os.Args = []string{";", "sh", "-c", "touch /tmp/luptests/main"}
	main()
	if _, err := os.Stat("/tmp/luptests/main"); err != nil {
		t.Error("Failed main")
	} else {
		if err := os.Remove("/tmp/luptests/main"); err != nil {
			t.Error("Failed to delete main testfile")
		}
	}
}
