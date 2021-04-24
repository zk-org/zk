package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/strings"
)

// Init creates a notebook in the given directory
type Init struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the notebook."`
}

func (cmd *Init) Run(container *cli.Container) error {
	opts, err := startInitWizard()
	if err != nil {
		if err == terminal.InterruptErr {
			return nil
		}
		return err
	}

	fmt.Println()

	notebook, err := container.Notebooks.Init(cmd.Directory, opts)
	if err != nil {
		return err
	}

	force := false
	_, err = notebook.Index(force)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(cmd.Directory)
	if err != nil {
		path = cmd.Directory
	}

	fmt.Printf("Initialized a notebook in %v\n", path)
	return nil
}

func startInitWizard() (core.InitOpts, error) {
	answers := struct {
		WikiLink bool
		Tags     []string
	}{}

	hashtag := "#hashtag"
	multiwordTag := "#Bear's multi-word tag#"
	colonTag := ":colon:tag:"

	questions := []*survey.Question{
		{
			Name: "wikilink",
			Prompt: &survey.Confirm{
				Message: "Do you prefer [[WikiLinks]] over regular Markdown links?",
				Default: false,
			},
		},
		{
			Name: "tags",
			Prompt: &survey.MultiSelect{
				Message: "Choose your favorite inline tag syntaxes:",
				Options: []string{hashtag, multiwordTag, colonTag},
			},
		},
	}

	var opts core.InitOpts
	err := survey.Ask(questions, &answers)
	if err != nil {
		return opts, err
	}

	opts.WikiLinks = answers.WikiLink

	opts.Hashtags = strings.InList(answers.Tags, hashtag)
	opts.MultiwordTags = strings.InList(answers.Tags, multiwordTag)
	opts.ColonTags = strings.InList(answers.Tags, colonTag)

	return opts, nil
}
