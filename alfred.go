package main

// Package is called aw
import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	aw "github.com/deanishe/awgo"
)

// Workflow is the main API
var (
	alfredQuery string
	_           = os.Setenv("alfred_workflow_bundleid", "1")
	_           = os.Setenv("alfred_workflow_cache", "1")
	_           = os.Setenv("alfred_workflow_data", "1")
	wf          *aw.Workflow
	maxResults  = 100
	iconSet     map[string]string
)

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	wf = aw.New(aw.HelpURL("http://www.github.com/danielecook/gg"),
		aw.MaxResults(maxResults))
}

// Your workflow starts here
func run() {

	libsummary := librarySummary()
	iconSet = loadIcons()

	args := wf.Args()
	argSet := strings.Join(args[1:], "")

	// Run sync operation in the background
	wf.RunInBackground("sync", exec.Command("gg", "sync"))

	if len(args) > 0 {
		alfredQuery = strings.Join(args[1:], "")
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
			Subtitle(fmt.Sprintf("%v Starred", libsummary.starred))

		wf.NewItem("owner").
			Icon(randomOwnerIcon()).
			Autocomplete(":").
			Subtitle(fmt.Sprintf("%v Owners", libsummary.owners))
		wf.NewItem("sync")
		wf.NewItem("set-editor")
		wf.NewItem("login")
	} else {
		switch {
		case strings.HasPrefix(alfredQuery, "#"):
			fieldSummary("Tags")
		case strings.HasPrefix(alfredQuery, "~"):
			fieldSummary("Language")
		case strings.HasPrefix(alfredQuery, ":"):
			fieldSummary("Owner")
		default:
			queryGistsAlfred(alfredQuery)
		}

		//wf.NewItem(argSet)
	}
	// Send results to Alfred
	wf.SendFeedback()
}

func runAlfred() {
	outputFormat = "alfred"
	wf.Run(run)
}
