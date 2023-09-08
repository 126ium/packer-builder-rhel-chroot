package chroot

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepPrepareOutputDir struct {
	success bool
}

func (s *StepPrepareOutputDir) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if !config.DontRsync {
		if _, err := os.Stat(config.MountPath); err == nil {
			if !config.PackerForce {
				err := fmt.Errorf("Output directory already exists: %s", config.MountPath)
				return halt(state, err)
			}

			ui.Say("Deleting previous output directory...")
			os.RemoveAll(config.MountPath)
		}

		ui.Say("Creating output directory...")
		if err := os.MkdirAll(config.MountPath, 0755); err != nil {
			return halt(state, err)
		}
	}

	if _, err := os.Stat(config.MountPath); err != nil {
		err := fmt.Errorf("Output directory does not exits: %s", config.MountPath)
		return halt(state, err)
	}

	s.success = true

	return multistep.ActionContinue
}

func (s *StepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Got here in cleanup")

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled || !halted {
		ui.Say("Ok lets return from cleanup")
		return
	}

	if cancelled || halted {
		config := state.Get("config").(*Config)

		ui.Say("Deleting output directory...")
		for i := 0; i < 5; i++ {
			err := os.RemoveAll(config.MountPath)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}
