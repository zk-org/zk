package zk

// IDOptions holds the options used to generate an ID.
type IDOptions struct {
	Length  int
	Charset Charset
	Case    Case
}

// Charset is a set of characters.
type Charset []rune

var (
	// CharsetAlphanum is a charset containing letters and numbers.
	CharsetAlphanum = Charset("0123456789abcdefghijklmnopqrstuvwxyz")
	// CharsetAlphanum is a charset containing hexadecimal characters.
	CharsetHex = Charset("0123456789abcdef")
	// CharsetLetters is a charset containing only letters.
	CharsetLetters = Charset("abcdefghijklmnopqrstuvwxyz")
	// CharsetNumbers is a charset containing only numbers.
	CharsetNumbers = Charset("0123456789")
)

// Case represents the letter case to use when generating an ID.
type Case int

const (
	CaseLower Case = iota + 1
	CaseUpper
	CaseMixed
)
