package sqlite

import (
	"fmt"
	"testing"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func testLinkDAO(t *testing.T, callback func(tx Transaction, dao *LinkDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewLinkDAO(tx, &util.NullLogger))
	})
}

type linkRow struct {
	SourceId                         core.NoteID
	TargetId                         *core.NoteID
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
		var sourceId int64
		var targetId *int64
		err = rows.Scan(&sourceId, &targetId, &row.Title, &row.Href, &row.Type, &row.IsExternal, &row.Rels, &row.Snippet, &row.SnippetStart, &row.SnippetEnd)
		assert.Nil(t, err)
		row.SourceId = core.NoteID(sourceId)
		if targetId != nil {
			row.TargetId = idPointer(*targetId)
		}
		links = append(links, row)
	}
	rows.Close()
	assert.Nil(t, rows.Err())

	return links
}
