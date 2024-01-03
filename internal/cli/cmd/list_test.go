package cmd

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestListFormatDefault(t *testing.T) {
	cmd := List{}
	assert.Equal(t, cmd.noteTemplate(), `{{style "title" title}} {{style "path" path}} ({{format-date created "elapsed"}})

{{list snippets}}`)
}

func TestListFormatPredefined(t *testing.T) {
	test := func(format, expectedTemplate string) {
		cmd := List{Format: format}
		assert.Equal(t, cmd.noteTemplate(), expectedTemplate)
	}

	// Known formats
	test("json", `{{json .}}`)
	test("jsonl", `{{json .}}`)
	test("path", `{{path}}`)
	test("link", `{{link}}`)

	test("oneline", `{{style "title" title}} {{style "path" path}} ({{format-date created "elapsed"}})`)

	test("short", `{{style "title" title}} {{style "path" path}} ({{format-date created "elapsed"}})

{{list snippets}}`)

	test("medium", `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}

{{list snippets}}`)

	test("long", `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}
Modified: {{format-date modified "short"}}

{{list snippets}}`)

	test("full", `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}
Modified: {{format-date modified "short"}}
Tags: {{join tags ", "}}

{{prepend "  " body}}
`)

	// Predefined formats are case sensitive.
	test("Path", "Path")
}

func TestListFormatCustom(t *testing.T) {
	test := func(format, expectedTemplate string) {
		cmd := List{Format: format}
		assert.Equal(t, cmd.noteTemplate(), expectedTemplate)
	}

	// Custom formats are used literally.
	test("{{title}}", "{{title}}")
	// \n and \t in custom formats are expanded.
	test(`{{title}}\t{{path}}\n{{snippet}}`, "{{title}}\t{{path}}\n{{snippet}}")
}
