package core

import (
	"path/filepath"
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
			actual, err := formatter(LinkFormatterContext{
				Filename: "filename",
				Path:     "path",
				RelPath:  path,
				AbsPath:  "abs-path",
				Title:    title,
			})
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
			actual, err := formatter(LinkFormatterContext{
				Filename: "filename",
				Path:     "path",
				RelPath:  path,
				AbsPath:  "abs-path",
				Title:    "title",
			})
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
			actual, err := formatter(LinkFormatterContext{
				Filename: "filename",
				Path:     path,
				RelPath:  "rel-path",
				AbsPath:  "abs-path",
				Title:    "title",
			})
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
	newTester := func(encodePath, dropExtension bool) func(path, title string, expected LinkFormatterContext) {
		return func(path, title string, expected LinkFormatterContext) {
			loader := newTemplateLoaderMock()
			template := loader.SpyString("custom")

			formatter, err := NewLinkFormatter(MarkdownConfig{
				LinkFormat:        "custom",
				LinkEncodePath:    encodePath,
				LinkDropExtension: dropExtension,
			}, loader)
			assert.Nil(t, err)

			actual, err := formatter(LinkFormatterContext{
				Filename: filepath.Base(path),
				Path:     path,
				AbsPath:  "/" + path,
				RelPath:  "../" + path,
				Title:    title,
			})
			assert.Nil(t, err)
			assert.Equal(t, actual, "custom")
			assert.Equal(t, template.Contexts, []interface{}{expected})
		}
	}

	test := newTester(false, false)
	test("path/to note.md", "", LinkFormatterContext{
		Filename: "to note.md",
		Path:     "path/to note.md",
		AbsPath:  "/path/to note.md",
		RelPath:  "../path/to note.md",
	})
	test("", "", LinkFormatterContext{
		Filename: ".",
		Path:     "",
		AbsPath:  "/",
		RelPath:  "../",
	})
	test("path/to note.md", "An interesting subject", LinkFormatterContext{
		Filename: "to note.md",
		Path:     "path/to note.md",
		AbsPath:  "/path/to note.md",
		RelPath:  "../path/to note.md",
		Title:    "An interesting subject",
	})
	test(`path/(no\te).md`, `An [interesting] \subject`, LinkFormatterContext{
		Filename: `(no\te).md`,
		Path:     `path/(no\te).md`,
		AbsPath:  `/path/(no\te).md`,
		RelPath:  `../path/(no\te).md`,
		Title:    `An [interesting] \subject`,
	})
	test = newTester(true, false)
	test("path/to note.md", "An interesting subject", LinkFormatterContext{
		Filename: "to%20note.md",
		Path:     "path/to%20note.md",
		AbsPath:  "/path/to%20note.md",
		RelPath:  "../path/to%20note.md",
		Title:    "An interesting subject",
	})
	test = newTester(false, true)
	test("path/to note.md", "An interesting subject", LinkFormatterContext{
		Filename: "to note",
		Path:     "path/to note",
		AbsPath:  "/path/to note",
		RelPath:  "../path/to note",
		Title:    "An interesting subject",
	})
	test = newTester(true, true)
	test("path/to note.md", "An interesting subject", LinkFormatterContext{
		Filename: "to%20note",
		Path:     "path/to%20note",
		AbsPath:  "/path/to%20note",
		RelPath:  "../path/to%20note",
		Title:    "An interesting subject",
	})
}
