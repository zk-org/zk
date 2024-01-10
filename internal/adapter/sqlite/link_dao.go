package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

// LinkDAO persists links in the SQLite database.
type LinkDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	addLinkStmt        *LazyStmt
	removeLinksStmt    *LazyStmt
	updateTargetIDStmt *LazyStmt
}

// NewLinkDAO creates a new instance of a DAO working on the given database
// transaction.
func NewLinkDAO(tx Transaction, logger util.Logger) *LinkDAO {
	return &LinkDAO{
		tx:     tx,
		logger: logger,

		// Add a new link.
		addLinkStmt: tx.PrepareLazy(`
			INSERT INTO links (source_id, target_id, title, href, type, external, rels, snippet, snippet_start, snippet_end)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`),

		// Remove all the outbound links of a note.
		removeLinksStmt: tx.PrepareLazy(`
			DELETE FROM links
			 WHERE source_id = ?
		`),

		updateTargetIDStmt: tx.PrepareLazy(`
			UPDATE links
			   SET target_id = ?
			 WHERE id = ?
		`),
	}
}

// Add inserts all the outbound links of the given note.
func (d *LinkDAO) Add(links []core.ResolvedLink) error {
	for _, link := range links {
		sourceID := noteIDToSQL(link.SourceID)
		targetID := noteIDToSQL(link.TargetID)

		_, err := d.addLinkStmt.Exec(sourceID, targetID, link.Title, link.Href, link.Type, link.IsExternal, joinLinkRels(link.Rels), link.Snippet, link.SnippetStart, link.SnippetEnd)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAll removes all the outbound links of the given note.
func (d *LinkDAO) RemoveAll(id core.NoteID) error {
	_, err := d.removeLinksStmt.Exec(noteIDToSQL(id))
	return err
}

// SetTargetID updates the target note of a link.
func (d *LinkDAO) SetTargetID(id core.LinkID, targetID core.NoteID) error {
	_, err := d.updateTargetIDStmt.Exec(noteIDToSQL(targetID), linkIDToSQL(id))
	return err
}

// joinLinkRels will concatenate a list of rels into a SQLite ready string.
// Each rel is delimited by \x01 for easy matching in queries.
func joinLinkRels(rels []core.LinkRelation) string {
	if len(rels) == 0 {
		return ""
	}
	delimiter := "\x01"
	res := delimiter
	for _, rel := range rels {
		res += string(rel) + delimiter
	}
	return res
}

// FindInternal returns all the links internal to the notebook.
func (d *LinkDAO) FindInternal() ([]core.ResolvedLink, error) {
	return d.findWhere("external = 0")
}

// FindBetweenNotes returns all the links existing between the given notes.
func (d *LinkDAO) FindBetweenNotes(ids []core.NoteID) ([]core.ResolvedLink, error) {
	idsString := joinNoteIDs(ids, ",")
	return d.findWhere(fmt.Sprintf("source_id IN (%s) AND target_id IN (%s)", idsString, idsString))
}

// findWhere returns all the links, filtered by the given where query.
func (d *LinkDAO) findWhere(where string) ([]core.ResolvedLink, error) {
	links := make([]core.ResolvedLink, 0)

	query := `
		SELECT id, source_id, source_path, target_id, target_path, title, href, type, external, rels, snippet, snippet_start, snippet_end
		  FROM resolved_links
	`

	if where != "" {
		query += "\nWHERE " + where
	}

	rows, err := d.tx.Query(query)
	if err != nil {
		return links, err
	}
	defer rows.Close()

	for rows.Next() {
		link, err := d.scanLink(rows)
		if err != nil {
			d.logger.Err(err)
			continue
		}
		if link != nil {
			links = append(links, *link)
		}
	}

	return links, nil
}

func (d *LinkDAO) scanLink(row RowScanner) (*core.ResolvedLink, error) {
	var (
		id, sourceID, snippetStart, snippetEnd     int
		targetID                                   sql.NullInt64
		sourcePath, title, href, linkType, snippet string
		external                                   bool
		targetPath, rels                           sql.NullString
	)

	err := row.Scan(
		&id, &sourceID, &sourcePath, &targetID, &targetPath, &title, &href,
		&linkType, &external, &rels, &snippet, &snippetStart, &snippetEnd,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &core.ResolvedLink{
			ID:         core.LinkID(id),
			SourceID:   core.NoteID(sourceID),
			SourcePath: sourcePath,
			TargetID:   core.NoteID(targetID.Int64),
			TargetPath: targetPath.String,
			Link: core.Link{
				Title:        title,
				Href:         href,
				Type:         core.LinkType(linkType),
				IsExternal:   external,
				Rels:         core.LinkRels(parseListFromNullString(rels)...),
				Snippet:      snippet,
				SnippetStart: snippetStart,
				SnippetEnd:   snippetEnd,
			},
		}, nil
	}
}
