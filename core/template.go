package core

// TemplateLoader parses a given string template.
type TemplateLoader interface {
	Load(template string) (Template, error)
	LoadFile(path string) (Template, error)
}

// Template renders strings using a given context.
type Template interface {
	Render(context interface{}) (string, error)
}

// TemplateFunc is an adapter to use a function as a Template.
type TemplateFunc func(context interface{}) (string, error)

func (f TemplateFunc) Render(context interface{}) (string, error) {
	return f(context)
}

// NullTemplate is a Template returning always an empty string.
var NullTemplate = nullTemplate{}

type nullTemplate struct{}

func (t nullTemplate) Render(context interface{}) (string, error) {
	return "", nil
}
