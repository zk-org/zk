package core

// templateLoaderMock implements an in-memory TemplateLoader for testing.
type templateLoaderMock struct {
	templates     map[string]*templateSpy
	fileTemplates map[string]*templateSpy
	styler        Styler
}

func newTemplateLoaderMock() *templateLoaderMock {
	return &templateLoaderMock{
		templates:     map[string]*templateSpy{},
		fileTemplates: map[string]*templateSpy{},
		styler:        &stylerMock{},
	}
}

func (m *templateLoaderMock) Spy(template string, result func(context any) string) *templateSpy {
	spy := newTemplateSpy(result)
	spy.styler = m.styler
	m.templates[template] = spy
	return spy
}

func (m *templateLoaderMock) SpyString(content string) *templateSpy {
	spy := newTemplateSpyString(content)
	spy.styler = m.styler
	m.templates[content] = spy
	return spy
}

func (m *templateLoaderMock) SpyFile(path string, content string) *templateSpy {
	spy := newTemplateSpyString(content)
	spy.styler = m.styler
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
	Result   func(any) string
	Contexts []any
	styler   Styler
}

func newTemplateSpy(result func(any) string) *templateSpy {
	return &templateSpy{
		Contexts: make([]any, 0),
		Result:   result,
	}
}

func newTemplateSpyString(result string) *templateSpy {
	return &templateSpy{
		Contexts: make([]any, 0),
		Result:   func(_ any) string { return result },
	}
}

func (m *templateSpy) Styler() Styler {
	return m.styler
}

func (m *templateSpy) Render(context any) (string, error) {
	m.Contexts = append(m.Contexts, context)
	return m.Result(context), nil
}
