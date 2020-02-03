package main

import (
	"fmt"
	"os"

	. "github.com/logrusorgru/aurora"
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
	return fmt.Sprintf("----%-25v\n", Blue(Bold(header[:60])))
}

func outputPipe() bool {
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func yesNo(s bool) string {
	if s {
		return "yes"
	} else {
		return "no"
	}
}
