package chroot

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCompressImage struct{}

func (s *StepCompressImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cmdWrapper := state.Get("command_wrapper").(CommandWrapper)
	imageName := state.Get("image_name").(string)
	mountPath := state.Get("mount_path").(string)

	ui.Say("Compressing image...")

    // mksquashfs ${1} squashfs.img.TMP -comp xz -b 1048576 -Xbcj x86 -Xdict-size 100%

	cmd := fmt.Sprintf("mksquashfs %s %s -comp xz -b 1048576 -Xbcj x86 -Xdict-size 100%%", mountPath, imageName)
	ui.Say(cmd)
	cmd, err := cmdWrapper(cmd)
	if err != nil {
		err := fmt.Errorf("Error creating compression command: %s", err)
		return halt(state, err)
	}

	log.Printf("Compression command: %s", cmd)

	shell := NewShellCommand(cmd)
	shell.Stderr = new(bytes.Buffer)
	if err := shell.Run(); err != nil {
		err := fmt.Errorf("Error compressing image: %s\n%s", err, shell.Stderr)
		return halt(state, err)
	}


	return multistep.ActionContinue
}

func (s *StepCompressImage) Cleanup(state multistep.StateBag) {}
