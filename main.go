package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/insanemal/packer-builder-rhel-chroot/rhel/chroot"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	server.RegisterBuilder(chroot.NewBuilder())
	server.Serve()
}
