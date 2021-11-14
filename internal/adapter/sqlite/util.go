package sqlite

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/errors"
)

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type RowQuerier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// escapeLikeTerm returns the given term after escaping any LIKE-significant
// characters with the given escapeChar.
// This is meant to be used with the ESCAPE keyword:
// https://www.sqlite.org/lang_expr.html
func escapeLikeTerm(term string, escapeChar rune) string {
	escape := func(term string, char string) string {
		return strings.ReplaceAll(term, char, string(escapeChar)+char)
	}
	return escape(escape(escape(term, string(escapeChar)), "%"), "_")
}

func noteIDToSQL(id core.NoteID) sql.NullInt64 {
	if id.IsValid() {
		return sql.NullInt64{Int64: int64(id), Valid: true}
	} else {
		return sql.NullInt64{}
	}
}

func joinNoteIDs(ids []core.NoteID, delimiter string) string {
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
