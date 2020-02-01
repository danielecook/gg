package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
	. "github.com/logrusorgru/aurora"
	"github.com/schollz/progressbar/v2"
	"golang.org/x/oauth2"
)

// global background context
var ctx = context.Background()

// libConfig - Global library configuration
var libConfig = fmt.Sprintf("%s/config.json", getLibraryDirectory())

// Global library configuration
var libPath = fmt.Sprintf("%s/library.json", getLibraryDirectory())

type configuration struct {
	AuthToken  string `json:"token"`
	LastUpdate string `json:"last_update"`
}

// Snippet - Used to store gist data
type Snippet struct {
	// The ID is actually the github Node ID which is unique to the given commit
	ID          string                                  `json:"id"`
	Description string                                  `json:"description"`
	Public      bool                                    `json:"public"`
	Files       map[github.GistFilename]github.GistFile `json:"files"`
	Comments    int                                     `json:"comments"`
	CreatedAt   time.Time                               `json:"CreatedAt"`
	UpdatedAt   time.Time                               `json:"UpdatedAt"`
	Snippet     string                                  `json:"Snippet"`
	URL         string                                  `json:"URL"`
	Commit      string                                  `json:"commit"`
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Fetch the raw text for a gist
func fetchContent(url string) *string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var result = string(content)
	return &result
}

func getLibraryDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(filepath.Join(usr.HomeDir, ".gg"))
}

func deleteLibrary() {
	os.RemoveAll(getLibraryDirectory())
}

func initializeLibrary(AuthToken string, rebuild bool) bool {
	if rebuild {
		deleteLibrary()
		// Reload index
		DbIdx = *openDb()
	}
	var config = configuration{
		AuthToken: string(AuthToken),
	}
	_ = os.Mkdir(getLibraryDirectory(), 0755)
	out, err := json.Marshal(config)
	check(err)
	err = ioutil.WriteFile(libConfig, out, 0644)
	check(err)

	return true
}

func getConfig() configuration {
	// Check that config exists
	if _, err := os.Stat(libConfig); os.IsNotExist(err) {
		errMsg := "No config found. Run 'gg login'"
		ThrowError(errMsg, 2)
	}
	jsonFile, err := os.Open(libConfig)
	if err != nil {
		errMsg := "JSON Parse Error. Run 'gg login'"
		ThrowError(errMsg, 2)
	}
	defer jsonFile.Close()

	configData, _ := ioutil.ReadAll(jsonFile)
	var config configuration
	json.Unmarshal(configData, &config)
	return config
}

// authenticate - Setup user authentication with github token
func authenticate() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getConfig().AuthToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func updateLibrary() {
	client := authenticate()

	// Add a spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	s.Writer = os.Stderr
	s.Start() // Start the spinner

	/*
		List User Gists
	*/
	var allGists []*github.Gist
	page := 1

	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{Page: 0, PerPage: 100},
	}

	for {

		gists, resp, err := client.Gists.List(ctx, "", opt)
		check(err)

		allGists = append(allGists, gists...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

		errlog.Printf("Listing [total=%v] [page=%v]\n", len(allGists), page)
		page++
	}
	errlog.Printf("Listing complete [total=%v]\n", len(allGists))
	s.Stop()

	/*
		Parse Library
	*/
	var Library []*Snippet
	Library = make([]*Snippet, len(allGists))

	errlog.Println(Bold("Loading Gists"))

	// Dump previous DB to determine whether
	// gists need to be updated.
	existingGists := dumpDb()
	existingGistIds := make([]string, len(existingGists.Hits))
	for idx, gist := range existingGists.Hits {
		existingGistIds[idx] = gist.ID
	}

	currentGistIds := make([]string, len(allGists))

	// Initialize progress bar
	bar := progressbar.New(len(allGists))
	for idx, gist := range allGists {
		gistID := fmt.Sprintf("%v::%v", gist.GetID(), gist.GetUpdatedAt())
		currentGistIds[idx] = gistID

		items := make(map[github.GistFilename]github.GistFile)
		// Check if gist has already been loaded
		// if not, download files.
		if contains(existingGistIds, gistID) == false {
			for k := range gist.Files {
				var updated = gist.Files[k]
				var url = string(*gist.Files[k].RawURL)
				updated.Content = fetchContent(url)
				items[k] = updated
			}
		}

		bar.Add(1)
		var f = Snippet{
			ID:          gistID,
			Description: gist.GetDescription(),
			Public:      gist.GetPublic(),
			Files:       items,
			Comments:    gist.GetComments(),
			CreatedAt:   gist.GetCreatedAt(),
			UpdatedAt:   gist.GetUpdatedAt(),
			URL:         gist.GetHTMLURL(),
		}
		// Store gist in db
		Library[idx] = &f
	}

	batch := DbIdx.NewBatch()
	// Delete gist IDs that no longer exist
	for _, existingId := range existingGistIds {
		if contains(currentGistIds, existingId) == false {
			fmt.Printf("REMOVING %v", existingId)
			fmt.Println(batch.Size())
			batch.Delete(existingId)
		}
	}

	for _, gist := range Library {
		// Store gist in db if it is not there already
		if contains(existingGistIds, gist.ID) == false {
			batch.Index(gist.ID, gist)
		}
	}

	// Execute database updates
	DbIdx.Batch(batch)

	out, err := json.Marshal(Library)
	check(err)
	err = ioutil.WriteFile(libPath, out, 0644)
	check(err)

	docCount, err := DbIdx.DocCount()
	fmt.Println()
	errlog.Println(Bold(Green(fmt.Sprintf("Loaded %v gists", docCount))))
}
