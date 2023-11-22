package editor

import (
	"os"
	"testing"

	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestEditorUsesZkEditorFirst(t *testing.T) {
	os.Setenv("ZK_EDITOR", "zk-editor")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")

	editor, err := NewEditor(opt.NewString("custom-editor"), "")
	assert.Nil(t, err)
	assert.Equal(t, editor.editor, "zk-editor")
}

func TestEditorFallsbackOnUserConfig(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")

	editor, err := NewEditor(opt.NewString("custom-editor"), "")
	assert.Nil(t, err)
	assert.Equal(t, editor.editor, "custom-editor")
}

func TestEditorFallsbackOnVisual(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")

	editor, err := NewEditor(opt.NullString, "")
	assert.Nil(t, err)
	assert.Equal(t, editor.editor, "visual")
}

func TestEditorFallsbackOnEditor(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Unsetenv("VISUAL")
	os.Setenv("EDITOR", "editor")

	editor, err := NewEditor(opt.NullString, "")
	assert.Nil(t, err)
	assert.Equal(t, editor.editor, "editor")
}

func TestEditorFailsWhenUnset(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")

	editor, err := NewEditor(opt.NullString, "")
	assert.Err(t, err, "no editor set in config")
	assert.Nil(t, editor)
}
