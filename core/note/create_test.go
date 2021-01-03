package note

import (
	"fmt"
	"testing"

	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
)

func TestCreate(t *testing.T) {
	filenameTemplate := spyTemplateString("filename")
	bodyTemplate := spyTemplateString("body")

	res, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.DirConfig{
					Extension: "md",
					Extra: map[string]string{
						"hello": "world",
					},
				},
			},
			Title:   opt.NewString("Note title"),
			Content: opt.NewString("Note content"),
		},
		createDeps{
			filenameTemplate: &filenameTemplate,
			bodyTemplate:     &bodyTemplate,
			genId:            func() string { return "abc" },
			validatePath:     func(path string) (bool, error) { return true, nil },
		},
	)

	// Check the created note.
	assert.Nil(t, err)
	assert.Equal(t, res, &createdNote{
		path:    "/test/log/filename.md",
		content: "body",
	})

	// Check that the templates received the proper render contexts.
	assert.Equal(t, filenameTemplate.Contexts, []renderContext{{
		ID:      "abc",
		Title:   "Note title",
		Content: "Note content",
		Dir:     "log",
		Extra: map[string]string{
			"hello": "world",
		},
	}})
	assert.Equal(t, bodyTemplate.Contexts, []renderContext{{
		ID:           "abc",
		Title:        "Note title",
		Content:      "Note content",
		Dir:          "log",
		Filename:     "filename.md",
		FilenameStem: "filename",
		Extra: map[string]string{
			"hello": "world",
		},
	}})
}

func TestCreateTriesUntilValidPath(t *testing.T) {
	filenameTemplate := spyTemplate(func(context renderContext) string {
		return context.ID
	})
	bodyTemplate := spyTemplateString("body")

	res, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.DirConfig{
					Extension: "md",
				},
			},
			Title: opt.NewString("Note title"),
		},
		createDeps{
			filenameTemplate: &filenameTemplate,
			bodyTemplate:     &bodyTemplate,
			genId:            incrementingID(),
			validatePath: func(path string) (bool, error) {
				return path == "/test/log/3.md", nil
			},
		},
	)

	// Check the created note.
	assert.Nil(t, err)
	assert.Equal(t, res, &createdNote{
		path:    "/test/log/3.md",
		content: "body",
	})

	assert.Equal(t, filenameTemplate.Contexts, []renderContext{
		{
			ID:    "1",
			Title: "Note title",
			Dir:   "log",
		},
		{
			ID:    "2",
			Title: "Note title",
			Dir:   "log",
		},
		{
			ID:    "3",
			Title: "Note title",
			Dir:   "log",
		},
	})
}

func TestCreateErrorWhenNoValidPaths(t *testing.T) {
	_, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.DirConfig{
					Extension: "md",
				},
			},
		},
		createDeps{
			filenameTemplate: core.TemplateFunc(func(context interface{}) (string, error) {
				return "filename", nil
			}),
			bodyTemplate: core.NullTemplate,
			genId:        func() string { return "abc" },
			validatePath: func(path string) (bool, error) { return false, nil },
		},
	)

	assert.Err(t, err, "/test/log/filename.md: note already exists")
}

func spyTemplate(result func(renderContext) string) TemplateSpy {
	return TemplateSpy{
		Contexts: make([]renderContext, 0),
		Result:   result,
	}
}

func spyTemplateString(result string) TemplateSpy {
	return TemplateSpy{
		Contexts: make([]renderContext, 0),
		Result:   func(_ renderContext) string { return result },
	}
}

type TemplateSpy struct {
	Result   func(renderContext) string
	Contexts []renderContext
}

func (m *TemplateSpy) Render(context interface{}) (string, error) {
	renderContext := context.(renderContext)
	m.Contexts = append(m.Contexts, renderContext)
	return m.Result(renderContext), nil
}

func incrementingID() func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("%d", i)
	}
}
