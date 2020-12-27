package handlebars

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/date"
	"github.com/mickael-menu/zk/util/fixtures"
)

func init() {
	logger := log.New(os.Stderr, "zk: warning: ", 0)
	date := date.NewFrozen(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	Init("en", logger, &date)
}

func TestRenderString(t *testing.T) {
	sut := NewRenderer()
	res, err := sut.Render("Goodbye, {{name}}", map[string]string{"name": "Ed"})
	assert.Nil(t, err)
	assert.Equal(t, res, "Goodbye, Ed")
}

func TestRenderFile(t *testing.T) {
	sut := NewRenderer()
	res, err := sut.RenderFile(fixtures.Path("template.txt"), map[string]string{"name": "Thom"})
	assert.Nil(t, err)
	assert.Equal(t, res, "Hello, Thom\n")
}

func TestUnknownVariable(t *testing.T) {
	sut := NewRenderer()
	res, err := sut.Render("Hi, {{unknown}}!", nil)
	assert.Nil(t, err)
	assert.Equal(t, res, "Hi, !")
}

func TestDoesntEscapeHTML(t *testing.T) {
	sut := NewRenderer()

	res, err := sut.Render("Salut, {{name}}!", map[string]string{"name": "l'ami"})
	assert.Nil(t, err)
	assert.Equal(t, res, "Salut, l'ami!")

	res, err = sut.RenderFile(fixtures.Path("unescape.txt"), map[string]string{"name": "l'ami"})
	assert.Nil(t, err)
	assert.Equal(t, res, "Salut, l'ami!\n")
}

func TestSlugHelper(t *testing.T) {
	sut := NewRenderer()
	// block
	res, err := sut.Render("{{#slug}}This will be slugified!{{/slug}}", nil)
	assert.Nil(t, err)
	assert.Equal(t, res, "this-will-be-slugified")
	// inline
	res, err = sut.Render(`{{slug "This will be slugified!"}}`, nil)
	assert.Nil(t, err)
	assert.Equal(t, res, "this-will-be-slugified")
}

func TestDateHelper(t *testing.T) {
	sut := NewRenderer()

	test := func(format string, expected string) {
		res, err := sut.Render(fmt.Sprintf("{{date '%s'}}", format), nil)
		assert.Nil(t, err)
		assert.Equal(t, res, expected)
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
