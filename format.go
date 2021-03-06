package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	aw "github.com/deanishe/awgo"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
)

func caseInsensitiveReplace(subject string, search string, replace string) string {
	searchRegex := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
	pos := searchRegex.FindStringIndex(subject)
	if pos != nil {
		start, stop := pos[0], pos[1]
		return subject[:start] + highlightText.Sprintf(subject[start:stop]) + subject[stop:len(subject)]
	}
	return subject
}

func highlightTerms(s string, terms []string) string {
	for _, term := range terms {
		s = caseInsensitiveReplace(s, term, highlightText.Sprint(term))
	}
	return s
}

func removeFields(header []string, tableData [][]string, omitFields []string) ([]string, [][]string) {
	// Removes fields from the header and table

	// Remove fields from header
	n := 0
	for _, field := range header {
		if contains(omitFields, field) == false {
			header[n] = field
			n++
		}
	}
	header = header[:n]

	// Remove fields from table
	for ridx := range tableData {
		n = 0
		for idx, field := range header {
			if contains(omitFields, field) == false {
				tableData[ridx][n] = tableData[ridx][idx]
				n++
			}
		}
		tableData[ridx] = tableData[ridx][:n]
	}
	return header, tableData
}

// Generate a result table
func resultTable(results *bleve.SearchResult, isQuery bool, highlightTermSet []string) {

	/*
		Format
	*/
	// Terminal Window size
	var xsize, _, _ = terminal.GetSize(0)

	var colWidth int

	tableData := make([][]string, len(results.Hits))
	for idx, gist := range results.Hits {

		updatedAt := strings.Split(gist.Fields["UpdatedAt"].(string), "T")[0]

		tableData[idx] = []string{
			fmt.Sprintf("%v", gist.Fields["IDX"]),
			ifelse(gist.Fields["Starred"].(string) == "T", "⭐", ""),
			ifelse(gist.Fields["Public"].(string) == "F", "🔒", ""),
			highlightTerms(fmt.Sprintf("%.60v", gist.Fields["Description"].(string)), highlightTermSet),
			highlightTerms(fmt.Sprintf("%v", gist.Fields["Filename"]), highlightTermSet),
			highlightTerms(fmt.Sprintf("%v", gist.Fields["Language"]), highlightTermSet),
			highlightTerms(gist.Fields["Owner"].(string), highlightTermSet),
			string(fmt.Sprintf("%v", gist.Fields["NLines"].(float64))),
			updatedAt,
		}

		if isQuery {
			tableData[idx] = append(tableData[idx], fmt.Sprintf("%1.3f", gist.Score))
		}
	}

	// Render results
	table := tablewriter.NewWriter(os.Stdout)

	table.SetAutoWrapText(false)

	/*
		Header
	*/
	var header = []string{"ID", "⭐", "🔒", "Description", "Filename", "Language", "Owner", "n", "Updated"}
	if isQuery {
		header = append(header, "Score")
	}

	colWidth = (xsize / len(header))

	var omitFields []string
	switch {
	case between(xsize, 100, 125):
		omitFields = []string{"Updated", "Owner"}
	case between(xsize, 80, 100):
		omitFields = []string{"Updated", "Owner", "Language"}
	}
	if xsize < 125 {
		header, tableData = removeFields(header, tableData, omitFields)
	}

	table.SetAutoFormatHeaders(false)
	table.SetHeader(header)

	var headerColors []tablewriter.Colors
	headerColors = make([]tablewriter.Colors, len(header))
	for i := 0; i < len(header); i++ {
		headerColors[i] = tablewriter.Colors{tablewriter.Bold}
	}
	table.SetHeaderColor(headerColors...)
	table.SetHeaderLine(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.SetColWidth(colWidth)
	table.SetColMinWidth(3, int(float32(colWidth)*2.5))
	table.SetColumnSeparator("\t")
	table.SetCenterSeparator("\t")
	// Give Description 2x width
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.AppendBulk(tableData)
	//table.SetTablePadding("")
	table.SetAutoWrapText(false)
	table.Render()
	blueText.Printf("Showing %v Hit%s of %v Results\n", len(results.Hits), ifelse(results.Total != 1, "s", ""), results.Total)
}

/*
	Summarize a field
*/
func fieldSummaryTable(field string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{field, "Count"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)
	table.SetHeaderLine(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.AppendBulk(data)
	table.SetColumnSeparator("\t")
	table.SetCenterSeparator("\t")
	table.Render()
}

func queryGistsAlfred(alfredQuery string) {
	squery.limit = 100
	squery.status = "all"
	switch {
	// Starred gists
	case strings.HasPrefix(alfredQuery, "⭐"):
		squery.starred = true
		var subQuery = strings.SplitAfter(alfredQuery, " ")
		if len(subQuery) > 0 {
			squery.term = subQuery[0]
		}
	default:
		squery.term = alfredQuery
	}
	ls(&squery)
}

func fieldSummaryAlfred(field string, data [][]string) {
	var qPrefix string
	var tagFmt string
	var icon *aw.Icon
	for _, row := range data {
		switch {
		case field == "Tags":
			qPrefix = "#"
			icon = tagIcon
		case field == "Language":
			qPrefix = "~"
			icon = resolveIcon(row[0])
		case field == "Owner":
			qPrefix = ":"
			icon = randomOwnerIcon()
		}
		tagFmt = qPrefix + strings.ToLower(row[0])
		var subQuery = strings.SplitAfter(alfredQuery, " ")
		// Need to split this logic out for other args...
		var prefixMatch = strings.HasPrefix(tagFmt, strings.ToLower(alfredQuery)) || (tagFmt == strings.Trim(subQuery[0], " "))
		var isMatched = (alfredQuery[len(alfredQuery)-1] == ' ') || (len(subQuery) > 1)
		if (prefixMatch == true) && (isMatched == false) {
			wf.NewItem(fmt.Sprintf("%s%v", qPrefix, row[0])).
				Icon(icon).
				Autocomplete(fmt.Sprintf("%s%v ", qPrefix, row[0])).
				Subtitle(fmt.Sprintf("%v gists", row[1]))
		} else if (prefixMatch == true) && (isMatched == true) {
			// Query for results
			squery.term = subQuery[1] // if empty (""); term does nothing
			squery.status = "all"
			squery.limit = 100
			if qPrefix == "#" {
				squery.tag = row[0]
			} else if qPrefix == "~" {
				squery.language = row[0]
			} else if qPrefix == ":" {
				squery.owner = row[0]
			}

			ls(&squery) // invokes resultListAlfred
		}
	}
}

func resultListAlfred(results *bleve.SearchResult) {

	for _, gist := range results.Hits {

		// Get Private and Starred States
		isStarred := ifelse(gist.Fields["Starred"].(string) == "T", "⭐", "")
		isPrivate := ifelse(gist.Fields["Public"].(string) == "F", "🔒", "")

		statusText := ""
		if len(isStarred) > 0 || len(isPrivate) > 0 {
			statusText = fmt.Sprintf("%s%s| ", isStarred, isPrivate)
		}

		// Parse/format filenames
		filenameText := ""
		switch val := gist.Fields["Filename"].(type) {
		case string:
			filenameText = val
		case []interface{}:
			for _, s := range val {
				filenameText += s.(string) + " "
			}
		}
		filenameText = ifelse(filenameText != "", fmt.Sprintf("%s | ", filenameText), "")

		// Get Gist Content
		gistText := GistToText(gist)
		subtitleText := fmt.Sprintf("%s%s%s", statusText, filenameText, truncateString(gistText, 100))

		icon := resolveIcon(gist.Fields["Language"])
		it := wf.NewItem(gist.Fields["Description"].(string)).
			Icon(icon).
			Quicklook(gist.Fields["URL"].(string)).
			Copytext(gistText).
			Subtitle(subtitleText).
			Arg(gistText).
			Var("title", fmt.Sprintf("'%s'", gist.Fields["Description"].(string))).
			Valid(true)

		it.Cmd().
			Subtitle("Open Gist in browser").
			Arg(gist.Fields["URL"].(string)).
			Valid(true)

		it.Opt().
			Subtitle("Edit locally").
			Arg(fmt.Sprintf("%v", int(gist.Fields["IDX"].(float64)))).
			Valid(true)

	}
}

func fieldSummary(field string) {
	// Calculates frequencies for a given field
	facet := bleve.NewFacetRequest(field, 100000)
	query := query.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.AddFacet("count", facet)
	searchResults, err := dbIdx.Search(searchRequest)
	if err != nil {
		panic(err)
	}

	// term with highest occurrences in field name
	data := make([][]string, searchResults.Facets["count"].Terms.Len())
	for idx, val := range searchResults.Facets["count"].Terms {
		data[idx] = []string{val.Term, strconv.Itoa(val.Count)}
	}

	if outputFormat == "console" {
		fieldSummaryTable(field, data)
	} else if outputFormat == "alfred" {
		fieldSummaryAlfred(field, data)
	}
}

func outputGist(gistIdx int) {
	gist := lookupGist(gistIdx)
	fileset := parseGistFiles(gist)

	for _, file := range fileset {
		var xsize, _, _ = terminal.GetSize(0)
		var line = strings.Repeat("-", xsize-len(file["filename"])-50)
		var isPrivate string
		if gist.Fields["Public"] == "false" {
			isPrivate = "🔒"
		} else {
			isPrivate = "-"
		}
		if outputPipe() {
			fmt.Print(file["content"])
		} else {
			errlog.Printf("%s%s%s%s", greenText.Sprint(file["filename"]), isPrivate, line, file["language"])
			highlight(os.Stdout, file["filename"], file["content"], "terminal16m", "fruity")
			fmt.Fprintf(os.Stderr, "\n\n")
		}
	}
}

func fetchGistContent(gistIdx int) string {
	gist := lookupGist(gistIdx)
	fileset := parseGistFiles(gist)
	var result string
	for _, file := range fileset {
		result += file["content"]
	}
	return result
}
