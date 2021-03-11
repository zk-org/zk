package os

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

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

// Getenv returns an optional String for the environment variable with given
// key.
func GetOptEnv(key string) opt.String {
	return opt.NewNotEmptyString(os.Getenv(key))
}

// Env returns a map of environment variables.
func Env() map[string]string {
	env := map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		env[pair[0]] = pair[1]
	}
	return env
}
