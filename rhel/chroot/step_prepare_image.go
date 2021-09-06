package chroot

import (
	"context"
	"bytes"
	"fmt"
	"io"
	"log"
    "strings"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepPrepareImage struct {
	imagePath string
}

func (s *StepPrepareImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	sourcePath, err := filepath.Abs(config.SourceRPM)
	if err != nil {
		err := fmt.Errorf("Error formatting source image path: %s", err)
		return halt(state, err)
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		err := fmt.Errorf("Source RPM not found: %s", sourcePath)
		return halt(state, err)
	}

	log.Printf("Distro Release RPM: %s", sourcePath)
	ui.Say("Inital Chroot setup...")

	chrootDir := filepath.Abs(config.MountPath)

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
	cmd := fmt.Sprintf("rpm --root %s -ihv %s", chrootDir, RPMList)
	cmd, err = cmdWrapper(cmd)
	if err != nil {
		err := fmt.Errorf("Error formating RPM command: %s", err)
		return halt(state, err)
	 }

	shell := NewShellCommand(cmd)
	shell.Stderr = new(bytes.Buffer)
	if err := shell.Run(); err != nil {
		err := fmt.Errorf("Error running rpm to init DB: %s\n%s", sourcePath, err)
		return halt(state, err)
	}

	cmd := fmt.Sprintf("yum install -u --installroot=%s yum", chrootDir)
	cmd, err = cmdWrapper(cmd)
	if err != nil {
		err := fmt.Errorf("Error formating Yum command: %s", err)
		return halt(state, err)
	 }

	shell := NewShellCommand(cmd)
	shell.Stderr = new(bytes.Buffer)
	if err := shell.Run(); err != nil {
		err := fmt.Errorf("Error installing Yum: %s", err)
		return halt(state, err)
	}



	return multistep.ActionContinue
}

func (s *StepPrepareImage) Cleanup(state multistep.StateBag) {}

