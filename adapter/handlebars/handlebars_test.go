package handlebars

import (
	"testing"
	"time"

	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/fixtures"
)

func init() {
	Init("en", &util.NullLogger)
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

func TestPrependHelper(t *testing.T) {
	// inline
	testString(t, "{{prepend '> ' 'A quote'}}", nil, "> A quote")

	// block
	testString(t, "{{#prepend '> '}}A quote{{/prepend}}", nil, "> A quote")
	testString(t, "{{#prepend '> '}}A quote on\nseveral lines{{/prepend}}", nil, "> A quote on\n> several lines")
}

func TestSlugHelper(t *testing.T) {
	// inline
	testString(t,
		`{{slug "This will be slugified!"}}`,
		nil,
		"this-will-be-slugified",
	)
	// block
	testString(t,
		"{{#slug}}This will be slugified!{{/slug}}",
		nil,
		"this-will-be-slugified",
	)
}

func TestDateHelper(t *testing.T) {
	context := map[string]interface{}{"now": time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}
	testString(t, "{{date now}}", context, "2009-11-17")
	testString(t, "{{date now 'short'}}", context, "11/17/2009")
	testString(t, "{{date now 'medium'}}", context, "Nov 17, 2009")
	testString(t, "{{date now 'long'}}", context, "November 17, 2009")
	testString(t, "{{date now 'full'}}", context, "Tuesday, November 17, 2009")
	testString(t, "{{date now 'year'}}", context, "2009")
	testString(t, "{{date now 'time'}}", context, "20:34")
	testString(t, "{{date now 'timestamp'}}", context, "200911172034")
	testString(t, "{{date now 'timestamp-unix'}}", context, "1258490098")
	testString(t, "{{date now 'cust: %Y-%m'}}", context, "cust: 2009-11")
	testString(t, "{{date now 'elapsed'}}", context, "12 years ago")
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
