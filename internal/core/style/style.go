package style

// Styler stylizes text according to predefined styling rules.
//
// A rule key can be either semantic, e.g. "title" or explicit, e.g. "red".
type Styler interface {
	Style(text string, rules ...Rule) (string, error)
	MustStyle(text string, rules ...Rule) string
}

// Rule is a key representing a single styling rule.
type Rule string

// Predefined styling rules.
var (
	// Title of a note.
	RuleTitle = Rule("title")
	// Path to notebook file.
	RulePath = Rule("path")
	// Searched for term in a note.
	RuleTerm = Rule("term")
	// Element to emphasize, for example the short version of a prompt response: [y]es.
	RuleEmphasis = Rule("emphasis")
	// Element to understate, for example the content of the note in fzf.
	RuleUnderstate = Rule("understate")

	RuleBold          = Rule("bold")
	RuleItalic        = Rule("italic")
	RuleFaint         = Rule("faint")
	RuleUnderline     = Rule("underline")
	RuleStrikethrough = Rule("strikethrough")
	RuleBlink         = Rule("blink")
	RuleReverse       = Rule("reverse")
	RuleHidden        = Rule("hidden")

	RuleBlack   = Rule("black")
	RuleRed     = Rule("red")
	RuleGreen   = Rule("green")
	RuleYellow  = Rule("yellow")
	RuleBlue    = Rule("blue")
	RuleMagenta = Rule("magenta")
	RuleCyan    = Rule("cyan")
	RuleWhite   = Rule("white")

	RuleBlackBg   = Rule("black-bg")
	RuleRedBg     = Rule("red-bg")
	RuleGreenBg   = Rule("green-bg")
	RuleYellowBg  = Rule("yellow-bg")
	RuleBlueBg    = Rule("blue-bg")
	RuleMagentaBg = Rule("magenta-bg")
	RuleCyanBg    = Rule("cyan-bg")
	RuleWhiteBg   = Rule("white-bg")

	RuleBrightBlack   = Rule("bright-black")
	RuleBrightRed     = Rule("bright-red")
	RuleBrightGreen   = Rule("bright-green")
	RuleBrightYellow  = Rule("bright-yellow")
	RuleBrightBlue    = Rule("bright-blue")
	RuleBrightMagenta = Rule("bright-magenta")
	RuleBrightCyan    = Rule("bright-cyan")
	RuleBrightWhite   = Rule("bright-white")

	RuleBrightBlackBg   = Rule("bright-black-bg")
	RuleBrightRedBg     = Rule("bright-red-bg")
	RuleBrightGreenBg   = Rule("bright-green-bg")
	RuleBrightYellowBg  = Rule("bright-yellow-bg")
	RuleBrightBlueBg    = Rule("bright-blue-bg")
	RuleBrightMagentaBg = Rule("bright-magenta-bg")
	RuleBrightCyanBg    = Rule("bright-cyan-bg")
	RuleBrightWhiteBg   = Rule("bright-white-bg")
)

// NullStyler is a Styler with no styling rules.
var NullStyler = nullStyler{}

type nullStyler struct{}

func (s nullStyler) Style(text string, rule ...Rule) (string, error) {
	return text, nil
}
