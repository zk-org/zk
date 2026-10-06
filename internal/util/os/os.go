package os

import (
	"os"
	"strings"

	"github.com/zk-org/zk/internal/util/ptr"
)

// GetOptEnv returns an optional String for the environment variable with given
// key.
func GetOptEnv(key string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return ptr.NotEmptyString(value)
	}
	return nil
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
