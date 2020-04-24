package main

import (
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

	// Languages
	languages = map[string]*aw.Icon{
		"py":     {"icons/python.png", aw.IconTypeImage},
		"python": {"icons/python.png", aw.IconTypeImage},

		"rb":   {"icons/ruby.png", aw.IconTypeImage},
		"ruby": {"icons/ruby.png", aw.IconTypeImage},

		"c": {"icons/c.png", aw.IconTypeImage},

		"cpp": {"icons/c++.png", aw.IconTypeImage},
		"c++": {"icons/c++.png", aw.IconTypeImage},

		"sh": {"icons/bash.png", aw.IconTypeImage},
		"r":  {"icons/r.png", aw.IconTypeImage},

		"md":       {"icons/markdown.png", aw.IconTypeImage},
		"markdown": {"icons/markdown.png", aw.IconTypeImage},

		"tsv":  {"icons/data.png", aw.IconTypeImage},
		"csv":  {"icons/data.png", aw.IconTypeImage},
		"data": {"icons/data.png", aw.IconTypeImage},
	}
)

func resolveIcon(s string) *aw.Icon {
	//log.Println(s)
	if v, ok := languages[s]; ok {
		return v
	}
	return &aw.Icon{}
}
