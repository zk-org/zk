package cmd

import (
	"fmt"
	"path/filepath"
	"os"

	"github.com/zk-org/zk/internal/cli"
	"github.com/pelletier/go-toml"
)

// Notebook is the command group for notebook operations.
type Notebook struct {
	Context ContextCmd `cmd group:"notebook" help:"Manage notebook contexts."`
}

// ContextCmd is the command group for context operations.
type ContextCmd struct {
	Add ContextAdd `cmd help:"Associate a project directory with the current notebook."`
}

// ContextAdd adds a directory as a context for the current notebook.
type ContextAdd struct {
	Path string `arg optional type:"path" default:"." help:"Project directory to associate."`
}

func (cmd *ContextAdd) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return fmt.Errorf("this command must be run within a notebook: %w", err)
	}

	targetPath, err := filepath.Abs(cmd.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	configPath := filepath.Join(notebook.Path, ".zk", "config.toml")
	
	// Load the config file using toml.Tree to preserve as much as possible
	// (though go-toml v1 might still struggle with comments if we rewrite the whole tree, 
	// but using Tree is better than struct marshaling)
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	tree, err := toml.LoadBytes(configContent)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get current contexts
	var contexts []string
	if tree.Has("notebook.contexts") {
		currentContexts := tree.Get("notebook.contexts")
		if cArr, ok := currentContexts.([]interface{}); ok {
			for _, c := range cArr {
				if s, ok := c.(string); ok {
					contexts = append(contexts, s)
				}
			}
		}
	}

	// Check if already exists
	for _, c := range contexts {
		if c == targetPath {
			fmt.Printf("Context already exists: %s\n", targetPath)
			return nil
		}
	}

	// Add new context
	contexts = append(contexts, targetPath)
	
	// Update tree
	// go-toml v1 Set needs the full key path.
	// We might need to ensure [notebook] table exists, but it usually does.
	tree.Set("notebook.contexts", contexts)

	// Write back
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	_, err = tree.WriteTo(f)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Added context: %s\n", targetPath)
	return nil
}
