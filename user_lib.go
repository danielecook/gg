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
	"regexp"
	"sort"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
	. "github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v2"
	"golang.org/x/oauth2"
)

// global background context
var ctx = context.Background()

// libConfig - Global library configuration
var libConfig = fmt.Sprintf("%s/config.json", getLibraryDirectory())

// Global library configuration
var libPath = fmt.Sprintf("%s/library.json", getLibraryDirectory())
var libTagsPath = fmt.Sprintf("%s/tags.json", getLibraryDirectory())

type configuration struct {
	AuthToken  string `json:"token"`
	LastUpdate string `json:"last_update"`
}

// Snippet - Used to store gist data
type Snippet struct {
	// The ID is actually the github Node ID which is unique to the given commit
	ID          string                                  `json:"ID"`
	Description string                                  `json:"Description"`
	Public      bool                                    `json:"Public"`
	Files       map[github.GistFilename]github.GistFile `json:"Files"`
	Tags        []string                                `json:"Tags"`
	Comments    int                                     `json:"Comments"`
	CreatedAt   time.Time                               `json:"CreatedAt"`
	UpdatedAt   time.Time                               `json:"UpdatedAt"`
	Snippet     string                                  `json:"Snippet"`
	URL         string                                  `json:"URL"`
	Commit      string                                  `json:"Commit"`
}

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Extract tags from the description
func parseTags(s string) []string {
	re := regexp.MustCompile(`#([A-Za-z0-9]+)`)
	r := re.FindAllStringSubmatch(s, -1)
	var tagSet []string
	if len(r) > 0 {
		for _, tags := range r {
			tagSet = append(tagSet, tags[1])
		}
		return tagSet
	}
	return []string{}
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
	var LibraryTags []string
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
		tags := parseTags(gist.GetDescription())
		for _, t := range tags {
			LibraryTags = append(LibraryTags, t)
		}
		var f = Snippet{
			ID:          gistID,
			Description: gist.GetDescription(),
			Public:      gist.GetPublic(),
			Files:       items,
			Tags:        tags,
			Comments:    gist.GetComments(),
			CreatedAt:   gist.GetCreatedAt(),
			UpdatedAt:   gist.GetUpdatedAt(),
			URL:         gist.GetHTMLURL(),
		}

		// Store gist in db
		Library[idx] = &f
	}

	errlog.Println(Bold("Indexing Gists"))
	batch := DbIdx.NewBatch()

	// Delete gist IDs that no longer exist
	for _, existingId := range existingGistIds {
		if contains(currentGistIds, existingId) == false {
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

	/*
		Store JSON
	*/
	// Tags
	tagCounts := counter(LibraryTags)
	var Tags []*Tag
	for key, val := range tagCounts {
		Tags = append(Tags, &Tag{Name: key, Count: val})
	}
	out, err := json.Marshal(Tags)
	check(err)
	err = ioutil.WriteFile(libTagsPath, out, 0644)
	check(err)

	// Library
	out, err = json.Marshal(Library)
	check(err)
	err = ioutil.WriteFile(libPath, out, 0644)
	check(err)

	docCount, err := DbIdx.DocCount()
	fmt.Println()
	errlog.Println(Bold(Green(fmt.Sprintf("Loaded %v gists", docCount))))
}

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

// List all tags in library
func listTags() {
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
