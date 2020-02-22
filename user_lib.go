package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blevesearch/bleve/search"
	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
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
	AuthToken string `json:"token"`
	Login     string `json:"login"`
}

type gistSort []*github.Gist

func (e gistSort) Len() int {
	return len(e)
}

func (e gistSort) Less(i, j int) bool {
	return e[i].UpdatedAt.Before(e[j].GetUpdatedAt())
}

func (e gistSort) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// Snippet - Used to store gist data
type Snippet struct {
	// The ID is actually the github Node ID which is unique to the given commit
	// IDX is A convenience numeric ID for listing individual snippets
	ID          string                                  `json:"ID"`
	GistID      string                                  `json:"GistID"`
	IDX         int                                     `json:"IDX"`
	Owner       string                                  `json:"Owner"`
	Description string                                  `json:"Description"`
	Public      string                                  `json:"Public"`
	Starred     string                                  `json:"Starred"`
	Files       map[github.GistFilename]github.GistFile `json:"Files"`
	NFiles      int                                     `json:"NFiles"`
	Language    []string                                `json:"Language"`
	Filename    []string                                `json:"Filename"`
	Tags        []string                                `json:"Tags"`
	Comments    int                                     `json:"Comments"`
	CreatedAt   time.Time                               `json:"CreatedAt"`
	UpdatedAt   time.Time                               `json:"UpdatedAt"`
	URL         string                                  `json:"URL"`
}

// Generate list of IDs for gists
func idMap(gistSet []*github.Gist) map[string]*github.Gist {
	m := make(map[string]*github.Gist)
	var key string
	for _, gist := range gistSet {
		key = getGistRecID(gist)
		m[key] = gist
	}
	return m
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
		dbIdx = *openDb()
	}

	client, _ := authenticate(AuthToken)
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		ThrowError(fmt.Sprintf("Error authenticating: %s", resp), 1)
	}

	var config = configuration{
		AuthToken: string(AuthToken),
		Login:     string(user.GetLogin()),
	}
	_ = os.Mkdir(getLibraryDirectory(), 0755)
	out, err := json.Marshal(config)
	check(err)
	err = ioutil.WriteFile(libConfig, out, 0644)
	check(err)

	return true
}

func getConfig() (configuration, error) {
	// Check that config exists
	var blankConfig configuration
	if _, err := os.Stat(libConfig); os.IsNotExist(err) {
		return blankConfig, errors.New("No config found. Run 'gg sync --token <github token>'")
	}
	jsonFile, err := os.Open(libConfig)
	if err != nil {
		return blankConfig, errors.New("JSON Parse Error. Run 'gg sync --rebuild'")
	}
	defer jsonFile.Close()

	configData, _ := ioutil.ReadAll(jsonFile)
	var config configuration
	json.Unmarshal(configData, &config)
	return config, nil
}

// authenticate - Setup user authentication with github token
func authenticate(authToken string) (*github.Client, string) {
	var login string
	if authToken == "" {
		config, err := getConfig()
		if err != nil {
			ThrowError(err.Error(), 1)
		}
		authToken = config.AuthToken
		login = config.Login
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	client.UserAgent = "github.com/danielecook/gg"
	return client, login
}

func newGist(fileSet map[string]string, description string, public bool) {
	client, _ := authenticate("")

	var gist github.Gist
	files := map[github.GistFilename]github.GistFile{}

	for fname, item := range fileSet {
		fset := new(string)
		*fset = fname
		files[github.GistFilename(fname)] = github.GistFile{
			Content:  &item,
			Filename: fset,
		}
	}

	gist.Description = &description
	gist.Files = files
	gist.Public = &public

	resultGist, _, err := client.Gists.Create(ctx, &gist)
	if err != nil {
		ThrowError(fmt.Sprintf("Error: %s", err), 1)
	}
	// Add record to database
	gistDbRec := gistDbRecord(resultGist, nextIdx(), []string{})
	dbIdx.Index(gistDbRec.ID, gistDbRec)
	// Print URL on success
	boldUnderline.Println(*resultGist.HTMLURL)
}

func editGist(gistID int) {
	// TODO: Split out template portion/editing for creating new gists...
	client, _ := authenticate("")

	dbGist := lookupGist(gistID)
	gistFiles := parseGistFilesStruct(dbGist)

	dbStarred := dbGist.Fields["Starred"].(string) == "T"

	var params = Snippet{
		Description: dbGist.Fields["Description"].(string),
		Starred:     dbGist.Fields["Starred"].(string),
		Public:      dbGist.Fields["Public"].(string),
		Files:       gistFiles,
	}

	// Use first filename ext
	var ext string
	for _, item := range gistFiles {
		ext = filepath.Ext(*item.Filename)
		break
	}

	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("gist.*.%s", ext))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	t, err := template.New("tname").Funcs(template.FuncMap{
		"Deref": func(i *string) string { return *i },
		"fname_line": func(fname string) string {
			return fmt.Sprintf("%s%s", fname, strings.Repeat("-", (100-len(fname))))
		},
	}).Parse(string(gistTemplate))
	check(err)
	buf := new(bytes.Buffer)
	t.Execute(buf, params)

	// Write the header of the file
	err = ioutil.WriteFile(tmpfile.Name(), buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	var cmd *exec.Cmd
	var editor = "subl"
	// editorPath := os.Getenv("EDITOR")
	if editor == "subl" {
		cmd = exec.Command("subl", "--wait", fmt.Sprintf("%s", tmpfile.Name()))
	} else if editor == "nano" {
		cmd = exec.Command("nano", "-t", tmpfile.Name())
		cmd.Stdin = os.Stdout
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	edit, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		ThrowError("Error reading output", 1)
	}

	var starred bool
	var eGist github.Gist
	eGist, starred, err = parseGistTemplate(string(edit))

	// If filenames were removed from the original, set them to NULL to delete.
	for fname := range gistFiles {
		if (eGist.Files[fname] == github.GistFile{}) {
			var nullFile *string = nil
			eGist.Files[fname] = github.GistFile{
				Filename: nullFile,
			}
		}
	}

	if err != nil {
		// Reload template here with comment
	}

	greenText.Printf("Saving %v [%v]\n", int(dbGist.Fields["IDX"].(float64)), dbGist.Fields["GistID"].(string))
	resultGist, response, err := client.Gists.Edit(ctx, dbGist.Fields["GistID"].(string), &eGist)
	if err != nil {
		fmt.Printf("Error: %v", err)
		ThrowError(response.String(), 1)
	}

	// If star status has changed, update
	var starErr error
	var starIds []string
	if dbStarred != starred {
		if starred {
			_, starErr = client.Gists.Star(ctx, resultGist.GetID())
		} else {
			_, starErr = client.Gists.Unstar(ctx, resultGist.GetID())
		}
		if starErr != nil {
			ThrowError("Error updating star", 1)
		}
		starIds = []string{getGistRecID(resultGist)}
	}
	// Delete the old record, and insert the new record below.
	// Retain the same 'IDX' as before.
	batch := dbIdx.NewBatch()
	editGistDbRec := gistDbRecord(resultGist, int(dbGist.Fields["IDX"].(float64)), starIds)
	batch.Index(editGistDbRec.ID, editGistDbRec)
	batch.Delete(dbGist.ID)
	dbIdx.Batch(batch)

	boldUnderline.Println(*resultGist.HTMLURL)
}

func rmGist(gistID int) {
	client, _ := authenticate("")

	gist := lookupGist(gistID)
	_, err := client.Gists.Delete(ctx, gist.Fields["GistID"].(string))
	if err != nil {
		ThrowError(fmt.Sprintf("Error: %s", err), 1)
	}
	// Print URL on success
	msg := fmt.Sprintf("Removed %s\n", gist.Fields["GistID"].(string))

	// Remove from search index
	dbIdx.Delete(gist.ID)
	successMsg(msg)
}

func getGistRecID(gist *github.Gist) string {
	return fmt.Sprintf("%v::%v", gist.GetID(), gist.GetUpdatedAt())
}

func gistDbRecord(gist *github.Gist, idx int, starIDs []string) Snippet {
	gistRecID := getGistRecID(gist)
	ch := make(chan *string, 50)
	items := make(map[github.GistFilename]github.GistFile)
	// Check if gist has already been loaded
	// if not, download files.
	filenames := []string{}
	languages := []string{}
	// Check whether document exists
	if _, err := dbIdx.Document(gistRecID); err == nil {
		for k := range gist.Files {
			var updated = gist.Files[k]
			// If RawURL is nil, the gist was generated
			// locally and does not need to be retrieved.
			if gist.Files[k].RawURL != nil {
				var url = string(*gist.Files[k].RawURL)
				go fetchContent(url, ch)
				updated.Content = <-ch
			}
			items[k] = updated
			if gist.Files[k].Filename != nil {
				filenames = append(filenames, *gist.Files[k].Filename)
			}
			if gist.Files[k].Language != nil {
				languages = append(languages, *gist.Files[k].Language)
			}
		}
	}
	tags := parseTags(gist.GetDescription())
	var sn = Snippet{
		ID:          getGistRecID(gist),
		GistID:      gist.GetID(),
		IDX:         idx,
		Owner:       string(gist.GetOwner().GetLogin()),
		Description: gist.GetDescription(),
		Public:      trueFalse(gist.GetPublic()),
		Files:       items,
		Language:    languages,
		Filename:    filenames,
		Starred:     trueFalse(contains(starIDs, gistRecID)),
		NFiles:      len(items),
		Tags:        tags,
		Comments:    gist.GetComments(),
		CreatedAt:   gist.GetCreatedAt(),
		UpdatedAt:   gist.GetUpdatedAt(),
		URL:         gist.GetHTMLURL(),
	}
	return sn
}

func updateLibrary() {
	client, username := authenticate("")

	// Add a spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	s.Writer = os.Stderr
	s.Start() // Start the spinner

	/*
		List User Gists
	*/
	var allGists []*github.Gist
	var starredGists []*github.Gist

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

	since, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 1987")
	opt = &github.GistListOptions{
		ListOptions: github.ListOptions{Page: 0, PerPage: 100},
		Since:       since,
	}

	gistMap := idMap(allGists)
	// Get starred gists
	for {
		gists, resp, err := client.Gists.ListStarred(ctx, opt)
		check(err)
		starredGists = append(starredGists, gists...)

		// Append starred gists not present in the library
		for _, gist := range starredGists {
			key := fmt.Sprintf("%v::%v", gist.GetID(), gist.GetUpdatedAt())
			if gistMap[key] == nil {
				allGists = append(allGists, gist)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

		errlog.Printf("Fetching starred [total=%v] [page=%v]\n", len(allGists), page)
		page++
	}

	starIDs := make([]string, len(starredGists))
	for idx, gist := range starredGists {
		starIDs[idx] = getGistRecID(gist)
	}

	errlog.Printf("Listing complete [total=%v]\n", len(allGists))
	s.Stop()

	sort.Sort(gistSort(allGists))

	// Parse library
	var Library []*Snippet
	Library = make([]*Snippet, len(allGists))
	boldMsg(fmt.Sprintf("Loading Gists for %s\n", username))

	// Dump previous DB to determine whether
	// gists need to be updated.
	existingGists := dumpDb()
	existingGistIds := make([]string, len(existingGists.Hits))
	for idx, gist := range existingGists.Hits {
		existingGistIds[idx] = gist.ID
	}
	// Initialize progress bar
	bar := progressbar.New(len(allGists))

	currentGistIds := make([]string, len(allGists))
	// Not sure if this concurrent method is working in parallel or not...
	idStart := nextIdx()
	for i, gist := range allGists {
		// Store gist in db
		var gistDbRec Snippet
		if contains(existingGistIds, getGistRecID(gist)) == false {
			// Calculate nextIdx so IDs are static unless
			// a rebuild is performed.
			gistDbRec = gistDbRecord(gist, idStart+i, starIDs)
		}
		currentGistIds[i] = getGistRecID(gist)
		Library[i] = &gistDbRec
		bar.Add(1)
	}

	boldMsg("Indexing Gists\n")
	batch := dbIdx.NewBatch()

	// Delete gist IDs that no longer exist
	for _, existingID := range existingGistIds {
		if contains(currentGistIds, existingID) == false {
			batch.Delete(existingID)
		}
	}

	for _, gist := range Library {
		// Store gist in db if it is not there already
		if contains(existingGistIds, gist.ID) == false {
			batch.Index(gist.ID, gist)
		}
	}

	// Execute database updates
	dbIdx.Batch(batch)

	/*
		Store JSON
	*/

	// Library
	out, err := json.Marshal(Library)
	check(err)
	err = ioutil.WriteFile(libPath, out, 0644)
	check(err)

	docCount, err := dbIdx.DocCount()
	fmt.Println()
	successMsg(fmt.Sprintf("Loaded %v gist%s\n", docCount, ifelse(docCount == 1, "", "s")))
}

func libExists() bool {
	if _, err := os.Stat(libConfig); os.IsNotExist(err) {
		return false
	}
	return true
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

func parseGistFilesStruct(gist *search.DocumentMatch) map[github.GistFilename]github.GistFile {
	// Converts gistfiles struct for use in Snippet
	files := parseGistFiles(gist)
	var result map[github.GistFilename]github.GistFile
	result = make(map[github.GistFilename]github.GistFile, len(files))
	for _, item := range files {
		var content = item["content"]
		var fname = item["filename"]
		var fset = github.GistFilename(fname)
		result[fset] = github.GistFile{
			Content:  &content,
			Filename: &fname,
		}
	}
	return result
}
