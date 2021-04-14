package markdown

import (
	"testing"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/test/assert"
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

func TestParseHashtags(t *testing.T) {
	test := func(source string, tags []string) {
		content := parseWithOptions(t, source, ParserOpts{
			HashtagEnabled:      true,
			MultiWordTagEnabled: false,
		})
		assert.Equal(t, content.Tags, tags)
	}

	test("", []string{})
	test("#", []string{})
	test("##", []string{})
	test("# No tags around here", []string{})
	test("#single-hashtag", []string{"single-hashtag"})
	test("a #tag in the middle", []string{"tag"})
	test("#multiple #hashtags", []string{"multiple", "hashtags"})
	test("#multiple#hashtags", []string{})
	// Unicode hashtags
	test("#libellé-français, #日本語ハッシュタグ", []string{"libellé-français", "日本語ハッシュタグ"})
	// Punctuation breaking tags
	test(
		"#a #b, #c; #d. #e! #f? #g* #h\", #i(, #j), #k[, #l], #m{, #n}",
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"},
	)
	// Authorized special characters
	test("#a/@'~-_$%&+=: end", []string{"a/@'~-_$%&+=:"})
	// Escape punctuation and space
	test(`#an\ \\espaced\ tag\!`, []string{`an \espaced tag!`})
	// Hashtags containing only numbers and dots are invalid
	test("#123, #1.2.3", []string{})
	// Must not be preceded by a hash or any other valid hashtag character
	test("##invalid also#invalid", []string{})
	// Bear's multi multi-word tags are disabled
	test("#multi word# end", []string{"multi"})
}

func TestParseWordtags(t *testing.T) {
	test := func(source string, tags []string) {
		content := parseWithOptions(t, source, ParserOpts{
			HashtagEnabled:      true,
			MultiWordTagEnabled: true,
		})
		assert.Equal(t, content.Tags, tags)
	}

	test("", []string{})
	test("#", []string{})
	test("##", []string{})
	test("# No tags around here", []string{})
	test("#single-hashtag", []string{"single-hashtag"})
	test("a #tag in the middle", []string{"tag"})
	test("#multiple #hashtags", []string{"multiple", "hashtags"})
	test("#multiple#hashtags", []string{"multiple"})
	// Unicode hashtags
	test("#libellé-français, #日本語ハッシュタグ", []string{"libellé-français", "日本語ハッシュタグ"})
	// Punctuation breaking tags
	test(
		"#a #b, #c; #d. #e! #f? #g* #h\", #i(, #j), #k[, #l], #m{, #n}",
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"},
	)
	// Authorized special characters
	test("#a/@'~-_$%&+=: end", []string{"a/@'~-_$%&+=:"})
	// Escape punctuation and space
	test(`#an\ \\espaced\ tag\!`, []string{`an \espaced tag!`})
	// Leading and trailing spaces are trimmed
	test(`#\ \	tag\	\   end`, []string{`tag`})
	// Hashtags containing only numbers and dots are invalid
	test("#123, #1.2.3", []string{})
	// Must not be preceded by a hash or any other valid hashtag character
	test("##invalid also#invalid", []string{})
	// Bear's multi multi-word tags
	test("#multi word#", []string{"multi word"})
	test("#surrounded# end", []string{"surrounded"})
	test("#multi word#end", []string{"multi word"})
	test("#multi word #other", []string{"multi", "other"})
	test("#multi word# #other", []string{"multi word", "other"})
	test("#multi word##other", []string{"multi word"})
	test("a #multi word# in the middle", []string{"multi word"})
	test("a #multi word#, and a #tag", []string{"multi word", "tag"})
	test("#multi, word#", []string{"multi"})
}

func TestParseColontags(t *testing.T) {
	test := func(source string, tags []string) {
		content := parseWithOptions(t, source, ParserOpts{
			ColontagEnabled: true,
		})
		assert.Equal(t, content.Tags, tags)
	}

	test("", []string{})
	test(":", []string{})
	test("::", []string{})
	test("not:valid:", []string{})
	test(":no tags:", []string{})
	test(": no-tags:", []string{})
	test(":no-tags :", []string{})
	test(":single-colontag:", []string{"single-colontag"})
	test("a :tag: in the middle", []string{"tag"})
	test(":multiple:colontags:", []string{"multiple", "colontags"})
	test(":multiple::colontags:", []string{"multiple"})
	test(":multiple: :colontags:", []string{"multiple", "colontags"})
	test(":multiple:,:colontags:", []string{"multiple", "colontags"})
	test(":more:than:two:colontags:", []string{"more", "than", "two", "colontags"})
	test(":multiple :colontags", []string{})
	test(":multiple :colontags:", []string{"colontags"})
	// Unicode colontags
	test(":libellé-français:日本語ハッシュタグ:", []string{"libellé-français", "日本語ハッシュタグ"})
	// Punctuation is not allowed
	test(":a : :b,: :c;: :d.: :e!: :f?: :g*: :h\": :i(: :j): :k[: :l]: :m{: :n}:", []string{})
	// Authorized special characters
	test(":#a/@'~-_$%&+=: end", []string{"#a/@'~-_$%&+="})
	// Escape punctuation and space
	test(`:an\ \\espaced\ tag\!:`, []string{`an \espaced tag!`})
	// Leading and trailing spaces are trimmed
	test(`:\ \	tag\	\ :`, []string{`tag`})
	// A colontag containing only numbers is valid
	test(":123:1.2.3:", []string{"123"})
	// Must not be preceded by a : or any other valid colontag character
	test("::invalid also:invalid:", []string{})
}

func TestParseMixedTags(t *testing.T) {
	test := func(source string, tags []string) {
		content := parseWithOptions(t, source, ParserOpts{
			HashtagEnabled:      true,
			MultiWordTagEnabled: true,
			ColontagEnabled:     true,
		})
		assert.Equal(t, content.Tags, tags)
	}

	test(":colontag: #tag #word tag#", []string{"colontag", "tag", "word tag"})
	test(":#colontag: #:tag: #:word:tag:#", []string{"#colontag", ":tag:", ":word:tag:"})
}

func TestParseTagsFromFrontmatter(t *testing.T) {
	test := func(source string, tags []string) {
		content := parse(t, source)
		assert.Equal(t, content.Tags, tags)
	}

	test(`---
Tags:
    - "#tag1"
    - tag 2
---

Body
`, []string{"tag1", "tag 2"})

	test(`---
Keywords: [keyword1, "#keyword 2"]
---

Body
`, []string{"keyword1", "keyword 2"})

	test(`---
tags: [tag1, "   tag 2  "]
keywords:
    - keyword1  
    - keyword 2
---

Body
`, []string{"tag1", "tag 2", "keyword1", "keyword 2"})

	// When a string, parse space-separated tags.
	test(`---
Tags: "tag1   #tag-2"
Keywords: kw1 kw2 kw3
---

Body
`, []string{"tag1", "tag-2", "kw1", "kw2", "kw3"})
}

func TestParseTagsIgnoresDuplicates(t *testing.T) {
	test := func(source string, tags []string) {
		content := parse(t, source)
		assert.Equal(t, content.Tags, tags)
	}

	test(`---
Tags: [tag1, "#tag1", tag2]
---

#tag1 #tag2 #tag3 #tag3 :tag2:
`, []string{"tag1", "tag2", "tag3"})
}

func TestParseLinks(t *testing.T) {
	test := func(source string, links []core.Link) {
		content := parse(t, source)
		assert.Equal(t, content.Links, links)
	}

	test("", []core.Link{})
	test("No links around here", []core.Link{})

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
`, []core.Link{
		{
			Title:        "link",
			Href:         "heading",
			Rels:         []core.LinkRelation{},
			IsExternal:   false,
			Snippet:      "Heading with a [link](heading)",
			SnippetStart: 3,
			SnippetEnd:   33,
		},
		{
			Title:      "multiple links",
			Href:       "stripped-formatting",
			Rels:       []core.LinkRelation{},
			IsExternal: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
			SnippetStart: 35,
			SnippetEnd:   222,
		},
		{
			Title:      "relative",
			Href:       "../other",
			Rels:       []core.LinkRelation{},
			IsExternal: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
			SnippetStart: 35,
			SnippetEnd:   222,
		},
		{
			Title:      "one relation",
			Href:       "one",
			Rels:       core.LinkRels("rel-1"),
			IsExternal: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
			SnippetStart: 35,
			SnippetEnd:   222,
		},
		{
			Title:      "several relations",
			Href:       "several",
			Rels:       core.LinkRels("rel-1", "rel-2"),
			IsExternal: false,
			Snippet: `Paragraph containing [multiple **links**](stripped-formatting), here's one [relative](../other).
A link can have [one relation](one "rel-1") or [several relations](several "rel-1 rel-2").`,
			SnippetStart: 35,
			SnippetEnd:   222,
		},
		{
			Title:        "https://inline-link.com",
			Href:         "https://inline-link.com",
			IsExternal:   true,
			Rels:         []core.LinkRelation{},
			Snippet:      "An https://inline-link.com and http://another-inline-link.com.",
			SnippetStart: 224,
			SnippetEnd:   286,
		},
		{
			Title:        "http://another-inline-link.com",
			Href:         "http://another-inline-link.com",
			IsExternal:   true,
			Rels:         []core.LinkRelation{},
			Snippet:      "An https://inline-link.com and http://another-inline-link.com.",
			SnippetStart: 224,
			SnippetEnd:   286,
		},
		{
			Title:        "Wiki link",
			Href:         "Wiki link",
			IsExternal:   false,
			Rels:         []core.LinkRelation{},
			Snippet:      "A [[Wiki link]] is surrounded by [[2-brackets | two brackets]].",
			SnippetStart: 288,
			SnippetEnd:   351,
		},
		{
			Title:        "two brackets",
			Href:         "2-brackets",
			IsExternal:   false,
			Rels:         []core.LinkRelation{},
			Snippet:      "A [[Wiki link]] is surrounded by [[2-brackets | two brackets]].",
			SnippetStart: 288,
			SnippetEnd:   351,
		},
		{
			Title:        `esca]]ped [chara\cters`,
			Href:         `esca]]ped [chara\cters`,
			IsExternal:   false,
			Rels:         []core.LinkRelation{},
			Snippet:      `It can contain [[esca]\]ped \[chara\\cters]].`,
			SnippetStart: 353,
			SnippetEnd:   398,
		},
		{
			Title:        "Folgezettel link",
			Href:         "Folgezettel link",
			IsExternal:   false,
			Rels:         core.LinkRels("down"),
			Snippet:      "A [[[Folgezettel link]]] is surrounded by three brackets.",
			SnippetStart: 400,
			SnippetEnd:   457,
		},
		{
			Title:        "trailing hash",
			Href:         "trailing hash",
			IsExternal:   false,
			Rels:         core.LinkRels("down"),
			Snippet:      "Neuron also supports a [[trailing hash]]# for Folgezettel links.",
			SnippetStart: 459,
			SnippetEnd:   523,
		},
		{
			Title:        "leading hash",
			Href:         "leading hash",
			IsExternal:   false,
			Rels:         core.LinkRels("up"),
			Snippet:      "A #[[leading hash]] is used for #uplinks.",
			SnippetStart: 525,
			SnippetEnd:   566,
		},
		{
			Title:        "Trailing link",
			Href:         "trailing",
			IsExternal:   false,
			Rels:         core.LinkRels("down"),
			Snippet:      "Neuron links with titles: [[trailing|Trailing link]]# #[[leading |  Leading link]]",
			SnippetStart: 568,
			SnippetEnd:   650,
		},
		{
			Title:        "Leading link",
			Href:         "leading",
			IsExternal:   false,
			Rels:         core.LinkRels("up"),
			Snippet:      "Neuron links with titles: [[trailing|Trailing link]]# #[[leading |  Leading link]]",
			SnippetStart: 568,
			SnippetEnd:   650,
		},
		{
			Title:        "External links",
			Href:         "http://example.com",
			Rels:         []core.LinkRelation{},
			IsExternal:   true,
			Snippet:      `[External links](http://example.com) are marked [as such](ftp://domain).`,
			SnippetStart: 652,
			SnippetEnd:   724,
		},
		{
			Title:        "as such",
			Href:         "ftp://domain",
			Rels:         []core.LinkRelation{},
			IsExternal:   true,
			Snippet:      `[External links](http://example.com) are marked [as such](ftp://domain).`,
			SnippetStart: 652,
			SnippetEnd:   724,
		},
	})
}

func TestParseMetadataFromFrontmatter(t *testing.T) {
	test := func(source string, expectedMetadata map[string]interface{}) {
		content := parse(t, source)
		assert.Equal(t, content.Metadata, expectedMetadata)
	}

	test("", map[string]interface{}{})
	test("# A title", map[string]interface{}{})
	test("---\n---\n# A title", map[string]interface{}{})
	test(`---
title: A title
tags:
  - tag1
  - "tag 2"
nested:
  key: value
---

Paragraph
`, map[string]interface{}{
		"title": "A title",
		"tags":  []interface{}{"tag1", "tag 2"},
		"nested": map[string]interface{}{
			"key": "value",
		},
	})
}

func parse(t *testing.T, source string) core.ParsedNote {
	return parseWithOptions(t, source, ParserOpts{
		HashtagEnabled:      true,
		MultiWordTagEnabled: true,
		ColontagEnabled:     true,
	})
}

func parseWithOptions(t *testing.T, source string, options ParserOpts) core.ParsedNote {
	content, err := NewParser(options).Parse(source)
	assert.Nil(t, err)
	return *content
}
