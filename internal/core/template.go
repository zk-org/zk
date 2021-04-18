package core

// Template produces a string using a given context.
type Template interface {

	// Styler used to format the templates content.
	Styler() Styler

	// Render generates this template using the given variable context.
	Render(context interface{}) (string, error)
}

// TemplateFunc is an adapter to use a function as a Template.
type TemplateFunc func(context interface{}) (string, error)

// Styler implements Template.
func (f TemplateFunc) Styler() Styler {
	return NullStyler
}

// Render implements Template.
func (f TemplateFunc) Render(context interface{}) (string, error) {
	return f(context)
}

// NullTemplate is a Template always returning an empty string.
var NullTemplate = nullTemplate{}

type nullTemplate struct{}

func (t nullTemplate) Styler() Styler {
	return NullStyler
}

func (t nullTemplate) Render(context interface{}) (string, error) {
	return "", nil
}

// TemplateLoader parses a string into a new Template instance.
type TemplateLoader interface {
	// LoadTemplate creates a Template instance from a string template.
	LoadTemplate(template string) (Template, error)

	// LoadTemplate creates a Template instance from a template stored in the
	// file at the given path.
	// The path may be relative to template directories registered to the loader.
	LoadTemplateAt(path string) (Template, error)
}

// TemplateLoaderFactory creates a new instance of an implementation of the
// TemplateLoader port.
type TemplateLoaderFactory func(language string) (TemplateLoader, error)

// NullTemplateLoader a TemplateLoader always returning a NullTemplate.
var NullTemplateLoader = nullTemplateLoader{}

type nullTemplateLoader struct{}

func (t nullTemplateLoader) LoadTemplate(template string) (Template, error) {
	return &NullTemplate, nil
}

func (t nullTemplateLoader) LoadTemplateAt(path string) (Template, error) {
	return &NullTemplate, nil
}
