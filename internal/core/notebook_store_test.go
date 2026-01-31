package core

import (
	"testing"
)

func TestResolveNotebookFromContext(t *testing.T) {
	fs := newFileStorageMock("/home/user", []string{
		"/home/user/notes",
		"/home/user/notes/.zk",
		"/home/user/personal",
		"/home/user/personal/.zk",
	})

	// Notebook 1: matches /home/user/project
	fs.Write("/home/user/notes/.zk/config.toml", []byte(`
[notebook]
contexts = ["/home/user/project"]
`))

	// Notebook 2: matches /home/user/work
	fs.Write("/home/user/personal/.zk/config.toml", []byte(`
[notebook]
contexts = ["/home/user/work"]
`))

	config := NewDefaultConfig()
	config.Notebooks = []string{"/home/user/notes", "/home/user/personal"}

	store := NewNotebookStore(config, NotebookStorePorts{
		FS: fs,
	})

	tests := []struct {
		cwd      string
		expected string
		found    bool
	}{
		{
			cwd:      "/home/user/project",
			expected: "/home/user/notes",
			found:    true,
		},
		{
			cwd:      "/home/user/project/subdir",
			expected: "/home/user/notes",
			found:    true,
		},
		{
			cwd:      "/home/user/work",
			expected: "/home/user/personal",
			found:    true,
		},
		{
			cwd:      "/home/user/other",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.cwd, func(t *testing.T) {
			path, found, err := store.ResolveNotebookFromContext(tt.cwd)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if found && path != tt.expected {
				t.Errorf("expected path=%v, got %v", tt.expected, path)
			}
		})
	}
}
