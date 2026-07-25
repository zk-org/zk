package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
	"gopkg.in/yaml.v3"
)

// loadFixtures inserts the YAML fixture files found in dir into the given
// SQLite database. Each .yml/.yaml file is expected to be named after a table
// and to contain a YAML sequence of records to insert.
//
// Foreign keys are deferred for the duration of the load so that records can
// be inserted in any order.
func loadFixtures(t *testing.T, db *sql.DB, dir string) {
	entries, err := os.ReadDir(dir)
	assert.Nil(t, err)

	tx, err := db.Begin()
	assert.Nil(t, err)
	defer tx.Rollback()

	_, err = tx.Exec("PRAGMA defer_foreign_keys = ON")
	assert.Nil(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		table := strings.TrimSuffix(name, ext)
		data, err := os.ReadFile(filepath.Join(dir, name))
		assert.Nil(t, err)

		var records []map[string]any
		err = yaml.Unmarshal(data, &records)
		assert.Nil(t, err)

		for _, record := range records {
			insertFixtureRecord(t, tx, table, record)
		}
	}

	assert.Nil(t, tx.Commit())
}

func insertFixtureRecord(t *testing.T, tx *sql.Tx, table string, record map[string]any) {
	type column struct {
		name        string
		value       any
		placeholder string
	}

	columns := make([]column, 0, len(record))
	for name, value := range record {
		columns = append(columns, column{
			name:        name,
			value:       value,
			placeholder: "?",
		})
	}

	sort.Slice(columns, func(i, j int) bool {
		return columns[i].name < columns[j].name
	})

	names := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	values := make([]any, len(columns))
	for i, col := range columns {
		names[i] = col.name
		placeholders[i] = col.placeholder
		values[i] = col.value
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(names, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := tx.Exec(query, values...)
	assert.Nil(t, err)
}
