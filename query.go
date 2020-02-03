package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/olekukonko/tablewriter"
)

/*
	Tags
*/

func fetchTags() map[string]int {
	var tagSet []Tag
	tagCounts := make(map[string]int)
	tagsFile, err := os.Open(libTagsPath)
	if err != nil {
		ThrowError("No Tags Summary File Found; Update library", 1)
	}
	defer tagsFile.Close()
	jsonParser := json.NewDecoder(tagsFile)
	jsonParser.Decode(&tagSet)
	for _, tag := range tagSet {
		tagCounts[tag.Name] = tag.Count
	}
	return tagCounts
}

// ListTags - all tags in library
func ListTags() {
	tags := fetchTags()
	keys := make([]string, 0, len(tags))
	for tag := range tags {
		keys = append(keys, tag)
	}
	sort.Slice(keys, func(i, j int) bool { return tags[keys[i]] > tags[keys[j]] })

	data := make([][]string, len(tags))
	for idx, key := range keys {
		data[idx] = []string{key, fmt.Sprintf("%x", tags[key])}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tag", "Count"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.AppendBulk(data)
	table.Render()
}

type LibSummary struct {
	gists uint64
	files int
}

func librarySummary() LibSummary {
	dc, _ := DbIdx.DocCount()
	q := query.NewMatchAllQuery()
	sr := bleve.NewSearchRequest(q)
	results, err := DbIdx.Search(sr)
	if err != nil {
		fmt.Println("No Results")
	}
	var nfiles int
	for _, gist := range results.Hits {
		nfiles += gist.Fields["NFiles"].(int)
	}
	return LibSummary{gists: dc, files: nfiles}
}

/*
	Langauge
*/
// func languageSummary() {
// 	q := query.NewMatchAllQuery()
// 	sr = bleve.NewSearchRequest(q)
// 	sr.Size = int(dc)
// 	sr.SortBy([]string{"UpdatedAt"})
// }

// ls - the primary query interface
func ls(searchTerm string, sortBy string, tag string, language string, starred bool, status string) {
	var qstring string

	// Consider reworking filtering here to be done manually...
	if searchTerm != "" {
		qstring = fmt.Sprintf("%s", searchTerm)
	}

	if tag != "" {
		qstring = fmt.Sprintf("+Tags:%v %s", tag, qstring)
	}

	if language != "" {
		qstring = fmt.Sprintf("+Language:%v %s", language, qstring)
	}

	fmt.Printf(qstring)
	// if status == "public" {
	// 	qstring = fmt.Sprintf("%s +Public", qstring)
	// } else if status == "private" {
	// 	qstring = fmt.Sprintf("%s -Public", qstring)
	// } else if status != "all" {
	// 	ThrowError("--public must be 'all', 'public', or 'private'", 1)
	// }

	var isQuery bool
	var sr *bleve.SearchRequest
	dc, _ := DbIdx.DocCount()
	// dump when no query params present
	if searchTerm == "" && tag == "" && language == "" && status == "all" {
		q := query.NewMatchAllQuery()
		sr = bleve.NewSearchRequest(q)
		sr.Size = int(dc)
		sr.SortBy([]string{"UpdatedAt"})
		isQuery = false
	} else {
		q := query.NewQueryStringQuery(qstring)
		sr = bleve.NewSearchRequest(q)
		//sr.Highlight = bleve.NewHighlightWithStyle("ansi")
		sr.Size = 50
		isQuery = true
	}

	sr.Fields = []string{"*"}
	results, err := DbIdx.Search(sr)

	if err != nil {
		fmt.Println("No Results")
	}
	resultTable(results.Hits, isQuery)
}

// Perform fuzzy search
func fuzzySearch(searchTerm string) {
	var isQuery bool
	var sr *bleve.SearchRequest
	q := query.NewFuzzyQuery(searchTerm)
	sr = bleve.NewSearchRequest(q)
	sr.Size = 10
	isQuery = true

	sr.Fields = []string{"*"}
	results, err := DbIdx.Search(sr)
	if err != nil {
		fmt.Println("No Results")
	}
	resultTable(results.Hits, isQuery)
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
	searchResults, err := DbIdx.Search(sr)
	if err != nil {
		panic(err)
	}
	return searchResults.Hits[0]
}
