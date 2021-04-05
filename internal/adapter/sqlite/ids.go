package sqlite

import "github.com/mickael-menu/zk/internal/core"

type SQLNoteID int64

// IsValid implements core.NoteID.
func (id SQLNoteID) IsValid() bool {
	return id > 0
}

type SQLCollectionID int64

// IsValid implements core.CollectionID.
func (id SQLCollectionID) IsValid() bool {
	return id > 0
}

type SQLNoteCollectionID int64

// IsValid implements core.NoteCollectionID.
func (id SQLNoteCollectionID) IsValid() bool {
	return id > 0
}

func SQLNoteIDsFromCoreIDs(ids []core.NoteID) []SQLNoteID {
	res := make([]SQLNoteID, len(ids))
	for i := range ids {
		res[i] = ids[i].(SQLNoteID)
	}
	return res
}
