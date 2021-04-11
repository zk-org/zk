package core

/*

var Now = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

func TestNotebookNewNote(t *testing.T) {
	filenameTemplate := NewTemplateSpyString("filename")
	bodyTemplate := NewTemplateSpyString("body")

	// notebook := NewNotebook(path string, config Config, ports NotebookPorts)

	res, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.GroupConfig{
					Note: zk.NoteConfig{
						Extension: "md",
					},
					Extra: map[string]string{
						"hello": "world",
					},
				},
			},
			Title:   opt.NewString("Note title"),
			Content: opt.NewString("Note content"),
		},
		createDeps{
			filenameTemplate: filenameTemplate,
			bodyTemplate:     bodyTemplate,
			genId:            func() string { return "abc" },
			validatePath:     func(path string) (bool, error) { return true, nil },
			now:              Now,
		},
	)

	// Check the created note.
	assert.Nil(t, err)
	assert.Equal(t, res, &createdNote{
		path:    "/test/log/filename.md",
		content: "body",
	})

	// Check that the templates received the proper render contexts.
	assert.Equal(t, filenameTemplate.Contexts, []interface{}{renderContext{
		ID:      "abc",
		Title:   "Note title",
		Content: "Note content",
		Dir:     "log",
		Extra: map[string]string{
			"hello": "world",
		},
		Now: Now,
		Env: os.Env(),
	}})
	assert.Equal(t, bodyTemplate.Contexts, []interface{}{renderContext{
		ID:           "abc",
		Title:        "Note title",
		Content:      "Note content",
		Dir:          "log",
		Filename:     "filename.md",
		FilenameStem: "filename",
		Extra: map[string]string{
			"hello": "world",
		},
		Now: Now,
		Env: os.Env(),
	}})
}

func TestNotebookNewNoteTriesUntilValidPath(t *testing.T) {
	filenameTemplate := NewTemplateSpy(func(context interface{}) string {
		return context.(renderContext).ID
	})
	bodyTemplate := NewTemplateSpyString("body")

	res, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.GroupConfig{
					Note: zk.NoteConfig{
						Extension: "md",
					},
				},
			},
			Title: opt.NewString("Note title"),
		},
		createDeps{
			filenameTemplate: filenameTemplate,
			bodyTemplate:     bodyTemplate,
			genId:            incrementingID(),
			validatePath: func(path string) (bool, error) {
				return path == "/test/log/3.md", nil
			},
			now: Now,
		},
	)

	// Check the created note.
	assert.Nil(t, err)
	assert.Equal(t, res, &createdNote{
		path:    "/test/log/3.md",
		content: "body",
	})

	assert.Equal(t, filenameTemplate.Contexts, []interface{}{
		renderContext{
			ID:    "1",
			Title: "Note title",
			Dir:   "log",
			Now:   Now,
			Env:   os.Env(),
		},
		renderContext{
			ID:    "2",
			Title: "Note title",
			Dir:   "log",
			Now:   Now,
			Env:   os.Env(),
		},
		renderContext{
			ID:    "3",
			Title: "Note title",
			Dir:   "log",
			Now:   Now,
			Env:   os.Env(),
		},
	})
}

func TestCreateErrorWhenNoValidPaths(t *testing.T) {
	_, err := create(
		CreateOpts{
			Dir: zk.Dir{
				Name: "log",
				Path: "/test/log",
				Config: zk.GroupConfig{
					Note: zk.NoteConfig{
						Extension: "md",
					},
				},
			},
		},
		createDeps{
			filenameTemplate: TemplateFunc(func(context interface{}) (string, error) {
				return "filename", nil
			}),
			bodyTemplate: NullTemplate,
			genId:        func() string { return "abc" },
			validatePath: func(path string) (bool, error) { return false, nil },
			now:          Now,
		},
	)

	assert.Err(t, err, "/test/log/filename.md: note already exists")
}

// incrementingID returns a generator of incrementing string ID.
func incrementingID() func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("%d", i)
	}
}
*/
