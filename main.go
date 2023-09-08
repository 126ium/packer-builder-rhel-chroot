package main

import (
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"os"
	"packer-builder-rhel-chroot/rhel/chroot"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("rhel-chroot", new(chroot.Builder))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
