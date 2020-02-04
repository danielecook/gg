package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/blevesearch/bleve/search"
	. "github.com/logrusorgru/aurora"
	"github.com/pkg/browser"
	"github.com/urfave/cli"
)

var errlog = log.New(os.Stderr, "", 0)

func main() {

	// These strings are reserved for commands
	// and cannot be searched.
	var queryReserve = []string{"new", "r", "open", "edit", "search", "login", "update", "tag", "tags", "-h", "--help", "help", "ls"}
	var searchTerm string

	app := cli.NewApp()

	app.Name = "gg"
	app.Usage = "A tool for Github Gists\n\n\t gg <search term> - quick search\n\n\t gg <ID> - retrieve gist"
	app.Version = "0.0.1"
	app.EnableBashCompletion = true

	app.Authors = []cli.Author{
		{
			Name:  "Daniel Cook",
			Email: "danielecook@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:                   "new",
			Usage:                  "Create a new gist",
			Category:               "Gists",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				var fileSet map[string]string
				fileSet = make(map[string]string)
				if c.Bool("clipboard") {
					/* New from clipboard */
					fmt.Println("clip")
					content, err := clipboard.ReadAll()
					if err != nil {
						ThrowError("Error reading from clipboard", 1)
					}
					fileSet[c.String("filename")] = content
				} else if inputPipe() {
					/* New from stdin */
					fmt.Println("input")
					bytes, err := ioutil.ReadAll(os.Stdin)
					if err != nil {
						ThrowError("Error reading from stdin", 1)
					}
					content := string(bytes)
					fileSet[c.String("filename")] = content
				} else {
					/* New from list of files */
					if c.NArg() > 0 {
						if c.String("filename") != "" {
							ThrowError("Cannot use --filename with files", 1)
						}
						for _, fname := range c.Args() {
							bytes, err := ioutil.ReadFile(fname)
							if err != nil {
								ThrowError(fmt.Sprintf("Error reading %s", fname), 1)
							}
							fileSet[fname] = string(bytes)
						}
					}
				}
				if len(fileSet) == 0 {
					ThrowError("No content supplied (use --clipboard, stdin, or files)", 1)
				}
				createGist(fileSet, c.String("description"), c.Bool("private") == false)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d, description",
					Usage: "Set the description for gist",
				},
				cli.StringFlag{
					Name:  "f, filename",
					Usage: "Set the filename with --clipboard or stdin",
				},
				cli.BoolFlag{
					Name: "p, private",
				},
				cli.BoolFlag{
					Name: "c, clipboard",
				},
			},
		},
		{
			Name:                   "open",
			Usage:                  "Open gist in browser",
			Category:               "Gists",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				var gist *search.DocumentMatch
				if v, err := strconv.Atoi(c.Args().Get(0)); err == nil {
					gist = lookupGist(v)
				} else {
					ThrowError("Invalid Index", 1)
				}
				browser.OpenURL(gist.Fields["URL"].(string))
				return nil
			},
		},
		{
			Name:                   "edit",
			Usage:                  "Edit a gist using $EDITOR",
			Category:               "Gists",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:      "login",
			Usage:     "Login and Setup your gist library",
			UsageText: "\n\t\tgg login [Authentication Token KEY]\n",
			Category:  "Library",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "authentication_token",
					EnvVar: "TOKEN",
				},
				cli.BoolFlag{
					Name:  "r, rebuild",
					Usage: "Rebuild library",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("authentication_token") == "" {
					/* gg login */
					errMsg := Bold(Red("\n\tError - Please supply your Authentication Token\n"))
					return cli.NewExitError(errMsg, 2)
				}
				if len(c.String("authentication_token")) != 40 {
					/* gg login <wrong_length> */
					return cli.NewExitError(Bold(Red("\n\tThe API Key should be 40 characters\n\n")), 32)
				}
				/* Store Token */
				initializeLibrary(c.String("authentication_token"), c.Bool("rebuild"))
				updateLibrary()
				return nil
			},
		},
		{
			Name:      "update",
			Usage:     "Update gist library",
			UsageText: "\n\t\tgg update\n",
			Category:  "Library",
			Action: func(c *cli.Context) error {
				updateLibrary()
				return nil
			},
		},
		{
			Name:                   "r",
			Usage:                  "Retrieve a single gist",
			UsageText:              "\n\t\tgg ls [options] [query]\n\n\t\tquery - Searches most fields",
			Category:               "Query",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				if v, err := strconv.Atoi(c.Args().Get(0)); err == nil {
					if c.Bool("clipboard") {
						clipboard.WriteAll(fetchGistContent(v))
					}
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "c, clipboard",
					Usage: "Copy to clipboard",
				},
			},
		},
		{
			Name:                   "ls",
			Usage:                  "List, Search and filter",
			UsageText:              "\n\t\tgg ls [options] [query]\n\n\t\tquery - Searches most fields",
			Category:               "Query",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				if v, err := strconv.Atoi(c.Args().Get(0)); err == nil {
					outputGist(v)
				} else {
					// build search term
					for i := 0; i <= c.NArg(); i++ {
						searchTerm += " " + c.Args().Get(i)
					}
					searchTerm = strings.Trim(searchTerm, " ")
					ls(searchTerm, "", c.String("tag"), c.String("language"), c.Bool("starred"), c.String("status"))
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "t, tag",
					Value: "",
					Usage: "Filter by tag (omit the # prefix)",
				},
				cli.StringFlag{
					Name:  "l, language",
					Value: "",
					Usage: "Filter by language (case-insensitive)",
				},
				cli.BoolFlag{
					Name:  "s, starred",
					Usage: "Filter by starred snippets",
				},
				cli.BoolFlag{
					Name:  "f, forked",
					Usage: "Filter by forked snippets",
				},
				cli.StringFlag{
					Name:  "status",
					Value: "all",
					Usage: "Filter by (all|public|private)",
				},
				cli.BoolFlag{
					Name:  "w, syntax",
					Usage: "Output with syntax highlighting",
				},
				cli.BoolFlag{
					Name:  "o, output",
					Usage: "Output content of each snippet",
				},
			},
		},
		{
			Name:      "search",
			Usage:     "Use fuzzy search to find Gist",
			UsageText: "\n\t\tgg search query\n",
			Category:  "Query",
			Action: func(c *cli.Context) error {
				for i := 0; i <= c.NArg(); i++ {
					searchTerm += " " + c.Args().Get(i)
				}
				searchTerm = strings.Trim(searchTerm, " ")
				fmt.Println(searchTerm)
				fuzzySearch(searchTerm)
				return nil
			},
		},
		{
			Name:      "tag",
			Usage:     "List or query tag",
			UsageText: "\n\t\tgg tag [tag name] [query]\n",
			Category:  "Query",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "query",
					Value: "",
					Usage: "Filter by tag (omit the # prefix)",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().First() == "" {
					ListTags()
				} else {
					for i := 1; i <= c.NArg(); i++ {
						searchTerm += " " + c.Args().Get(i)
					}
					searchTerm = strings.Trim(searchTerm, " ")
					ls(searchTerm, "", c.Args().Get(0), "", false, "")
				}
				return nil
			},
		},
		{
			Name:      "language",
			Usage:     "List or query language",
			UsageText: "\n\t\tgg language [tag name] [query]\n",
			Category:  "Query",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "query",
					Value: "",
					Usage: "Filter by language",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().First() == "" {
					fieldSummary("Language")
				} else {
					for i := 1; i <= c.NArg(); i++ {
						searchTerm += " " + c.Args().Get(i)
					}
					searchTerm = strings.Trim(searchTerm, " ")
					ls(searchTerm, "", "", c.Args().Get(0), false, "")
				}
				return nil
			},
		},
	}

	var a string
	if len(os.Args) > 1 {
		a = os.Args[1]
	}
	args := os.Args
	if _, err := strconv.Atoi(a); err == nil {
		args = insert(args, 1, "o")
	} else {
		args = os.Args
	}
	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}

}
