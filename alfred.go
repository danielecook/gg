package main

// Package is called aw
import (
	"fmt"
	"log"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
)

// Workflow is the main API
var alfredQuery string
var _ = os.Setenv("alfred_workflow_bundleid", "1")
var _ = os.Setenv("alfred_workflow_cache", "1")
var _ = os.Setenv("alfred_workflow_data", "1")

var wf *aw.Workflow

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	wf = aw.New()
}

// Your workflow starts here
func run() {

	libsummary := librarySummary()

	args := wf.Args()
	argSet := strings.TrimSpace(strings.Join(args[1:], ""))
	if len(args) > 0 {
		alfredQuery = strings.TrimSpace(strings.Join(args[1:], ""))
	}
	if len(argSet) == 0 {
		wf.NewItem("Tags").
			Icon(tagIcon).
			Autocomplete("#").
			Subtitle(fmt.Sprintf("%v Tags", libsummary.tags))

		wf.NewItem("Language").
			Icon(languageIcon).
			Autocomplete("~").
			Subtitle(fmt.Sprintf("%v Languages", libsummary.languages))

		wf.NewItem("Starred").
			Icon(starIcon).
			Autocomplete("⭐").
			Subtitle(fmt.Sprintf("%v Languages", libsummary.languages))

		wf.NewItem("owner")
		wf.NewItem("sync")
		wf.NewItem("set-editor")
		wf.NewItem("login")
	} else {
		log.Println(alfredQuery)
		log.Println(strings.HasPrefix(alfredQuery, "#"))
		switch {
		case strings.HasPrefix(alfredQuery, "#"):
			fieldSummary("Tags")
		case strings.HasPrefix(alfredQuery, "~"):
			fieldSummary("Language")

		}

		wf.NewItem(argSet)
	}
	// Send results to Alfred
	wf.SendFeedback()
}

func runAlfred() {
	outputFormat = "alfred"
	wf.Run(run)
}
