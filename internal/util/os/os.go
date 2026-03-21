package os

import (
	"os"
	"strings"

	"github.com/zk-org/zk/internal/util/opt"
)

// GetOptEnv returns an optional String for the environment variable with given
// key.
func GetOptEnv(key string) opt.String {
	if value, ok := os.LookupEnv(key); ok {
		return opt.NewNotEmptyString(value)
	}
	return opt.NullString
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
