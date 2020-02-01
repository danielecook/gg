package main

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
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

// Format search results as table
func formatResults(results *bleve.SearchResult) {
	for _, gist := range results.Hits {
		fmt.Println(gist)
	}
}

// ls - the primary query interface
func ls(searchTerm string, sortBy string, tag string) *bleve.SearchResult {
	var qstring string

	if searchTerm != "" {
		qstring = fmt.Sprintf("%s", searchTerm)
	}

	if tag != "" {
		qstring = fmt.Sprintf("%s +Tags:%v", qstring, tag)
	}

	fmt.Println(qstring)

	q := query.NewQueryStringQuery(qstring)
	sr := bleve.NewSearchRequest(q)

	//	q := query.NewMatchAllQuery()
	//	sr = bleve.NewSearchRequest(q)
	//}
	dc, _ := DbIdx.DocCount()
	sr.Fields = []string{"*"}
	sr.Size = int(dc)
	//sr.SortBy([]string{"UpdatedAt"})
	results, err := DbIdx.Search(sr)
	if err != nil {
		fmt.Println("No Results")
	}
	// Implement tag filtering
	for _, gist := range results.Hits {
		fmt.Println(gist.Fields["Description"])
	}
	return results
}
