package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func unique(e []string) []string {
	r := []string{}

	for _, s := range e {
		if !contains(r[:], s) {
			r = append(r, s)
		}
	}
	return r
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func counter(arr []string) map[string]int {
	var count = make(map[string]int)
	for _, x := range arr {
		count[x]++
	}
	return count
}

func insert(slice []string, index int, value string) []string {
	slice = append(slice, "")
	copy(slice[index+1:], slice[index:])
	slice[index] = value
	return slice
}

func filenameHeader(filename string) string {
	/*
	   Prints filename header
	*/
	header := filename + "--------------------------------------------------------------------------"
	return fmt.Sprintf("----%-25v\n", blueText.Sprint(header[:60]))
}

func outputPipe() bool {
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

/* True if data is coming in from stdin */
func inputPipe() bool {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode()&os.ModeCharDevice) == 0 && stat.Size() > 0 {
		return true
	}
	return false
}

func ifelse(s bool, t string, f string) string {
	if s {
		return t
	}
	return f
}

func parseTags(s string) []string {
	// Extract tags from string field
	re := regexp.MustCompile(`#([A-Za-z0-9]+)`)
	r := re.FindAllStringSubmatch(s, -1)
	var tagSet []string
	if len(r) > 0 {
		for _, tags := range r {
			tagSet = append(tagSet, tags[1])
		}
		return tagSet
	}
	return []string{}
}

func fetchContent(url string, ch chan *string) {
	// Fetch raw content from a URL
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var result = string(content)
	ch <- &result
}

// True / False

func parseTrueFalse(s string) string {
	switch {
	case strings.ToLower(s) == "t":
		return "T"
	case strings.ToLower(s) == "f":
		return "F"
	case strings.ToLower(s) == "true":
		return "T"
	case strings.ToLower(s) == "false":
		return "F"
	}
	return "Error"
}

func trueFalse(s bool) string {
	// Convert true and false to 'T' and 'F' b/c of
	// bleve search index limitations
	if s {
		return "T"
	}
	return "F"
}
