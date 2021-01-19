package note

import (
	"fmt"
	"io"
	"os"

	"github.com/mickael-menu/zk/core/style"
	"github.com/mickael-menu/zk/core/templ"
	"github.com/mickael-menu/zk/util/opt"
)

type ListOpts struct {
	Format opt.String
	FinderOpts
}

type ListDeps struct {
	BasePath  string
	Finder    Finder
	Templates templ.Loader
	Styler    style.Styler
}

// List finds notes matching given criteria and formats them according to user
// preference.
func List(opts ListOpts, deps ListDeps, out io.Writer) (int, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	formatter, err := NewFormatter(deps.BasePath, wd, opts.Format, deps.Templates, deps.Styler)
	if err != nil {
		return 0, err
	}

	return deps.Finder.Find(opts.FinderOpts, func(note Match) error {
		ft, err := formatter.Format(note)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, ft)
		return err
	})
}
