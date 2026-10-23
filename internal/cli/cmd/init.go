package cmd

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
)

// Init creates a notebook in the given directory
type Init struct {
	Directory string
}

// InitSpec describes the `zk init` command.
func InitSpec(container *cli.Container) *spec.Command {
	cmd := Init{}

	return &spec.Command{
		Name:  "init",
		Short: "Create a new notebook in the given directory.",
		Group: GroupNotebook,
		Args: []spec.Arg{
			{
				Name:          "directory",
				Help:          "Directory containing the notebook.",
				Optional:      true,
				Default:       ".",
				Path:          true,
				Target:        &cmd.Directory,
				CompleteFiles: true,
			},
		},
		Run: func() error { return cmd.Run(container) },
	}
}

func (cmd *Init) Run(container *cli.Container) error {
	opts, err := newInitOpts(container)
	if err != nil {
		if err == terminal.InterruptErr {
			return nil
		}
		return err
	}

	fmt.Fprintln(container.Stdout)

	notebook, err := container.Notebooks.Init(cmd.Directory, opts)
	if err != nil {
		return err
	}

	index := Index{Quiet: true}
	err = index.RunWithNotebook(container, notebook)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(cmd.Directory)
	if err != nil {
		path = cmd.Directory
	}

	fmt.Fprintf(container.Stdout, "Initialized a notebook in %v\n", path)
	return nil
}

func newInitOpts(container *cli.Container) (core.InitOpts, error) {
	if container.Terminal.NoInput {
		return core.NewDefaultInitOpts(), nil
	} else {
		return startInitWizard()
	}
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

	opts.Hashtags = slices.Contains(answers.Tags, hashtag)
	opts.MultiwordTags = slices.Contains(answers.Tags, multiwordTag)
	opts.ColonTags = slices.Contains(answers.Tags, colonTag)

	return opts, nil
}
