package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

// TaggedNoteDAO persists tag-note associations to the SQLite database.

type TaggedNoteDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	createAssociationStmt  *LazyStmt
	removeAssociationsStmt *LazyStmt
	generalQuery           string
	filteredQuery          string
}

func NewTaggedNoteDAO(tx Transaction, logger util.Logger) *TaggedNoteDAO {
	return &TaggedNoteDAO{
		tx:     tx,
		logger: logger,

		createAssociationStmt: tx.PrepareLazy(`
			INSERT INTO tags (tag_id, note_id, pos)
			VALUES (?, ?, ?)
		`),

		removeAssociationsStmt: tx.PrepareLazy(`
			DELETE FROM tags
			WHERE note_id = ?
		`),

		generalQuery: `
				SELECT c.name, n.title, n.path, GROUP_CONCAT(t.pos)
				FROM collections c
				JOIN tags t
				ON c.id = tag_id
				JOIN notes n
				ON t.note_id = n.id
				WHERE c.kind = '` + string(core.CollectionKindTag) + `'
				GROUP BY c.name, n.title, n.path
		`,

		filteredQuery: `
				SELECT c.name, n.title, n.path, GROUP_CONCAT(t.pos)
				FROM collections c
				JOIN tags t
				ON c.id = tag_id
				JOIN notes n
				ON t.note_id = n.id
				WHERE c.kind =  '` + string(core.CollectionKindTag) + `'
				AND c.name = ?
				GROUP BY c.name, n.title, n.path
		`,
	}
}

func parsePositions(positionsCat string) []int {
	positionsCatTrimmed := strings.TrimSpace(positionsCat)
	positions := strings.Split(positionsCatTrimmed, ",")

	result := make([]int, 0, len(positions))

	for _, position := range positions {
		var num int
		_, err := fmt.Sscanf(position, "%d", &num)
		if err == nil {
			result = append(result, num)
		}
	}
	return result
}

func (d *TaggedNoteDAO) scanTaggedNotes(rows *sql.Rows) ([]core.TaggedNoteDetail, error) {
	tagDetails := []core.TaggedNoteDetail{}
	for rows.Next() {
		var tagName string
		var title string
		var path string
		var positions string
		err := rows.Scan(&tagName, &title, &path, &positions)
		if err != nil {
			return tagDetails, err
		}

		tagDetails = append(tagDetails, core.TaggedNoteDetail{
			Tag:       tagName,
			Title:     title,
			Path:      path,
			Positions: parsePositions(positions),
		})
	}
	return tagDetails, nil
}

func (d *TaggedNoteDAO) buildQuery(query string, sorters []core.TaggedNoteSorter) string {
	orderTerms := []string{}
	for _, sorter := range sorters {
		orderTerms = append(orderTerms, taggedNoteOrderTerm(sorter))
	}

	if len(orderTerms) > 0 {
		query += "ORDER BY " + strings.Join(orderTerms, ", ") + "\n"
	}
	return query
}

func (d *TaggedNoteDAO) FindTaggedNotes(tagName string, sorters []core.TaggedNoteSorter) ([]core.TaggedNoteDetail, error) {
	query := d.filteredQuery
	query = d.buildQuery(query, sorters)
	rows, err := d.tx.Query(query, tagName)
	if err != nil {
		return []core.TaggedNoteDetail{}, err
	}
	defer rows.Close()
	return d.scanTaggedNotes(rows)
}

func (d *TaggedNoteDAO) FindTaggedNotesAll(sorters []core.TaggedNoteSorter) ([]core.TaggedNoteDetail, error) {
	query := d.generalQuery
	query = d.buildQuery(query, sorters)
	rows, err := d.tx.Query(query)
	if err != nil {
		return []core.TaggedNoteDetail{}, err
	}
	defer rows.Close()
	return d.scanTaggedNotes(rows)
}

func (d *TaggedNoteDAO) RemoveAssociations(noteID core.NoteID) error {
	if !noteID.IsValid() {
		return fmt.Errorf("note ID (%d) not valid", noteID)
	}

	_, err := d.removeAssociationsStmt.Exec(noteID)
	if err != nil {
		return fmt.Errorf("failed to remove tag associations of note %d: %w", noteID, err)
	}

	return nil
}

func (d *TaggedNoteDAO) CreateAssociation(noteID core.NoteID, tagID core.CollectionID, pos int) error {
	if !noteID.IsValid() || !tagID.IsValid() {
		return fmt.Errorf("note ID (%d) or tag ID (%d) not valid", noteID, tagID)
	}

	_, err := d.createAssociationStmt.Exec(tagID, noteID, pos)
	if err != nil {
		return fmt.Errorf("failed to create tag associations for note %d: %w", noteID, err)
	}

	return nil
}

func taggedNoteOrderTerm(sorter core.TaggedNoteSorter) string {
	order := " ASC"
	if !sorter.Ascending {
		order = " DESC"
	}

	switch sorter.Field {
	case core.TagSortName:
		return "c.name COLLATE NOCASE" + order
	case core.TagSortTitle:
		return "n.title COLLATE NOCASE" + order
	default:
		panic(fmt.Sprintf("%v: unknown core.TaggedNoteSortField", sorter.Field))
	}
}
