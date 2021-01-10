package note

import (
	"os"
	"testing"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/test/assert"
	"github.com/mickael-menu/zk/util/opt"
)

func TestEditorUsesUserConfigFirst(t *testing.T) {
	os.Setenv("VISUAL", "editor")
	zk := zk.Zk{Config: zk.Config{Editor: opt.NewString("custom-editor")}}

	assert.Equal(t, editor(&zk), opt.NewString("custom-editor"))
}

func TestEditorFallsbackOnVisual(t *testing.T) {
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NewString("visual"))
}

func TestEditorFallsbackOnEditor(t *testing.T) {
	os.Unsetenv("VISUAL")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NewString("editor"))
}

func TestEditorWhenUnset(t *testing.T) {
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NullString)
}
