package rand

import (
	"math/rand"
	"time"
	"unicode"

	"github.com/mickael-menu/zk/core/zk"
)

// NewIDGenerator returns a function generating string IDs using the given options.
// Inspired by https://www.calhoun.io/creating-random-strings-in-go/
func NewIDGenerator(options zk.IDOptions) func() string {
	if options.Length < 1 {
		panic("IDOptions.Length must be at least 1")
	}

	var charset []rune
	for _, char := range options.Charset {
		switch options.Case {
		case zk.CaseLower:
			charset = append(charset, unicode.ToLower(char))
		case zk.CaseUpper:
			charset = append(charset, unicode.ToUpper(char))
		case zk.CaseMixed:
			charset = append(charset, unicode.ToLower(char))
			charset = append(charset, unicode.ToUpper(char))
		default:
			panic("unknown zk.Case value")
		}
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func() string {
		buf := make([]rune, options.Length)
		for i := range buf {
			buf[i] = charset[rand.Intn(len(charset))]
		}

		return string(buf)
	}
}
