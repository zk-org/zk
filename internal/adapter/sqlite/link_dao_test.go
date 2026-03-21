package sqlite

import (
	"fmt"
	"testing"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/test/assert"
)

type linkRow struct {
	SourceID                         core.NoteID
	TargetID                         *core.NoteID
	Href, Type, Title, Rels, Snippet string
	SnippetStart, SnippetEnd         int
	IsExternal                       bool
}

func queryLinkRows(t *testing.T, q RowQuerier, where string) []linkRow {
	links := make([]linkRow, 0)

	rows, err := q.Query(fmt.Sprintf(`
		SELECT source_id, target_id, title, href, type, external, rels, snippet, snippet_start, snippet_end
		  FROM links
		 WHERE %v
		 ORDER BY id
	`, where))
	assert.Nil(t, err)

	for rows.Next() {
		var row linkRow
		var sourceID int64
		var targetID *int64
		err = rows.Scan(&sourceID, &targetID, &row.Title, &row.Href, &row.Type, &row.IsExternal, &row.Rels, &row.Snippet, &row.SnippetStart, &row.SnippetEnd)
		assert.Nil(t, err)
		row.SourceID = core.NoteID(sourceID)
		if targetID != nil {
			row.TargetID = idPointer(*targetID)
		}
		links = append(links, row)
	}
	rows.Close()
	assert.Nil(t, rows.Err())

	return links
}
