package cmd

import (
	"sort"
	"strings"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
)

// Completers for the values zk can only know by looking at the notebook. They
// all fail quietly: a front-end asking for suggestions is never a good place
// to report that a notebook couldn't be opened.

// completeTags suggests the tags used in the current notebook.
func completeTags(container *cli.Container) spec.Completer {
	return func(prefix string) []string {
		notebook, err := container.CurrentNotebook()
		if err != nil {
			return nil
		}
		tags, err := notebook.FindCollections(core.CollectionKindTag, nil)
		if err != nil {
			return nil
		}
		names := []string{}
		for _, tag := range tags {
			if strings.HasPrefix(tag.Name, prefix) {
				names = append(names, tag.Name)
			}
		}
		return names
	}
}

// completeNotePaths suggests the paths of the notes in the current notebook.
func completeNotePaths(container *cli.Container) spec.Completer {
	return func(prefix string) []string {
		notebook, err := container.CurrentNotebook()
		if err != nil {
			return nil
		}
		notes, err := notebook.FindMinimalNotes(core.NoteFindOpts{})
		if err != nil {
			return nil
		}
		paths := []string{}
		for _, note := range notes {
			if strings.HasPrefix(note.Path, prefix) {
				paths = append(paths, note.Path)
			}
		}
		return paths
	}
}

// completeGroups suggests the note groups the configuration declares.
func completeGroups(container *cli.Container) spec.Completer {
	return func(prefix string) []string {
		notebook, err := container.CurrentNotebook()
		if err != nil {
			return nil
		}
		return completeKeys(notebook.Config.Groups, prefix)
	}
}

func completeKeys[V any](values map[string]V, prefix string) []string {
	keys := []string{}
	for key := range values {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// Completers for the values which are the same in every notebook.

func completeNoteSortTerms() spec.Completer {
	return spec.Values(
		"created", "created-", "created+",
		"modified", "modified-", "modified+",
		"path", "path-", "path+",
		"title", "title-", "title+",
		"random",
		"word-count", "word-count-", "word-count+",
	)
}

func completeTagSortTerms() spec.Completer {
	return spec.Values("name", "name-", "name+", "note-count", "note-count-", "note-count+")
}

func completeMatchStrategies() spec.Completer {
	return spec.Values("fts", "re", "exact")
}

// filteringFlags returns the filtering flags, with the suggestions zk can
// offer for each of them.
func filteringFlags(container *cli.Container, f *cli.Filtering) []spec.Flag {
	paths := completeNotePaths(container)

	flags := f.Flags()
	for i := range flags {
		switch flags[i].Name {
		case "tag":
			flags[i].Complete = completeTags(container)
		case "sort":
			flags[i].Complete = completeNoteSortTerms()
		case "match-strategy":
			flags[i].Complete = completeMatchStrategies()
		case "exclude", "mention", "mentioned-by", "link-to", "no-link-to",
			"linked-by", "no-linked-by", "related":
			flags[i].Complete = paths
		}
	}
	return flags
}
