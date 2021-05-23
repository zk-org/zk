package core

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestMarkdownLinkFormatter(t *testing.T) {
	newTester := func(encodePath, dropExtension bool) func(path, title, expected string) {
		formatter, err := NewLinkFormatter(MarkdownConfig{
			LinkFormat:        "markdown",
			LinkEncodePath:    encodePath,
			LinkDropExtension: dropExtension,
		}, &NullTemplateLoader)
		assert.Nil(t, err)

		return func(path, title, expected string) {
			actual, err := formatter(path, title)
			assert.Nil(t, err)
			assert.Equal(t, actual, expected)
		}
	}

	test := newTester(false, false)
	test("path/to note.md", "", "[](path/to note.md)")
	test("", "", "[]()")
	test("path/to note.md", "An interesting subject", "[An interesting subject](path/to note.md)")
	test(`path/(no\te).md`, `An [interesting] \subject`, `[An [interesting\] \\subject](path/(no\\te\).md)`)
	test = newTester(true, false)
	test("path/to note.md", "An interesting subject", "[An interesting subject](path/to%20note.md)")
	test(`path/(no\te).md`, `An [interesting] \subject`, `[An [interesting\] \\subject](path/%28no%5Cte%29.md)`)
	test = newTester(false, true)
	test("path/to note.md", "An interesting subject", "[An interesting subject](path/to note)")
	test = newTester(true, true)
	test("path/to note.md", "An interesting subject", "[An interesting subject](path/to%20note)")
}

func TestMarkdownLinkFormatterOnlyHref(t *testing.T) {
	newTester := func(encodePath, dropExtension bool) func(path, expected string) {
		formatter, err := NewMarkdownLinkFormatter(MarkdownConfig{
			LinkFormat:        "markdown",
			LinkEncodePath:    encodePath,
			LinkDropExtension: dropExtension,
		}, true)
		assert.Nil(t, err)

		return func(path, expected string) {
			actual, err := formatter(path, "")
			assert.Nil(t, err)
			assert.Equal(t, actual, expected)
		}
	}

	test := newTester(false, false)
	test("path/to note.md", "(path/to note.md)")
	test("", "()")
	test("path/to note.md", "(path/to note.md)")
	test(`path/(no\te).md`, `(path/(no\\te\).md)`)
	test = newTester(true, false)
	test("path/to note.md", "(path/to%20note.md)")
	test(`path/(no\te).md`, `(path/%28no%5Cte%29.md)`)
	test = newTester(false, true)
	test("path/to note.md", "(path/to note)")
	test = newTester(true, true)
	test("path/to note.md", "(path/to%20note)")
}

func TestWikiLinkFormatter(t *testing.T) {
	newTester := func(encodePath, dropExtension bool) func(path, title, expected string) {
		formatter, err := NewLinkFormatter(MarkdownConfig{
			LinkFormat:        "wiki",
			LinkEncodePath:    encodePath,
			LinkDropExtension: dropExtension,
		}, &NullTemplateLoader)
		assert.Nil(t, err)

		return func(path, title, expected string) {
			actual, err := formatter(path, title)
			assert.Nil(t, err)
			assert.Equal(t, actual, expected)
		}
	}

	test := newTester(false, false)
	test("", "", "[[]]")
	test("path/to note.md", "title", "[[path/to note.md]]")
	test(`path/[no\te].md`, "title", `[[path/[no\\te].md]]`)
	test(`path/[[no\te]].md`, "title", `[[path/[[no\\te\]].md]]`)
	test = newTester(true, false)
	test("path/to note.md", "title", "[[path/to%20note.md]]")
	test(`path/[no\te].md`, "title", "[[path/%5Bno%5Cte%5D.md]]")
	test(`path/[[no\te]].md`, "title", "[[path/%5B%5Bno%5Cte%5D%5D.md]]")
	test = newTester(false, true)
	test("path/to note.md", "title", "[[path/to note]]")
	test = newTester(true, true)
	test("path/to note.md", "title", "[[path/to%20note]]")
}

func TestCustomLinkFormatter(t *testing.T) {
	newTester := func(encodePath, dropExtension bool) func(path, title string, expected customLinkRenderContext) {
		return func(path, title string, expected customLinkRenderContext) {
			loader := newTemplateLoaderMock()
			template := loader.SpyString("custom")

			formatter, err := NewLinkFormatter(MarkdownConfig{
				LinkFormat:        "custom",
				LinkEncodePath:    encodePath,
				LinkDropExtension: dropExtension,
			}, loader)
			assert.Nil(t, err)

			actual, err := formatter(path, title)
			assert.Nil(t, err)
			assert.Equal(t, actual, "custom")
			assert.Equal(t, template.Contexts, []interface{}{expected})
		}
	}

	test := newTester(false, false)
	test("path/to note.md", "", customLinkRenderContext{Path: "path/to note.md"})
	test("", "", customLinkRenderContext{})
	test("path/to note.md", "An interesting subject", customLinkRenderContext{
		Title: "An interesting subject",
		Path:  "path/to note.md",
	})
	test(`path/(no\te).md`, `An [interesting] \subject`, customLinkRenderContext{
		Title: `An [interesting] \subject`,
		Path:  `path/(no\te).md`,
	})
	test = newTester(true, false)
	test("path/to note.md", "An interesting subject", customLinkRenderContext{
		Title: "An interesting subject",
		Path:  "path/to%20note.md",
	})
	test = newTester(false, true)
	test("path/to note.md", "An interesting subject", customLinkRenderContext{
		Title: "An interesting subject",
		Path:  "path/to note",
	})
	test = newTester(true, true)
	test("path/to note.md", "An interesting subject", customLinkRenderContext{
		Title: "An interesting subject",
		Path:  "path/to%20note",
	})
}
