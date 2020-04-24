package main

import (
	aw "github.com/deanishe/awgo"
)

var (
	newIcon        = &aw.Icon{"icons/new.png", aw.IconTypeImage}
	tagIcon        = &aw.Icon{"icons/tag.png", aw.IconTypeImage}
	collectionIcon = &aw.Icon{"icons/collection.png", aw.IconTypeImage}
	languageIcon   = &aw.Icon{"icons/language.png", aw.IconTypeImage}
	token_icon     = &aw.Icon{"icons/token.png", aw.IconTypeImage}
	starIcon       = &aw.Icon{"icons/star.png", aw.IconTypeImage}
	forkIcon       = &aw.Icon{"icons/forked.png", aw.IconTypeImage}
	latestIcon     = &aw.Icon{"icons/latest.png", aw.IconTypeImage}
	iconAvailable  = &aw.Icon{"icons/update-available.png", aw.IconTypeImage}

	// Languages
	languages = map[string]*aw.Icon{
		"py":     {"icn/python.png", aw.IconTypeImage},
		"python": {"icn/python.png", aw.IconTypeImage},

		"rb":   {"icn/ruby.png", aw.IconTypeImage},
		"ruby": {"icn/ruby.png", aw.IconTypeImage},

		"c": {"icn/c.png", aw.IconTypeImage},

		"cpp": {"icn/c++.png", aw.IconTypeImage},
		"c++": {"icn/c++.png", aw.IconTypeImage},

		"sh": {"icn/bash.png", aw.IconTypeImage},
		"r":  {"icn/r.png", aw.IconTypeImage},

		"md":       {"icn/markdown.png", aw.IconTypeImage},
		"markdown": {"icn/markdown.png", aw.IconTypeImage},

		"tsv":  {"icn/data.png", aw.IconTypeImage},
		"csv":  {"icn/data.png", aw.IconTypeImage},
		"data": {"icn/data.png", aw.IconTypeImage},
	}
)

func resolve_icon(s string) *aw.Icon {
	//log.Println(s)
	if v, ok := languages[s]; ok {
		return v
	}
	return &aw.Icon{}
}
