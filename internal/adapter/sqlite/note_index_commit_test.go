package sqlite

import (
	"testing"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// Commit hands the transaction a NoteIndex built from the receiver. That index
// must carry every field of the receiver: the whole indexing pass runs inside
// this transaction (see Notebook.IndexWithCallback), and linkMatchesPath reads
// notebookPath to resolve a link relative to its source note. Dropping the
// field left the transaction resolving links against an empty notebook path.
func TestNoteIndexCommitPreservesFields(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	index := NewNoteIndex("/notebook", db, &util.NullLogger, "markdown")

	err := index.Commit(func(idx core.NoteIndex) error {
		inner, ok := idx.(*NoteIndex)
		assert.True(t, ok)
		assert.Equal(t, inner.notebookPath, "/notebook")
		assert.Equal(t, inner.extension, "markdown")
		return nil
	})
	assert.Nil(t, err)
}
