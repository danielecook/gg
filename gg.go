package main

import (
	"fmt"
	"log"
	"os"

	. "github.com/logrusorgru/aurora"
	"github.com/pkg/browser"
	"github.com/urfave/cli"
)

var errlog = log.New(os.Stderr, "", 0)

func main() {

	var token string
	var rebuild bool
	fmt.Println(token)

	//client := github.NewClient(nil)
	app := cli.NewApp()

	app.Name = "gg"
	app.Usage = "A tool for Github Gists"
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
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				browser.OpenURL("https://gist.github.com")
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
					Name:        "authentication_token",
					EnvVar:      "TOKEN",
					Destination: &token,
				},
				cli.BoolFlag{
					Name:        "r, rebuild",
					Usage:       "Rebuild library",
					Destination: &rebuild,
				},
			},
			Action: func(c *cli.Context) error {
				if token == "" {
					/* gg login */
					errMsg := Bold(Red("\n\tError - Please supply your Authentication Token\n"))
					return cli.NewExitError(errMsg, 2)
				}
				if len(token) != 40 {
					/* gg login <wrong_length> */
					return cli.NewExitError(Bold(Red("\n\tThe API Key should be 40 characters\n\n")), 32)
				}
				/* Store Token */
				initializeLibrary(token, rebuild)
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
			Name:                   "ls",
			Usage:                  "List and filter",
			UsageText:              "\n\t\tsq ls [options] [query]\n\n\t\tquery - Searches most fields",
			Category:               "Snippets",
			UseShortOptionHandling: true,
			Action: func(c *cli.Context) error {
				ls()
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
			Name:      "tags",
			Usage:     "List library tags",
			UsageText: "\n\t\tgg tags\n",
			Category:  "Query",
			Action: func(c *cli.Context) error {
				listTags()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
