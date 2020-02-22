package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
)

type libSummary struct {
	gists   uint64
	files   int
	starred int
}

type searchQuery struct {
	term     string
	sort     string
	tag      string
	owner    string
	language string
	starred  bool
	status   string
	limit    int
	debug    bool
}

// Used to allow more flexibility when specifying sort.
var sortMap = map[string]string{
	"":            "-UpdatedAt",
	"id":          "IDX",
	"owner":       "Owner",
	"description": "Description",
	"public":      "-Public",
	"-public":     "Public",
	"private":     "Public",
	"-private":    "-Public",
	"filename":    "Filename",
	"language":    "Language",
	"tag":         "Tags",
	"tags":        "Tags",
	"-starred":    "Starred",
	"starred":     "-Starred",
	"created":     "CreatedAt",
	"createdat":   "CreatedAt",
	"updated":     "UpdatedAt",
	"n":           "NLines",
}

func librarySummary() libSummary {
	dc, _ := dbIdx.DocCount()
	q := query.NewMatchAllQuery()
	sr := bleve.NewSearchRequest(q)
	sr.Size = int(dc)
	sr.Fields = []string{"NFiles", "Starred"}
	results, err := dbIdx.Search(sr)
	if err != nil {
		errorMsg("No Results")
	}
	boolInt := map[bool]int{false: 0, true: 1}
	var nfiles int
	var nstarred int
	for _, gist := range results.Hits {
		nfiles += int(gist.Fields["NFiles"].(float64))
		nstarred += boolInt[gist.Fields["Starred"].(string) == "T"]
	}
	return libSummary{gists: dc, files: nfiles, starred: nstarred}
}

// ls - the primary query interface
func ls(search *searchQuery) {
	var qstring string
	var highlightTermSet []string
	// Consider reworking filtering here to be done manually...
	if search.term != "" {
		qstring = fmt.Sprintf("%s", search.term)
		// TODO: Fix term splitting
		// TODO: Handle highlighting at field-level when filtering.
		debugMsg(fmt.Sprint(strings.Split(search.term, " ")))
		highlightTermSet = append(highlightTermSet, strings.Split(search.term, " ")...)
		debugMsg(fmt.Sprintf("highlight- %+v", highlightTermSet))
		debugMsg("TERM")
	}

	if search.tag != "" {
		qstring = fmt.Sprintf("+Tags:%v %s", search.tag, qstring)
		highlightTermSet = append(highlightTermSet, "#"+search.tag)
	}

	if search.language != "" {
		qstring = fmt.Sprintf("+Language:%v %s", search.language, qstring)
		highlightTermSet = append(highlightTermSet, search.language)
	}

	if search.starred {
		qstring = fmt.Sprintf("+Starred:T %s", qstring)
	}

	if search.owner != "" {
		qstring = fmt.Sprintf("+Owner:%v %s", search.owner, qstring)
		highlightTermSet = append(highlightTermSet, search.owner)
	}

	if search.status == "public" {
		qstring = fmt.Sprintf("+Public:T %s", qstring)
	} else if search.status == "private" {
		qstring = fmt.Sprintf("+Public:F %s", qstring)
	} else if search.status != "all" {
		ThrowError("--public must be 'all', 'public', or 'private'", 1)
	}

	debugMsg(fmt.Sprintf("Query: %s", qstring))
	debugMsg(fmt.Sprintf("%+v", search))

	qstring = strings.Trim(qstring, " ")
	var isQuery bool
	var sr *bleve.SearchRequest

	// dump when no query params present
	if search.term == "" && qstring == "" && search.status == "all" {
		q := query.NewMatchAllQuery()
		sr = bleve.NewSearchRequest(q)
		sr.Size = search.limit
		isQuery = false
	} else {
		q := query.NewQueryStringQuery(qstring)
		sr = bleve.NewSearchRequest(q)
		//sr.Highlight = bleve.NewHighlightWithStyle("ansi")
		sr.Size = search.limit
		isQuery = true
	}

	// Handle sorting
	if search.sort != "" {
		sortBy := sortMap[search.sort]
		if sortBy == "" && search.sort[0] == '-' {
			sortBy = fmt.Sprintf("-%s", sortMap[strings.Trim(search.sort, "-")])
		}
		sr.SortBy([]string{sortBy, "-_score"})
	} else if isQuery == true {
		sr.SortBy([]string{"-_score"})
	} else if isQuery == false {
		sr.SortBy([]string{"-UpdatedAt"})
	}

	sr.Fields = []string{"*"}

	results, err := dbIdx.Search(sr)
	if err != nil || len(results.Hits) == 0 {
		// If no results, try fuzzy search
		debugMsg(search.term)
		fuzzySearch(search.term)
		os.Exit(0)
	}
	resultTable(results, isQuery, highlightTermSet)
}

// Perform fuzzy search
func fuzzySearch(searchTerm string) {
	var isQuery bool
	var sr *bleve.SearchRequest
	q := query.NewFuzzyQuery(searchTerm)
	q.SetFuzziness(2)
	sr = bleve.NewSearchRequest(q)
	sr.Size = 10
	isQuery = true

	sr.Fields = []string{"*"}
	results, err := dbIdx.Search(sr)
	if err != nil {
		errorMsg("No Results\n")
		os.Exit(0)
	}
	resultTable(results, isQuery, []string{})
}

func highlight(out io.Writer, filename string, content string, formatter string, style string) {
	/*
		Highlights code

		Formatters:
			html
			json
			noop
			terminal
			terminal16m
			terminal256
			tokens
	*/
	lexer := lexers.Match(filename)
	if lexer == nil || lexer.Config().Name == "plaintext" {
		lexer = lexers.Analyse(content)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}
	quick.Highlight(out, content, lexer.Config().Name, formatter, style)
}

func lookupGist(gistIdx int) *search.DocumentMatch {
	q := query.NewQueryStringQuery(fmt.Sprintf("IDX:%v", gistIdx))
	sr := bleve.NewSearchRequest(q)
	sr.Fields = []string{"*"}
	searchResults, err := dbIdx.Search(sr)
	if err != nil || len(searchResults.Hits) == 0 {
		errorMsg(fmt.Sprintf("%d is not a valid ID\n", gistIdx))
		os.Exit(0)
	}
	return searchResults.Hits[0]
}

// Returns the next IDX to use
func nextIdx() int {
	dc, _ := dbIdx.DocCount()
	if dc == 0 {
		return 0
	}
	q := query.NewMatchAllQuery()
	sr := bleve.NewSearchRequest(q)
	sr.Fields = []string{"IDX"}
	sr.Size = int(dc)
	results, err := dbIdx.Search(sr)
	if err != nil {
		ThrowError("Index error", 1)
	}
	gistIdxSet := make([]int, len(results.Hits))
	for idx, gist := range results.Hits {
		gistIdxSet[idx] = int(gist.Fields["IDX"].(float64))
	}
	sort.Ints(gistIdxSet)
	return gistIdxSet[len(gistIdxSet)-1] + 1
}
