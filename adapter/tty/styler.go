package tty

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/mickael-menu/zk/core/style"
)

// Styler is a text styler using ANSI escape codes to be used with a TTY.
type Styler struct{}

func NewStyler() *Styler {
	return &Styler{}
}

func (s *Styler) Style(text string, rules ...style.Rule) (string, error) {
	if text == "" {
		return text, nil
	}
	attrs, err := s.attributes(expandThemeAliases(rules))
	if err != nil {
		return "", err
	}
	if len(attrs) == 0 {
		return text, nil
	}
	return color.New(attrs...).Sprint(text), nil
}

func (s *Styler) MustStyle(text string, rules ...style.Rule) string {
	text, err := s.Style(text, rules...)
	if err != nil {
		panic(err.Error())
	}
	return text
}

// FIXME: User config
var themeAliases = map[style.Rule][]style.Rule{
	"title": {"bold", "yellow"},
	"path":  {"underline", "cyan"},
	"term":  {"red"},
}

func expandThemeAliases(rules []style.Rule) []style.Rule {
	expanded := make([]style.Rule, 0)
	for _, rule := range rules {
		aliases, ok := themeAliases[rule]
		if ok {
			aliases = expandThemeAliases(aliases)
			for _, alias := range aliases {
				expanded = append(expanded, alias)
			}

		} else {
			expanded = append(expanded, rule)
		}
	}

	return expanded
}

var attrsMapping = map[style.Rule]color.Attribute{
	style.RuleBold:          color.Bold,
	style.RuleFaint:         color.Faint,
	style.RuleItalic:        color.Italic,
	style.RuleUnderline:     color.Underline,
	style.RuleBlink:         color.BlinkSlow,
	style.RuleReverse:       color.ReverseVideo,
	style.RuleHidden:        color.Concealed,
	style.RuleStrikethrough: color.CrossedOut,

	style.RuleBlack:   color.FgBlack,
	style.RuleRed:     color.FgRed,
	style.RuleGreen:   color.FgGreen,
	style.RuleYellow:  color.FgYellow,
	style.RuleBlue:    color.FgBlue,
	style.RuleMagenta: color.FgMagenta,
	style.RuleCyan:    color.FgCyan,
	style.RuleWhite:   color.FgWhite,

	style.RuleBlackBg:   color.BgBlack,
	style.RuleRedBg:     color.BgRed,
	style.RuleGreenBg:   color.BgGreen,
	style.RuleYellowBg:  color.BgYellow,
	style.RuleBlueBg:    color.BgBlue,
	style.RuleMagentaBg: color.BgMagenta,
	style.RuleCyanBg:    color.BgCyan,
	style.RuleWhiteBg:   color.BgWhite,

	style.RuleBrightBlack:   color.FgHiBlack,
	style.RuleBrightRed:     color.FgHiRed,
	style.RuleBrightGreen:   color.FgHiGreen,
	style.RuleBrightYellow:  color.FgHiYellow,
	style.RuleBrightBlue:    color.FgHiBlue,
	style.RuleBrightMagenta: color.FgHiMagenta,
	style.RuleBrightCyan:    color.FgHiCyan,
	style.RuleBrightWhite:   color.FgHiWhite,

	style.RuleBrightBlackBg:   color.BgHiBlack,
	style.RuleBrightRedBg:     color.BgHiRed,
	style.RuleBrightGreenBg:   color.BgHiGreen,
	style.RuleBrightYellowBg:  color.BgHiYellow,
	style.RuleBrightBlueBg:    color.BgHiBlue,
	style.RuleBrightMagentaBg: color.BgHiMagenta,
	style.RuleBrightCyanBg:    color.BgHiCyan,
	style.RuleBrightWhiteBg:   color.BgHiWhite,
}

func (s *Styler) attributes(rules []style.Rule) ([]color.Attribute, error) {
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
