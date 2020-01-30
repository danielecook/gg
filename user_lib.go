package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var libConfig = fmt.Sprintf("%s/config.json", getLibraryDirectory())
var libPath = fmt.Sprintf("%s/library.json", getLibraryDirectory())
var libDb = fmt.Sprintf("%s/db", getLibraryDirectory())

type configuration struct {
	AuthToken  string `json:"token"`
	LastUpdate string `json:"last_update"`
}

type Snippet struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Public      bool      `json:"public"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
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
	fmt.Println(index)

	/* Fetch list of users gists */
	var allGists []*github.Gist
	var Library []*Snippet
	for {
		gists, resp, err := client.Gists.List(ctx, "", opt)
		check(err)

		allGists = append(allGists, gists...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		fmt.Printf("GETTING NEXT PAGE2\n")
	}
	Library = make([]*Snippet, len(allGists))
	for idx, gist := range allGists {
		var f = Snippet{ID: gist.GetID(),
			Description: gist.GetDescription(),
			Public:      gist.GetPublic(),
			Comments:    gist.GetComments(),
			CreatedAt:   gist.GetCreatedAt(),
			UpdatedAt:   gist.GetUpdatedAt(),
		}
		// Store gist in db
		index.Index(gist.GetID(), f)
		Library[idx] = &f
		fmt.Printf("%+v\n", f)
	}
	out, err := json.Marshal(allGists)
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
	fmt.Println(searchResults)
	// for _, gist := range allGists {
	// 	//gist2, resp, err := client.Gists.Get(ctx, *gist.ID)
	// 	//check(err)
	// 	//fmt.Printf("%s", gist2)
	// 	//fmt.Printf("%s", resp)
	// 	//fmt.Printf("%s\n\n", github.Stringify(gist))
	// }
}
