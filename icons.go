package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/deanishe/awgo"
)

var (
	newIcon        = &aw.Icon{"icons/new.png", aw.IconTypeImage}
	tagIcon        = &aw.Icon{"icons/tag.png", aw.IconTypeImage}
	collectionIcon = &aw.Icon{"icons/collection.png", aw.IconTypeImage}
	languageIcon   = &aw.Icon{"icons/language.png", aw.IconTypeImage}
	tokenIcon      = &aw.Icon{"icons/token.png", aw.IconTypeImage}
	starIcon       = &aw.Icon{"icons/star.png", aw.IconTypeImage}
	forkIcon       = &aw.Icon{"icons/forked.png", aw.IconTypeImage}
	latestIcon     = &aw.Icon{"icons/latest.png", aw.IconTypeImage}
	iconAvailable  = &aw.Icon{"icons/update-available.png", aw.IconTypeImage}
)

func loadIcons() map[string]string {
	iconDir, err := os.Open("./icons/")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer iconDir.Close()
	iconFile, err := iconDir.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}
	iconSet = make(map[string]string, len(iconFile))
	for _, val := range iconFile {
		iconSet[strings.Replace(filepath.Base(val), ".png", "", 1)] = fmt.Sprintf("icons/%s", val)
	}
	return iconSet
}

func resolveIcon(i interface{}) *aw.Icon {
	/*
		If it is a language query always
	*/

	// If it is a language query return that result
	if v, ok := iconSet[strings.ToLower(squery.language)]; ok {
		return &aw.Icon{Type: aw.IconTypeImage,
			Value: v}
	}

	// Otherwise convert string and return
	{
		switch t := i.(type) {
		case string:
			if v, ok := iconSet[strings.ToLower(t)]; ok {
				return &aw.Icon{Type: aw.IconTypeImage,
					Value: v}
			}
		}
	}
	return &aw.Icon{}
}
