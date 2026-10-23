package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	dateutil "github.com/zk-org/zk/internal/util/date"
	"github.com/zk-org/zk/internal/util/ptr"
)

// New adds a new note to the notebook.
type New struct {
	Directory   string
	Interactive bool
	Title       string
	Date        string
	Group       string
	Extra       map[string]string
	Template    string
	PrintPath   bool
	DryRun      bool
	ID          string
}

// NewSpec describes the `zk new` command.
func NewSpec(container *cli.Container) *spec.Command {
	cmd := New{}

	return &spec.Command{
		Name:  "new",
		Short: "Create a new note in the given notebook directory.",
		Group: GroupNotes,
		Args: []spec.Arg{
			{
				Name:          "directory",
				Help:          "Directory in which to create the note.",
				Optional:      true,
				Default:       ".",
				Target:        &cmd.Directory,
				CompleteFiles: true,
			},
		},
		Flags: []spec.Flag{
			{Name: "interactive", Short: "i", Target: &cmd.Interactive,
				Help: "Read contents from standard input."},
			{Name: "title", Short: "t", Placeholder: "TITLE", Target: &cmd.Title,
				Help: "Title of the new note."},
			{Name: "date", Placeholder: "DATE", Target: &cmd.Date,
				Help: "Set the current date."},
			{Name: "group", Short: "g", Placeholder: "NAME", Target: &cmd.Group,
				Help:     "Name of the config group this note belongs to. Takes precedence over the config of the directory.",
				Complete: completeGroups(container)},
			{Name: "extra", Target: &cmd.Extra,
				Help: "Extra variables passed to the templates."},
			{Name: "template", Placeholder: "PATH", Target: &cmd.Template,
				Help: "Custom template used to render the note.", CompleteFiles: true},
			{Name: "print-path", Short: "p", Target: &cmd.PrintPath,
				Help: "Print the path of the created note instead of editing it."},
			{Name: "dry-run", Short: "n", Target: &cmd.DryRun,
				Help: "Don't actually create the note. Instead, prints its content on stdout and the generated path on stderr."},
			{Name: "id", Placeholder: "ID", Target: &cmd.ID,
				Help: "Skip id generation and use provided value."},
		},
		Run: func() error { return cmd.Run(container) },
	}
}

func (cmd *New) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	var content []byte
	if cmd.Interactive {
		content, err = io.ReadAll(container.Stdin)
		if err != nil {
			return err
		}
	}

	date := time.Now()
	if cmd.Date != "" {
		date, err = dateutil.TimeFromNatural(cmd.Date)
		if err != nil {
			return err
		}
	}

	note, err := notebook.NewNote(core.NewNoteOpts{
		Title:     ptr.NotEmptyString(cmd.Title),
		Content:   string(content),
		Directory: ptr.NotEmptyString(cmd.Directory),
		Group:     ptr.NotEmptyString(cmd.Group),
		Template:  ptr.NotEmptyString(cmd.Template),
		Extra:     cmd.Extra,
		Date:      date,
		DryRun:    cmd.DryRun,
		ID:        cmd.ID,
	})

	if cmd.DryRun {
		if err != nil {
			return err
		}
		path := filepath.Join(notebook.Path, note.Path)
		fmt.Fprintln(container.Stderr, path)
		fmt.Fprint(container.Stdout, note.RawContent)
		return nil
	}

	var path string
	if err == nil {
		path = filepath.Join(notebook.Path, note.Path)
	} else {
		var noteExists core.ErrNoteExists
		if !errors.As(err, &noteExists) {
			return err
		}

		if confirmed, _ := container.Terminal.Confirm(
			fmt.Sprintf("%s already exists, do you want to edit this note instead?", noteExists.Name),
			true,
		); !confirmed {
			// abort...
			return nil
		}

		path = noteExists.Path
	}

	if cmd.PrintPath {
		fmt.Fprintf(container.Stdout, "%+v\n", path)
		return nil
	} else {
		editor, err := container.NewNoteEditor(notebook)
		if err != nil {
			return err
		}
		if err := editor.Open(path); err != nil {
			return err
		}

		// When the notebook is not scanned before each command, the index is
		// not refreshed with the note that was just written.
		if !notebook.Config.Notebook.Scan {
			if err := notebook.RefreshNote(path); err != nil {
				return err
			}
		}

		return nil
	}
}
