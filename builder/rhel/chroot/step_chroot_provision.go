package chroot

import (
	"context"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepChrootProvision struct{}

func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	cmdWrapper := state.Get("command_wrapper").(CommandWrapper)

	comm := &Communicator{
		Chroot:     mountPath,
		CmdWrapper: cmdWrapper,
	}

	log.Println("Running the provision hook")
	if err := hook.Run(ctx, packer.HookProvision, ui, comm, nil); err != nil {
		return halt(state, err)
	}

	return multistep.ActionContinue
}

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
