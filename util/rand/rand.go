package rand

import (
	"math/rand"
	"time"
	"unicode"
)

var (
	// AlphanumCharset is a charset containing letters and numbers.
	AlphanumCharset = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	// AlphanumCharset is a charset containing hexadecimal characters.
	HexCharset = []rune("0123456789abcdef")
	// LettersCharset is a charset containing only letters.
	LettersCharset = []rune("abcdefghijklmnopqrstuvwxyz")
	// NumbersCharset is a charset containing only numbers.
	NumbersCharset = []rune("0123456789")
)

// Case represents the letter case to use when generating a string.
type Case int

const (
	LowerCase Case = iota
	UpperCase
	MixedCase
)

// IDOpts holds the options used to generate a random ID.
type IDOpts struct {
	Length  int
	Charset []rune
	Case    Case
}

// GenID creates a new random string ID using the given options.
// Inspired by https://www.calhoun.io/creating-random-strings-in-go/
func GenID(options IDOpts) string {
	if options.Length < 1 {
		panic("IDOpts.Length must be at least 1")
	}

	var charset []rune
	for _, char := range options.Charset {
		switch options.Case {
		case LowerCase:
			charset = append(charset, unicode.ToLower(char))
		case UpperCase:
			charset = append(charset, unicode.ToUpper(char))
		case MixedCase:
			charset = append(charset, unicode.ToLower(char))
			charset = append(charset, unicode.ToUpper(char))
		default:
			panic("unknown rand.Case value")
		}
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]rune, options.Length)
	for i := range buf {
		buf[i] = charset[rand.Intn(len(charset))]
	}

	return string(buf)
}
