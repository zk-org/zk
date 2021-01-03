package file

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/fixtures"
)

var root = fixtures.Path("walk")

func TestWalkRootDir(t *testing.T) {
	dir := zk.Dir{Name: "", Path: root}
	res := toSlice(Walk(dir, "md", &util.NullLogger))
	assert.Equal(t, res, []Metadata{
		{
			Path:     Path{Dir: "", Filename: "a.md", Abs: filepath.Join(root, "a.md")},
			Modified: date("2021-01-03T11:30:26.069257899+01:00"),
		},
		{
			Path:     Path{Dir: "", Filename: "b.md", Abs: filepath.Join(root, "b.md")},
			Modified: date("2021-01-03T11:30:27.545667767+01:00"),
		},
		{
			Path:     Path{Dir: "dir1", Filename: "a.md", Abs: filepath.Join(root, "dir1/a.md")},
			Modified: date("2021-01-03T11:31:18.961628888+01:00"),
		},
		{
			Path:     Path{Dir: "dir1", Filename: "b.md", Abs: filepath.Join(root, "dir1/b.md")},
			Modified: date("2021-01-03T11:31:24.692881103+01:00"),
		},
		{
			Path:     Path{Dir: "dir1/dir1", Filename: "a.md", Abs: filepath.Join(root, "dir1/dir1/a.md")},
			Modified: date("2021-01-03T11:31:27.900472856+01:00"),
		},
		{
			Path:     Path{Dir: "dir2", Filename: "a.md", Abs: filepath.Join(root, "dir2/a.md")},
			Modified: date("2021-01-03T11:31:51.001827456+01:00"),
		},
	})
}

func TestWalkSubDir(t *testing.T) {
	dir := zk.Dir{Name: "dir1", Path: filepath.Join(root, "dir1")}
	res := toSlice(Walk(dir, "md", &util.NullLogger))
	assert.Equal(t, res, []Metadata{
		{
			Path:     Path{Dir: "dir1", Filename: "a.md", Abs: filepath.Join(root, "dir1/a.md")},
			Modified: date("2021-01-03T11:31:18.961628888+01:00"),
		},
		{
			Path:     Path{Dir: "dir1", Filename: "b.md", Abs: filepath.Join(root, "dir1/b.md")},
			Modified: date("2021-01-03T11:31:24.692881103+01:00"),
		},
		{
			Path:     Path{Dir: "dir1/dir1", Filename: "a.md", Abs: filepath.Join(root, "dir1/dir1/a.md")},
			Modified: date("2021-01-03T11:31:27.900472856+01:00"),
		},
	})
}

func TestWalkSubSubDir(t *testing.T) {
	dir := zk.Dir{Name: "dir1/dir1", Path: filepath.Join(root, "dir1/dir1")}
	res := toSlice(Walk(dir, "md", &util.NullLogger))
	assert.Equal(t, res, []Metadata{
		{
			Path:     Path{Dir: "dir1/dir1", Filename: "a.md", Abs: filepath.Join(root, "dir1/dir1/a.md")},
			Modified: date("2021-01-03T11:31:27.900472856+01:00"),
		},
	})
}

func date(s string) time.Time {
	date, _ := time.Parse(time.RFC3339, s)
	return date
}

func toSlice(c <-chan Metadata) []Metadata {
	s := make([]Metadata, 0)
	for fm := range c {
		s = append(s, fm)
	}
	return s
}
