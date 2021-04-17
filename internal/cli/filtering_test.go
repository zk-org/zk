package cli

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestExpandNamedFiltersNone(t *testing.T) {
	f := Filtering{
		Path:           []string{"path1"},
		Limit:          10,
		Interactive:    true,
		Match:          "match query",
		Exclude:        []string{"excl-path1", "excl-path2"},
		Tag:            []string{"tag1", "tag2"},
		Mention:        []string{"mention1", "mention2"},
		MentionedBy:    []string{"note1", "note2"},
		LinkTo:         []string{"link1", "link2"},
		NoLinkTo:       []string{"link3", "link4"},
		LinkedBy:       []string{"linked1", "linked2"},
		NoLinkedBy:     []string{"linked3", "linked4"},
		Related:        []string{"related1", "related2"},
		MaxDistance:    2,
		Created:        "yesterday",
		CreatedBefore:  "two days ago",
		CreatedAfter:   "three days ago",
		Modified:       "tomorrow",
		ModifiedBefore: "two days",
		ModifiedAfter:  "three days",
		Sort:           []string{"title", "created"},
	}

	res, err := f.ExpandNamedFilters(
		map[string]string{
			"recents": "--created-after '2 weeks ago'",
			"journal": "log --sort created",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.Equal(t, res, f)
}

// ExpandNamedFilters: list options are concatenated.
func TestExpandNamedFiltersJoinLists(t *testing.T) {
	f := Filtering{
		Path:        []string{"path1", "f1", "f2"},
		Exclude:     []string{"excl-path1", "excl-path2"},
		Tag:         []string{"tag1", "tag2"},
		Mention:     []string{"mention1", "mention2"},
		MentionedBy: []string{"note1", "note2"},
		LinkTo:      []string{"link1", "link2"},
		NoLinkTo:    []string{"link3", "link4"},
		LinkedBy:    []string{"linked1", "linked2"},
		NoLinkedBy:  []string{"linked3", "linked4"},
		Related:     []string{"related1", "related2"},
		Sort:        []string{"title", "created"},
	}

	res, err := f.ExpandNamedFilters(
		map[string]string{
			"f1": "path2 --exclude excl-path3 -x excl-path4 --tag tag3 -t tag4 --mention mention3,mention4 --mentioned-by note3",
			"f2": "--link-to link5 --no-link-to link6 --linked-by linked5 --no-linked-by linked6 --related related3 --related related4 --sort random-",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.Equal(t, res.Path, []string{"path1", "path2"})
	assert.Equal(t, res.Exclude, []string{"excl-path1", "excl-path2", "excl-path3", "excl-path4"})
	assert.Equal(t, res.Tag, []string{"tag1", "tag2", "tag3", "tag4"})
	assert.Equal(t, res.Mention, []string{"mention1", "mention2", "mention3", "mention4"})
	assert.Equal(t, res.MentionedBy, []string{"note1", "note2", "note3"})
	assert.Equal(t, res.LinkTo, []string{"link1", "link2", "link5"})
	assert.Equal(t, res.NoLinkTo, []string{"link3", "link4", "link6"})
	assert.Equal(t, res.LinkedBy, []string{"linked1", "linked2", "linked5"})
	assert.Equal(t, res.NoLinkedBy, []string{"linked3", "linked4", "linked6"})
	assert.Equal(t, res.Related, []string{"related1", "related2", "related3", "related4"})
	assert.Equal(t, res.Sort, []string{"title", "created", "random-"})
}

// ExpandNamedFilters: boolean options are computed with disjunction.
func TestExpandNamedFiltersJoinBools(t *testing.T) {
	f := Filtering{
		Path: []string{"path1", "f1", "f2"},
	}

	res, err := f.ExpandNamedFilters(
		map[string]string{
			"f1": "--exact-match --interactive --orphan",
			"f2": "--recursive",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.True(t, res.ExactMatch)
	assert.True(t, res.Interactive)
	assert.True(t, res.Orphan)
	assert.True(t, res.Recursive)
}

// ExpandNamedFilters: non-zero integer and non-empty string options take precedence over named filters.
func TestExpandNamedFiltersJoinLitterals(t *testing.T) {
	f1 := Filtering{Path: []string{"f1", "f2"}}
	res1, err := f1.ExpandNamedFilters(
		map[string]string{
			"f1": "--limit 42 --created 'yesterday' --created-before '2 days ago' --created-after '3 days ago'",
			"f2": "--max-distance 24 --modified 'tomorrow' --modified-before '2 days' --modified-after '3 days'",
		},
		[]string{},
	)
	assert.Nil(t, err)
	assert.Equal(t, res1.Limit, 42)
	assert.Equal(t, res1.MaxDistance, 24)
	assert.Equal(t, res1.Created, "yesterday")
	assert.Equal(t, res1.CreatedBefore, "2 days ago")
	assert.Equal(t, res1.CreatedAfter, "3 days ago")
	assert.Equal(t, res1.Modified, "tomorrow")
	assert.Equal(t, res1.ModifiedBefore, "2 days")
	assert.Equal(t, res1.ModifiedAfter, "3 days")

	f2 := Filtering{
		Path:           []string{"f1", "f2"},
		Limit:          10,
		MaxDistance:    20,
		Created:        "last week",
		CreatedBefore:  "two weeks ago",
		CreatedAfter:   "three weeks ago",
		Modified:       "next week",
		ModifiedBefore: "two weeks",
		ModifiedAfter:  "three weeks",
	}
	res2, err := f2.ExpandNamedFilters(
		map[string]string{
			"f1": "--limit 42 --created 'yesterday' --created-before '2 days ago' --created-after '3 days ago'",
			"f2": "--max-distance 24 --modified 'tomorrow' --modified-before '2 days' --modified-after '3 days'",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.Equal(t, res2.Limit, 10)
	assert.Equal(t, res2.MaxDistance, 20)
	assert.Equal(t, res2.Created, "last week")
	assert.Equal(t, res2.CreatedBefore, "two weeks ago")
	assert.Equal(t, res2.CreatedAfter, "three weeks ago")
	assert.Equal(t, res2.Modified, "next week")
	assert.Equal(t, res2.ModifiedBefore, "two weeks")
	assert.Equal(t, res2.ModifiedAfter, "three weeks")
}

// ExpandNamedFilters: Match option predicates are cumulated with AND.
func TestExpandNamedFiltersJoinMatch(t *testing.T) {
	f := Filtering{
		Path:  []string{"f1", "f2"},
		Match: "(chocolate OR caramel)",
	}

	res, err := f.ExpandNamedFilters(
		map[string]string{
			"f1": "--match banana",
			"f2": "--match apple",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.Equal(t, res.Match, "(((chocolate OR caramel)) AND (banana)) AND (apple)")
}

func TestExpandNamedFiltersExpandsRecursively(t *testing.T) {
	f := Filtering{
		Path: []string{"path1", "journal", "recents"},
	}

	res, err := f.ExpandNamedFilters(
		map[string]string{
			"recents":      "--created-after '2 weeks ago'",
			"journal":      "journal sort-created",
			"sort-created": "--sort created",
		},
		[]string{},
	)

	assert.Nil(t, err)
	assert.Equal(t, res.Path, []string{"path1", "journal"})
	assert.Equal(t, res.CreatedAfter, "2 weeks ago")
	assert.Equal(t, res.Sort, []string{"created"})
}

func TestExpandNamedFiltersReportsParsingError(t *testing.T) {
	f := Filtering{Path: []string{"f1"}}

	_, err := f.ExpandNamedFilters(
		map[string]string{
			"f1": "--test",
		},
		[]string{},
	)

	assert.Err(t, err, "failed to expand named filter `f1`: unknown flag --test")
}
