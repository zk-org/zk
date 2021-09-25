package core

type noteContentParserMock struct {
	results map[string]*NoteContent
}

func newNoteContentParserMock(results map[string]*NoteContent) *noteContentParserMock {
	return &noteContentParserMock{
		results: results,
	}
}

func (p *noteContentParserMock) ParseNoteContent(content string) (*NoteContent, error) {
	if note, ok := p.results[content]; ok {
		return note, nil
	}
	return &NoteContent{}, nil
}
