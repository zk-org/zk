package kongcli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// The same command tree the Cobra front-end is tested with, to show that a
// spec is read the same way whichever library presents it.
type testApp struct {
	app *spec.App
	out *bytes.Buffer
	ran string

	format string
	tag    []string
	limit  int
	paths  []string
	title  string
}

func newTestApp() *testApp {
	t := &testApp{out: &bytes.Buffer{}}

	list := &spec.Command{
		Name:  "list",
		Short: "List notes matching the given criteria.",
		Group: "notes",
		Args: []spec.Arg{
			{Name: "path", Help: "Paths to look in.", Optional: true, Variadic: true, Target: &t.paths},
		},
		Flags: []spec.Flag{
			{Name: "format", Short: "f", Group: spec.GroupFormat, Placeholder: "TEMPLATE",
				Help: "How to print each note.", Target: &t.format},
			{Name: "tag", Short: "t", Group: spec.GroupFilter,
				Help: "Find notes with the given tags.", Target: &t.tag},
			{Name: "limit", Short: "n", Group: spec.GroupFilter, Placeholder: "COUNT",
				Help: "Limit the number of notes found.", Target: &t.limit},
		},
	}
	list.Run = func() error {
		t.ran = "list"
		fmt.Fprintf(t.out, "format=%q tag=%v limit=%d paths=%v", t.format, t.tag, t.limit, t.paths)
		return nil
	}

	graph := &spec.Command{
		Name:  "graph",
		Short: "Produce a graph of the notes.",
		Group: "notes",
		Flags: []spec.Flag{
			{Name: "format", Short: "f", Group: spec.GroupFormat, Required: true,
				OneOf: []string{"json"}, Help: "Format of the graph.", Target: &t.format},
		},
		Run: func() error { t.ran = "graph"; return nil },
	}

	tagList := &spec.Command{
		Name:  "list",
		Short: "List all the note tags.",
		Flags: []spec.Flag{{Name: "title", Target: &t.title, Help: "A title."}},
	}
	tagList.Run = func() error {
		t.ran = "tag list"
		fmt.Fprintf(t.out, "title=%q", t.title)
		return nil
	}

	t.app = &spec.App{
		Name:    "zk",
		Version: "v1.2.3",
		Commands: []*spec.Command{
			list,
			graph,
			{
				Name:         "tag",
				Short:        "Manage the note tags.",
				Group:        "notes",
				Children:     []*spec.Command{tagList},
				DefaultChild: "list",
			},
		},
		CommandGroups: []spec.Group{{ID: "notes", Title: "NOTES", Description: "Edit or browse your notes"}},
		FlagGroups: []spec.Group{
			{ID: spec.GroupFormat, Title: "Formatting"},
			{ID: spec.GroupFilter, Title: "Filtering"},
		},
	}
	return t
}

func (t *testApp) run(args ...string) error {
	return Run(t.app, args, nil)
}

func TestRunFillsInTheSpecTargets(t *testing.T) {
	app := newTestApp()

	err := app.run("list", "-f", "path", "--tag", "a,b", "-n", "3", "notes", "inbox")

	assert.Nil(t, err)
	assert.Equal(t, app.ran, "list")
	assert.Equal(t, app.out.String(), `format="path" tag=[a b] limit=3 paths=[notes inbox]`)
}

func TestRunCallsTheHookBeforeTheCommand(t *testing.T) {
	app := newTestApp()
	hooked := ""

	err := Run(app.app, []string{"list", "--tag", "x"}, func(cmd *spec.Command) error {
		hooked = cmd.Name
		assert.Equal(t, app.tag, []string{"x"})
		assert.Equal(t, app.ran, "")
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, hooked, "list")
	assert.Equal(t, app.ran, "list")
}

func TestRunFallsBackOnTheDefaultChild(t *testing.T) {
	app := newTestApp()
	assert.Nil(t, app.run("tag"))
	assert.Equal(t, app.ran, "tag list")

	app = newTestApp()
	assert.Nil(t, app.run("tag", "--title", "Hello"))
	assert.Equal(t, app.ran, "tag list")
	assert.Equal(t, app.out.String(), `title="Hello"`)
}

// Kong reports these in its own words, which are the ones zk has always
// printed.
func TestRunReportsErrors(t *testing.T) {
	test := func(args []string, expected string) {
		app := newTestApp()
		assert.Err(t, app.run(args...), expected)
		assert.Equal(t, app.ran, "")
	}

	test([]string{"graph"}, "missing flags: --format=STRING")
	test([]string{"graph", "--format", "xml"}, `--format must be one of "json" but got "xml"`)
	test([]string{"list", "--nope"}, "unknown flag --nope")
	test([]string{"nope"}, "unexpected argument nope")
}
