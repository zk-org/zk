package note

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/mickael-menu/zk/adapter/tty"
	"github.com/mickael-menu/zk/core/style"
	"github.com/mickael-menu/zk/util/strings"
)

func WithMatchFilter(callback func(func(Match) error)) ([]string, error) {
	styler := tty.NewStyler()

	return withFilter(func(w io.Writer) {
		callback(func(m Match) error {
			fmt.Fprintf(w, "%v\x01  %v  %v\n",
				m.Path,
				styler.MustStyle(m.Title, style.Rule("yellow")),
				styler.MustStyle(strings.JoinLines(m.Body), style.Rule("faint")),
			)
			return nil
		})
	})
}

func withFilter(callback func(w io.Writer)) ([]string, error) {
	zkBin, err := os.Executable()
	if err != nil {
		return []string{}, err
	}

	cmd := exec.Command(
		"fzf",
		"--delimiter", "\x01",
		"--tiebreak", "begin",
		"--ansi",
		"--exact",
		"--height", "100%",
		// FIXME: Use it to create a new note? Like notational velocity
		// "--print-query",
		// Make sure the path and titles are always visible
		"--no-hscroll",
		"--tabstop", "4",
		// Don't highlight search terms
		"--color", "hl:-1,hl+:-1",
		// "--preview", `bat -p --theme Nord --color always {1}`,
		"--preview", zkBin+" list -f {{raw-content}} {1}",
		"--preview-window", "noborder:wrap",
	)
	cmd.Stderr = os.Stderr

	w, err := cmd.StdinPipe()
	if err != nil {
		return []string{}, err
	}
	go func() {
		callback(w)
		w.Close()
	}()

	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	return strings.SplitLines(string(output)), nil
}
