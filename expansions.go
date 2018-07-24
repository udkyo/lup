package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func backref(s string, curTerms []string) string {
	if regexp.MustCompile(`^[0-9]+$`).MatchString(s) {
		n, _ := strconv.Atoi(s)
		if n > len(curTerms) {
			fmt.Fprintf(os.Stderr, "Invalid backref, %s is greater than the number of groups.\n", s)
			os.Exit(4)
		} else {
			s = curTerms[n-1]
		}
	}
	return s
}

func expand(s string, externalPath string, inSingles bool, inDoubles bool) (r []string) {
	r = expandPaths(expandLines(expandRanges([]string{s})), externalPath)
	for i := range r {
		if !inSingles && !inDoubles {
			r[i] = strings.Replace(r[i], "'", "\\'", -1)
			r[i] = strings.Replace(r[i], "\"", "\\\"", -1)
		}
		if inSingles {
			r[i] = strings.Replace(r[i], "'", "'\\''", -1)
		}
	}
	return
}

func expandLines(words []string) (expanded []string) {
	for _, word := range words {
		if strings.HasPrefix(word, "lines:") {
			file, err := os.Open(word[6:])
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				st := scanner.Text()
				expanded = append(expanded, addSlashes(st))
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
		}
	}
	if len(expanded) == 0 {
		expanded = words
	}
	return
}

func expandRanges(words []string) (expanded []string) {
	for _, word := range words {
		re := regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)`)
		res := re.FindAllStringSubmatch(word, -1)
		for m, x := range res {
			if len(x) == 0 {
				break
			}
			first, _ := strconv.Atoi(res[m][1])
			last, _ := strconv.Atoi(res[m][2])

			if last < first {
				for i := first; i >= last; i-- {
					expanded = append(expanded, strconv.Itoa(i))
				}
			} else if first < last {
				for i := first; i <= last; i++ {
					expanded = append(expanded, strconv.Itoa(i))
				}
			} else {
				if first == last {
					fmt.Fprintf(os.Stderr, "Integer range starts and ends on the same number.")
					os.Exit(6)
				}
			}
		}
	}
	if len(expanded) == 0 {
		expanded = words
	}
	return
}

func expandPaths(words []string, externalPath string) (s []string) {
	var done bool
	for _, word := range words {
		for _, marker := range []string{"files", "dirs", "all"} {
			if strings.HasPrefix(word, marker+":") {
				s = append(s, getNodes(strings.Replace(word, marker+":", "", 1), externalPath, marker)...)
				done = true
			}
		}
	}
	if !done {
		s = words
	}
	return
}

func getNodes(directivePath string, externalPath string, kind string) (s []string) {
	var nodes []string
	var tainted bool
	var rel bool

	if !strings.HasPrefix(directivePath, "/") {
		rel = true
	}
	if !rel && externalPath != "" {
		fmt.Fprintf(os.Stderr, "Can't mix and match immediately preceeding paths and group paths in files/dirs/all directives")
		os.Exit(7)
	}

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
	fullPath = unescapeGlobChars(fullPath)
	contents, err := filepath.Glob(fullPath)
	if err != nil {
		errOn(err, "Globbing error", 7)
	}
	if len(contents) == 0 {
		fmt.Fprintf(os.Stderr, "No nodes matched (kind:"+kind+" / directivePath:"+directivePath+" / externalPath:"+externalPath+")")
		os.Exit(8)
	}
	for _, f := range contents {
		nodeStat, err := os.Stat(f)
		if err != nil {
			errOn(err, "Couldn't stat", 9)
		}
		if tainted {
			f = strings.Replace(f, " ", "\\ ", -1)
		} else {
			if filepath.Dir(f) != "." && rel {
				s, err := filepath.Rel(externalPath, f)
				if err != nil {
					errOn(err, "Path error", 10)
				}
				f = s
			} else {
				f = filepath.Base(f)
			}
			f = strings.Replace(f, " ", "\\ ", -1)
		}
		switch mode := nodeStat.Mode(); {
		case mode.IsDir() && (kind == "all" || kind == "dirs"):
			nodes = append(nodes, f)
		case mode.IsRegular() && (kind == "all" || kind == "files"):
			nodes = append(nodes, f)
		}
	}
	return nodes
}
