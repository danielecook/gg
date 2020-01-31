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

// libConfig - Global library configuration
var libConfig = fmt.Sprintf("%s/config.json", getLibraryDirectory())

// Global library configuration
var libPath = fmt.Sprintf("%s/library.json", getLibraryDirectory())
var libDb = fmt.Sprintf("%s/db", getLibraryDirectory())

type configuration struct {
	AuthToken  string `json:"token"`
	LastUpdate string `json:"last_update"`
}

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

// parseLibrary
func parseLibrary(allGists []*github.Gist) []*Snippet {
	var Library []*Snippet
	Library = make([]*Snippet, len(allGists))

	errlog.Println(Bold("Loading Gists"))

	// Initialize progress bar
	bar := progressbar.New(len(allGists))
	for idx, gist := range allGists {
		items := make(map[github.GistFilename]github.GistFile)
		for k, _ := range gist.Files {
			var updated = gist.Files[k]
			var url = string(*gist.Files[k].RawURL)
			updated.Content = fetchContent(url)
			items[k] = updated
		}
		bar.Add(1)
		var f = Snippet{ID: gist.GetID(),
			Description: gist.GetDescription(),
			Public:      gist.GetPublic(),
			Files:       items,
			Comments:    gist.GetComments(),
			CreatedAt:   gist.GetCreatedAt(),
			UpdatedAt:   gist.GetUpdatedAt(),
			URL:         gist.GetHTMLURL(),
		}
		// Store gist in db
		//index.Index(gist.GetID(), f)
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

func setupLibrary(AuthToken string) bool {
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

func authenticate() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getConfig().AuthToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{Page: 0, PerPage: 100},
	}

	// Create search index with bleve
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(libDb, mapping)
	if err != nil {
		fmt.Println(err)
	}

	/* Use spinner for initial fetching of gist list */
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	s.Writer = os.Stderr
	s.Start() // Start the spinner
	/* Fetch list of users gists */
	var allGists []*github.Gist
	var Library []*Snippet
	page := 1
	for {

		gists, resp, err := client.Gists.List(ctx, "", opt)
		check(err)

		allGists = append(allGists, gists...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

		errlog.Printf("Loading [total=%v] [page=%v]\n", len(allGists), page)
		page += 1

	}
	errlog.Printf("Fetching complete [total=%v]\n", len(allGists))
	s.Stop()
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
