package paths

import (
	"testing"

	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/fixtures"
)

func TestWalk(t *testing.T) {
	var path = fixtures.Path("walk")

	testEqual(t, Walk(path, "md", &util.NullLogger), []string{
		"a.md",
		"b.md",
		"dir1/a.md",
		"dir1/b.md",
		"dir1/dir1/a.md",
		"dir2/a.md",
	})
}

func testEqual(t *testing.T, actual <-chan Metadata, expected []string) {
	popExpected := func() (string, bool) {
		if len(expected) == 0 {
			return "", false
		}
		item := expected[0]
		expected = expected[1:]
		return item, true
	}

	for act := range actual {
		exp, ok := popExpected()
		if !ok {
			t.Errorf("More paths available than expected")
			return
		}
		assert.Equal(t, act.Path, exp)
		assert.NotNil(t, act.Modified)
	}

	if len(expected) > 0 {
		t.Errorf("Missing expected paths: %v", expected)
	}
}
