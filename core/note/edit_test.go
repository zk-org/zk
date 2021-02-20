package note

import (
	"os"
	"testing"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestEditorUsesZkEditorFirst(t *testing.T) {
	os.Setenv("ZK_EDITOR", "zk-editor")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{Config: zk.Config{
		Tool: zk.ToolConfig{
			Editor: opt.NewString("custom-editor"),
		},
	}}

	assert.Equal(t, editor(&zk), opt.NewString("zk-editor"))
}

func TestEditorFallsbackOnUserConfig(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{Config: zk.Config{
		Tool: zk.ToolConfig{
			Editor: opt.NewString("custom-editor"),
		},
	}}

	assert.Equal(t, editor(&zk), opt.NewString("custom-editor"))
}

func TestEditorFallsbackOnVisual(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Setenv("VISUAL", "visual")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NewString("visual"))
}

func TestEditorFallsbackOnEditor(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Unsetenv("VISUAL")
	os.Setenv("EDITOR", "editor")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NewString("editor"))
}

func TestEditorWhenUnset(t *testing.T) {
	os.Unsetenv("ZK_EDITOR")
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
	zk := zk.Zk{}

	assert.Equal(t, editor(&zk), opt.NullString)
}
