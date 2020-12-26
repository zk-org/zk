package rand

var (
	// AlphanumCharset is a charset containing letters and numbers.
	AlphanumCharset = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	// AlphanumCharset is a charset containing hexadecimal characters.
	HexCharset = []rune("0123456789abcdef")
	// LettersCharset is a charset containing only letters.
	LettersCharset = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	// NumbersCharset is a charset containing only numbers.
	NumbersCharset = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
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
	Charset []rune
	Length  int
	Case    Case
}

// GenID creates a new random string ID using the given options.
func GenID(options IDOpts) string {
	if options.Length < 1 {
		panic("IDOpts.Length must be at least 1")
	}

	return ""
}
