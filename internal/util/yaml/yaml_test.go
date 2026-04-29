package yaml

import (
	"bytes"
	"encoding/json"
	"testing"
)

// helper to find the byte offset of a substring in source, used to
// assert tag positions without hardcoding magic numbers.
func mustFindOffset(t *testing.T, source []byte, tag string) int {
	t.Helper()
	idx := bytes.Index(source, []byte(tag))
	if idx == -1 {
		t.Fatalf("tag %q not found in source", tag)
	}
	return idx
}

func TestParseFrontmatterBasic(t *testing.T) {
	src := []byte(`---
title: Hello World
date: 2024-01-15
draft: false
count: 42
---

Body text here.
`)
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		key  string
		want string
	}{
		{"title", "Hello World"},
		{"date", "2024-01-15"},
		{"draft", "false"},
		{"count", "42"},
	}
	for _, tc := range cases {
		v, ok := fm.GetValue(tc.key)
		if !ok {
			t.Errorf("key %q not found in Values", tc.key)
			continue
		}
		if v.(string) != tc.want {
			t.Errorf("key %q: got %q, want %q", tc.key, v, tc.want)
		}
	}
}

func TestParseFrontmatterKeysAreLowercased(t *testing.T) {
	src := []byte("---\nTitle: My Note\nAUTHOR: Alice\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, key := range []string{"title", "author"} {
		if _, ok := fm.GetValue(key); !ok {
			t.Errorf("expected lowercased key %q to be present", key)
		}
	}
	for _, key := range []string{"Title", "AUTHOR"} {
		if _, ok := fm.GetValue(key); ok {
			t.Errorf("expected original-case key %q to be absent", key)
		}
	}
}

func TestParseFrontmatterMalformedYAML(t *testing.T) {
	src := []byte("---\n: bad: yaml: here\n---\n")
	_, err := ParseFrontmatter(src)
	if err == nil {
		t.Error("expected an error for malformed YAML, got nil")
	}
}

func TestGetTitleString_Present(t *testing.T) {
	src := []byte("---\ntitle: My Great Note\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := fm.GetTitleString()
	if title.IsNull() {
		t.Fatal("expected title to be found")
	}
	if title.String() != "My Great Note" {
		t.Errorf("got %q, want %q", title, "My Great Note")
	}
}

func TestTagsSingleScalarSpaceSeparated(t *testing.T) {
	src := []byte("---\ntags: foo bar baz\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantNames := []string{"foo", "bar", "baz"}
	if len(fm.Tags) != len(wantNames) {
		t.Fatalf("got %d tags, want %d: %v", len(fm.Tags), len(wantNames), fm.Tags)
	}
	for i, tag := range fm.Tags {
		if tag.Name != wantNames[i] {
			t.Errorf("tag[%d]: got %q, want %q", i, tag.Name, wantNames[i])
		}
	}
}

func TestTagsSequenceFormat(t *testing.T) {
	src := []byte("---\ntags:\n  - alpha\n  - beta\n  - gamma\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantNames := []string{"alpha", "beta", "gamma"}
	if len(fm.Tags) != len(wantNames) {
		t.Fatalf("got %d tags, want %d: %v", len(fm.Tags), len(wantNames), fm.Tags)
	}
	for i, tag := range fm.Tags {
		if tag.Name != wantNames[i] {
			t.Errorf("tag[%d]: got %q, want %q", i, tag.Name, wantNames[i])
		}
	}
}

func TestTagsAllSupportedKeywords(t *testing.T) {
	for _, keyword := range []string{"tag", "tags", "keyword", "keywords"} {
		t.Run(keyword, func(t *testing.T) {
			src := []byte("---\n" + keyword + ":\n  - one\n  - two\n---\n")
			fm, err := ParseFrontmatter(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fm.Tags) != 2 {
				t.Errorf("got %d tags, want 2", len(fm.Tags))
			}
		})
	}
}

func TestTagPositionsScalarTags(t *testing.T) {
	src := []byte("---\ntags: foo bar\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(fm.Tags))
	}

	for _, tag := range fm.Tags {
		wantPos := mustFindOffset(t, src, tag.Name)
		if tag.Pos != wantPos {
			t.Errorf("tag %q: Pos=%d, want %d", tag.Name, tag.Pos, wantPos)
		}
	}
}

func TestTagPositionsSequenceTagsAlt(t *testing.T) {
	src := []byte("---\ntags: [foo, bar]\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(fm.Tags))
	}

	for _, tag := range fm.Tags {
		wantPos := mustFindOffset(t, src, tag.Name)
		if tag.Pos != wantPos {
			t.Errorf("tag %q: Pos=%d, want %d", tag.Name, tag.Pos, wantPos)
		}
	}
}

func TestTagPositionsSequenceTags(t *testing.T) {
	src := []byte("---\ntags:\n  - alpha\n  - beta\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(fm.Tags))
	}

	for _, tag := range fm.Tags {
		wantPos := mustFindOffset(t, src, tag.Name)
		if tag.Pos != wantPos {
			t.Errorf("tag %q: Pos=%d, want %d", tag.Name, tag.Pos, wantPos)
		}
	}
}

func TestTagPositionsMultilineDocument(t *testing.T) {
	src := []byte("---\ntitle: My Note\ndate: 2024-01-01\ntags:\n  - golang\n  - testing\n---\n\nBody.\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(fm.Tags))
	}

	for _, tag := range fm.Tags {
		wantPos := mustFindOffset(t, src, tag.Name)
		if tag.Pos != wantPos {
			t.Errorf("tag %q: Pos=%d, want %d", tag.Name, tag.Pos, wantPos)
		}
	}
}

func TestValuesJSONSerializable(t *testing.T) {
	src := []byte(`---
title: JSON Test
count: 7
flag: true
list:
  - a
  - b
nested:
  x: 1
tags:
  - go
---
`)
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = json.Marshal(fm.Values)
	if err != nil {
		t.Errorf("Values is not JSON-serializable: %v", err)
	}
}

func TestValuesJSONRoundtrip(t *testing.T) {
	src := []byte("---\ntitle: Round Trip\nauthor: Alice\n---\n")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(fm.Values)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundtripped map[string]any
	if err := json.Unmarshal(data, &roundtripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	for _, key := range []string{"title", "author"} {
		orig, _ := fm.Values[key]
		rt, ok := roundtripped[key]
		if !ok {
			t.Errorf("key %q missing after round-trip", key)
			continue
		}
		if orig != rt {
			t.Errorf("key %q: original=%v, after round-trip=%v", key, orig, rt)
		}
	}
}

func TestValuesJSONNilSafe(t *testing.T) {
	// Document with no frontmatter — Values is nil; marshalling should not panic.
	src := []byte("# No frontmatter here")
	fm, err := ParseFrontmatter(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = json.Marshal(fm.Values)
	if err != nil {
		t.Errorf("marshalling nil Values failed: %v", err)
	}
}
