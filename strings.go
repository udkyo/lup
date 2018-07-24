package main

import (
	"fmt"
	"strings"
)

var (
	globChars = []rune{'*', '?', '!', '{', '}'}
)

type state struct {
	on bool
}

func (t *state) toggle(src rune, match rune, disableOnNoMatch bool) {
	if disableOnNoMatch == true {
		if src != match {
			t.on = false
		} else {
			t.on = !t.on
		}
	} else if src == match {
		t.on = !t.on
	}
}

func lupGroup(i int) (s string) {
	return fmt.Sprintf("##LUP_GROUP_%d##", i)
}

func isHidden(text string) (b bool, s string) {
	if strings.HasPrefix(text, hider) {
		b = true
		s = text[2:]
	} else {
		s = text
	}
	return b, s
}

func addSlashes(word string) string {
	delimiter := '@'
	word = strings.Replace(word, string(delimiter), "\\"+string(delimiter), -1)
	word = strings.Replace(word, ",", "\\,", -1)
	return word
}

func stripSlashes(word string) string {
	word = strings.Replace(word, "\\"+string(delimiter), string(delimiter), -1)
	word = strings.Replace(word, "\\,", ",", -1)
	return word
}

func runeIn(group []rune, r rune) (b bool) {
	for _, x := range group {
		if string(x) == string(r) {
			b = true
		}
	}
	return b
}

func hasGlobs(text string) (b bool) {
	for _, c := range text {
		if runeIn(globChars, c) {
			b = true
		}
	}
	return b
}

func isEscaped(text string, n int) (b bool) {
	if n > 0 && text[n-1] == '\\' {
		b = true
	}
	return b
}

func unescapeGlobChars(text string) string {
	for _, char := range globChars {
		text = strings.Replace(text, "\\"+string(char), string(char), -1)
	}
	return text
}
