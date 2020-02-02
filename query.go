package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	. "github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
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

// ls - the primary query interface
func ls(searchTerm string, sortBy string, tag string) {
	var qstring string
	if searchTerm != "" {
		qstring = fmt.Sprintf("%s", searchTerm)
	}

	if tag != "" {
		qstring = fmt.Sprintf("%s +Tags:%v", qstring, tag)
	}

	var isQuery bool
	var sr *bleve.SearchRequest
	dc, _ := DbIdx.DocCount()
	// dump when no query params present
	if (searchTerm == "") && tag == "" {
		q := query.NewMatchAllQuery()
		sr = bleve.NewSearchRequest(q)
		sr.Highlight = bleve.NewHighlight()
		sr.Size = int(dc)
		sr.SortBy([]string{"UpdatedAt"})
		isQuery = false
	} else {
		errlog.Println(qstring)
		q := query.NewQueryStringQuery(qstring)
		sr = bleve.NewSearchRequest(q)
		sr.Size = 5
		isQuery = true
	}
	sr.Fields = []string{"*"}
	results, err := DbIdx.Search(sr)
	if err != nil {
		fmt.Println("No Results")
	}
	tableData := make([][]string, len(results.Hits))
	for idx, gist := range results.Hits {
		tableData[idx] = []string{
			fmt.Sprintf("%v", gist.Fields["IDX"]),
			gist.Fields["Description"].(string),
			fmt.Sprintf("%v", gist.Fields["Filename"]),
			fmt.Sprintf("%v", gist.Fields["Language"]),
			gist.Fields["Owner"].(string),
			gist.Fields["UpdatedAt"].(string),
		}
		if isQuery {
			tableData[idx] = append(tableData[idx], fmt.Sprintf("%1.3f", gist.Score))
		}
	}

	// Terminal Window size
	var xsize, _, _ = terminal.GetSize(0)
	var header = []string{"ID", "Description", "Filename", "Language", "Author", "Updated"}
	if isQuery {
		header = append(header, "Score")
	}
	var colWidth int
	colWidth = (xsize / len(header))
	// Render results
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetColWidth(colWidth)
	// Give Description 2x width
	table.SetColMinWidth(2, colWidth*2)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.AppendBulk(tableData)
	table.Render()
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

func outputGist(gistIdx int) {
	gist := lookupGist(gistIdx)
	// Parse bleve index which flattens results
	keys := reflect.ValueOf(gist.Fields).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	var fsplit []string
	var fileset = map[string]map[string]string{}
	for idx := range strkeys {
		fsplit = strings.Split(strkeys[idx], ".")
		if fsplit[0] == "Files" {
			field := fsplit[len(fsplit)-1]
			filename := strings.Join(fsplit[1:len(fsplit)-1], ".")
			value := gist.Fields[strkeys[idx]]
			if fileset[filename] == nil {
				fileset[filename] = map[string]string{}
			}
			fileset[filename][field] = fmt.Sprintf("%v", value)
		}
	}

	for _, file := range fileset {
		var xsize, _, _ = terminal.GetSize(0)
		var line = strings.Repeat("-", xsize-len(file["filename"])-50)
		if outputPipe() {
			fmt.Print(file["content"])
		} else {
			errlog.Printf("%s%s%s", Green(Bold(file["filename"])), line, file["language"])
			highlight(os.Stdout, file["filename"], file["content"], "terminal16m", "fruity")
			fmt.Fprintf(os.Stderr, "\n\n")
		}
	}

}
