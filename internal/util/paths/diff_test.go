package paths

import (
	"errors"
	"testing"
	"time"

	"github.com/zk-org/zk/internal/util/test/assert"
)

var date1 = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
var date2 = time.Date(2012, 10, 20, 12, 34, 58, 651387237, time.UTC)
var date3 = time.Date(2014, 12, 10, 3, 34, 58, 651387237, time.UTC)

func TestDiffEmpty(t *testing.T) {
	source := []Metadata{}
	target := []Metadata{}
	test(t, source, target, false, []DiffChange{})
}

func TestNoDiff(t *testing.T) {
	files := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
		{
			Path:     "b/1",
			Modified: date3,
		},
	}

	test(t, files, files, false, []DiffChange{
		{Path: "a/1", Kind: DiffUnchanged},
		{Path: "a/2", Kind: DiffUnchanged},
		{Path: "b/1", Kind: DiffUnchanged},
	})
}

func TestDiff(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
		{
			Path:     "b/1",
			Modified: date3,
		},
	}

	target := []Metadata{
		{
			// Date changed
			Path:     "a/1",
			Modified: date1.Add(time.Hour),
		},
		// 2 is added
		{
			// 3 is removed
			Path:     "a/3",
			Modified: date3,
		},
		{
			// No change
			Path:     "b/1",
			Modified: date3,
		},
	}

	test(t, source, target, false, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffModified,
		},
		{
			Path: "a/2",
			Kind: DiffAdded,
		},
		{
			Path: "a/3",
			Kind: DiffRemoved,
		},
		{
			Path: "b/1",
			Kind: DiffUnchanged,
		},
	})
}

func TestDiffForceModified(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
		{
			Path:     "b/1",
			Modified: date3,
		},
	}

	target := []Metadata{
		{
			// Date changed
			Path:     "a/1",
			Modified: date1.Add(time.Hour),
		},
		// 2 is added
		{
			// 3 is removed
			Path:     "a/3",
			Modified: date3,
		},
		{
			// No change
			Path:     "b/1",
			Modified: date3,
		},
	}

	test(t, source, target, true, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffModified,
		},
		{
			Path: "a/2",
			Kind: DiffAdded,
		},
		{
			Path: "a/3",
			Kind: DiffRemoved,
		},
		{
			// Forced modified
			Path: "b/1",
			Kind: DiffModified,
		},
	})
}

func TestDiffWithMoreInSource(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
	}

	target := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
	}

	test(t, source, target, false, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffUnchanged,
		},
		{
			Path: "a/2",
			Kind: DiffAdded,
		},
	})
}

func TestDiffWithMoreInTarget(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
	}

	target := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
	}

	test(t, source, target, false, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffUnchanged,
		},
		{
			Path: "a/2",
			Kind: DiffRemoved,
		},
	})
}

func TestDiffEmptySource(t *testing.T) {
	source := []Metadata{}

	target := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
	}

	test(t, source, target, false, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffRemoved,
		},
		{
			Path: "a/2",
			Kind: DiffRemoved,
		},
	})
}

func TestDiffEmptyTarget(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
	}

	target := []Metadata{}

	test(t, source, target, false, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffAdded,
		},
		{
			Path: "a/2",
			Kind: DiffAdded,
		},
	})
}

func TestDiffCancellation(t *testing.T) {
	source := []Metadata{
		{
			Path:     "a/1",
			Modified: date1,
		},
		{
			Path:     "a/2",
			Modified: date2,
		},
	}

	target := []Metadata{}

	received := make([]DiffChange, 0)
	_, err := Diff(toChannel(source), toChannel(target), false, func(change DiffChange) error {
		received = append(received, change)

		if len(received) == 1 {
			return errors.New("cancelled")
		} else {
			return nil
		}
	})

	assert.Equal(t, received, []DiffChange{
		{
			Path: "a/1",
			Kind: DiffAdded,
		},
	})
	assert.Err(t, err, "cancelled")
}

func test(t *testing.T, source, target []Metadata, forceModified bool, expected []DiffChange) {
	expectedCount := len(source)
	received := make([]DiffChange, 0)
	actualCount, err := Diff(toChannel(source), toChannel(target), forceModified, func(change DiffChange) error {
		received = append(received, change)
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, actualCount, expectedCount)
	assert.Equal(t, received, expected)
}

func toChannel(fm []Metadata) <-chan Metadata {
	c := make(chan Metadata)
	go func() {
		for _, m := range fm {
			c <- m
		}
		close(c)
	}()
	return c
}
