package fixtures

import (
	"path/filepath"
	"runtime"
)

// Path returns the absolute path to the given fixture.
func Path(name string) string {
	_, callerPath, _, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get the caller's path")
	}
	return filepath.Join(filepath.Dir(callerPath), "testdata", name)
}
