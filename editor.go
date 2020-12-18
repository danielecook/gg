package main

import (
	"github.com/AlecAivazis/survey/v2"
)

var editorSurvey = []*survey.Question{
	{
		Name: "editor",
		Prompt: &survey.Select{
			Message: "Choose an editor:",
			Options: []string{"subl", "nano", "vim", "micro"},
			Default: "subl",
		},
	},
}
