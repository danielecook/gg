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

	"github.com/blevesearch/bleve"
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
	ID          string                                  `json:"id"`
	Description string                                  `json:"description"`
	Public      bool                                    `json:"public"`
	Files       map[github.GistFilename]github.GistFile `json:"files"`
	Content     string                                  `json:"content"`
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

// parseLibrary
// Loads raw text from gists
// and parses
func parseLibrary(allGists []*github.Gist) []*Snippet {
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

	// Initialize progress bar
	bar := progressbar.New(len(allGists))
	for idx, gist := range allGists {
		items := make(map[github.GistFilename]github.GistFile)
		// Check if gist has already been loaded
		// if not, download files.
		if contains(existingGistIds, gist.GetNodeID()) == false {
			for k := range gist.Files {
				var updated = gist.Files[k]
				var url = string(*gist.Files[k].RawURL)
				updated.Content = fetchContent(url)
				items[k] = updated
			}
		}
		bar.Add(1)
		var f = Snippet{ID: gist.GetNodeID(),
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
	return Library
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

func initializeLibrary(AuthToken string) bool {
	deleteLibrary()
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
	jsonFile, err := os.Open(libConfig)
	check(err)
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
	/* Fetch list of users gists */
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

		errlog.Printf("Loading [total=%v] [page=%v]\n", len(allGists), page)
		page++
	}
	errlog.Printf("Fetching complete [total=%v]\n", len(allGists))
	s.Stop()

	// Parse the resulting library
	var Library []*Snippet
	Library = parseLibrary(allGists)

	for _, snippet := range Library {
		// Store gist in db
		index.Index(snippet.ID, snippet)
	}
	out, err := json.Marshal(Library)
	check(err)
	err = ioutil.WriteFile(libPath, out, 0644)
	check(err)

	query := bleve.NewMatchQuery("annotation")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(index.DocCount())
	fmt.Println(searchResults)
}
