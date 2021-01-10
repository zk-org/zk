package paths

import (
	"testing"

	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/fixtures"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestWalk(t *testing.T) {
	var path = fixtures.Path("walk")

	actual := make([]string, 0)
	for m := range Walk(path, "md", &util.NullLogger) {
		assert.NotNil(t, m.Modified)
		actual = append(actual, m.Path)
	}

	assert.Equal(t, actual, []string{
		"a.md",
		"b.md",
		"dir1/a.md",
		"dir1/b.md",
		"dir1/dir1/a.md",
		"dir2/a.md",
	})
}
