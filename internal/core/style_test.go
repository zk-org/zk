package core

import "fmt"

// stylerMock implements core.Styler by doing the transformation:
// "hello", "red" -> "red(hello)"
type stylerMock struct{}

func (s *stylerMock) Style(text string, rules ...Style) (string, error) {
	return s.MustStyle(text, rules...), nil
}

func (s *stylerMock) MustStyle(text string, rules ...Style) string {
	for _, rule := range rules {
		text = fmt.Sprintf("%s(%s)", rule, text)
	}
	return text
}
