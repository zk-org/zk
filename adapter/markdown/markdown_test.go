package markdown

import (
	"testing"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestParseTitle(t *testing.T) {
	test := func(source string, expectedTitle string) {
		content := parse(t, source)
		assert.Equal(t, content.Title, opt.NewNotEmptyString(expectedTitle))
	}

	test("", "")
	test("#  ", "")
	test("#A title", "")
	test("   # A title", "A title")
	test("# A title", "A title")
	test("#   A  title   ", "A  title")
	test("## A title", "A title")
	test("Paragraph \n\n## A title\nBody", "A title")
	test("# Heading 1\n## Heading 1.a\n# Heading 2", "Heading 1")
	test("## Small Heading\n# Bigger Heading", "Bigger Heading")
	test("# A **title** with [formatting](http://stripped)", "A title with formatting")

	// From a YAML frontmatter
	test(`---
Title:     A title
Tags:
    - tag1
    - tag2
---

# Heading
`, "A title")
	test(`---
title: lowercase key
---
Paragraph
`, "lowercase key")
}

func TestParseBody(t *testing.T) {
	test := func(source string, expectedBody string) {
		content := parse(t, source)
		assert.Equal(t, content.Body, opt.NewNotEmptyString(expectedBody))
	}

	test("", "")
	test("# A title\n    \n", "")
	test("Paragraph \n\n# A title", "")
	test("Paragraph \n\n# A title\nBody", "Body")

	test(
		`## Small Heading
# Bigger Heading
     
## Smaller Heading
Body
several lines
# Body heading

Paragraph:

* item1
* item2
    
`,
		`## Smaller Heading
Body
several lines
# Body heading

Paragraph:

* item1
* item2`,
	)
	test(`---
title: A title
---

Paragraph
`, "Paragraph")
}

func TestParseLead(t *testing.T) {
	test := func(source string, expectedLead string) {
		content := parse(t, source)
		assert.Equal(t, content.Lead, opt.NewNotEmptyString(expectedLead))
	}

	test("", "")
	test("# A title\n    \n", "")

	test(
		`Paragraph
# A title`,
		"",
	)

	test(
		`Paragraph 
# A title
Lead`,
		"Lead",
	)

	test(
		`# A title
Lead
multiline

other`,
		"Lead\nmultiline",
	)

	test(
		`# A title

Lead
multiline

## Heading`,
		"Lead\nmultiline",
	)

	test(
		`# A title

## Heading
Lead
multiline

other`,
		`## Heading
Lead
multiline`,
	)

	test(
		`# A title

* item1
* item2

Paragraph`,
		`* item1
* item2`,
	)
}

func TestParseLinks(t *testing.T) {
	test := func(source string, links []note.Link) {
		content := parse(t, source)
		assert.Equal(t, content.Links, links)
	}

	test("", []note.Link{})
	test("No links around here", []note.Link{})

	test(`
# Heading with a [link](heading)

Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").

An https://inline-link.com and http://another-inline-link.com.

A [[Wiki link]] is surrounded by [[2-brackets | two brackets]].

It can contain [[esca]\]ped \[chara\\cters]].

A [[[Folgezettel link]]] is surrounded by three brackets.

Neuron also supports a [[trailing hash]]# for Folgezettel links.

A #[[leading hash]] is used for #uplinks.

Neuron links with titles: [[trailing|Trailing link]]# #[[leading |  Leading link]]

[External links](http://example.com) are marked [as such](ftp://domain).
`, []note.Link{
		{
			Title:    "link",
			Href:     "heading",
			Rels:     []string{},
			External: false,
			Snippet:  "Heading with a [link](heading)",
		},
		{
			Title:    "multiple links",
			Href:     "stripped-formatting",
			Rels:     []string{},
			External: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
		},
		{
			Title:    "relative",
			Href:     "../other",
			Rels:     []string{},
			External: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
		},
		{
			Title:    "one relation",
			Href:     "one",
			Rels:     []string{"rel-1"},
			External: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
		},
		{
			Title:    "several relations",
			Href:     "several",
			Rels:     []string{"rel-1", "rel-2"},
			External: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
		},
		{
			Title:    "https://inline-link.com",
			Href:     "https://inline-link.com",
			External: true,
			Rels:     []string{},
			Snippet:  "An https://inline-link.com and http://another-inline-link.com.",
		},
		{
			Title:    "http://another-inline-link.com",
			Href:     "http://another-inline-link.com",
			External: true,
			Rels:     []string{},
			Snippet:  "An https://inline-link.com and http://another-inline-link.com.",
		},
		{
			Title:    "Wiki link",
			Href:     "Wiki link",
			External: false,
			Rels:     []string{},
			Snippet:  "A [[Wiki link]] is surrounded by [[2-brackets | two brackets]].",
		},
		{
			Title:    "two brackets",
			Href:     "2-brackets",
			External: false,
			Rels:     []string{},
			Snippet:  "A [[Wiki link]] is surrounded by [[2-brackets | two brackets]].",
		},
		{
			Title:    `esca]]ped [chara\cters`,
			Href:     `esca]]ped [chara\cters`,
			External: false,
			Rels:     []string{},
			Snippet:  `It can contain [[esca]\]ped \[chara\\cters]].`,
		},
		{
			Title:    "Folgezettel link",
			Href:     "Folgezettel link",
			External: false,
			Rels:     []string{"down"},
			Snippet:  "A [[[Folgezettel link]]] is surrounded by three brackets.",
		},
		{
			Title:    "trailing hash",
			Href:     "trailing hash",
			External: false,
			Rels:     []string{"down"},
			Snippet:  "Neuron also supports a [[trailing hash]]# for Folgezettel links.",
		},
		{
			Title:    "leading hash",
			Href:     "leading hash",
			External: false,
			Rels:     []string{"up"},
			Snippet:  "A #[[leading hash]] is used for #uplinks.",
		},
		{
			Title:    "Trailing link",
			Href:     "trailing",
			External: false,
			Rels:     []string{"down"},
			Snippet:  "Neuron links with titles: [[trailing|Trailing link]]# #[[leading |  Leading link]]",
		},
		{
			Title:    "Leading link",
			Href:     "leading",
			External: false,
			Rels:     []string{"up"},
			Snippet:  "Neuron links with titles: [[trailing|Trailing link]]# #[[leading |  Leading link]]",
		},
		{
			Title:    "External links",
			Href:     "http://example.com",
			Rels:     []string{},
			External: true,
			Snippet:  `[External links](http://example.com) are marked [as such](ftp://domain).`,
		},
		{
			Title:    "as such",
			Href:     "ftp://domain",
			Rels:     []string{},
			External: true,
			Snippet:  `[External links](http://example.com) are marked [as such](ftp://domain).`,
		},
	})
}

func parse(t *testing.T, source string) note.Content {
	content, err := NewParser().Parse(source)
	assert.Nil(t, err)
	return *content
}
