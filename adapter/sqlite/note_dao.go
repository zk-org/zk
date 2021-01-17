package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/fts5"
	"github.com/mickael-menu/zk/util/paths"
)

// NoteDAO persists notes in the SQLite database.
// It implements the core ports note.Indexer and note.Finder.
type NoteDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	indexedStmt *LazyStmt
	addStmt     *LazyStmt
	updateStmt  *LazyStmt
	removeStmt  *LazyStmt
	existsStmt  *LazyStmt
}

func NewNoteDAO(tx Transaction, logger util.Logger) *NoteDAO {
	return &NoteDAO{
		tx:     tx,
		logger: logger,
		indexedStmt: tx.PrepareLazy(`
			SELECT path, modified from notes
			 ORDER BY sortable_path ASC
		`),
		addStmt: tx.PrepareLazy(`
			INSERT INTO notes (path, sortable_path, title, lead, body, raw_content, word_count, checksum, created, modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`),
		updateStmt: tx.PrepareLazy(`
			UPDATE notes
			   SET title = ?, lead = ?, body = ?, raw_content = ?, word_count = ?, checksum = ?, modified = ?
			 WHERE path = ?
		`),
		removeStmt: tx.PrepareLazy(`
			DELETE FROM notes
			 WHERE path = ?
		`),
		existsStmt: tx.PrepareLazy(`
			SELECT EXISTS (SELECT 1 FROM notes WHERE path = ?)
		`),
	}
}

func (d *NoteDAO) Indexed() (<-chan paths.Metadata, error) {
	wrap := errors.Wrapper("failed to get indexed notes")

	rows, err := d.indexedStmt.Query()
	if err != nil {
		return nil, wrap(err)
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
				d.logger.Err(wrap(err))
			}

			c <- paths.Metadata{
				Path:     path,
				Modified: modified,
			}
		}

		err = rows.Err()
		if err != nil {
			d.logger.Err(wrap(err))
		}
	}()

	return c, nil
}

func (d *NoteDAO) Add(note note.Metadata) error {
	// For sortable_path, we replace in path / by the shortest non printable
	// character available to make it sortable. Without this, sorting by the
	// path would be a lexicographical sort instead of being the same order
	// returned by filepath.Walk.
	// \x01 is used instead of \x00, because SQLite treats \x00 as and end of
	// string.
	sortablePath := strings.ReplaceAll(note.Path, "/", "\x01")

	_, err := d.addStmt.Exec(
		note.Path, sortablePath, note.Title, note.Lead, note.Body, note.RawContent, note.WordCount, note.Checksum,
		note.Created, note.Modified,
	)
	return errors.Wrapf(err, "%v: can't add note to the index", note.Path)
}

func (d *NoteDAO) Update(note note.Metadata) error {
	wrap := errors.Wrapperf("%v: failed to update note index", note.Path)

	exists, err := d.exists(note.Path)
	if err != nil {
		return wrap(err)
	}
	if !exists {
		return wrap(errors.New("note not found in the index"))
	}

	_, err = d.updateStmt.Exec(
		note.Title, note.Lead, note.Body, note.RawContent, note.WordCount, note.Checksum, note.Modified,
		note.Path,
	)
	return errors.Wrapf(err, "%v: failed to update note index", note.Path)
}

func (d *NoteDAO) Remove(path string) error {
	wrap := errors.Wrapperf("%v: failed to remove note index", path)

	exists, err := d.exists(path)
	if err != nil {
		return wrap(err)
	}
	if !exists {
		return wrap(errors.New("note not found in the index"))
	}

	_, err = d.removeStmt.Exec(path)
	return wrap(err)
}

func (d *NoteDAO) exists(path string) (bool, error) {
	row, err := d.existsStmt.QueryRow(path)
	if err != nil {
		return false, err
	}
	var exists bool
	row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (d *NoteDAO) Find(opts note.FinderOpts, callback func(note.Match) error) (int, error) {
	rows, err := d.findRows(opts)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++

		var (
			id, wordCount                          int
			title, lead, body, rawContent, snippet string
			path, checksum                         string
			created, modified                      time.Time
		)

		err := rows.Scan(&id, &path, &title, &lead, &body, &rawContent, &wordCount, &created, &modified, &checksum, &snippet)
		if err != nil {
			d.logger.Err(err)
			continue
		}

		callback(note.Match{
			Snippet: snippet,
			Metadata: note.Metadata{
				Path:       path,
				Title:      title,
				Lead:       lead,
				Body:       body,
				RawContent: rawContent,
				WordCount:  wordCount,
				Created:    created,
				Modified:   modified,
				Checksum:   checksum,
			},
		})
	}

	return count, nil
}

type findQuery struct {
	SnippetCol string
	WhereExprs []string
	OrderTerms []string
	Args       []interface{}
}

func (d *NoteDAO) findRows(opts note.FinderOpts) (*sql.Rows, error) {
	snippetCol := `n.lead`
	whereExprs := make([]string, 0)
	orderTerms := make([]string, 0)
	args := make([]interface{}, 0)

	for _, filter := range opts.Filters {
		switch filter := filter.(type) {

		case note.MatchFilter:
			snippetCol = `snippet(notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20) as snippet`
			orderTerms = append(orderTerms, `bm25(notes_fts, 1000.0, 500.0, 1.0)`)
			whereExprs = append(whereExprs, "notes_fts MATCH ?")
			args = append(args, fts5.ConvertQuery(string(filter)))

		case note.PathFilter:
			if len(filter) == 0 {
				break
			}
			globs := make([]string, 0)
			for _, path := range filter {
				globs = append(globs, "n.path GLOB ?")
				args = append(args, path+"*")
			}
			whereExprs = append(whereExprs, strings.Join(globs, " OR "))

		case note.ExcludePathFilter:
			if len(filter) == 0 {
				break
			}
			globs := make([]string, 0)
			for _, path := range filter {
				globs = append(globs, "n.path NOT GLOB ?")
				args = append(args, path+"*")
			}
			whereExprs = append(whereExprs, strings.Join(globs, " AND "))

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

		default:
			panic(fmt.Sprintf("%v: unknown filter type", filter))
		}
	}

	for _, sorter := range opts.Sorters {
		orderTerms = append(orderTerms, orderTerm(sorter))
	}
	orderTerms = append(orderTerms, `n.title ASC`)

	query := "SELECT n.id, n.path, n.title, n.lead, n.body, n.raw_content, n.word_count, n.created, n.modified, n.checksum, " + snippetCol

	query += `
FROM notes n
JOIN notes_fts
ON n.id = notes_fts.rowid`

	if len(whereExprs) > 0 {
		query += "\nWHERE " + strings.Join(whereExprs, "\nAND ")
	}

	query += "\nORDER BY " + strings.Join(orderTerms, ", ")

	if opts.Limit > 0 {
		query += fmt.Sprintf("\nLIMIT %d", opts.Limit)
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
