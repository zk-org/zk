package core

// templateLoaderSpy implements TemplateLoader and saves the render
// contexts provided to the templates it creates.
//
// The generated Template returns the template used to create them without
// modification.
type templateLoaderSpy struct {
	Contexts []interface{}
}

func newTemplateLoaderSpy() *templateLoaderSpy {
	return &templateLoaderSpy{
		Contexts: make([]interface{}, 0),
	}
}

func (l *templateLoaderSpy) LoadTemplate(template string) (Template, error) {
	return newTemplateSpy(func(context interface{}) string {
		l.Contexts = append(l.Contexts, context)
		return template
	}), nil
}

func (l *templateLoaderSpy) LoadTemplateAt(path string) (Template, error) {
	panic("not implemented")
}

// templateSpy implements Template and saves the provided render contexts.
type templateSpy struct {
	Result   func(interface{}) string
	Contexts []interface{}
}

func newTemplateSpy(result func(interface{}) string) *templateSpy {
	return &templateSpy{
		Contexts: make([]interface{}, 0),
		Result:   result,
	}
}

func newTemplateSpyString(result string) *templateSpy {
	return &templateSpy{
		Contexts: make([]interface{}, 0),
		Result:   func(_ interface{}) string { return result },
	}
}

func (m *templateSpy) Styler() Styler {
	return NullStyler
}

func (m *templateSpy) Render(context interface{}) (string, error) {
	m.Contexts = append(m.Contexts, context)
	return m.Result(context), nil
}
