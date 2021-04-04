package lsp

import "strings"

type workspace struct {
	folders []string
}

func newWorkspace() *workspace {
	return &workspace{
		folders: []string{},
	}
}

func (w *workspace) addFolder(folder string) {
	folder = strings.TrimPrefix(folder, "file://")
	w.folders = append(w.folders, folder)
}

func (w *workspace) removeFolder(folder string) {
	folder = strings.TrimPrefix(folder, "file://")
	for i, f := range w.folders {
		if f == folder {
			w.folders = append(w.folders[:i], w.folders[i+1:]...)
			break
		}
	}
}
