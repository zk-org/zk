package term

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/zk-org/zk/internal/core"
)

// Style implements core.Styler using ANSI escape codes to be used with a terminal.
func (t *Terminal) Style(text string, rules ...core.Style) (string, error) {
	if text == "" {
		return text, nil
	}
	attrs, err := attributes(expandThemeAliases(rules))
	if err != nil {
		return "", err
	}
	if len(attrs) == 0 {
		return text, nil
	}
	return color.New(attrs...).Sprint(text), nil
}

func (t *Terminal) MustStyle(text string, rules ...core.Style) string {
	text, err := t.Style(text, rules...)
	if err != nil {
		panic(err.Error())
	}
	return text
}

// FIXME: User config
var themeAliases = map[core.Style][]core.Style{
	"title":      {"bold", "yellow"},
	"path":       {"underline", "cyan"},
	"term":       {"red"},
	"emphasis":   {"bold", "cyan"},
	"understate": {"faint"},
}

func expandThemeAliases(rules []core.Style) []core.Style {
	expanded := make([]core.Style, 0)
	for _, rule := range rules {
		aliases, ok := themeAliases[rule]
		if ok {
			aliases = expandThemeAliases(aliases)
			expanded = append(expanded, aliases...)

		} else {
			expanded = append(expanded, rule)
		}
	}

	return expanded
}

var attrsMapping = map[core.Style]color.Attribute{
	core.StyleBold:          color.Bold,
	core.StyleFaint:         color.Faint,
	core.StyleItalic:        color.Italic,
	core.StyleUnderline:     color.Underline,
	core.StyleBlink:         color.BlinkSlow,
	core.StyleReverse:       color.ReverseVideo,
	core.StyleHidden:        color.Concealed,
	core.StyleStrikethrough: color.CrossedOut,

	core.StyleBlack:   color.FgBlack,
	core.StyleRed:     color.FgRed,
	core.StyleGreen:   color.FgGreen,
	core.StyleYellow:  color.FgYellow,
	core.StyleBlue:    color.FgBlue,
	core.StyleMagenta: color.FgMagenta,
	core.StyleCyan:    color.FgCyan,
	core.StyleWhite:   color.FgWhite,

	core.StyleBlackBg:   color.BgBlack,
	core.StyleRedBg:     color.BgRed,
	core.StyleGreenBg:   color.BgGreen,
	core.StyleYellowBg:  color.BgYellow,
	core.StyleBlueBg:    color.BgBlue,
	core.StyleMagentaBg: color.BgMagenta,
	core.StyleCyanBg:    color.BgCyan,
	core.StyleWhiteBg:   color.BgWhite,

	core.StyleBrightBlack:   color.FgHiBlack,
	core.StyleBrightRed:     color.FgHiRed,
	core.StyleBrightGreen:   color.FgHiGreen,
	core.StyleBrightYellow:  color.FgHiYellow,
	core.StyleBrightBlue:    color.FgHiBlue,
	core.StyleBrightMagenta: color.FgHiMagenta,
	core.StyleBrightCyan:    color.FgHiCyan,
	core.StyleBrightWhite:   color.FgHiWhite,

	core.StyleBrightBlackBg:   color.BgHiBlack,
	core.StyleBrightRedBg:     color.BgHiRed,
	core.StyleBrightGreenBg:   color.BgHiGreen,
	core.StyleBrightYellowBg:  color.BgHiYellow,
	core.StyleBrightBlueBg:    color.BgHiBlue,
	core.StyleBrightMagentaBg: color.BgHiMagenta,
	core.StyleBrightCyanBg:    color.BgHiCyan,
	core.StyleBrightWhiteBg:   color.BgHiWhite,
}

func attributes(rules []core.Style) ([]color.Attribute, error) {
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
