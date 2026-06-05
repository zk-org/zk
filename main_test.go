package main

import (
	"os"
	"os/user"
	"testing"

	"github.com/zk-org/zk/internal/cli"
)

func Test_parseDirsPathExpansion(t *testing.T) {
	user, _ := user.Current()
	_ = os.Setenv("FOO", "/home/foo")
	tests := []struct {
		name     string
		args     []string
		wantDirs cli.Dirs
	}{
		{
			name:     "With tilde (working dir)",
			args:     []string{"--working-dir=~/notes"},
			wantDirs: cli.Dirs{WorkingDir: user.HomeDir + "/notes"},
		},
		{
			name:     "With tilde (notebook dir)",
			args:     []string{"--notebook-dir=~/notes"},
			wantDirs: cli.Dirs{NotebookDir: user.HomeDir + "/notes"},
		},
		{
			name:     "With absolute path",
			args:     []string{"--notebook-dir=" + user.HomeDir + "/notes"},
			wantDirs: cli.Dirs{NotebookDir: user.HomeDir + "/notes"},
		},
		{
			name:     "With environment variable in path",
			args:     []string{"--notebook-dir=${FOO}/notes"},
			wantDirs: cli.Dirs{NotebookDir: os.Getenv("FOO") + "/notes"},
		},
		{
			name:     "With environment variable and two string argument",
			args:     []string{"--notebook-dir", "${FOO}/notes"},
			wantDirs: cli.Dirs{NotebookDir: os.Getenv("FOO") + "/notes"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliDirs, _, gotErr := parseDirs(tt.args)
			workDir := cliDirs.WorkingDir
			noteDir := cliDirs.NotebookDir
			if gotErr != nil {
				t.Errorf("parseDirs() failed: %v", gotErr)
			}
			if workDir != tt.wantDirs.WorkingDir {
				t.Errorf("parseDirs() got WorkingDir = %v, want %v", cliDirs.WorkingDir, tt.wantDirs.WorkingDir)
			}
			if noteDir != tt.wantDirs.NotebookDir {
				t.Errorf("parseDirs() got NotebookDir = %v, want %v", cliDirs.NotebookDir, tt.wantDirs.NotebookDir)
			}
		})
	}
}
