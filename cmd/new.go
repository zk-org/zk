package cmd

import "github.com/mickael-menu/zk/core/zk"

type New struct {
	Directory string `arg optional name:"directory" default:"."`
}

func (cmd *New) Run() error {
	return zk.Open(cmd.Directory)
}
