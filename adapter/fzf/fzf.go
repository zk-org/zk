package fzf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	stringsutil "github.com/mickael-menu/zk/util/strings"
)

// fzf exit codes
var (
	exitInterrupted = 130
	exitNoMatch     = 1
)

// Opts holds the options used to run fzf.
type Opts struct {
	// Preview command executed by fzf when hovering a line.
	PreviewCmd opt.String
	// Amount of space between two non-empty fields.
	Padding int
	// Delimiter used by fzf between fields.
	Delimiter string
}

// Fzf filters a set of fields using fzf.
//
// After adding all the fields with Add, use Selection to get the filtered
// results.
type Fzf struct {
	opts Opts

	// Fields selection or error result.
	err       error
	selection [][]string

	done      chan bool
	cmd       *exec.Cmd
	pipe      io.WriteCloser
	closeOnce sync.Once
}

// New runs a fzf instance.
//
// To show a preview of each line, provide a previewCmd which will be executed
// by fzf.
func New(opts Opts) (*Fzf, error) {
	// \x01 is a convenient delimiter because not visible in the output and
	// most likely not part of the fields themselves.
	if opts.Delimiter == "" {
		opts.Delimiter = "\x01"
	}

	args := []string{
		"--delimiter", opts.Delimiter,
		"--tiebreak", "begin",
		"--ansi",
		"--exact",
		"--tabstop", "4",
		"--height", "100%",
		"--layout", "reverse",
		// FIXME: Use it to create a new note? Like notational velocity
		// "--print-query",
		// Make sure the path and titles are always visible
		"--no-hscroll",
		// Don't highlight search terms
		"--color", "hl:-1,hl+:-1",
		// "--preview-window", "noborder:wrap",
	}
	if !opts.PreviewCmd.IsNull() {
		args = append(args, "--preview", opts.PreviewCmd.String())
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stderr = os.Stderr

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	f := Fzf{
		opts:      opts,
		cmd:       cmd,
		pipe:      pipe,
		closeOnce: sync.Once{},
		done:      done,
		selection: make([][]string, 0),
	}

	go func() {
		defer func() {
			close(done)
			f.close()
		}()

		output, err := cmd.Output()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok &&
				err.ExitCode() != exitInterrupted &&
				err.ExitCode() != exitNoMatch {
				f.err = errors.Wrap(err, "failed to filter interactively the output with fzf, try again without --interactive or make sure you have a working fzf installation")
			}
		} else {
			f.parseSelection(string(output))
		}
	}()

	return &f, nil
}

// parseSelection extracts the fields from fzf's output.
func (f *Fzf) parseSelection(output string) {
	f.selection = make([][]string, 0)
	lines := stringsutil.SplitLines(string(output))
	for _, line := range lines {
		fields := strings.Split(line, f.opts.Delimiter)
		// Trim padding
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}
		f.selection = append(f.selection, fields)
	}
}

// Add appends a new line of fields to fzf input.
func (f *Fzf) Add(fields []string) error {
	line := ""
	for i, field := range fields {
		if i > 0 {
			line += f.opts.Delimiter

			if field != "" && f.opts.Padding > 0 {
				line += strings.Repeat(" ", f.opts.Padding)
			}
		}
		line += field
	}
	if line == "" {
		return nil
	}

	_, err := fmt.Fprintln(f.pipe, line)
	return err
}

// Selection returns the field lines selected by the user through fzf.
func (f *Fzf) Selection() ([][]string, error) {
	f.close()
	<-f.done
	return f.selection, f.err
}

func (f *Fzf) close() error {
	var err error
	f.closeOnce.Do(func() {
		err = f.pipe.Close()
	})
	return err
}
