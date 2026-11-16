package spec

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestPlaceholderFor(t *testing.T) {
	test := func(flag Flag, expected string) {
		assert.Equal(t, PlaceholderFor(flag), expected)
	}

	var (
		flag  bool
		count int
		text  string
		list  []string
		pairs map[string]string
	)

	// A boolean flag takes no value.
	test(Flag{Name: "orphan", Target: &flag}, "")
	// A declared placeholder wins over everything else.
	test(Flag{Name: "limit", Placeholder: "COUNT", Target: &count}, "COUNT")
	test(Flag{Name: "match-strategy", Placeholder: "STRATEGY", Default: "fts", Target: &text}, "STRATEGY")
	// Otherwise a default value stands in for it, quoted.
	test(Flag{Name: "delimiter", Default: "\n", Target: &text}, `"\n"`)
	test(Flag{Name: "footer", Default: `\n`, Target: &text}, `"\\n"`)
	// Otherwise the kind of value the flag takes.
	test(Flag{Name: "header", Target: &text}, "STRING")
	test(Flag{Name: "max-distance", Target: &count}, "INT")
	// A repeatable flag says so, and is named after itself.
	test(Flag{Name: "tag", Target: &list}, "TAG,...")
	test(Flag{Name: "match", Placeholder: "QUERY", Target: &list}, "QUERY,...")
	test(Flag{Name: "extra", Target: &pairs}, "KEY=VALUE,...")
}

func TestParseFlagsSplitsOnCommas(t *testing.T) {
	var tags []string
	flags := []Flag{{Name: "tag", Short: "t", Target: &tags}}

	rest, err := ParseFlags(flags, []string{"--tag", "a,b", "-t", "c", "path"})
	assert.Nil(t, err)
	assert.Equal(t, tags, []string{"a", "b", "c"})
	assert.Equal(t, rest, []string{"path"})
}

// A value is split on commas verbatim: reading it as CSV, as pflag's own
// string slice does, would mangle the quotes and reject this outright.
func TestParseFlagsKeepsQuotes(t *testing.T) {
	var match []string
	flags := []Flag{{Name: "match", Target: &match}}

	_, err := ParseFlags(flags, []string{"--match", `["न", "म", "स्", "ते"]`})
	assert.Nil(t, err)
	assert.Equal(t, match, []string{`["न"`, ` "म"`, ` "स्"`, ` "ते"]`})
}

func TestParseFlagsAppliesDefaults(t *testing.T) {
	var strategy string
	var limit int
	flags := []Flag{
		{Name: "match-strategy", Default: "fts", Target: &strategy},
		{Name: "limit", Default: 10, Target: &limit},
	}

	_, err := ParseFlags(flags, nil)
	assert.Nil(t, err)
	assert.Equal(t, strategy, "fts")
	assert.Equal(t, limit, 10)
}

func TestParseFlagsReportsErrorsLikeZk(t *testing.T) {
	var limit int
	flags := []Flag{{Name: "limit", Short: "n", Target: &limit}}

	_, err := ParseFlags(flags, []string{"--limit", "a"})
	assert.Err(t, err, `--limit: expected a valid 64 bit int but got "a"`)

	_, err = ParseFlags(flags, []string{"--nope"})
	assert.Err(t, err, "unknown flag --nope")
}

func TestMapFlagSplitsPairs(t *testing.T) {
	var extra map[string]string
	flags := []Flag{{Name: "extra", Target: &extra}}

	_, err := ParseFlags(flags, []string{"--extra", "a=1,b=2", "--extra", "c=3"})
	assert.Nil(t, err)
	assert.Equal(t, extra, map[string]string{"a": "1", "b": "2", "c": "3"})

	_, err = ParseFlags([]Flag{{Name: "extra", Target: &extra}}, []string{"--extra", "nope"})
	assert.Err(t, err, `expected a KEY=VALUE pair but got "nope"`)
}

func TestArgLabels(t *testing.T) {
	var one string
	var many []string

	assert.Equal(t, Arg{Name: "uri", Target: &one}.Label(), "<uri>")
	assert.Equal(t, Arg{Name: "directory", Optional: true, Target: &one}.Label(), "[<directory>]")
	assert.Equal(t, Arg{Name: "path", Optional: true, Variadic: true, Target: &many}.Label(), "[<path> ...]")

	cmd := &Command{Args: []Arg{{Name: "path", Optional: true, Variadic: true, Target: &many}}}
	assert.Equal(t, cmd.UsageArgs(), "[<path> ...]")
}

func TestBindArgs(t *testing.T) {
	var directory string
	cmd := &Command{Args: []Arg{
		{Name: "directory", Optional: true, Default: "here", Target: &directory},
	}}

	assert.Nil(t, cmd.BindArgs([]string{"there"}))
	assert.Equal(t, directory, "there")

	directory = ""
	assert.Nil(t, cmd.BindArgs(nil))
	assert.Equal(t, directory, "here")

	assert.Err(t, cmd.BindArgs([]string{"a", "b"}), "unexpected argument b")
}

func TestBindArgsReportsMissing(t *testing.T) {
	var uri string
	cmd := &Command{Args: []Arg{{Name: "uri", Target: &uri}}}

	assert.Err(t, cmd.BindArgs(nil), `expected "<uri>"`)
}

func TestBindArgsVariadic(t *testing.T) {
	var paths []string
	cmd := &Command{Args: []Arg{
		{Name: "path", Optional: true, Variadic: true, Target: &paths},
	}}

	assert.Nil(t, cmd.BindArgs([]string{"a", "b"}))
	assert.Equal(t, paths, []string{"a", "b"})
}

func TestValidateFlags(t *testing.T) {
	var format string
	flags := []Flag{
		{Name: "format", Short: "f", Required: true, OneOf: []string{"json"}, Target: &format},
	}

	never := func(string) bool { return false }
	always := func(string) bool { return true }

	assert.Err(t, ValidateFlags(flags, never), "missing flags: --format=STRING")

	format = "json"
	assert.Nil(t, ValidateFlags(flags, always))

	format = "xml"
	assert.Err(t, ValidateFlags(flags, always), `--format must be one of "json" but got "xml"`)
}
