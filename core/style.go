package core

// Styler stylizes text according to predefined styling rules.
//
// A rule key can be either semantic, e.g. "title" or explicit, e.g. "red".
type Styler interface {
	Style(text string, rules ...StyleRule) (string, error)
}

// StyleRule is a key representing a single styling rule.
type StyleRule string

// NullStyler is a Styler with no styling rules.
var NullStyler = nullStyler{}

type nullStyler struct{}

func (s nullStyler) Style(text string, rule ...StyleRule) (string, error) {
	return text, nil
}
