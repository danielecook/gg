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
	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
)

var errlog = log.New(os.Stderr, "", 0)

var greenText = color.New(color.FgGreen).Add(color.Bold)
var blueText = color.New(color.FgBlue).Add(color.Bold)

func successMsg(s string) {
	c := greenText.FprintFunc()
	c(os.Stderr, s)
}

func errorMsg(s string) {
	c := color.New(color.FgRed).Add(color.Bold).FprintFunc()
	c(os.Stderr, s)
}

func boldMsg(s string) {
	c := color.New(color.Bold).FprintFunc()
	c(os.Stderr, s)
}

// Flags

func main() {

	limitFlag := cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Value:   10,
		Usage:   "Max number of results to display",
	}

	var searchTerm string

	app := cli.NewApp()

	app.Name = "gg"
	app.Usage = "A tool for Github Gists\n\n\t gg <ID> - retrieve gist"
	app.Version = "0.0.1"
	app.EnableBashCompletion = true

	app.Authors = []*cli.Author{
		{
			Name:  "Daniel Cook",
			Email: "danielecook@gmail.com",
		},
	}

	app.Commands = []*cli.Command{
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
					content, err := clipboard.ReadAll()
					if err != nil {
						ThrowError("Error reading from clipboard", 1)
					}
					fileSet[c.String("filename")] = content
				} else if inputPipe() {
					/* New from stdin */
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
						for _, fname := range c.Args().Slice() {
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
				&cli.StringFlag{
					Name:  "d, description",
					Usage: "Set the description for gist",
				},
				&cli.StringFlag{
					Name:  "f, filename",
					Usage: "Set the filename with --clipboard or stdin",
				},
				&cli.BoolFlag{
					Name: "p, private",
				},
				&cli.BoolFlag{
					Name: "c, clipboard",
				},
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
			Name:      "sync",
			Usage:     "Login and fetch your gist library",
			UsageText: "\n\t\tgg sync [Authentication Token]\n",
			Category:  "Library",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "token",
					Usage:   "Required on first login; This is stored in .gg/config.json",
					EnvVars: []string{"TOKEN"},
				},
				&cli.BoolFlag{
					Name:  "r, rebuild",
					Usage: "Clear and rebuild library",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("token") != "" || c.Bool("rebuild") {
					/* gg login */
					initializeLibrary(c.String("token"), c.Bool("rebuild"))
				}
				updateLibrary()
				return nil
			},
		},
		{
			Name:                   "open",
			Aliases:                []string{"o"},
			Usage:                  "Copy or output a single gist",
			UsageText:              "\n\t\tgg o [options] [gists...]\n\n\t\t",
			Category:               "Query",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				if v, err := strconv.Atoi(c.Args().Get(0)); err == nil {
					if c.Bool("clipboard") {
						clipboard.WriteAll(fetchGistContent(v))
						successMsg("Copied to clipboard")
					} else {
						for g := range c.Args().Slice() {
							if v, err := strconv.Atoi(c.Args().Get(g)); err == nil {
								outputGist(v)
							} else {
								errorMsg(fmt.Sprintf("%v is an invalid ID", c.Args().Get(g)))
							}
						}
					}
				}
				return nil
			},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "c, clipboard",
					Usage: "Copy to clipboard. Only works for first gist.",
				},
			},
		},
		{
			Name:                   "rm",
			Usage:                  "Delete gists",
			UsageText:              "\n\t\tgg rm [options] [gists...]\n\n\t\t",
			Category:               "Query",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				for g := range c.Args().Slice() {
					if v, err := strconv.Atoi(c.Args().Get(g)); err == nil {
						rmGist(v)
					} else {
						errorMsg(fmt.Sprintf("%v is an invalid ID", c.Args().Get(g)))
					}
				}
				return nil
			},
		},
		{
			Name:                   "web",
			Aliases:                []string{"w"},
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
					ls(searchTerm, "", c.String("tag"), c.String("language"), c.Bool("starred"), c.String("status"), c.Int("limit"))
				}
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "t, tag",
					Value: "",
					Usage: "Filter by tag (omit the # prefix)",
				},
				&cli.StringFlag{
					Name:  "language",
					Value: "",
					Usage: "Filter by language (case-insensitive)",
				},
				&cli.BoolFlag{
					Name:  "s, starred",
					Usage: "Filter by starred snippets",
				},
				&cli.BoolFlag{
					Name:  "f, forked",
					Usage: "Filter by forked snippets",
				},
				&cli.StringFlag{
					Name:  "status",
					Value: "all",
					Usage: "Filter by (all|public|private)",
				},
				&cli.BoolFlag{
					Name:  "o, output",
					Usage: "Output content of each snippet",
				},
				&limitFlag,
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
				fuzzySearch(searchTerm)
				return nil
			},
		},
		{
			Name:      "tag",
			Aliases:   []string{"tags"},
			Usage:     "List or query tag",
			UsageText: "\n\t\tgg tag [tag name] [query]\n",
			Category:  "Query",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "query",
					Value: "",
					Usage: "Filter by tag (omit the # prefix)",
				},
				&limitFlag,
				&cli.StringFlag{
					Name:  "status",
					Value: "all",
					Usage: "Filter by (all|public|private)",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().First() == "" {
					fieldSummary("Tags")
				} else {
					for i := 1; i <= c.NArg(); i++ {
						searchTerm += " " + c.Args().Get(i)
					}
					searchTerm = strings.Trim(searchTerm, " ")
					ls(searchTerm, "", c.Args().Get(0), "", false, c.String("status"), c.Int("limit"))
				}
				return nil
			},
		},
		{
			Name:      "language",
			Aliases:   []string{"languages"},
			Usage:     "List or query language",
			UsageText: "\n\t\tgg language [language name] [query]\n",
			Category:  "Query",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "query",
					Value: "",
					Usage: "Filter by language",
				},
				&limitFlag,
				&cli.StringFlag{
					Name:  "status",
					Value: "all",
					Usage: "Filter by (all|public|private)",
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
					ls(searchTerm, "", "", c.Args().Get(0), false, c.String("status"), c.Int("limit"))
				}
				return nil
			},
		},
		{
			Name:      "owner",
			Usage:     "List or query owner",
			UsageText: "\n\t\tgg language [owner] [query]\n",
			Category:  "Query",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "query",
					Value: "",
					Usage: "Filter by owner",
				},
				&limitFlag,
				&cli.StringFlag{
					Name:  "status",
					Value: "all",
					Usage: "Filter by (all|public|private)",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().First() == "" {
					fieldSummary("Owner")
				} else {
					for i := 1; i <= c.NArg(); i++ {
						searchTerm += " " + c.Args().Get(i)
					}
					searchTerm = strings.Trim(searchTerm, " ")
					ls(searchTerm, "", "", c.Args().Get(0), false, c.String("status"), c.Int("limit"))
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
