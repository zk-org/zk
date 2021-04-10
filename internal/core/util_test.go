package core

import (
	"fmt"
	"time"
)

var Now = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

// TemplateLoaderSpy implements TemplateLoader and saves the render
// contexts provided to the templates it creates.
//
// The generated Template returns the template used to create them without
// modification.
type TemplateLoaderSpy struct {
	Contexts []interface{}
}

func NewTemplateLoaderSpy() *TemplateLoaderSpy {
	return &TemplateLoaderSpy{
		Contexts: make([]interface{}, 0),
	}
}

func (l *TemplateLoaderSpy) LoadTemplate(template string) (Template, error) {
	return NewTemplateSpy(func(context interface{}) string {
		l.Contexts = append(l.Contexts, context)
		return template
	}), nil
}

func (l *TemplateLoaderSpy) LoadTemplateAt(path string) (Template, error) {
	panic("not implemented")
}

// TemplateSpy implements Template and saves the provided render contexts.
type TemplateSpy struct {
	Result   func(interface{}) string
	Contexts []interface{}
}

func NewTemplateSpy(result func(interface{}) string) *TemplateSpy {
	return &TemplateSpy{
		Contexts: make([]interface{}, 0),
		Result:   result,
	}
}

func NewTemplateSpyString(result string) *TemplateSpy {
	return &TemplateSpy{
		Contexts: make([]interface{}, 0),
		Result:   func(_ interface{}) string { return result },
	}
}

func (m *TemplateSpy) Styler() Styler {
	return NullStyler
}

func (m *TemplateSpy) Render(context interface{}) (string, error) {
	m.Contexts = append(m.Contexts, context)
	return m.Result(context), nil
}

// StylerMock implements core.Styler by doing the transformation:
// "hello", "red" -> "red(hello)"
type StylerMock struct{}

func (s *StylerMock) Style(text string, rules ...Style) (string, error) {
	return s.MustStyle(text, rules...), nil
}

func (s *StylerMock) MustStyle(text string, rules ...Style) string {
	for _, rule := range rules {
		text = fmt.Sprintf("%s(%s)", rule, text)
	}
	return text
}
