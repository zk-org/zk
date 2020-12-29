package handlebars

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/date"
	"github.com/mickael-menu/zk/util/fixtures"
)

func init() {
	date := date.NewFrozen(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	Init("en", &util.NullLogger, &date)
}

func testString(t *testing.T, template string, context interface{}, expected string) {
	sut := NewLoader()

	templ, err := sut.Load(template)
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func testFile(t *testing.T, name string, context interface{}, expected string) {
	sut := NewLoader()

	templ, err := sut.LoadFile(fixtures.Path(name))
	assert.Nil(t, err)

	actual, err := templ.Render(context)
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestRenderString(t *testing.T) {
	testString(t,
		"Goodbye, {{name}}",
		map[string]string{"name": "Ed"},
		"Goodbye, Ed",
	)
}

func TestRenderFile(t *testing.T) {
	testFile(t,
		"template.txt",
		map[string]string{"name": "Thom"},
		"Hello, Thom\n",
	)
}

func TestUnknownVariable(t *testing.T) {
	testString(t,
		"Hi, {{unknown}}!",
		nil,
		"Hi, !",
	)
}

func TestDoesntEscapeHTML(t *testing.T) {
	testString(t,
		"Salut, {{name}}!",
		map[string]string{"name": "l'ami"},
		"Salut, l'ami!",
	)

	testFile(t,
		"unescape.txt",
		map[string]string{"name": "l'ami"},
		"Salut, l'ami!\n",
	)
}

func TestSlugHelper(t *testing.T) {
	// block
	testString(t,
		"{{#slug}}This will be slugified!{{/slug}}",
		nil,
		"this-will-be-slugified",
	)
	// inline
	testString(t,
		`{{slug "This will be slugified!"}}`,
		nil,
		"this-will-be-slugified",
	)
}

func TestDateHelper(t *testing.T) {
	// Default
	testString(t, "{{date}}", nil, "2009-11-17")

	test := func(format string, expected string) {
		testString(t, fmt.Sprintf("{{date '%s'}}", format), nil, expected)
	}

	test("short", "11/17/2009")
	test("medium", "Nov 17, 2009")
	test("long", "November 17, 2009")
	test("full", "Tuesday, November 17, 2009")
	test("year", "2009")
	test("time", "20:34")
	test("timestamp", "200911172034")
	test("timestamp-unix", "1258490098")
	test("cust: %Y-%m", "cust: 2009-11")
}

func TestShellHelper(t *testing.T) {
	// block is passed as piped input
	testString(t,
		`{{#sh "tr '[a-z]' '[A-Z]'"}}Hello, world!{{/sh}}`,
		nil,
		"HELLO, WORLD!",
	)
	// inline
	testString(t,
		`{{sh "echo 'Hello, world!'"}}`,
		nil,
		"Hello, world!\n",
	)
}
