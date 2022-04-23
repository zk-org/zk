package fzf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	stringsutil "github.com/mickael-menu/zk/internal/util/strings"
)

// ErrCancelled is returned when the user cancelled fzf.
var ErrCancelled = errors.New("cancelled")

// fzf exit codes
var (
	exitInterrupted = 130
	exitNoMatch     = 1
)

// Opts holds the options used to run fzf.
type Opts struct {
	// Preview command executed by fzf when hovering a line.
	PreviewCmd opt.String
	// Optionally preovide additional arguments, taken from the config `fzf-additional-args` property.
	AdditionalArgs opt.String
	// Amount of space between two non-empty fields.
	Padding int
	// Delimiter used by fzf between fields.
	Delimiter string
	// List of key bindings enabled in fzf.
	Bindings []Binding
}

// Binding represents a keyboard shortcut bound to an action in fzf.
type Binding struct {
	// Keyboard shortcut, e.g. `ctrl-n`.
	Keys string
	// fzf action, see `man fzf`.
	Action string
	// Description which will be displayed as a fzf header if not empty.
	Description string
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

	rawDefaultOpts := os.Getenv("FZF_DEFAULT_OPTS")
	defaultFzfArgs, err := shellquote.Split(rawDefaultOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "can't split FZF_DEFAULT_OPTS: %s", rawDefaultOpts)
	}

	args := []string{
		"--delimiter", opts.Delimiter,
		"--tiebreak", "begin",
		"--ansi",
		"--exact",
		"--tabstop", "4",
		"--height", "100%",
		"--layout", "reverse",
		//"--info", "inline",
		// Make sure the path and titles are always visible
		"--no-hscroll",
		// Don't highlight search terms
		"--color", "hl:-1,hl+:-1",
		"--preview-window", "wrap",
	}

	args = append(defaultFzfArgs, args...)

	additionalArgs, err := shellquote.Split(opts.AdditionalArgs.String())
	if err != nil {
		return nil, errors.Wrapf(err, "can't split the fzf-options: %s", opts.AdditionalArgs.String())
	}
	args = append(args, additionalArgs...)

	header := ""
	binds := []string{}
	for _, binding := range opts.Bindings {
		if binding.Description != "" {
			header += binding.Keys + ": " + binding.Description + "\n"
		}
		binds = append(binds, binding.Keys+":"+binding.Action)
	}

	if header != "" {
		args = append(args, "--header", strings.TrimSpace(header))
	}
	if len(binds) > 0 {
		args = append(args, "--bind", strings.Join(binds, ","))
	}

	if !opts.PreviewCmd.IsNull() {
		args = append(args, "--preview", opts.PreviewCmd.String())
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf("interactive mode requires fzf, try without --interactive or install fzf from https://github.com/junegunn/fzf")
	}

	cmd := exec.Command(fzfPath, args...)
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
			exitErr, ok := err.(*exec.ExitError)
			switch {
			case ok && exitErr.ExitCode() == exitInterrupted:
				f.err = ErrCancelled
			case ok && exitErr.ExitCode() == exitNoMatch:
				break
			default:
				f.err = errors.Wrap(err, "failed to filter interactively the output with fzf, try again without --interactive or make sure you have a working fzf installation")
			}
		} else {
			f.parseSelection(output)
		}
	}()

	return &f, nil
}

// parseSelection extracts the fields from fzf's output.
func (f *Fzf) parseSelection(output []byte) {
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
