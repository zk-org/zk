package handlebars

import (
	"testing"

	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/fixtures"
)

func init() {
	Init()
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

func TestSlug(t *testing.T) {
	sut := NewRenderer()
	res, err := sut.Render("{{#slug}}This will be slugified!{{/slug}}", nil)
	assert.Nil(t, err)
	assert.Equal(t, res, "this-will-be-slugified")

	res, err = sut.Render(`{{slug "This will be slugified!"}}`, nil)
	assert.Nil(t, err)
	assert.Equal(t, res, "this-will-be-slugified")
}
