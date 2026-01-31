package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/paths"
)

// Notebook is the command group for notebook operations.
type Notebook struct {
	Context  ContextCmd       `cmd group:"notebook" help:"Manage notebook contexts."`
	Register NotebookRegister `cmd help:"Register a notebook in the global configuration."`
	List     NotebookList     `cmd help:"List registered notebooks."`
	Status   NotebookStatus   `cmd help:"Show the current notebook status."`
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
	
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	tree, err := toml.LoadBytes(configContent)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

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

	for _, c := range contexts {
		if c == targetPath {
			fmt.Printf("Context already exists: %s\n", targetPath)
			return nil
		}
	}

	contexts = append(contexts, targetPath)
	tree.Set("notebook.contexts", contexts)

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

// NotebookRegister registers a notebook path in the global configuration.
type NotebookRegister struct {
	Path string `arg optional type:"path" default:"." help:"Notebook path to register."`
}

func (cmd *NotebookRegister) Run(container *cli.Container) error {
	targetPath, err := filepath.Abs(cmd.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", targetPath)
	}

	// Check for .zk directory
	zkDir := filepath.Join(targetPath, ".zk")
	zkInfo, err := os.Stat(zkDir)
	if err != nil || !zkInfo.IsDir() {
		return fmt.Errorf("not a valid notebook: %s/.zk directory missing", targetPath)
	}

	configPath, err := container.GlobalConfigPath()
	if err != nil {
		return fmt.Errorf("failed to locate global config: %w", err)
	}

	var tree *toml.Tree
	configContent, err := os.ReadFile(configPath)
	if err == nil {
		tree, err = toml.LoadBytes(configContent)
		if err != nil {
			return fmt.Errorf("failed to parse global config file: %w", err)
		}
	} else if os.IsNotExist(err) {
		tree, _ = toml.Load("")
	} else {
		return fmt.Errorf("failed to read global config file: %w", err)
	}

	var notebooks []string
	if tree.Has("notebooks") {
		currentNotebooks := tree.Get("notebooks")
		if cArr, ok := currentNotebooks.([]interface{}); ok {
			for _, c := range cArr {
				if s, ok := c.(string); ok {
					notebooks = append(notebooks, s)
				}
			}
		}
	}

	for _, n := range notebooks {
		if n == targetPath {
			fmt.Printf("Notebook already registered: %s\n", targetPath)
			return nil
		}
	}

	notebooks = append(notebooks, targetPath)
	tree.Set("notebooks", notebooks)

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open global config file for writing: %w", err)
	}
	defer f.Close()

	_, err = tree.WriteTo(f)
	if err != nil {
		return fmt.Errorf("failed to write global config file: %w", err)
	}

	fmt.Printf("Registered notebook: %s\n", targetPath)
	return nil
}

// NotebookList lists registered notebooks.
type NotebookList struct {}

func (cmd *NotebookList) Run(container *cli.Container) error {
	notebooks := container.Config.Notebooks
	if len(notebooks) == 0 {
		fmt.Println("No notebooks registered.")
		return nil
	}

	for _, path := range notebooks {
		exists, _ := paths.Exists(path)
		status := ""
		if !exists {
			status = " (missing)"
		}
		fmt.Printf("%s%s\n", path, status)
	}
	return nil
}

// NotebookStatus shows the current notebook and discovery reason.
type NotebookStatus struct {}

func (cmd *NotebookStatus) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		fmt.Println("No active notebook.")
		return nil
	}

	fmt.Printf("Active Notebook: %s\n", notebook.Path)

	wd, _ := os.Getwd()
	foundPath, found, _ := container.Notebooks.ResolveNotebookFromContext(wd)
	
	// Determine the most likely source
	source := "Manual selection / Environment"
	
	if found && foundPath == notebook.Path {
		source = fmt.Sprintf("Context discovery (project: %s)", wd)
	} else if !container.Config.Notebook.Dir.IsNull() {
		defaultDir, _ := paths.ExpandPath(container.Config.Notebook.Dir.Unwrap())
		if defaultDir == notebook.Path {
			source = "Default configuration"
		}
	}
	
	if wd == notebook.Path {
		source = "Current directory"
	}

	fmt.Printf("Source: %s\n", source)
	return nil
}
