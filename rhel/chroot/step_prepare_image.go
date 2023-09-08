package chroot

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepPrepareImage struct {
	imagePath string
}

func (s *StepPrepareImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	cmdWrapper := state.Get("command_wrapper").(CommandWrapper)

	if config.NewImage {

		ui.Say("Inital Chroot setup...")

		chrootDir, err := filepath.Abs(config.MountPath)
		if err != nil {
			err := fmt.Errorf("Error formating MountPath command: %s", err)
			return halt(state, err)
		}

		cmd := fmt.Sprintf("rpm --root %s --initdb", chrootDir)
		cmd, err = cmdWrapper(cmd)
		if err != nil {
			err := fmt.Errorf("Error formating RPM command: %s", err)
			return halt(state, err)
		}

		shell := NewShellCommand(cmd)
		shell.Stderr = new(bytes.Buffer)
		if err := shell.Run(); err != nil {
			err := fmt.Errorf("Error running rpm to init DB: %s\n%s", err, shell.Stderr)
			return halt(state, err)
		}

		RPMList := strings.Join(config.BaseRPMS, " ")
		cmd = fmt.Sprintf("rpm --root %s -ihv %s", chrootDir, RPMList)
		cmd, err = cmdWrapper(cmd)
		if err != nil {
			err := fmt.Errorf("Error formating RPM command: %s", err)
			return halt(state, err)
		}

		shell = NewShellCommand(cmd)
		shell.Stderr = new(bytes.Buffer)
		if err := shell.Run(); err != nil {
			err := fmt.Errorf("Error running rpm to init DB: %s", err)
			return halt(state, err)
		}

		cmd = fmt.Sprintf("yum install -y --installroot=%s yum", chrootDir)
		cmd, err = cmdWrapper(cmd)
		if err != nil {
			err := fmt.Errorf("Error formating Yum command: %s", err)
			return halt(state, err)
		}

		shell = NewShellCommand(cmd)
		shell.Stderr = new(bytes.Buffer)
		if err := shell.Run(); err != nil {
			err := fmt.Errorf("Error installing Yum: %s", err)
			return halt(state, err)
		}

		ui.Say("I think it wworked...")

		s.imagePath = config.ImageName
		state.Put("mount_path", chrootDir)
		state.Put("image_name", config.ImageName)
		state.Put("image_path", config.OutputDir)

		return multistep.ActionContinue

	} else {

		ui.Say("Cloning existing image...")

		chrootDir, err := filepath.Abs(config.MountPath)
		if err != nil {
			err := fmt.Errorf("Error formatting MountPath command: %s", err)
			return halt(state, err)
		}

		sourceDir, err := filepath.Abs(config.BaseIamge)
		if err != nil {
			err := fmt.Errorf("Error formatting BaseImage command: %s", err)
			return halt(state, err)
		}

		srcDir, err := filepath.Abs(sourceDir)
		cmd := fmt.Sprintf("rsync -av %s/. %s", srcDir, chrootDir)
		if err != nil {
			err := fmt.Errorf("Error formatting rsync command: %s", err)
			return halt(state, err)
		}
		ui.Say(cmd)

		if !config.DontRsync {
			shell := NewShellCommand(cmd)
			shell.Stderr = new(bytes.Buffer)
			if err := shell.Run(); err != nil {
				err := fmt.Errorf("Error running rsync to clone image: %s\n%s", err, shell.Stderr)
				return halt(state, err)
			}
		} else {
			ui.Say("Using existing iamge without rsync clone")
		}

		ui.Say("I think it worked.....")

		s.imagePath = config.ImageName
		state.Put("mount_path", chrootDir)
		state.Put("image_name", config.ImageName)
		state.Put("image_path", config.OutputDir)

		return multistep.ActionContinue

	}
}

func (s *StepPrepareImage) Cleanup(state multistep.StateBag) {}
