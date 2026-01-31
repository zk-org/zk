package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoveryPrecedence(t *testing.T) {
	// Setup temporary directory structure
	rootDir, err := os.MkdirTemp("", "zk-discovery-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	// Define paths
	notebookFlag := filepath.Join(rootDir, "notebook-flag")
	notebookCwd := filepath.Join(rootDir, "notebook-cwd")
	notebookEnv := filepath.Join(rootDir, "notebook-env")
	notebookContext := filepath.Join(rootDir, "notebook-context")
	notebookConfig := filepath.Join(rootDir, "notebook-config")

	projectDir := filepath.Join(rootDir, "project")

	// Create directories
	dirs := []string{notebookFlag, notebookCwd, notebookEnv, notebookContext, notebookConfig, projectDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
		// Initialize notebooks (create .zk config to mark them valid if needed,
		// though zk often treats any dir with notes as notebook, strict mode might require init)
		// We'll just create a unique file in each to identify them.
	}

	createFile(t, notebookFlag, "flag.md")
	createFile(t, notebookCwd, "cwd.md")
	createFile(t, notebookEnv, "env.md")
	createFile(t, notebookContext, "context.md")
	createFile(t, notebookConfig, "config.md")

	// Setup Config for Context and Default
	// We need a global config file. zk looks in ~/.config/zk/config.toml or similar.
	// We can point ZK_CONFIG_DIR or similar?
	// Looking at code (not shown yet), zk uses standard config paths.
	// internal/cli/container.go usually handles this.
	// I might need to mock the config file by setting XDG_CONFIG_HOME.

	configDir := filepath.Join(rootDir, "config")
	if err := os.MkdirAll(filepath.Join(configDir, "zk"), 0755); err != nil {
		t.Fatal(err)
	}

	// config.toml content (Global Config)
	// 1. Register notebookContext in the list of known notebooks.
	// 2. Set notebookConfig as the default notebook.
	configContent := `
notebooks = ["` + escapePath(notebookContext) + `"]

[notebook]
dir = "` + escapePath(notebookConfig) + `"
`
	err = os.WriteFile(filepath.Join(configDir, "zk", "config.toml"), []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Notebook Context Config
	// This notebook claims ownership of projectDir context.
	notebookContextConfigDir := filepath.Join(notebookContext, ".zk")
	if err := os.MkdirAll(notebookContextConfigDir, 0755); err != nil {
		t.Fatal(err)
	}
	notebookContextContent := `
[notebook]
contexts = ["` + escapePath(projectDir) + `"]
`
	err = os.WriteFile(filepath.Join(notebookContextConfigDir, "config.toml"), []byte(notebookContextContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build zk binary
	zkBin := filepath.Join(rootDir, "zk")
	// Assuming test is run from project root or tests/ dir.
	// We'll try to find main.go
	mainPath, err := filepath.Abs("../main.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// Try current dir (if running from root)
		mainPath, err = filepath.Abs("main.go")
		if err != nil || os.IsNotExist(err) {
			t.Fatal("Could not find main.go")
		}
	}

	buildCmd := exec.Command("go", "build", "-tags", "fts5", "-o", zkBin, mainPath)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build zk: %v\n%s", err, out)
	}

	// Initialize notebooks
	// notebookContext is manually initialized with specific config
	for _, d := range []string{notebookFlag, notebookCwd, notebookEnv, notebookConfig} {
		initCmd := exec.Command(zkBin, "init", d, "--no-input")
		if out, err := initCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to init notebook %s: %v\n%s", d, err, out)
		}
	}

	// Helper to run zk
	runZk := func(t *testing.T, workDir string, envVarNotebook string, extraArgs ...string) string {
		cmdArgs := []string{"list", "--format", "{{path}}", "--quiet"}
		cmdArgs = append(cmdArgs, extraArgs...)

		cmd := exec.Command(zkBin, cmdArgs...)
		cmd.Dir = workDir

		env := os.Environ()
		// Filter out existing ZK envs to avoid pollution
		cleanEnv := []string{}
		for _, e := range env {
			if !strings.HasPrefix(e, "ZK_") && !strings.HasPrefix(e, "XDG_CONFIG_HOME=") {
				cleanEnv = append(cleanEnv, e)
			}
		}

		cleanEnv = append(cleanEnv, "XDG_CONFIG_HOME="+configDir)
		if envVarNotebook != "" {
			cleanEnv = append(cleanEnv, "ZK_NOTEBOOK_DIR="+envVarNotebook)
		}

		cmd.Env = cleanEnv

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("zk command failed: %v\nOutput: %s", err, out)
		}
		return strings.TrimSpace(string(out))
	}

	// Test 1: Flag Precedence (Highest)
	// Run from projectDir, with Env set, Config set. Flag should win.
	t.Run("Flag Precedence", func(t *testing.T) {
		out := runZk(t, projectDir, notebookEnv, "--notebook-dir", notebookFlag)
		if !strings.Contains(out, "flag.md") {
			t.Errorf("Expected flag.md, got: %s", out)
		}
	})

	// Test 2: CWD Precedence
	// Run FROM notebookCwd. Env set. Config set. CWD should win (if it's a notebook).
	t.Run("CWD Precedence", func(t *testing.T) {
		out := runZk(t, notebookCwd, notebookEnv)
		if !strings.Contains(out, "cwd.md") {
			t.Errorf("Expected cwd.md, got: %s", out)
		}
	})

	// Test 3: Env Var Precedence
	// Run from projectDir (context match). Env set. Env should win over context.
	t.Run("Env Precedence", func(t *testing.T) {
		out := runZk(t, projectDir, notebookEnv)
		if !strings.Contains(out, "env.md") {
			t.Errorf("Expected env.md, got: %s", out)
		}
	})

	// Test 4: Context Precedence
	// Run from projectDir. No Env. Context should win over Default Config.
	t.Run("Context Precedence", func(t *testing.T) {
		out := runZk(t, projectDir, "")
		if !strings.Contains(out, "context.md") {
			t.Errorf("Expected context.md, got: %s", out)
		}
	})

	// Test 5: Default Config Precedence (Lowest)
	// Run from rootDir (no context). No Env. Default Config should be used.
	t.Run("Default Config Precedence", func(t *testing.T) {
		out := runZk(t, rootDir, "")
		if !strings.Contains(out, "config.md") {
			t.Errorf("Expected config.md, got: %s", out)
		}
	})

	// Test 6: Notebook Context Add Command
	// 1. Create a new notebook and project dir
	notebookAdd := filepath.Join(rootDir, "notebook-add")
	projectAdd := filepath.Join(rootDir, "project-add")
	if err := os.MkdirAll(notebookAdd, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectAdd, 0755); err != nil {
		t.Fatal(err)
	}
	createFile(t, notebookAdd, "add.md")

	// Init notebook-add
	initCmd := exec.Command(zkBin, "init", notebookAdd, "--no-input")
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init notebook %s: %v\n%s", notebookAdd, err, out)
	}

	// 2. Run `zk notebook context add <projectAdd>` from inside notebookAdd
	t.Run("Context Add Command", func(t *testing.T) {
		// Run add command
		cmd := exec.Command(zkBin, "notebook", "context", "add", projectAdd)
		cmd.Dir = notebookAdd
		cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir) // Use same config env

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to add context: %v\n%s", err, out)
		}

		// 3. Verify config file content
		configPath := filepath.Join(notebookAdd, ".zk", "config.toml")
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), projectAdd) {
			t.Errorf("Config does not contain added path. Content:\n%s", content)
		}

		// 4. Verify discovery works from projectAdd
		// We need to register notebookAdd in global config so discovery sees it.
		// We can append it to the global config file we created earlier.
		// But wait, the global config is static in this test.
		// We can overwrite it or assume discovery works if we use `zk list`?
		// Discovery iterates `notebooks` list in global config.
		// So we MUST add `notebookAdd` to `config.toml` (global).

		// Update global config to include notebookAdd
		fullConfigContent := `
notebooks = ["` + escapePath(notebookContext) + `", "` + escapePath(notebookAdd) + `"]

[notebook]
dir = "` + escapePath(notebookConfig) + `"
`
		err = os.WriteFile(filepath.Join(configDir, "zk", "config.toml"), []byte(fullConfigContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Now check discovery from projectAdd
		listOut := runZk(t, projectAdd, "")
		if !strings.Contains(listOut, "add.md") {
			t.Errorf("Expected add.md when running from project-add, got: %s", listOut)
		}
	})
}

func createFile(t *testing.T, dir, name string) {
	err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func escapePath(p string) string {
	return strings.ReplaceAll(p, "\\", "\\\\")
}
