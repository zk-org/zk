package cobracli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// testApp is a small command tree standing in for zk's own, so that the
// front-end's contract with internal/cli/spec can be tested on its own.
type testApp struct {
	app *spec.App
	out *bytes.Buffer
	// ran names the command which was executed.
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
		Flags: []spec.Flag{
			{Name: "title", Target: &t.title, Help: "A title."},
		},
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

// A command's flags and arguments reach the targets its spec declared, and
// the command writes where it was told to, not to the process's own output.
func TestRunFillsInTheSpecTargets(t *testing.T) {
	app := newTestApp()

	err := app.run("list", "-f", "path", "--tag", "a,b", "-n", "3", "notes", "inbox")

	assert.Nil(t, err)
	assert.Equal(t, app.ran, "list")
	assert.Equal(t, app.out.String(), `format="path" tag=[a b] limit=3 paths=[notes inbox]`)
}

// The hook runs after the parameters are filled in and before the command, so
// that the caller can prepare whatever the command needs.
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

func TestRunReportsAFailingHook(t *testing.T) {
	app := newTestApp()

	err := Run(app.app, []string{"list"}, func(*spec.Command) error {
		return fmt.Errorf("no notebook here")
	})

	assert.Err(t, err, "no notebook here")
	assert.Equal(t, app.ran, "")
}

// `zk tag` runs `zk tag list`, with the arguments it was given.
func TestRunFallsBackOnTheDefaultChild(t *testing.T) {
	app := newTestApp()
	assert.Nil(t, app.run("tag"))
	assert.Equal(t, app.ran, "tag list")

	app = newTestApp()
	assert.Nil(t, app.run("tag", "--title", "Hello"))
	assert.Equal(t, app.ran, "tag list")
	assert.Equal(t, app.out.String(), `title="Hello"`)

	// Naming the subcommand explicitly still works.
	app = newTestApp()
	assert.Nil(t, app.run("tag", "list"))
	assert.Equal(t, app.ran, "tag list")
}

func TestRunReportsErrorsLikeZk(t *testing.T) {
	test := func(args []string, expected string) {
		app := newTestApp()
		assert.Err(t, app.run(args...), expected)
		assert.Equal(t, app.ran, "")
	}

	test([]string{"graph"}, "missing flags: --format=STRING")
	test([]string{"graph", "--format", "xml"}, `--format must be one of "json" but got "xml"`)
	test([]string{"list", "--limit", "a"}, `--limit: expected a valid 64 bit int but got "a"`)
	test([]string{"list", "--nope"}, "unknown flag --nope")
	test([]string{"nope"}, "unexpected argument nope")
	test([]string{"graph", "--format", "json", "extra"}, "unexpected argument extra")
}

// The help is zk's own, laid out in sections, and never printed on error.
func TestHelpIsRenderedInSections(t *testing.T) {
	app := newTestApp()
	root := (&builder{app: app.app, specs: map[*cobra.Command]*spec.Command{}}).build()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"list", "--help"})

	assert.Nil(t, root.Execute())

	help := out.String()
	for _, expected := range []string{
		"Usage: zk list [<path> ...]",
		"Arguments:\n  [<path> ...]    Paths to look in.",
		"Flags:\n  -h, --help ",
		"Formatting\n  -f, --format=TEMPLATE",
		"Filtering\n  -t, --tag=TAG,...",
	} {
		if !bytes.Contains(out.Bytes(), []byte(expected)) {
			t.Errorf("Expected the help to contain:\n%s\n---\nGot:\n%s", expected, help)
		}
	}
}
