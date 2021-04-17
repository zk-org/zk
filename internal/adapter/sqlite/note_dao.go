package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/fts5"
	"github.com/mickael-menu/zk/internal/util/icu"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/paths"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
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
	findIdByPathPrefixStmt *LazyStmt
	findByIdStmt           *LazyStmt
	addLinkStmt            *LazyStmt
	setLinksTargetStmt     *LazyStmt
	removeLinksStmt        *LazyStmt
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

		// Find a note ID from a prefix of its path.
		findIdByPathPrefixStmt: tx.PrepareLazy(`
			SELECT id FROM notes
			 WHERE path LIKE ? || '%'
			 ORDER BY LENGTH(path) ASC
		`),

		// Find a note from its ID.
		findByIdStmt: tx.PrepareLazy(`
			SELECT id, path, title, lead, body, raw_content, word_count, created, modified, metadata, checksum, tags, lead AS snippet
			  FROM notes_with_metadata
			 WHERE id = ?
		`),

		// Add a new link.
		addLinkStmt: tx.PrepareLazy(`
			INSERT INTO links (source_id, target_id, title, href, external, rels, snippet, snippet_start, snippet_end)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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

	id := core.NoteID(lastId)
	err = d.addLinks(id, note)
	return id, err
}

// Update modifies an existing note.
func (d *NoteDAO) Update(note core.Note) (core.NoteID, error) {
	id, err := d.findIdByPath(note.Path)
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
	if err != nil {
		return id, err
	}

	_, err = d.removeLinksStmt.Exec(d.idToSql(id))
	if err != nil {
		return id, err
	}

	err = d.addLinks(id, note)
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

// addLinks inserts all the outbound links of the given note.
func (d *NoteDAO) addLinks(id core.NoteID, note core.Note) error {
	for _, link := range note.Links {
		targetId, err := d.findIdByPathPrefix(link.Href)
		if err != nil {
			return err
		}

		_, err = d.addLinkStmt.Exec(id, d.idToSql(targetId), link.Title, link.Href, link.IsExternal, joinLinkRels(link.Rels), link.Snippet, link.SnippetStart, link.SnippetEnd)
		if err != nil {
			return err
		}
	}

	_, err := d.setLinksTargetStmt.Exec(int64(id), note.Path)
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

// Remove deletes the note with the given path from the index.
func (d *NoteDAO) Remove(path string) error {
	id, err := d.findIdByPath(path)
	if err != nil {
		return err
	}
	if !id.IsValid() {
		return errors.New("note not found in the index")
	}

	_, err = d.removeStmt.Exec(id)
	return err
}

func (d *NoteDAO) findIdByPath(path string) (core.NoteID, error) {
	row, err := d.findIdByPathStmt.QueryRow(path)
	if err != nil {
		return core.NoteID(0), err
	}
	return idForRow(row)
}

func (d *NoteDAO) findIdsByPathPrefixes(paths []string) ([]core.NoteID, error) {
	ids := make([]core.NoteID, 0)
	for _, path := range paths {
		id, err := d.findIdByPathPrefix(path)
		if err != nil {
			return ids, err
		}
		if id.IsValid() {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return ids, fmt.Errorf("could not find notes at: " + strings.Join(paths, ", "))
	}
	return ids, nil
}

func (d *NoteDAO) findIdByPathPrefix(path string) (core.NoteID, error) {
	row, err := d.findIdByPathPrefixStmt.QueryRow(path)
	if err != nil {
		return 0, err
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

func (d *NoteDAO) FindMinimal(opts core.NoteFindOpts) ([]core.MinimalNote, error) {
	notes := make([]core.MinimalNote, 0)

	opts, err := d.expandMentionsIntoMatch(opts)
	if err != nil {
		return notes, err
	}

	rows, err := d.findRows(opts, true)
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

func (d *NoteDAO) scanMinimalNote(row RowScanner) (*core.MinimalNote, error) {
	var (
		id          int
		path, title string
	)

	err := row.Scan(&id, &path, &title)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &core.MinimalNote{
			ID:    core.NoteID(id),
			Path:  path,
			Title: title,
		}, nil
	}
}

// Find returns all the notes matching the given criteria.
func (d *NoteDAO) Find(opts core.NoteFindOpts) ([]core.ContextualNote, error) {
	notes := make([]core.ContextualNote, 0)

	opts, err := d.expandMentionsIntoMatch(opts)
	if err != nil {
		return notes, err
	}

	rows, err := d.findRows(opts, false)
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

func (d *NoteDAO) scanNote(row RowScanner) (*core.ContextualNote, error) {
	var (
		id, wordCount                 int
		title, lead, body, rawContent string
		snippets, tags                sql.NullString
		path, metadataJSON, checksum  string
		created, modified             time.Time
	)

	err := row.Scan(
		&id, &path, &title, &lead, &body, &rawContent, &wordCount,
		&created, &modified, &metadataJSON, &checksum, &tags, &snippets,
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

// parseListFromNullString splits a 0-separated string.
func parseListFromNullString(str sql.NullString) []string {
	list := []string{}
	if str.Valid && str.String != "" {
		list = strings.Split(str.String, "\x01")
		list = strutil.RemoveDuplicates(list)
	}
	return list
}

// expandMentionsIntoMatch finds the titles associated with the notes in opts.Mention to
// expand them into the opts.Match predicate.
func (d *NoteDAO) expandMentionsIntoMatch(opts core.NoteFindOpts) (core.NoteFindOpts, error) {
	if opts.Mention == nil {
		return opts, nil
	}
	if opts.ExactMatch {
		return opts, fmt.Errorf("--exact-match and --mention cannot be used together")
	}

	// Find the IDs for the mentioned paths.
	ids, err := d.findIdsByPathPrefixes(opts.Mention)
	if err != nil {
		return opts, err
	}

	// Exclude the mentioned notes from the results.
	for _, id := range ids {
		opts = opts.ExcludingID(id)
	}

	// Find their titles.
	titlesQuery := "SELECT title, metadata FROM notes WHERE id IN (" + d.joinIds(ids, ",") + ")"
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
	match := opts.Match.String()
	match += " " + strings.Join(mentionQueries, " OR ")
	opts.Match = opt.NewString(match)

	return opts, nil
}

func (d *NoteDAO) findRows(opts core.NoteFindOpts, minimal bool) (*sql.Rows, error) {
	snippetCol := `n.lead`
	joinClauses := []string{}
	whereExprs := []string{}
	additionalOrderTerms := []string{}
	args := []interface{}{}
	groupBy := ""

	transitiveClosure := false
	maxDistance := 0

	setupLinkFilter := func(paths []string, direction int, negate, recursive bool) error {
		ids, err := d.findIdsByPathPrefixes(paths)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		idsList := "(" + d.joinIds(ids, ",") + ")"

		linksSrc := "links"

		if recursive {
			transitiveClosure = true
			linksSrc = "transitive_closure"
		}

		if !negate {
			if direction != 0 {
				snippetCol = "GROUP_CONCAT(REPLACE(l.snippet, l.title, '<zk:match>' || l.title || '</zk:match>'), '\x01')"
			}

			joinOns := make([]string, 0)
			if direction <= 0 {
				joinOns = append(joinOns, fmt.Sprintf(
					"(n.id = l.target_id AND l.source_id IN %s)", idsList,
				))
			}
			if direction >= 0 {
				joinOns = append(joinOns, fmt.Sprintf(
					"(n.id = l.source_id AND l.target_id IN %s)", idsList,
				))
			}

			joinClauses = append(joinClauses, fmt.Sprintf(
				"LEFT JOIN %s l ON %s",
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

	if !opts.Match.IsNull() {
		if opts.ExactMatch {
			whereExprs = append(whereExprs, `n.raw_content LIKE '%' || ? || '%' ESCAPE '\'`)
			args = append(args, escapeLikeTerm(opts.Match.String(), '\\'))
		} else {
			snippetCol = `snippet(fts_match.notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
			joinClauses = append(joinClauses, "JOIN notes_fts fts_match ON n.id = fts_match.rowid")
			additionalOrderTerms = append(additionalOrderTerms, `bm25(fts_match.notes_fts, 1000.0, 500.0, 1.0)`)
			whereExprs = append(whereExprs, "fts_match.notes_fts MATCH ?")
			args = append(args, fts5.ConvertQuery(opts.Match.String()))
		}
	}

	if opts.IncludePaths != nil {
		regexes := make([]string, 0)
		for _, path := range opts.IncludePaths {
			regexes = append(regexes, "n.path REGEXP ?")
			args = append(args, pathRegex(path))
		}
		whereExprs = append(whereExprs, strings.Join(regexes, " OR "))
	}

	if opts.ExcludePaths != nil {
		regexes := make([]string, 0)
		for _, path := range opts.ExcludePaths {
			regexes = append(regexes, "n.path NOT REGEXP ?")
			args = append(args, pathRegex(path))
		}
		whereExprs = append(whereExprs, strings.Join(regexes, " AND "))
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
		ids, err := d.findIdsByPathPrefixes(opts.MentionedBy)
		if err != nil {
			return nil, err
		}

		// Exclude the mentioning notes from the results.
		for _, id := range ids {
			opts = opts.ExcludingID(id)
		}

		snippetCol = `snippet(nsrc.notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
		joinClauses = append(joinClauses, "JOIN notes_fts nsrc ON nsrc.rowid IN ("+d.joinIds(ids, ",")+") AND nsrc.notes_fts MATCH mention_query(n.title, n.metadata)")
	}

	if opts.LinkedBy != nil {
		filter := opts.LinkedBy
		maxDistance = filter.MaxDistance
		err := setupLinkFilter(filter.Paths, -1, filter.Negate, filter.Recursive)
		if err != nil {
			return nil, err
		}
	}

	if opts.LinkTo != nil {
		filter := opts.LinkTo
		maxDistance = filter.MaxDistance
		err := setupLinkFilter(filter.Paths, 1, filter.Negate, filter.Recursive)
		if err != nil {
			return nil, err
		}
	}

	if opts.Related != nil {
		maxDistance = 2
		err := setupLinkFilter(opts.Related, 0, false, true)
		if err != nil {
			return nil, err
		}
		groupBy += " HAVING MIN(l.distance) = 2"
	}

	if opts.Orphan {
		whereExprs = append(whereExprs, `n.id NOT IN (
			SELECT target_id FROM links WHERE target_id IS NOT NULL
		)`)
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

	if opts.ExcludeIDs != nil {
		whereExprs = append(whereExprs, "n.id NOT IN ("+d.joinIds(opts.ExcludeIDs, ",")+")")
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
		orderTerms = append([]string{"l.distance"}, orderTerms...)

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

	query += "SELECT n.id, n.path, n.title"
	if !minimal {
		query += fmt.Sprintf(", n.lead, n.body, n.raw_content, n.word_count, n.created, n.modified, n.metadata, n.checksum, n.tags, %s AS snippet", snippetCol)
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
	case core.NoteSortPathLength:
		return "LENGTH(path)" + order
	default:
		panic(fmt.Sprintf("%v: unknown core.NoteSortField", sorter.Field))
	}
}

// pathRegex returns an ICU regex to match the files in the folder at given
// `path`, or any file having `path` for prefix.
func pathRegex(path string) string {
	path = icu.EscapePattern(path)
	return path + "[^/]*|" + path + "/.+"
}

func (d *NoteDAO) idToSql(id core.NoteID) sql.NullInt64 {
	if id.IsValid() {
		return sql.NullInt64{Int64: int64(id), Valid: true}
	} else {
		return sql.NullInt64{}
	}
}

func (d *NoteDAO) joinIds(ids []core.NoteID, delimiter string) string {
	strs := make([]string, 0)
	for _, i := range ids {
		strs = append(strs, strconv.FormatInt(int64(i), 10))
	}
	return strings.Join(strs, delimiter)
}

func unmarshalMetadata(metadataJSON string) (metadata map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(metadataJSON), &metadata)
	err = errors.Wrapf(err, "cannot parse note metadata from JSON: %s", metadataJSON)
	return
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

type RowScanner interface {
	Scan(dest ...interface{}) error
}
