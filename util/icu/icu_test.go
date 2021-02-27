package icu

import (
	"testing"

	"github.com/mickael-menu/zk/util/test/assert"
)

func TestEscapePAttern(t *testing.T) {
	tests := map[string]string{
		`foo bar`: `foo bar`,
		`\a`:      `\\a`,
		`.`:       `\.`,
		`^`:       `\^`,
		`$`:       `\$`,
		`(`:       `\(`,
		`)`:       `\)`,
		`[`:       `\[`,
		`]`:       `\]`,
		`{`:       `\{`,
		`}`:       `\}`,
		`|`:       `\|`,
		`*`:       `\*`,
		`+`:       `\+`,
		`?`:       `\?`,
		`(?:[A-Za-z0-9]+[._]?){1,}[A-Za-z0-9]+\@(?:(?:[A-Za-z0-9]+[-]?){1,}[A-Za-z0-9]+\.){1,}`: `\(\?:\[A-Za-z0-9\]\+\[\._\]\?\)\{1,\}\[A-Za-z0-9\]\+\\@\(\?:\(\?:\[A-Za-z0-9\]\+\[-\]\?\)\{1,\}\[A-Za-z0-9\]\+\\\.\)\{1,\}`,
	}

	for input, expected := range tests {
		assert.Equal(t, EscapePattern(input), expected)
	}
}
