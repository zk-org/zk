package core

// templateLoaderMock implements an in-memory TemplateLoader for testing.
type templateLoaderMock struct {
	templates     map[string]*templateSpy
	fileTemplates map[string]*templateSpy
}

func newTemplateLoaderMock() *templateLoaderMock {
	return &templateLoaderMock{
		templates:     map[string]*templateSpy{},
		fileTemplates: map[string]*templateSpy{},
	}
}

func (m *templateLoaderMock) Spy(template string, result func(context interface{}) string) *templateSpy {
	spy := newTemplateSpy(result)
	m.templates[template] = spy
	return spy
}

func (m *templateLoaderMock) SpyString(content string) *templateSpy {
	spy := newTemplateSpyString(content)
	m.templates[content] = spy
	return spy
}

func (m *templateLoaderMock) SpyFile(path string, content string) *templateSpy {
	spy := newTemplateSpyString(content)
	m.fileTemplates[path] = spy
	return spy
}

func (l *templateLoaderMock) LoadTemplate(template string) (Template, error) {
	tpl, ok := l.templates[template]
	if !ok {
		panic("no template spy for content: " + template)
	}
	return tpl, nil
}

func (l *templateLoaderMock) LoadTemplateAt(path string) (Template, error) {
	tpl, ok := l.fileTemplates[path]
	if !ok {
		panic("no template spy for path: " + path)
	}
	return tpl, nil
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
