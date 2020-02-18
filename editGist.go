package main

import (
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
{{end}}`)

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

func parseGistTemplate(s string) (github.Gist, bool, error) {
	// Parses GistTemplate and returns
	// a gist object
	result := strings.Split(s, "\n")
	nlines := len(result)
	var filename string
	var fileContent string
	var description string
	var starred bool
	var public bool
	var err error
	var items map[github.GistFilename]github.GistFile
	items = make(map[github.GistFilename]github.GistFile, 1)
	for idx, line := range result {
		if idx <= 5 {
			switch {
			case strings.HasPrefix(line, "# description:"):
				description = cleanLine(line)
			case strings.HasPrefix(line, "# starred:"):
				starred, err = parseTrueFalse(cleanLine(line))
			case strings.HasPrefix(line, "# public:"):
				public, err = parseTrueFalse(cleanLine(line))
			}
			if err != nil {
				return github.Gist{}, false, err
			}
		}
		if idx > 5 {
			switch {
			case strings.HasSuffix(line, "::>>>"):
				// Store content
				appendFile(items, filename, fileContent)

				// Initate new file
				fileContent = ""
				filename = strings.Trim(line[0:len(line)-5], "-")
			default:
				fileContent += line + "\n"
			}
		}
		if idx == nlines-1 {
			appendFile(items, filename, strings.TrimSuffix(fileContent, "\n"))
		}

	}

	// Need to handle starring manually

	var resultGist = github.Gist{
		Description: &description,
		Public:      &public,
		Files:       items,
	}

	return resultGist, starred, nil
}
