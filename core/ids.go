package core

type NoteId int64

func (id NoteId) IsValid() bool {
	return id > 0
}

type CollectionId int64

func (id CollectionId) IsValid() bool {
	return id > 0
}
