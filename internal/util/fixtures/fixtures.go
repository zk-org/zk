package fixtures

import (
	"os"
	"path/filepath"
)

// Path returns the absolute path to the given fixture.
func Path(name string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic("failed to obtain current working directory")
	}
	return filepath.Join(cwd, "testdata", name)
}
