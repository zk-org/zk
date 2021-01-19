package note

import (
	"fmt"
	"time"

	"github.com/mickael-menu/zk/core/style"
	"github.com/mickael-menu/zk/core/templ"
)

var Now = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

// TemplLoaderSpy implements templ.Loader and saves the render contexts
// provided to the templates it creates.
//
// The generated Renderer returns the template used to create them without
// modification.
type TemplLoaderSpy struct {
	Contexts []interface{}
}

func NewTemplLoaderSpy() *TemplLoaderSpy {
	return &TemplLoaderSpy{
		Contexts: make([]interface{}, 0),
	}
}

func (l *TemplLoaderSpy) Load(template string) (templ.Renderer, error) {
	return NewRendererSpy(func(context interface{}) string {
		l.Contexts = append(l.Contexts, context)
		return template
	}), nil
}

func (l *TemplLoaderSpy) LoadFile(path string) (templ.Renderer, error) {
	panic("not implemented")
}

// RendererSpy implements templ.Renderer and saves the provided render contexts.
type RendererSpy struct {
	Result   func(interface{}) string
	Contexts []interface{}
}

func NewRendererSpy(result func(interface{}) string) *RendererSpy {
	return &RendererSpy{
		Contexts: make([]interface{}, 0),
		Result:   result,
	}
}

func NewRendererSpyString(result string) *RendererSpy {
	return &RendererSpy{
		Contexts: make([]interface{}, 0),
		Result:   func(_ interface{}) string { return result },
	}
}

func (m *RendererSpy) Render(context interface{}) (string, error) {
	m.Contexts = append(m.Contexts, context)
	return m.Result(context), nil
}

// StylerMock implements core.Styler by doing the transformation:
// "hello", "red" -> "red(hello)"
type StylerMock struct{}

func (s *StylerMock) Style(text string, rules ...style.Rule) (string, error) {
	for _, rule := range rules {
		text = fmt.Sprintf("%s(%s)", rule, text)
	}
	return text, nil
}
