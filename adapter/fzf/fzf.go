package fzf

import (
	"io"
	"os"
	"os/exec"

	"github.com/mickael-menu/zk/util/strings"
)

func withFzf(callback func(fzf io.Writer) error) ([]string, error) {
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

	var callbackErr error
	go func() {
		callbackErr = callback(w)
		w.Close()
	}()

	output, err := cmd.Output()
	if callbackErr != nil {
		return []string{}, callbackErr
	}
	if err != nil {
		return []string{}, err
	}

	return strings.SplitLines(string(output)), nil
}
