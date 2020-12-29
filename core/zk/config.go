package zk

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mickael-menu/zk/util/errors"
)

// config holds the user configuration of a slip box.
type config struct {
	Filename string            `hcl:"filename,optional"`
	Template string            `hcl:"template,optional"`
	ID       *idConfig         `hcl:"id,block"`
	Editor   string            `hcl:"editor,optional"`
	Dirs     []dirConfig       `hcl:"dir,block"`
	Extra    map[string]string `hcl:"extra,optional"`
}

type dirConfig struct {
	Dir      string            `hcl:"dir,label"`
	Filename string            `hcl:"filename,optional"`
	Template string            `hcl:"template,optional"`
	ID       *idConfig         `hcl:"id,block"`
	Extra    map[string]string `hcl:"extra,optional"`
}

type idConfig struct {
	Charset string `hcl:"charset,optional"`
	Length  int    `hcl:"length,optional"`
	Case    string `hcl:"case,optional"`
}

// parseConfig creates a new Config instance from its HCL representation.
func parseConfig(content []byte) (*config, error) {
	var config config
	err := hclsimple.Decode(".zk/config.hcl", content, nil, &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}
	return &config, nil
}
