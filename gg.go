package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	. "github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

func main() {

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
				fmt.Printf("Great\n")
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
					Name: "authentication_token",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					/* sq login */
					err_msg := Bold(Red("\n\tError - Please supply your Authentication Token\n"))
					return cli.NewExitError(err_msg, 2)
				} else if c.NArg() >= 1 {
					var apiKey = c.Args().Get(0)
					if len(apiKey) != 40 {
						/* sq login <wrong_length> */
						return cli.NewExitError(Bold(Red("\n\tThe API Key should be 40 characters\n\n")), 32)
					}
					/* Store API Key */
					s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
					s.Writer = os.Stderr
					s.Start() // Start the spinner
					setupLibrary(apiKey)
					authenticate()
					s.Stop()

				} else {
					ThrowError("UnknownError")
				}
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
				fmt.Printf("GGGGG")
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
