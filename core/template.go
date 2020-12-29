package core

// Template renders strings using a given context.
type Template interface {
	Render(context interface{}) (string, error)
}

// TemplateLoader parses a given string template.
type TemplateLoader interface {
	Load(template string) (Template, error)
	LoadFile(path string) (Template, error)
}
