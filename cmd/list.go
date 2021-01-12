package cmd

import (
	"fmt"
	"time"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/tj/go-naturaldate"
)

// List displays notes matching a set of criteria.
type List struct {
	Path           []string `arg optional placeholder:"PATH"`
	Format         string   `help:"Pretty prints the list using the given format" short:"f" placeholder:"TEMPLATE"`
	Match          string   `help:"Terms to search for in the notes" short:"m" placeholder:"TERMS"`
	Limit          int      `help:"Limit the number of results" short:"l" placeholder:"MAX"`
	Created        string   `help:"Show only the notes created on the given date" placeholder:"DATE"`
	CreatedBefore  string   `help:"Show only the notes created before the given date" placeholder:"DATE"`
	CreatedAfter   string   `help:"Show only the notes created after the given date" placeholder:"DATE"`
	Modified       string   `help:"Show only the notes modified on the given date" placeholder:"DATE"`
	ModifiedBefore string   `help:"Show only the notes modified before the given date" placeholder:"DATE"`
	ModifiedAfter  string   `help:"Show only the notes modified after the given date" placeholder:"DATE"`
}

func (cmd *List) Run(container *Container) error {
	zk, err := zk.Open(".")
	if err != nil {
		return err
	}

	opts, err := cmd.ListOpts(zk)
	if err != nil {
		return errors.Wrapf(err, "incorrect arguments")
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	return db.WithTransaction(func(tx sqlite.Transaction) error {
		notes := sqlite.NewNoteDAO(tx, container.Logger)

		deps := note.ListDeps{
			BasePath:  zk.Path,
			Finder:    notes,
			Templates: container.TemplateLoader(zk.Config.Lang),
		}

		count, err := note.List(*opts, deps, printNote)
		if err == nil {
			fmt.Printf("\nFound %d result(s)\n", count)
		}

		return err
	})
}

func (cmd *List) ListOpts(zk *zk.Zk) (*note.ListOpts, error) {
	filters := make([]note.Filter, 0)

	paths := make([]string, 0)
	for _, p := range cmd.Path {
		path, err := zk.RelPath(p)
		if err == nil {
			paths = append(paths, path)
		}
	}
	if len(paths) > 0 {
		filters = append(filters, note.PathFilter(paths))
	}

	if cmd.Match != "" {
		filters = append(filters, note.MatchFilter(cmd.Match))
	}

	if cmd.Created != "" {
		date, err := parseDate(cmd.Created)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateOn,
		})
	}

	if cmd.CreatedBefore != "" {
		date, err := parseDate(cmd.CreatedBefore)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateBefore,
		})
	}

	if cmd.CreatedAfter != "" {
		date, err := parseDate(cmd.CreatedAfter)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateAfter,
		})
	}

	if cmd.Modified != "" {
		date, err := parseDate(cmd.Modified)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateOn,
		})
	}

	if cmd.ModifiedBefore != "" {
		date, err := parseDate(cmd.ModifiedBefore)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateBefore,
		})
	}

	if cmd.ModifiedAfter != "" {
		date, err := parseDate(cmd.ModifiedAfter)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateAfter,
		})
	}

	return &note.ListOpts{
		Format: opt.NewNotEmptyString(cmd.Format),
		FinderOpts: note.FinderOpts{
			Filters: filters,
			Limit:   cmd.Limit,
		},
	}, nil
}

func printNote(note string) error {
	_, err := fmt.Println(note)
	return err
}

func parseDate(date string) (time.Time, error) {
	// FIXME: support years
	return naturaldate.Parse(date, time.Now().UTC(), naturaldate.WithDirection(naturaldate.Past))
}
