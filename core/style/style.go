package style

// Styler stylizes text according to predefined styling rules.
//
// A rule key can be either semantic, e.g. "title" or explicit, e.g. "red".
type Styler interface {
	Style(text string, rules ...Rule) (string, error)
}

// Rule is a key representing a single styling rule.
type Rule string

// NullStyler is a Styler with no styling rules.
var NullStyler = nullStyler{}

type nullStyler struct{}

func (s nullStyler) Style(text string, rule ...Rule) (string, error) {
	return text, nil
}
