package util

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/mickael-menu/zk/util/opt"
)

// ReadStdinPipe returns the content of any piped input.
func ReadStdinPipe() (opt.String, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return opt.NullString, err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		// Not a pipe
		return opt.NullString, nil
	}

	reader := bufio.NewReader(os.Stdin)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return opt.NullString, err
	}

	return opt.NewNotEmptyString(string(bytes)), nil
}
