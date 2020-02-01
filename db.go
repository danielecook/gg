package main

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/pkg/errors"
)

// global search index
var DbIdx = *openDb()
var libDb = fmt.Sprintf("%s/db", getLibraryDirectory())

// openDb
// This function will initialize a new db if one does
// not exist or open an existing one.
func openDb() *bleve.Index {
	if _, err := os.Stat(libDb); os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		DbIdx, err := bleve.New(libDb, mapping)
		if err != nil {
			ThrowError("Error creating library", 1)
		}
		return &DbIdx
	}
	index, err := bleve.Open(libDb)
	if err != nil {
		ThrowError("Error opening library", 1)
	}
	return &index
}

type Note struct {
	ID         string
	Body       string
	Title      string                  // some short title of this note
	Fragments  search.FieldFragmentMap // only used for query results, show a snippet of text around found terms
	AccessTime int64                   // time.Unix(), see NewIndexMapping():accessTime_fmap for FieldMapping
}

const (
	BODY        = "Body"
	ACCESS_TIME = "AccessTime"
	TITLE       = "Title"

	DEFAULT_BATCH_SIZE = 1000
)

func queryGists(docIds []string) *bleve.SearchResult {
	sr := bleve.NewSearchRequest(query.NewDocIDQuery(docIds))
	results, err := DbIdx.Search(sr)
	if err != nil {
		return nil
	}
	return results
}

func dumpDb() *bleve.SearchResult {
	dc, _ := DbIdx.DocCount()
	sr := bleve.NewSearchRequest(query.NewMatchAllQuery())
	sr.Fields = []string{"*"}
	sr.Size = int(dc)
	results, err := DbIdx.Search(sr) // bleve/index_impl, bleve/search/collector/topn.Collect
	if err != nil {
		return nil
	}
	return results
}

func ls() ([]Note, error) {
	dc, _ := DbIdx.DocCount()
	sr := bleve.NewSearchRequest(query.NewMatchAllQuery())
	sr.Fields = []string{"*"}
	sr.Size = int(dc)
	results, err := DbIdx.Search(sr) // bleve/index_impl, bleve/search/collector/topn.Collect
	if err != nil {
		return nil, errors.Wrap(err, "search failed")
	}

	notes := make([]Note, len(results.Hits))
	for idx := range notes {
		notes[idx] = toNote(results.Hits[idx])
		var r = results.Hits[idx]
		fmt.Println(r.ID)
		fmt.Println(r.Fields["description"])
	}
	fmt.Println(len(results.Hits))

	return notes, nil
}

func toNote(doc *search.DocumentMatch) Note {
	note := Note{}
	note.ID = doc.ID
	if atime, ok := doc.Fields[ACCESS_TIME]; ok {
		if v, ok := atime.(float64); ok {
			note.AccessTime = int64(v)
		}
	}
	if body, ok := doc.Fields[BODY]; ok {
		if v, ok := body.(string); ok {
			note.Body = v
		}
	}
	if title, ok := doc.Fields[TITLE]; ok {
		if v, ok := title.(string); ok {
			note.Title = v
		}
	}
	if doc.Fragments != nil {
		note.Fragments = doc.Fragments
	}
	return note
}
