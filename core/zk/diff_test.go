package zk

import (
	"errors"
	"testing"
	"time"

	"github.com/mickael-menu/zk/util/assert"
)

var date1 = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
var date2 = time.Date(2012, 10, 20, 12, 34, 58, 651387237, time.UTC)
var date3 = time.Date(2014, 12, 10, 3, 34, 58, 651387237, time.UTC)
var date4 = time.Date(2016, 13, 11, 4, 34, 58, 651387237, time.UTC)

func TestDiffEmpty(t *testing.T) {
	source := []FileMetadata{}
	target := []FileMetadata{}
	test(t, source, target, []DiffChange{})
}

func TestNoDiff(t *testing.T) {
	files := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
		{
			Path:     Path{Dir: "b", Filename: "1"},
			Modified: date3,
		},
	}

	test(t, files, files, []DiffChange{})
}

func TestDiff(t *testing.T) {
	source := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
		{
			Path:     Path{Dir: "b", Filename: "1"},
			Modified: date3,
		},
	}

	target := []FileMetadata{
		{
			// Date changed
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1.Add(time.Hour),
		},
		// 2 is added
		{
			// 3 is removed
			Path:     Path{Dir: "a", Filename: "3"},
			Modified: date3,
		},
		{
			// No change
			Path:     Path{Dir: "b", Filename: "1"},
			Modified: date3,
		},
	}

	test(t, source, target, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "1"},
			Kind: DiffModified,
		},
		{
			Path: Path{Dir: "a", Filename: "2"},
			Kind: DiffAdded,
		},
		{
			Path: Path{Dir: "a", Filename: "3"},
			Kind: DiffRemoved,
		},
	})
}

func TestDiffWithMoreInSource(t *testing.T) {
	source := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
	}

	target := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
	}

	test(t, source, target, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "2"},
			Kind: DiffAdded,
		},
	})
}

func TestDiffWithMoreInTarget(t *testing.T) {
	source := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
	}

	target := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
	}

	test(t, source, target, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "2"},
			Kind: DiffRemoved,
		},
	})
}

func TestDiffEmptySource(t *testing.T) {
	source := []FileMetadata{}

	target := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
	}

	test(t, source, target, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "1"},
			Kind: DiffRemoved,
		},
		{
			Path: Path{Dir: "a", Filename: "2"},
			Kind: DiffRemoved,
		},
	})
}

func TestDiffEmptyTarget(t *testing.T) {
	source := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
	}

	target := []FileMetadata{}

	test(t, source, target, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "1"},
			Kind: DiffAdded,
		},
		{
			Path: Path{Dir: "a", Filename: "2"},
			Kind: DiffAdded,
		},
	})
}

func TestDiffCancellation(t *testing.T) {
	source := []FileMetadata{
		{
			Path:     Path{Dir: "a", Filename: "1"},
			Modified: date1,
		},
		{
			Path:     Path{Dir: "a", Filename: "2"},
			Modified: date2,
		},
	}

	target := []FileMetadata{}

	received := make([]DiffChange, 0)
	err := Diff(toChannel(source), toChannel(target), func(change DiffChange) error {
		received = append(received, change)

		if len(received) == 1 {
			return errors.New("cancelled")
		} else {
			return nil
		}
	})

	assert.Equal(t, received, []DiffChange{
		{
			Path: Path{Dir: "a", Filename: "1"},
			Kind: DiffAdded,
		},
	})
	assert.Err(t, err, "cancelled")
}

func test(t *testing.T, source, target []FileMetadata, expected []DiffChange) {
	received := make([]DiffChange, 0)
	err := Diff(toChannel(source), toChannel(target), func(change DiffChange) error {
		received = append(received, change)
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, received, expected)
}

func toChannel(fm []FileMetadata) <-chan FileMetadata {
	c := make(chan FileMetadata)
	go func() {
		for _, m := range fm {
			c <- m
		}
		close(c)
	}()
	return c
}
