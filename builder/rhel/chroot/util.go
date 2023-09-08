package chroot

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func halt(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Error(err.Error())

	state.Put("error", err)

	return multistep.ActionHalt
}
