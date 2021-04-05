package core

// Template produces a string using a given context.
type Template interface {
	Render(context interface{}) (string, error)
}

// TemplateFunc is an adapter to use a function as a Template.
type TemplateFunc func(context interface{}) (string, error)

// Render implements Template.
func (f TemplateFunc) Render(context interface{}) (string, error) {
	return f(context)
}

// NullTemplate is a Template always returning an empty string.
var NullTemplate = nullTemplate{}

type nullTemplate struct{}

func (t nullTemplate) Render(context interface{}) (string, error) {
	return "", nil
}

// TemplateLoader parses a string into a new Template instance.
type TemplateLoader interface {

	// LoadTemplate creates a Template instance from a string template.
	LoadTemplate(template string) (Template, error)

	// LoadTemplate creates a Template instance from a template stored in the
	// file at the given path.
	LoadTemplateAt(path string) (Template, error)
}
