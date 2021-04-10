package core

// Style is a key representing a single styling rule.
type Style string

// Predefined styling rules.
var (
	// Title of a note.
	StyleTitle = Style("title")
	// Path to notebook file.
	StylePath = Style("path")
	// Searched for term in a note.
	StyleTerm = Style("term")
	// Element to emphasize, for example the short version of a prompt response: [y]es.
	StyleEmphasis = Style("emphasis")
	// Element to understate, for example the content of the note in fzf.
	StyleUnderstate = Style("understate")

	StyleBold          = Style("bold")
	StyleItalic        = Style("italic")
	StyleFaint         = Style("faint")
	StyleUnderline     = Style("underline")
	StyleStrikethrough = Style("strikethrough")
	StyleBlink         = Style("blink")
	StyleReverse       = Style("reverse")
	StyleHidden        = Style("hidden")

	StyleBlack   = Style("black")
	StyleRed     = Style("red")
	StyleGreen   = Style("green")
	StyleYellow  = Style("yellow")
	StyleBlue    = Style("blue")
	StyleMagenta = Style("magenta")
	StyleCyan    = Style("cyan")
	StyleWhite   = Style("white")

	StyleBlackBg   = Style("black-bg")
	StyleRedBg     = Style("red-bg")
	StyleGreenBg   = Style("green-bg")
	StyleYellowBg  = Style("yellow-bg")
	StyleBlueBg    = Style("blue-bg")
	StyleMagentaBg = Style("magenta-bg")
	StyleCyanBg    = Style("cyan-bg")
	StyleWhiteBg   = Style("white-bg")

	StyleBrightBlack   = Style("bright-black")
	StyleBrightRed     = Style("bright-red")
	StyleBrightGreen   = Style("bright-green")
	StyleBrightYellow  = Style("bright-yellow")
	StyleBrightBlue    = Style("bright-blue")
	StyleBrightMagenta = Style("bright-magenta")
	StyleBrightCyan    = Style("bright-cyan")
	StyleBrightWhite   = Style("bright-white")

	StyleBrightBlackBg   = Style("bright-black-bg")
	StyleBrightRedBg     = Style("bright-red-bg")
	StyleBrightGreenBg   = Style("bright-green-bg")
	StyleBrightYellowBg  = Style("bright-yellow-bg")
	StyleBrightBlueBg    = Style("bright-blue-bg")
	StyleBrightMagentaBg = Style("bright-magenta-bg")
	StyleBrightCyanBg    = Style("bright-cyan-bg")
	StyleBrightWhiteBg   = Style("bright-white-bg")
)

// Styler stylizes text according to predefined styling rules.
//
// A rule key can be either semantic, e.g. "title" or explicit, e.g. "red".
type Styler interface {
	// Style formats the given text according to the provided styling rules.
	Style(text string, rules ...Style) (string, error)
	// Style formats the given text according to the provided styling rules,
	// panicking if the rules are unknown.
	MustStyle(text string, rules ...Style) string
}

// NullStyler is a Styler with no styling rules.
var NullStyler = nullStyler{}

type nullStyler struct{}

// Style implements Styler.
func (s nullStyler) Style(text string, rule ...Style) (string, error) {
	return text, nil
}

// MustStyle implements Styler.
func (s nullStyler) MustStyle(text string, rule ...Style) string {
	return text
}
