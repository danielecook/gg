package main

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

// global search index
var dbIdx = *openDb()
var libDb = fmt.Sprintf("%s/db", getLibraryDirectory())

// openDb
// This function will initialize a new db if one does
// not exist or open an existing one.
func openDb() *bleve.Index {
	if _, err := os.Stat(libDb); os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		dbIdx, err := bleve.New(libDb, mapping)
		if err != nil {
			ThrowError("Error creating library", 1)
		}
		return &dbIdx
	}
	index, err := bleve.Open(libDb)
	if err != nil {
		ThrowError("Error opening library", 1)
	}
	return &index
}

func queryGists(docIds []string) *bleve.SearchResult {
	sr := bleve.NewSearchRequest(query.NewDocIDQuery(docIds))
	results, err := dbIdx.Search(sr)
	if err != nil {
		return nil
	}
	return results
}

func dumpDb() *bleve.SearchResult {
	dc, _ := dbIdx.DocCount()
	sr := bleve.NewSearchRequest(query.NewMatchAllQuery())
	sr.Fields = []string{"*"}
	sr.Size = int(dc)
	results, err := dbIdx.Search(sr) // bleve/index_impl, bleve/search/collector/topn.Collect
	if err != nil {
		return nil
	}
	return results
}
