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
			Usage:     "Store your Authentication Token",
			UsageText: "\n\t\tgg login [Authentication Token KEY]\n",
			Category:  "Account",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "authentication_token",
					EnvVar:      "TOKEN",
					Destination: &token,
				},
			},
			Action: func(c *cli.Context) error {
				if token == "" {
					/* sq login */
					errMsg := Bold(Red("\n\tError - Please supply your Authentication Token\n"))
					return cli.NewExitError(errMsg, 2)
				}
				if len(token) != 40 {
					/* sq login <wrong_length> */
					return cli.NewExitError(Bold(Red("\n\tThe API Key should be 40 characters\n\n")), 32)
				}
				/* Store Token */
				initializeLibrary(token)
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
				// Running api_auth_user will update the library if possible
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
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
