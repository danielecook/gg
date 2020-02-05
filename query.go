package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/olekukonko/tablewriter"
)

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
	Summarize a field
*/
func fieldSummary(field string) {
	// Calculates frequencies for a given field
	facet := bleve.NewFacetRequest(field, 100000)
	query := query.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.AddFacet("count", facet)
	searchResults, err := DbIdx.Search(searchRequest)
	if err != nil {
		panic(err)
	}

	// term with highest occurrences in field name
	data := make([][]string, searchResults.Size())
	for idx, val := range searchResults.Facets["count"].Terms {
		data[idx] = []string{val.Term, strconv.Itoa(val.Count)}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{field, "Count"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.AppendBulk(data)
	table.Render()

}

// ls - the primary query interface
func ls(searchTerm string, sortBy string, tag string, language string, starred bool, status string, limit int) {
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

	if starred {
		qstring = fmt.Sprintf("+Starred:T %s", qstring)
	}

	if status == "public" {
		qstring = fmt.Sprintf("+Public:T %s", qstring)
	} else if status == "private" {
		qstring = fmt.Sprintf("+Public:F %s", qstring)
	} else if status != "all" {
		ThrowError("--public must be 'all', 'public', or 'private'", 1)
	}

	qstring = strings.Trim(qstring, " ")
	var isQuery bool
	var sr *bleve.SearchRequest
	//dc, _ := DbIdx.DocCount()
	// dump when no query params present
	if searchTerm == "" && qstring == "" && status == "all" {
		q := query.NewMatchAllQuery()
		sr = bleve.NewSearchRequest(q)
		sr.Size = limit
		sr.SortBy([]string{"-UpdatedAt"})
		isQuery = false
	} else {
		q := query.NewQueryStringQuery(qstring)
		sr = bleve.NewSearchRequest(q)
		//sr.Highlight = bleve.NewHighlightWithStyle("ansi")
		sr.Size = limit
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
