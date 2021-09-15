package chroot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCompressImage struct{}

func (s *StepCompressImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	cmdWrapper := state.Get("command_wrapper").(CommandWrapper)
	imageName := state.Get("image_name").(string)
	mountPath := state.Get("mount_path").(string)
	outputDir := config.OutputDir

	if config.ExportBuild {
		for _, srcChroot := range config.ExportFiles {
			srcPath := filepath.Join(mountPath,srcChroot)
			dstPath := filepath.Join(outputDir,imageName)

			ui.Message(fmt.Sprintf("Copying: %s", srcPath))

			cmd := fmt.Sprintf("cp -r %s %s", srcPath, dstPath)
			cmd, err := cmdWrapper(cmd)
			if err != nil {
				err := fmt.Errorf("Errorr creating copy command: %s", err)
				return halt(state, err)
			}

			ui.Say(fmt.Sprintf("Copy command: %s", cmd))

			shell := NewShellCommand(cmd)
			shell.Stderr = new(bytes.Buffer)
			if err := shell.Run(); err != nil {
				err := fmt.Errorf("Error copying file/s: %s\n%s", err, shell.Stderr)
				return halt(state, err)
			}
		}
	}


	if  config.MakeSquash {
		ui.Say("Compressing image...")


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

	}
	return multistep.ActionContinue
}

func (s *StepCompressImage) Cleanup(state multistep.StateBag) {}
