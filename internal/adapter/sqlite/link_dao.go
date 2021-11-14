package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
)

// LinkDAO persists links in the SQLite database.
type LinkDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	addLinkStmt        *LazyStmt
	setLinksTargetStmt *LazyStmt
	removeLinksStmt    *LazyStmt
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

		// Set links matching a given href and missing a target ID to the given
		// target ID.
		setLinksTargetStmt: tx.PrepareLazy(`
			UPDATE links
			   SET target_id = ?
			 WHERE target_id IS NULL AND external = 0 AND ? LIKE href || '%'
		`),

		// Remove all the outbound links of a note.
		removeLinksStmt: tx.PrepareLazy(`
			DELETE FROM links
			 WHERE source_id = ?
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

// SetTargetID updates the missing target_id for links matching the given href.
// FIXME: Probably doesn't work for all type of href (partial, wikilinks, etc.)
func (d *LinkDAO) SetTargetID(href string, id core.NoteID) error {
	_, err := d.setLinksTargetStmt.Exec(int64(id), href)
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

func (d *LinkDAO) FindBetweenNotes(ids []core.NoteID) ([]core.ResolvedLink, error) {
	links := make([]core.ResolvedLink, 0)

	idsString := joinNoteIDs(ids, ",")
	rows, err := d.tx.Query(fmt.Sprintf(`
		SELECT id, source_id, source_path, target_id, target_path, title, href, type, external, rels, snippet, snippet_start, snippet_end
		  FROM resolved_links
		 WHERE source_id IN (%s) AND target_id IN (%s)
	`, idsString, idsString))
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
		id, sourceID, targetID, snippetStart, snippetEnd       int
		sourcePath, targetPath, title, href, linkType, snippet string
		external                                               bool
		rels                                                   sql.NullString
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
			SourceID:   core.NoteID(sourceID),
			SourcePath: sourcePath,
			TargetID:   core.NoteID(targetID),
			TargetPath: targetPath,
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
