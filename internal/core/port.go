package core

// FileStorage is a port providing read and write access to a file storage.
type FileStorage interface {

	// Abs makes the given file path absolute if needed, using the FileStorage
	// working directory.
	Abs(path string) (string, error)

	// FileExists returns whether a file exists at the given file path.
	FileExists(path string) (bool, error)

	// DirExists returns whether a directory exists at the given file path.
	DirExists(path string) (bool, error)

	// Read returns the bytes content of the file at the given file path.
	Read(path string) ([]byte, error)

	// Write creates or overwrite the content at the given file path, creating
	// any intermediate directories if needed.
	Write(path string, content []byte) error
}

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
	// The path may be relative to template directories registered to the loader.
	LoadTemplateAt(path string) (Template, error)
}

// TemplateLoaderFactory creates a new instance of an implementation of the
// TemplateLoader port.
type TemplateLoaderFactory func(language string, lookupPaths []string) (TemplateLoader, error)
