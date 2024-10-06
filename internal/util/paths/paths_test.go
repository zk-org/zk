package paths

import (
	"os/user"
	"strings"
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestExpandPath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Error(err)
	}

	test := func(path string, expected string) {
		expanded, err := ExpandPath(path)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, expanded, expected)
	}
	home := usr.HomeDir

	s1 := []string{home, "foo"}
	homefoo := strings.Join(s1, "/")
	s2 := []string{"E.T phone", home}
	etph := strings.Join(s2, " ")

	// base cases
	test("~", home)
	test("~/", home)
	test("~/foo", homefoo)
	test("${HOME}/foo", homefoo)
	test("/usr/opt", "/usr/opt")

	// edge cases
	test("not a path", "not a path")
	test("E.T phone ${HOME}", etph)
}
