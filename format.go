package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/blevesearch/bleve/search"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
)

// Generate a result table
func resultTable(results []*search.DocumentMatch, isQuery bool) {
	tableData := make([][]string, len(results))
	for idx, gist := range results {

		updatedAt := strings.Split(gist.Fields["UpdatedAt"].(string), "T")[0]

		tableData[idx] = []string{
			fmt.Sprintf("%v", gist.Fields["IDX"]),
			ifelse(gist.Fields["Starred"].(string) == "T", "⭐", ""),
			ifelse(gist.Fields["Public"].(string) == "F", "🔒", ""),
			gist.Fields["Description"].(string),
			fmt.Sprintf("%v", gist.Fields["Filename"]),
			fmt.Sprintf("%v", gist.Fields["Language"]),
			gist.Fields["Owner"].(string),
			updatedAt,
		}
		if isQuery {
			tableData[idx] = append(tableData[idx], fmt.Sprintf("%1.3f", gist.Score))
		}
	}
	// Render results
	table := tablewriter.NewWriter(os.Stdout)

	/*
		Header
	*/
	var header = []string{"ID", "⭐", "🔒", "Description", "Filename", "Language", "Author", "Updated"}
	if isQuery {
		header = append(header, "Score")
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

	/*
		Format
	*/
	// Terminal Window size
	var xsize, _, _ = terminal.GetSize(0)

	var colWidth int
	colWidth = (xsize / len(header))
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
}

func parseGistFiles(gist *search.DocumentMatch) map[string]map[string]string {
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
	return fileset
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
