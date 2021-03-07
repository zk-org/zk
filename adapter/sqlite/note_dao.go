package sqlite

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/fts5"
	"github.com/mickael-menu/zk/util/icu"
	"github.com/mickael-menu/zk/util/paths"
	strutil "github.com/mickael-menu/zk/util/strings"
)

// NoteDAO persists notes in the SQLite database.
// It implements the core port note.Finder.
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
			INSERT INTO notes (path, sortable_path, title, lead, body, raw_content, word_count, checksum, created, modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`),

		// Update the content of a note.
		updateStmt: tx.PrepareLazy(`
			UPDATE notes
			   SET title = ?, lead = ?, body = ?, raw_content = ?, word_count = ?, checksum = ?, modified = ?
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
		`),

		// Add a new link.
		addLinkStmt: tx.PrepareLazy(`
			INSERT INTO links (source_id, target_id, title, href, external, rels, snippet)
			VALUES (?, ?, ?, ?, ?, ?, ?)
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
func (d *NoteDAO) Add(note note.Metadata) (core.NoteId, error) {
	// For sortable_path, we replace in path / by the shortest non printable
	// character available to make it sortable. Without this, sorting by the
	// path would be a lexicographical sort instead of being the same order
	// returned by filepath.Walk.
	// \x01 is used instead of \x00, because SQLite treats \x00 as and end of
	// string.
	sortablePath := strings.ReplaceAll(note.Path, "/", "\x01")

	res, err := d.addStmt.Exec(
		note.Path, sortablePath, note.Title, note.Lead, note.Body, note.RawContent, note.WordCount, note.Checksum,
		note.Created, note.Modified,
	)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return core.NoteId(0), err
	}

	id := core.NoteId(lastId)
	err = d.addLinks(id, note)
	return id, err
}

// Update modifies an existing note.
func (d *NoteDAO) Update(note note.Metadata) (core.NoteId, error) {
	id, err := d.findIdByPath(note.Path)
	if err != nil {
		return 0, err
	}
	if !id.IsValid() {
		return 0, errors.New("note not found in the index")
	}

	_, err = d.updateStmt.Exec(
		note.Title, note.Lead, note.Body, note.RawContent, note.WordCount, note.Checksum, note.Modified,
		note.Path,
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

// addLinks inserts all the outbound links of the given note.
func (d *NoteDAO) addLinks(id core.NoteId, note note.Metadata) error {
	for _, link := range note.Links {
		targetId, err := d.findIdByPathPrefix(link.Href)
		if err != nil {
			return err
		}

		_, err = d.addLinkStmt.Exec(id, d.idToSql(targetId), link.Title, link.Href, link.External, joinLinkRels(link.Rels), link.Snippet)
		if err != nil {
			return err
		}
	}

	_, err := d.setLinksTargetStmt.Exec(int64(id), note.Path)
	return err
}

// joinLinkRels will concatenate a list of rels into a SQLite ready string.
// Each rel is delimited by \x01 for easy matching in queries.
func joinLinkRels(rels []string) string {
	if len(rels) == 0 {
		return ""
	}
	delimiter := "\x01"
	return delimiter + strings.Join(rels, delimiter) + delimiter
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

func (d *NoteDAO) findIdByPath(path string) (core.NoteId, error) {
	row, err := d.findIdByPathStmt.QueryRow(path)
	if err != nil {
		return core.NoteId(0), err
	}
	return idForRow(row)
}

func (d *NoteDAO) findIdsByPathPrefixes(paths []string) ([]core.NoteId, error) {
	ids := make([]core.NoteId, 0)
	for _, path := range paths {
		id, err := d.findIdByPathPrefix(path)
		if err != nil {
			return ids, err
		}
		if id.IsValid() {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (d *NoteDAO) findIdByPathPrefix(path string) (core.NoteId, error) {
	row, err := d.findIdByPathPrefixStmt.QueryRow(path)
	if err != nil {
		return core.NoteId(0), err
	}
	return idForRow(row)
}

func idForRow(row *sql.Row) (core.NoteId, error) {
	var id sql.NullInt64
	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return core.NoteId(0), nil
	case err != nil:
		return core.NoteId(0), err
	default:
		return core.NoteId(id.Int64), nil
	}
}

// Find returns all the notes matching the given criteria.
func (d *NoteDAO) Find(opts note.FinderOpts) ([]note.Match, error) {
	matches := make([]note.Match, 0)

	rows, err := d.findRows(opts)
	if err != nil {
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, wordCount                 int
			title, lead, body, rawContent string
			snippets, tags                sql.NullString
			path, checksum                string
			created, modified             time.Time
		)

		err := rows.Scan(&id, &path, &title, &lead, &body, &rawContent, &wordCount, &created, &modified, &checksum, &tags, &snippets)
		if err != nil {
			d.logger.Err(err)
			continue
		}

		matches = append(matches, note.Match{
			Snippets: parseListFromNullString(snippets),
			Metadata: note.Metadata{
				Path:       path,
				Title:      title,
				Lead:       lead,
				Body:       body,
				RawContent: rawContent,
				WordCount:  wordCount,
				Links:      []note.Link{},
				Tags:       parseListFromNullString(tags),
				Created:    created,
				Modified:   modified,
				Checksum:   checksum,
			},
		})
	}

	return matches, nil
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

func (d *NoteDAO) findRows(opts note.FinderOpts) (*sql.Rows, error) {
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

	for _, filter := range opts.Filters {
		switch filter := filter.(type) {

		case note.MatchFilter:
			snippetCol = `snippet(notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
			joinClauses = append(joinClauses, "JOIN notes_fts ON n.id = notes_fts.rowid")
			additionalOrderTerms = append(additionalOrderTerms, `bm25(notes_fts, 1000.0, 500.0, 1.0)`)
			whereExprs = append(whereExprs, "notes_fts MATCH ?")
			args = append(args, fts5.ConvertQuery(string(filter)))

		case note.PathFilter:
			if len(filter) == 0 {
				break
			}
			regexes := make([]string, 0)
			for _, path := range filter {
				regexes = append(regexes, "n.path REGEXP ?")
				args = append(args, pathRegex(path))
			}
			whereExprs = append(whereExprs, strings.Join(regexes, " OR "))

		case note.ExcludePathFilter:
			if len(filter) == 0 {
				break
			}
			regexes := make([]string, 0)
			for _, path := range filter {
				regexes = append(regexes, "n.path NOT REGEXP ?")
				args = append(args, pathRegex(path))
			}
			whereExprs = append(whereExprs, strings.Join(regexes, " AND "))

		case note.LinkedByFilter:
			maxDistance = filter.MaxDistance
			err := setupLinkFilter(filter.Paths, -1, filter.Negate, filter.Recursive)
			if err != nil {
				return nil, err
			}

		case note.LinkingToFilter:
			maxDistance = filter.MaxDistance
			err := setupLinkFilter(filter.Paths, 1, filter.Negate, filter.Recursive)
			if err != nil {
				return nil, err
			}

		case note.RelatedFilter:
			maxDistance = 2
			err := setupLinkFilter(filter, 0, false, true)
			if err != nil {
				return nil, err
			}
			groupBy += " HAVING MIN(l.distance) = 2"

		case note.OrphanFilter:
			whereExprs = append(whereExprs, `n.id NOT IN (
				SELECT target_id FROM links WHERE target_id IS NOT NULL
			)`)

		case note.DateFilter:
			value := "?"
			field := "n." + dateField(filter)
			op, ignoreTime := dateDirection(filter)
			if ignoreTime {
				field = "date(" + field + ")"
				value = "date(?)"
			}

			whereExprs = append(whereExprs, fmt.Sprintf("%s %s %s", field, op, value))
			args = append(args, filter.Date)

		case note.InteractiveFilter:
			// No user interaction possible from here.
			break

		default:
			panic(fmt.Sprintf("%v: unknown filter type", filter))
		}
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

	query += fmt.Sprintf("SELECT n.id, n.path, n.title, n.lead, n.body, n.raw_content, n.word_count, n.created, n.modified, n.checksum, n.tags, %s AS snippet\n", snippetCol)

	query += "FROM notes_with_metadata n\n"

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

	// fmt.Println(query)
	// fmt.Println(args)
	return d.tx.Query(query, args...)
}

func dateField(filter note.DateFilter) string {
	switch filter.Field {
	case note.DateCreated:
		return "created"
	case note.DateModified:
		return "modified"
	default:
		panic(fmt.Sprintf("%v: unknown note.DateField", filter.Field))
	}
}

func dateDirection(filter note.DateFilter) (op string, ignoreTime bool) {
	switch filter.Direction {
	case note.DateOn:
		return "=", true
	case note.DateBefore:
		return "<", false
	case note.DateAfter:
		return ">=", false
	default:
		panic(fmt.Sprintf("%v: unknown note.DateDirection", filter.Direction))
	}
}

func orderTerm(sorter note.Sorter) string {
	order := " ASC"
	if !sorter.Ascending {
		order = " DESC"
	}

	switch sorter.Field {
	case note.SortCreated:
		return "n.created" + order
	case note.SortModified:
		return "n.modified" + order
	case note.SortPath:
		return "n.path" + order
	case note.SortRandom:
		return "RANDOM()"
	case note.SortTitle:
		return "n.title" + order
	case note.SortWordCount:
		return "n.word_count" + order
	default:
		panic(fmt.Sprintf("%v: unknown note.SortField", sorter.Field))
	}
}

// pathRegex returns an ICU regex to match the files in the folder at given
// `path`, or any file having `path` for prefix.
func pathRegex(path string) string {
	path = icu.EscapePattern(path)
	return path + "[^/]*|" + path + "/.+"
}

func (d *NoteDAO) idToSql(id core.NoteId) sql.NullInt64 {
	if id.IsValid() {
		return sql.NullInt64{Int64: int64(id), Valid: true}
	} else {
		return sql.NullInt64{}
	}
}

func (d *NoteDAO) joinIds(ids []core.NoteId, delimiter string) string {
	strs := make([]string, 0)
	for _, i := range ids {
		strs = append(strs, strconv.FormatInt(int64(i), 10))
	}
	return strings.Join(strs, delimiter)
}
