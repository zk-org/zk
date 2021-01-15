package pager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	osutil "github.com/mickael-menu/zk/util/os"
)

// Pager writes text to a TTY using the user's pager.
type Pager struct {
	io.WriteCloser
	done        chan bool
	isCloseable bool
	closeOnce   sync.Once
}

// PassthroughPager is a Pager writing the content directly to stdout without
// pagination.
var PassthroughPager = &Pager{
	WriteCloser: os.Stdout,
	isCloseable: false,
}

// New creates a pager.Pager to be used to write a paginated text to the TTY.
func New(logger util.Logger) (*Pager, error) {
	wrap := errors.Wrapper("failed to paginate the output, try again with --no-pager or fix your PAGER environment variable")

	pagerCmd := locatePager()
	if pagerCmd.IsNull() {
		return PassthroughPager, nil
	}

	args, err := shellquote.Split(pagerCmd.String())
	if err != nil {
		return nil, wrap(err)
	}
	cmd := exec.Command(args[0], args[1:]...)

	r, w, err := os.Pipe()
	if err != nil {
		return nil, wrap(err)
	}

	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	done := make(chan bool)
	pager := Pager{
		WriteCloser: w,
		done:        done,
		isCloseable: true,
		closeOnce:   sync.Once{},
	}

	go func() {
		defer close(done)

		err := cmd.Run()
		if err != nil {
			logger.Err(wrap(err))
			os.Exit(1)
		}
	}()

	return &pager, nil
}

// Close terminates the pager, waiting for the process to be finished before returning.
//
// We make sure Close is called only once, since we don't know how it is
// implemented in underlying writers.
func (p *Pager) Close() error {
	if !p.isCloseable {
		return nil
	}

	var err error
	p.closeOnce.Do(func() {
		err = p.WriteCloser.Close()
	})
	<-p.done
	return err
}

// WriteString sends the given text to the pager, ending with a newline.
func (p *Pager) WriteString(text string) error {
	_, err := fmt.Fprintln(p, text)
	return err
}

func locatePager() opt.String {
	return osutil.GetOptEnv("ZK_PAGER").
		Or(osutil.GetOptEnv("PAGER")).
		Or(locateDefaultPager())
}

var defaultPagers = []string{
	"less -FIRX", "more -R",
}

func locateDefaultPager() opt.String {
	for _, pager := range defaultPagers {
		parts, err := shellquote.Split(pager)
		if err != nil {
			continue
		}

		pager, err := exec.LookPath(parts[0])
		parts[0] = pager
		if err == nil {
			return opt.NewNotEmptyString(strings.Join(parts, " "))
		}
	}
	return opt.NullString
}
