package paths

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/fixtures"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestWalk(t *testing.T) {
	var path = fixtures.Path("walk")

	actual := make([]string, 0)
	for m := range Walk(path, "md", &util.NullLogger) {
		assert.NotNil(t, m.Modified)
		actual = append(actual, m.Path)
	}

	assert.Equal(t, actual, []string{
		"Dir3/a.md",
		"a.md",
		"b.md",
		"dir1/a.md",
		"dir1/b.md",
		"dir1/dir1/a.md",
		"dir1 a space/a.md",
		"dir2/a.md",
	})
}
