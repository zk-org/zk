package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/fts5"
	"github.com/zk-org/zk/internal/util/paths"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// NoteDAO persists notes in the SQLite database.
type NoteDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	indexedStmt            *LazyStmt
	addStmt                *LazyStmt
	updateStmt             *LazyStmt
	removeStmt             *LazyStmt
	findIdByPathStmt       *LazyStmt
	findIdsByPathRegexStmt *LazyStmt
	findByIdStmt           *LazyStmt
}

// NewNoteDAO creates a new instance of a DAO working on the given database
// transaction.
func NewNoteDAO(tx Transaction, logger util.Logger) *NoteDAO {
	return &NoteDAO{
		tx:     tx,
		logger: logger,

		// Get file info about all indexed notes.
		indexedStmt: tx.PrepareLazy(`
			SELECT path, modified from notes
			 ORDER BY sortable_path ASC
		`),

		// Add a new note to the index.
		addStmt: tx.PrepareLazy(`
			INSERT INTO notes (path, sortable_path, title, lead, body, raw_content, word_count, metadata, checksum, created, modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`),

		// Update the content of a note.
		updateStmt: tx.PrepareLazy(`
			UPDATE notes
			   SET title = ?, lead = ?, body = ?, raw_content = ?, word_count = ?, metadata = ?, checksum = ?, modified = ?
			 WHERE path = ?
		`),

		// Remove a note.
		removeStmt: tx.PrepareLazy(`
			DELETE FROM notes
			 WHERE id = ?
		`),

		// Find a note ID from its exact path.
		findIdByPathStmt: tx.PrepareLazy(`
			SELECT id FROM notes
			 WHERE path = ?
		`),

		// Find note IDs from a regex matching their path.
		findIdsByPathRegexStmt: tx.PrepareLazy(`
			SELECT id FROM notes
			 WHERE path REGEXP ?
				-- To find the best match possible, we sort by path length.
				-- See https://github.com/zk-org/zk/issues/23
			 ORDER BY LENGTH(path) ASC
		`),

		// Find a note from its ID.
		findByIdStmt: tx.PrepareLazy(`
			SELECT id, path, title, lead, body, raw_content, word_count, created, modified, metadata, checksum, tags, lead AS snippet
			  FROM notes_with_metadata
			 WHERE id = ?
		`),
	}
}

// Indexed returns file info of all indexed notes.
func (d *NoteDAO) Indexed() (<-chan paths.Metadata, error) {
	rows, err := d.indexedStmt.Query()
	if err != nil {
		return nil, err
	}

	c := make(chan paths.Metadata)
	go func() {
		defer close(c)
		defer rows.Close()
		var (
			path     string
			modified time.Time
		)

		for rows.Next() {
			err := rows.Scan(&path, &modified)
			if err != nil {
				d.logger.Err(err)
			}

			c <- paths.Metadata{
				Path:     path,
				Modified: modified,
			}
		}

		err = rows.Err()
		if err != nil {
			d.logger.Err(err)
		}
	}()

	return c, nil
}

// Add inserts a new note to the index.
func (d *NoteDAO) Add(note core.Note) (core.NoteID, error) {
	// For sortable_path, we replace in path / by the shortest non printable
	// character available to make it sortable. Without this, sorting by the
	// path would be a lexicographical sort instead of being the same order
	// returned by filepath.Walk.
	// \x01 is used instead of \x00, because SQLite treats \x00 as and end of
	// string.
	sortablePath := strings.ReplaceAll(note.Path, "/", "\x01")

	metadata := d.metadataToJSON(note)
	res, err := d.addStmt.Exec(
		note.Path, sortablePath, note.Title, note.Lead, note.Body,
		note.RawContent, note.WordCount, metadata, note.Checksum, note.Created,
		note.Modified,
	)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return core.NoteID(lastId), err
}

// Update modifies an existing note.
func (d *NoteDAO) Update(note core.Note) (core.NoteID, error) {
	id, err := d.FindIdByPath(note.Path)
	if err != nil {
		return 0, err
	}
	if !id.IsValid() {
		return 0, errors.New("note not found in the index")
	}

	metadata := d.metadataToJSON(note)
	_, err = d.updateStmt.Exec(
		note.Title, note.Lead, note.Body, note.RawContent, note.WordCount,
		metadata, note.Checksum, note.Modified, note.Path,
	)
	return id, err
}

func (d *NoteDAO) metadataToJSON(note core.Note) string {
	json, err := json.Marshal(note.Metadata)
	if err != nil {
		// Failure to serialize the metadata to JSON should not prevent the
		// note from being saved.
		d.logger.Err(errors.Wrapf(err, "cannot serialize note metadata to JSON: %s", note.Path))
		return "{}"
	}
	return string(json)
}

// Remove deletes the note with the given path from the index.
func (d *NoteDAO) Remove(path string) error {
	id, err := d.FindIdByPath(path)
	if err != nil {
		return err
	}
	if !id.IsValid() {
		return errors.New("note not found in the index")
	}

	_, err = d.removeStmt.Exec(id)
	return err
}

func (d *NoteDAO) FindIdByPath(path string) (core.NoteID, error) {
	row, err := d.findIdByPathStmt.QueryRow(path)
	if err != nil {
		return core.NoteID(0), err
	}
	return idForRow(row)
}

func idForRow(row *sql.Row) (core.NoteID, error) {
	var id sql.NullInt64
	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return core.NoteID(id.Int64), nil
	}
}

func (d *NoteDAO) findIdsByPathRegex(regex string) ([]core.NoteID, error) {
	ids := []core.NoteID{}
	rows, err := d.findIdsByPathRegexStmt.Query(regex)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullInt64
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, core.NoteID(id.Int64))
	}

	return ids, nil
}

func (d *NoteDAO) findIdWithStmt(stmt *LazyStmt, args ...interface{}) (core.NoteID, error) {
	row, err := stmt.QueryRow(args...)
	if err != nil {
		return core.NoteID(0), err
	}

	var id sql.NullInt64
	err = row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return core.NoteID(id.Int64), nil
	}
}

func (d *NoteDAO) FindIdByHref(href string, allowPartialHref bool) (core.NoteID, error) {
	ids, err := d.FindIdsByHref(href, allowPartialHref)
	if len(ids) == 0 || err != nil {
		return 0, err
	}
	return ids[0], nil
}

func (d *NoteDAO) findIdsByHrefs(hrefs []string, allowPartialHrefs bool) ([]core.NoteID, error) {
	ids := make([]core.NoteID, 0)
	for _, href := range hrefs {
		cids, err := d.FindIdsByHref(href, allowPartialHrefs)
		if err != nil {
			return ids, err
		}
		ids = append(ids, cids...)
	}
	return ids, nil
}

// FIXME: This logic is duplicated in NoteIndex.linkMatchesPath(). Maybe there's a way to share it using a custom SQLite function?
func (d *NoteDAO) FindIdsByHref(href string, allowPartialHref bool) ([]core.NoteID, error) {
	// Remove any anchor at the end of the HREF, since it's most likely
	// matching a sub-section in the note.
	href = strings.SplitN(href, "#", 2)[0]

	href = regexp.QuoteMeta(href)

	if allowPartialHref {
		ids, err := d.findIdsByPathRegex("^(.*/)?[^/]*" + href + "[^/]*$")
		if len(ids) > 0 || err != nil {
			return ids, err
		}

		ids, err = d.findIdsByPathRegex(".*" + href + ".*")
		if len(ids) > 0 || err != nil {
			return ids, err
		}
	}

	ids, err := d.findIdsByPathRegex("^(?:" + href + "[^/]*|" + href + "/.+)$")
	if len(ids) > 0 || err != nil {
		return ids, err
	}

	return []core.NoteID{}, nil
}

func (d *NoteDAO) FindMinimal(opts core.NoteFindOpts) ([]core.MinimalNote, error) {
	notes := make([]core.MinimalNote, 0)

	opts, err := d.expandMentionsIntoMatch(opts)
	if err != nil {
		return notes, err
	}

	rows, err := d.findRows(opts, noteSelectionMinimal)
	if err != nil {
		return notes, err
	}
	defer rows.Close()

	for rows.Next() {
		note, err := d.scanMinimalNote(rows)
		if err != nil {
			d.logger.Err(err)
			continue
		}
		if note != nil {
			notes = append(notes, *note)
		}
	}

	return notes, nil
}

// Find returns all the notes matching the given criteria.
func (d *NoteDAO) Find(opts core.NoteFindOpts) ([]core.ContextualNote, error) {
	notes := make([]core.ContextualNote, 0)

	opts, err := d.expandMentionsIntoMatch(opts)
	if err != nil {
		return notes, err
	}

	rows, err := d.findRows(opts, noteSelectionFull)
	if err != nil {
		return notes, err
	}
	defer rows.Close()

	for rows.Next() {
		note, err := d.scanNote(rows)
		if err != nil {
			d.logger.Err(err)
			continue
		}
		if note != nil {
			notes = append(notes, *note)
		}
	}

	return notes, nil
}

// parseListFromNullString splits a 0-separated string.
func parseListFromNullString(str sql.NullString) []string {
	list := []string{}
	if str.Valid && str.String != "" {
		list = strings.Split(str.String, "\x01")
		list = strutil.RemoveDuplicates(list)
		list = strutil.RemoveBlank(list)
	}
	return list
}

// expandMentionsIntoMatch finds the titles associated with the notes in opts.Mention to
// expand them into the opts.Match predicate.
func (d *NoteDAO) expandMentionsIntoMatch(opts core.NoteFindOpts) (core.NoteFindOpts, error) {
	if opts.Mention == nil {
		return opts, nil
	}
	if opts.MatchStrategy != core.MatchStrategyFts {
		return opts, fmt.Errorf("--mention can only be used with --match-strategy=fts")
	}

	// Find the IDs for the mentioned paths.
	ids, err := d.findIdsByHrefs(opts.Mention, true /* allowPartialHrefs */)
	if err != nil {
		return opts, err
	}
	if len(ids) == 0 {
		return opts, fmt.Errorf("could not find notes at: " + strings.Join(opts.Mention, ", "))
	}

	// Exclude the mentioned notes from the results.
	opts = opts.ExcludingIDs(ids)

	// Find their titles.
	titlesQuery := "SELECT title, metadata FROM notes WHERE id IN (" + joinNoteIDs(ids, ",") + ")"
	rows, err := d.tx.Query(titlesQuery)
	if err != nil {
		return opts, err
	}
	defer rows.Close()

	mentionQueries := []string{}

	for rows.Next() {
		var title, metadataJSON string
		err := rows.Scan(&title, &metadataJSON)
		if err != nil {
			return opts, err
		}

		mentionQueries = append(mentionQueries, buildMentionQuery(title, metadataJSON))
	}

	if len(mentionQueries) == 0 {
		return opts, nil
	}

	// Expand the mention queries in the match predicate.
	opts.Match = append(opts.Match, " ("+strings.Join(mentionQueries, " OR ")+") ")

	return opts, nil
}

// noteSelection represents the amount of column selected with findRows.
type noteSelection int

const (
	noteSelectionID noteSelection = iota + 1
	noteSelectionMinimal
	noteSelectionFull
)

func (d *NoteDAO) findRows(opts core.NoteFindOpts, selection noteSelection) (*sql.Rows, error) {
	snippetCol := `n.lead`
	joinClauses := []string{}
	whereExprs := []string{}
	additionalOrderTerms := []string{}
	args := []interface{}{}
	groupBy := ""

	transitiveClosure := false
	maxDistance := 0

	setupLinkFilter := func(tableAlias string, hrefs []string, direction int, negate, recursive bool) error {
		ids, err := d.findIdsByHrefs(hrefs, true /* allowPartialHrefs */)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return fmt.Errorf("could not find notes at: " + strings.Join(hrefs, ", "))
		}
		idsList := "(" + joinNoteIDs(ids, ",") + ")"

		linksSrc := "links"

		if recursive {
			transitiveClosure = true
			linksSrc = "transitive_closure"
			additionalOrderTerms = append(additionalOrderTerms, tableAlias+".distance")
		}

		if !negate {
			if direction != 0 {
				snippetCol = fmt.Sprintf("GROUP_CONCAT(REPLACE(%s.snippet, %[1]s.title, '<zk:match>' || %[1]s.title || '</zk:match>'), '\x01')", tableAlias)
			}

			joinOns := make([]string, 0)
			if direction <= 0 {
				joinOns = append(joinOns, fmt.Sprintf(
					"(n.id = %[1]s.target_id AND %[1]s.source_id IN %[2]s)", tableAlias, idsList,
				))
			}
			if direction >= 0 {
				joinOns = append(joinOns, fmt.Sprintf(
					"(n.id = %[1]s.source_id AND %[1]s.target_id IN %[2]s)", tableAlias, idsList,
				))
			}

			joinClauses = append(joinClauses, fmt.Sprintf(
				"LEFT JOIN %[2]s %[1]s ON %[3]s",
				tableAlias,
				linksSrc,
				strings.Join(joinOns, " OR "),
			))

			groupBy = "GROUP BY n.id"
		}

		idExpr := "n.id"
		if negate {
			idExpr += " NOT"
		}

		idSelects := make([]string, 0)
		if direction <= 0 {
			idSelects = append(idSelects, fmt.Sprintf(
				"    SELECT target_id FROM %s WHERE target_id IS NOT NULL AND source_id IN %s",
				linksSrc, idsList,
			))
		}
		if direction >= 0 {
			idSelects = append(idSelects, fmt.Sprintf(
				"    SELECT source_id FROM %s WHERE target_id IS NOT NULL AND target_id IN %s",
				linksSrc, idsList,
			))
		}

		idExpr += " IN (\n" + strings.Join(idSelects, "\n    UNION\n") + "\n)"

		whereExprs = append(whereExprs, idExpr)

		return nil
	}

	if 0 < len(opts.Match) {
		switch opts.MatchStrategy {
		case core.MatchStrategyExact:
			for _, match := range opts.Match {
				whereExprs = append(whereExprs, `n.raw_content LIKE '%' || ? || '%' ESCAPE '\'`)
				args = append(args, escapeLikeTerm(match, '\\'))
			}
		case core.MatchStrategyFts:
			snippetCol = `snippet(fts_match.notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
			joinClauses = append(joinClauses, "JOIN notes_fts fts_match ON n.id = fts_match.rowid")
			additionalOrderTerms = append(additionalOrderTerms, `bm25(fts_match.notes_fts, 1000.0, 500.0, 1.0)`)
			for _, match := range opts.Match {
				whereExprs = append(whereExprs, "fts_match.notes_fts MATCH ?")
				args = append(args, fts5.ConvertQuery(match))
			}
		case core.MatchStrategyRe:
			for _, match := range opts.Match {
				whereExprs = append(whereExprs, "n.raw_content REGEXP ?")
				args = append(args, match)
			}
			break
		}
	}

	if opts.IncludeHrefs != nil {
		ids, err := d.findIdsByHrefs(opts.IncludeHrefs, opts.AllowPartialHrefs)
		if err != nil {
			return nil, err
		}
		opts = opts.IncludingIDs(ids)
	}

	if opts.ExcludeHrefs != nil {
		ids, err := d.findIdsByHrefs(opts.ExcludeHrefs, opts.AllowPartialHrefs)
		if err != nil {
			return nil, err
		}
		opts = opts.ExcludingIDs(ids)
	}

	if opts.Tags != nil {
		separatorRegex := regexp.MustCompile(`(\ OR\ )|\|`)
		for _, tagsArg := range opts.Tags {
			tags := separatorRegex.Split(tagsArg, -1)

			negate := false
			globs := make([]string, 0)
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)

				if strings.HasPrefix(tag, "-") {
					negate = true
					tag = strings.TrimPrefix(tag, "-")
				} else if strings.HasPrefix(tag, "NOT") {
					negate = true
					tag = strings.TrimPrefix(tag, "NOT")
				}

				tag = strings.TrimSpace(tag)
				if len(tag) == 0 {
					continue
				}
				globs = append(globs, "t.name GLOB ?")
				args = append(args, tag)
			}

			if len(globs) == 0 {
				continue
			}
			if negate && len(globs) > 1 {
				return nil, fmt.Errorf("cannot negate a tag in a OR group: %s", tagsArg)
			}

			expr := "n.id"
			if negate {
				expr += " NOT"
			}
			expr += fmt.Sprintf(` IN (
SELECT note_id FROM notes_collections
WHERE collection_id IN (SELECT id FROM collections t WHERE kind = '%s' AND (%s))
)`,
				core.CollectionKindTag,
				strings.Join(globs, " OR "),
			)
			whereExprs = append(whereExprs, expr)
		}
	}

	if opts.MentionedBy != nil {
		ids, err := d.findIdsByHrefs(opts.MentionedBy, true /* allowPartialHrefs */)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("could not find notes at: " + strings.Join(opts.MentionedBy, ", "))
		}

		// Exclude the mentioning notes from the results.
		opts = opts.ExcludingIDs(ids)

		snippetCol = `snippet(nsrc.notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
		joinClauses = append(joinClauses, "JOIN notes_fts nsrc ON nsrc.rowid IN ("+joinNoteIDs(ids, ",")+") AND nsrc.notes_fts MATCH mention_query(n.title, n.metadata)")
	}

	if opts.LinkedBy != nil {
		filter := opts.LinkedBy
		maxDistance = filter.MaxDistance
		err := setupLinkFilter("l_by", filter.Hrefs, -1, filter.Negate, filter.Recursive)
		if err != nil {
			return nil, err
		}
	}

	if opts.LinkTo != nil {
		filter := opts.LinkTo
		maxDistance = filter.MaxDistance
		err := setupLinkFilter("l_to", filter.Hrefs, 1, filter.Negate, filter.Recursive)
		if err != nil {
			return nil, err
		}
	}

	if opts.Related != nil {
		maxDistance = 2
		err := setupLinkFilter("l_rel", opts.Related, 0, false, true)
		if err != nil {
			return nil, err
		}
		groupBy += " HAVING MIN(l_rel.distance) = 2"
	}

	if opts.Orphan {
		whereExprs = append(whereExprs, `n.id NOT IN (
			SELECT target_id FROM links WHERE target_id IS NOT NULL
		)`)
	}

	if opts.Tagless {
		whereExprs = append(whereExprs, `tags IS NULL`)
	}

	if opts.CreatedStart != nil {
		whereExprs = append(whereExprs, "created >= ?")
		args = append(args, opts.CreatedStart)
	}

	if opts.CreatedEnd != nil {
		whereExprs = append(whereExprs, "created < ?")
		args = append(args, opts.CreatedEnd)
	}

	if opts.ModifiedStart != nil {
		whereExprs = append(whereExprs, "modified >= ?")
		args = append(args, opts.ModifiedStart)
	}

	if opts.ModifiedEnd != nil {
		whereExprs = append(whereExprs, "modified < ?")
		args = append(args, opts.ModifiedEnd)
	}

	if opts.IncludeIDs != nil {
		whereExprs = append(whereExprs, "n.id IN ("+joinNoteIDs(opts.IncludeIDs, ",")+")")
	}

	if opts.ExcludeIDs != nil {
		whereExprs = append(whereExprs, "n.id NOT IN ("+joinNoteIDs(opts.ExcludeIDs, ",")+")")
	}

	orderTerms := []string{}
	for _, sorter := range opts.Sorters {
		orderTerms = append(orderTerms, orderTerm(sorter))
	}
	orderTerms = append(orderTerms, additionalOrderTerms...)
	orderTerms = append(orderTerms, `n.title ASC`)

	query := ""

	// Credit to https://inviqa.com/blog/storing-graphs-database-sql-meets-social-network
	if transitiveClosure {
		query += `WITH RECURSIVE transitive_closure(source_id, target_id, title, snippet, distance, path) AS (
    SELECT source_id, target_id, title, snippet,
           1 AS distance,
           '.' || source_id || '.' || target_id || '.' AS path
      FROM links
 
     UNION ALL
 
    SELECT tc.source_id, l.target_id, l.title, l.snippet,
           tc.distance + 1,
           tc.path || l.target_id || '.' AS path
      FROM links AS l
      JOIN transitive_closure AS tc
        ON l.source_id = tc.target_id
     WHERE tc.path NOT LIKE '%.' || l.target_id || '.%'`

		if maxDistance != 0 {
			query += fmt.Sprintf(" AND tc.distance < %d", maxDistance)
		}

		// Guard against infinite loops by limiting the number of recursions.
		query += "\n     LIMIT 100000"

		query += "\n)\n"
	}

	query += "SELECT n.id"
	if selection != noteSelectionID {
		query += ", n.path, n.title, n.metadata"
		if selection != noteSelectionMinimal {
			query += fmt.Sprintf(", n.lead, n.body, n.raw_content, n.word_count, n.created, n.modified, n.checksum, n.tags, %s AS snippet", snippetCol)
		}
	}

	query += "\nFROM notes_with_metadata n\n"

	for _, clause := range joinClauses {
		query += clause + "\n"
	}

	if len(whereExprs) > 0 {
		query += "WHERE " + strings.Join(whereExprs, "\nAND ") + "\n"
	}

	if groupBy != "" {
		query += groupBy + "\n"
	}

	query += "ORDER BY " + strings.Join(orderTerms, ", ") + "\n"

	if opts.Limit > 0 {
		query += fmt.Sprintf("LIMIT %d\n", opts.Limit)
	}

	// d.logger.Println(query)
	// d.logger.Println(args)

	return d.tx.Query(query, args...)
}

func (d *NoteDAO) scanNoteID(row RowScanner) (core.NoteID, error) {
	var id int
	err := row.Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return core.NoteID(id), nil
	}
}

func (d *NoteDAO) scanMinimalNote(row RowScanner) (*core.MinimalNote, error) {
	var (
		id                        int
		path, title, metadataJSON string
	)

	err := row.Scan(&id, &path, &title, &metadataJSON)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		metadata, err := unmarshalMetadata(metadataJSON)
		if err != nil {
			d.logger.Err(errors.Wrap(err, path))
		}

		return &core.MinimalNote{
			ID:       core.NoteID(id),
			Path:     path,
			Title:    title,
			Metadata: metadata,
		}, nil
	}
}

func (d *NoteDAO) scanNote(row RowScanner) (*core.ContextualNote, error) {
	var (
		id, wordCount                 int
		title, lead, body, rawContent string
		snippets, tags                sql.NullString
		path, metadataJSON, checksum  string
		created, modified             time.Time
	)

	err := row.Scan(
		&id, &path, &title, &metadataJSON, &lead, &body, &rawContent,
		&wordCount, &created, &modified, &checksum, &tags, &snippets,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		metadata, err := unmarshalMetadata(metadataJSON)
		if err != nil {
			d.logger.Err(errors.Wrap(err, path))
		}

		return &core.ContextualNote{
			Snippets: parseListFromNullString(snippets),
			Note: core.Note{
				ID:         core.NoteID(id),
				Path:       path,
				Title:      title,
				Lead:       lead,
				Body:       body,
				RawContent: rawContent,
				WordCount:  wordCount,
				Links:      []core.Link{},
				Tags:       parseListFromNullString(tags),
				Metadata:   metadata,
				Created:    created,
				Modified:   modified,
				Checksum:   checksum,
			},
		}, nil
	}
}

func orderTerm(sorter core.NoteSorter) string {
	order := " ASC"
	if !sorter.Ascending {
		order = " DESC"
	}

	switch sorter.Field {
	case core.NoteSortCreated:
		return "n.created" + order
	case core.NoteSortModified:
		return "n.modified" + order
	case core.NoteSortPath:
		return "n.path" + order
	case core.NoteSortRandom:
		return "RANDOM()"
	case core.NoteSortTitle:
		return "n.title" + order
	case core.NoteSortWordCount:
		return "n.word_count" + order
	default:
		panic(fmt.Sprintf("%v: unknown core.NoteSortField", sorter.Field))
	}
}

// buildMentionQuery creates an FTS5 predicate to match the given note's title
// (or aliases from the metadata) in the content of another note.
//
// It is exposed as a custom SQLite function as `mention_query()`.
func buildMentionQuery(title, metadataJSON string) string {
	titles := []string{}

	appendTitle := func(t string) {
		t = strings.TrimSpace(t)
		if t != "" {
			// Remove double quotes in the title to avoid tripping the FTS5 parser.
			titles = append(titles, `"`+strings.ReplaceAll(t, `"`, "")+`"`)
		}
	}

	appendTitle(title)

	// Support `aliases` key in the YAML frontmatter, like Obsidian:
	// https://publish.obsidian.md/help/How+to/Add+aliases+to+note
	metadata, err := unmarshalMetadata(metadataJSON)
	if err == nil {
		if aliases, ok := metadata["aliases"]; ok {
			switch aliases := aliases.(type) {
			case []interface{}:
				for _, alias := range aliases {
					appendTitle(fmt.Sprint(alias))
				}
			case string:
				appendTitle(aliases)
			}
		}
	}

	if len(titles) == 0 {
		// Return an arbitrary search term otherwise MATCH will find every note.
		// Not proud of this hack but it does the job.
		return "8b80252291ee418289cfc9968eb2961c"
	}

	return "(" + strings.Join(titles, " OR ") + ")"
}
