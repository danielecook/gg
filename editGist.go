package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

var gistTemplate = []byte(`# GIST FORM: Edit Metadata below
# ==============================
# description: {{ .Description }}
# starred: {{ .Starred }}
# public: {{ .Public }}
# ==============================
{{- range $elements := .Files }}
{{ fname_line (Deref $elements.Filename) }}::>>>
{{ (Deref $elements.Content ) -}}
{{end}}
`)

func cleanLine(s string) string {
	return strings.Trim(strings.Split(s, ":")[1], " ")
}

func appendFile(items map[github.GistFilename]github.GistFile, filename string, fileContent string) {
	if filename != "" && fileContent != "" {
		items[github.GistFilename(filename)] = github.GistFile{
			Filename: &filename,
			Content:  &fileContent,
		}
	}
}

func parseGistTemplate(s string) (Snippet, error) {
	result := strings.Split(s, "\n")
	nlines := len(result)
	var eGist Snippet
	var filename string
	var fileContent string
	var items map[github.GistFilename]github.GistFile
	items = make(map[github.GistFilename]github.GistFile, 1)
	for idx, line := range result {
		if idx <= 5 {
			switch {
			case strings.HasPrefix(line, "# description:"):
				fmt.Println("desc")
				eGist.Description = cleanLine(line)
			case strings.HasPrefix(line, "# starred:"):
				eGist.Starred = parseTrueFalse(cleanLine(line))
			case strings.HasPrefix(line, "# public:"):
				eGist.Public = parseTrueFalse(cleanLine(line))
			}
		}
		if idx > 5 {
			switch {
			case strings.HasSuffix(line, "::>>>"):
				// Store content
				appendFile(items, filename, fileContent)

				// Initate new file
				filename = strings.Trim(line[0:len(line)-5], "-")
			default:
				fileContent += line + "\n"
			}
		}
		if idx == nlines-1 {
			appendFile(items, filename, fileContent)
		}

	}
	fmt.Println(items)

	return Snippet{}, nil
}
