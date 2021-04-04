package templ

// Loader parses a given string template.
type Loader interface {
	Load(template string) (Renderer, error)
	LoadFile(path string) (Renderer, error)
}

// Renderer produces a string using a given context.
type Renderer interface {
	Render(context interface{}) (string, error)
}

// RendererFunc is an adapter to use a function as a Renderer.
type RendererFunc func(context interface{}) (string, error)

func (f RendererFunc) Render(context interface{}) (string, error) {
	return f(context)
}

// NullRenderer is a Renderer always returning an empty string.
var NullRenderer = nullRenderer{}

type nullRenderer struct{}

func (t nullRenderer) Render(context interface{}) (string, error) {
	return "", nil
}
