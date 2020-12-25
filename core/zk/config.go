package zk

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mickael-menu/zk/util/errors"
)

// Config holds the user configuration of a slip box.
type Config struct {
	root rootConfig
}

func (c1 Config) Equal(c2 Config) bool {
	return cmp.Equal(c1.root, c2.root)
}

type rootConfig struct {
	Editor    string            `hcl:"editor,optional"`
	Extension string            `hcl:"extension,optional"`
	Filename  string            `hcl:"filename,optional"`
	Template  string            `hcl:"template,optional"`
	RandomID  *randomIDConfig   `hcl:"random_id,block"`
	Dirs      []dirConfig       `hcl:"dir,block"`
	Ext       map[string]string `hcl:"ext,optional"`
}

type randomIDConfig struct {
	Charset string `hcl:"charset,optional"`
	Length  int    `hcl:"length,optional"`
	Case    string `hcl:"case,optional"`
}

type dirConfig struct {
	Dir      string            `hcl:"dir,label"`
	Filename string            `hcl:"filename,optional"`
	Template string            `hcl:"template,optional"`
	Ext      map[string]string `hcl:"ext,optional"`
}

// parseConfig creates a new Config instance from its HCL representation.
func parseConfig(content []byte) (*Config, error) {
	var root rootConfig
	err := hclsimple.Decode(".zk/config.hcl", content, nil, &root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}
	return &Config{root}, nil
}
