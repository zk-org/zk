package cmd

import (
	"fmt"

	"github.com/mickael-menu/zk/core/zk"
)

type New struct {
	Directory string `arg optional name:"directory" default:"."`
}

func (cmd *New) Run() error {
	zk, err := zk.Open(cmd.Directory)
	fmt.Printf("%+v\n", zk)
	return err
}
