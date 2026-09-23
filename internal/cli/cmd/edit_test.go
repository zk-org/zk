package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zk-org/zk/internal/adapter/fs"
	"github.com/zk-org/zk/internal/adapter/handlebars"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// Simulates the "create a note" fzf binding of `zk edit --interactive`, which
// creates a note in the directory resolved for the user's location.
func TestEditInteractiveNewNote(t *testing.T) {
	notebookDir := t.TempDir()

	for _, dir := range []string{"journal", ".zk/templates"} {
		err := os.MkdirAll(filepath.Join(notebookDir, dir), 0o700)
		assert.Equal(t, err, nil)
	}
	err := os.WriteFile(
		filepath.Join(notebookDir, ".zk/templates/default.md"),
		[]byte("Default: {{title}}\n"),
		0o600,
	)
	assert.Equal(t, err, nil)
	err = os.WriteFile(
		filepath.Join(notebookDir, ".zk/templates/journal.md"),
		[]byte("Journal: {{title}}\n"),
		0o600,
	)
	assert.Equal(t, err, nil)

	tests := []struct {
		name         string
		workingDir   string
		expectedPath string
		expectedBody string
	}{
		{
			name:         "a note created from a group dir is a note of that group",
			workingDir:   notebookDir + "/journal",
			expectedPath: "journal/test.md",
			expectedBody: "Journal: Test\n",
		},
		{
			name:         "a note created from the notebook root uses the root config",
			workingDir:   notebookDir,
			expectedPath: "test.md",
			expectedBody: "Default: Test\n",
		},
		{
			name:         "a note created from outside the notebook uses the root config",
			workingDir:   "/somewhere/else",
			expectedPath: "test.md",
			expectedBody: "Default: Test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := fs.NewFileStorage(tt.workingDir, &util.NullLogger)
			assert.Equal(t, err, nil)
			config := core.NewDefaultConfig()
			config.Note.BodyTemplatePath = opt.NewString("default.md")
			journalNote := config.Note
			journalNote.BodyTemplatePath = opt.NewString("journal.md")
			config.Groups["journal"] = core.GroupConfig{Paths: []string{"journal"}, Note: journalNote}
			notebook := core.NewNotebook(notebookDir, config, core.NotebookPorts{
				FS:                fs,
				NoteContentParser: markdown.NewParser(markdown.ParserOpts{}, &util.NullLogger),
				TemplateLoaderFactory: func(lang string) (core.TemplateLoader, error) {
					return handlebars.NewLoader(handlebars.LoaderOpts{
						LookupPaths: []string{filepath.Join(notebookDir, ".zk/templates")},
						Styler:      core.NullStyler,
					}), nil
				},
				IDGeneratorFactory: func(opts core.IDOptions) func() string {
					return func() string { return "test" }
				},
				OSEnv: func() map[string]string { return map[string]string{} },
			})

			dir := (&Edit{}).newNoteDir(notebook)
			note, err := notebook.NewNote(core.NewNoteOpts{
				Directory: opt.NewNotEmptyString(dir.Path),
				Title:     opt.NewNotEmptyString("Test"),
				DryRun:    true,
			})
			assert.Equal(t, err, nil)
			assert.Equal(t, note.Path, tt.expectedPath)
			assert.Equal(t, note.RawContent, tt.expectedBody)
		})
	}
}
