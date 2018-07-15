package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func getNodes(directivePath string, externalPath string, kind string) string {
	// used by expandPaths
	var nodes []string
	var prefix string
	var tainted bool
	var rel bool

	if isHidden(directivePath) {
		directivePath = directivePath[2:]
		prefix = hider
	}
	if !strings.HasPrefix(directivePath, "/") {
		rel = true
	}
	errOn(!rel && externalPath != "", "fatal: don't mix and match importing paths from outside @@ blocks and absolute paths inside them, no bueno (map: "+directivePath+")", 5)

	if hasGlobs(externalPath) {
		tainted = true
	}
	directivePath = unescapeGlobChars(directivePath)
	fullPath := directivePath
	if externalPath != "" {
		if !strings.HasSuffix(externalPath, "/") || isEscaped(externalPath, len(externalPath)-1) {
			fullPath = externalPath + "/" + directivePath
		} else {
			fullPath = externalPath + directivePath
		}
	}
	contents, err := filepath.Glob(fullPath)
	errIf(err)
	errOn(len(contents) == 0, "No nodes matched ("+kind+": "+directivePath+")", 3)
	for _, f := range contents {
		nodeStat, err := os.Stat(f)
		errIf(err)
		if tainted {
			f = strings.Replace(f, " ", "\\ ", -1)
		} else {
			if filepath.Dir(f) != "." && rel {
				s, err := filepath.Rel(externalPath, f)
				errIf(err)
				f = s
			} else {
				f = strings.Replace(filepath.Base(f), " ", "\\ ", -1)
			}
		}
		switch mode := nodeStat.Mode(); {
		case mode.IsDir() && (kind == "all" || kind == "dirs"):
			nodes = append(nodes, f)
		case mode.IsRegular() && (kind == "all" || kind == "files"):
			nodes = append(nodes, f)
		}
	}
	return prefix + strings.Join(nodes, ",")
}

func expandRanges(text string) string {
	expanded := ""
	capturing := -1
	word := ""
	newtext := text
	for n, char := range text {
		escaped := isEscaped(text, n)

		if (capturing == -1 && char == delimiter) && !escaped {
			capturing = n + 1
			expanded = ""
			word = ""
		} else if char == delimiter || char == ',' && capturing > -1 && !escaped {

			start := 0
			prefix := ""
			if isHidden(word) {
				start = 2
				prefix = "-:"
			}

			re := regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)`)
			res := re.FindAllStringSubmatch(word[start:], -1)

			for m, x := range res {
				if len(x) == 0 {
					break
				}
				first, _ := strconv.Atoi(res[m][1])
				last, _ := strconv.Atoi(res[m][2])

				if last < first {
					for i := first; i >= last; i-- {
						expanded += strconv.Itoa(i)
						if i > last {
							expanded += ","
						}
					}
					newtext = strings.Replace(newtext, word, prefix+expanded, 1)
				} else if first < last {
					for i := first; i <= last; i++ {
						expanded += strconv.Itoa(i)
						if i < last {
							expanded += ","
						}
					}
					newtext = strings.Replace(newtext, word, prefix+expanded, 1)
				} else {
					errOn(first == last, "Integer range starts and ends on the same number, please check", 3)
				}
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
		if char == ',' && n > 0 && !escaped {
			capturing = n
			word = ""
			expanded = ""
		}
	}
	return newtext
}

func expandLines(text string) string {
	capturing := -1
	word := ""
	newtext := text
	var lines []string
	start := -1
	prefix := ""

	for n, char := range text {
		escaped := isEscaped(text, n)
		if (capturing == -1 && char == delimiter) && !escaped {
			capturing = n + 1
			word = ""
		} else if char == delimiter || char == ',' && capturing > -1 && !escaped {
			if strings.HasPrefix(word, "lines:") || strings.HasPrefix(word, "-:lines:") {
				if strings.HasPrefix(word, "-:") {
					start = 8
					prefix = "-:"
				} else {
					start = 6
				}
				file, err := os.Open(word[start:])
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					lines = append(lines, addSlashes(scanner.Text()))
				}
				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
				newtext = strings.Replace(newtext, word, prefix+strings.Join(lines, ","), 1)
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
		if char == ',' && n > 0 && !escaped {
			capturing = n
			word = ""
		}
	}
	return newtext
}

func expandPaths(text string) string {
	capturing := -1
	word := ""
	newtext := text
	pathStart := -1
	path := ""

	for n, char := range text {
		escaped := isEscaped(text, n)

		if capturing == -1 {
			if pathStart == -1 && char == '/' && !escaped {
				pathStart = n
			}
		}
		if char == delimiter && !escaped {
			if pathStart > -1 {
				path = text[pathStart:n]
			}
			pathStart = -1
		}

		if (capturing == -1 && char == delimiter) && !escaped {
			capturing = n + 1
			word = ""
		} else if char == delimiter || char == ',' && capturing > -1 && !escaped {
			if hasGlobs(path) {
				newtext = strings.Replace(newtext, path, "", 1)
			}

			for _, marker := range []string{"files", "dirs", "all"} {
				if strings.HasPrefix(word, marker+":") || strings.HasPrefix(word, "-:"+marker+":") {
					replacement := getNodes(strings.Replace(word, marker+":", "", 1), path, marker)
					newtext = strings.Replace(newtext, word, replacement, 1)
				}
			}
			capturing = -1
		} else if capturing > -1 && n >= capturing {
			word += string(char)
		}
		if char == ',' && n > 0 && !escaped {
			capturing = n
			word = ""
		}
	}
	return newtext
}
