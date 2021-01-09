package tty

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mickael-menu/zk/core"
)

// Styler is a text styler using ANSI escape codes to be used with a TTY.
type Styler struct{}

func NewStyler() *Styler {
	return &Styler{}
}

// FIXME: Semantic rules
func (s *Styler) Style(text string, rules ...core.StyleRule) (string, error) {
	attrs, err := s.attributes(rules)
	if err != nil {
		return "", err
	}
	if len(attrs) == 0 {
		return text, nil
	}
	return color.New(attrs...).Sprint(text), nil
}

var attrsMapping = map[core.StyleRule]color.Attribute{
	"reset":         color.Reset,
	"bold":          color.Bold,
	"faint":         color.Faint,
	"italic":        color.Italic,
	"underline":     color.Underline,
	"blink-slow":    color.BlinkSlow,
	"blink-fast":    color.BlinkRapid,
	"hidden":        color.Concealed,
	"strikethrough": color.CrossedOut,

	"black":   color.FgBlack,
	"red":     color.FgRed,
	"green":   color.FgGreen,
	"yellow":  color.FgYellow,
	"blue":    color.FgBlue,
	"magenta": color.FgMagenta,
	"cyan":    color.FgCyan,
	"white":   color.FgWhite,

	"black-bg":   color.BgBlack,
	"red-bg":     color.BgRed,
	"green-bg":   color.BgGreen,
	"yellow-bg":  color.BgYellow,
	"blue-bg":    color.BgBlue,
	"magenta-bg": color.BgMagenta,
	"cyan-bg":    color.BgCyan,
	"white-bg":   color.BgWhite,

	"bright-black":   color.FgHiBlack,
	"bright-red":     color.FgHiRed,
	"bright-green":   color.FgHiGreen,
	"bright-yellow":  color.FgHiYellow,
	"bright-blue":    color.FgHiBlue,
	"bright-magenta": color.FgHiMagenta,
	"bright-cyan":    color.FgHiCyan,
	"bright-white":   color.FgHiWhite,

	"bright-black-bg":   color.BgHiBlack,
	"bright-red-bg":     color.BgHiRed,
	"bright-green-bg":   color.BgHiGreen,
	"bright-yellow-bg":  color.BgHiYellow,
	"bright-blue-bg":    color.BgHiBlue,
	"bright-magenta-bg": color.BgHiMagenta,
	"bright-cyan-bg":    color.BgHiCyan,
	"bright-white-bg":   color.BgHiWhite,
}

func (s *Styler) attributes(rules []core.StyleRule) ([]color.Attribute, error) {
	attrs := make([]color.Attribute, 0)

	for _, rule := range rules {
		attr, ok := attrsMapping[rule]
		if !ok {
			return attrs, fmt.Errorf("unknown styling rule: %v", rule)
		} else {
			attrs = append(attrs, attr)
		}
	}

	return attrs, nil
}
