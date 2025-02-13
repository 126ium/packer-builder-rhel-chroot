//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package chroot

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Config represents a configuration of builder.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputDir      string     `mapstructure:"output_directory"`
	WorkDir        string     `mapstructure:"tmp_directory"`
	ImageName      string     `mapstructure:"image_name"`
	MountPath      string     `mapstructure:"mount_path"`
	ExportFolder   string     `mapstructure:"export_folder"`
	MountOptions   []string   `mapstructure:"mount_options"`
	BaseRPMS       []string   `mapstructure:"base_rpms"`
	ChrootMounts   [][]string `mapstructure:"chroot_mounts"`
	CopyFiles      []string   `mapstructure:"copy_files"`
	ExportFiles    [][]string `mapstructure:"export_files"`
	CommandWrapper string     `mapstructure:"command_wrapper"`
	InitChroot     bool       `mapstructure:"init_chroot"`
	MakeSquash     bool       `mapstructure:"make_squash"`
	ExportBuild    bool       `mapstructure:"export_build"`
	NewImage       bool       `mapstructure:"new_image"`
	DontRsync      bool       `mapstructure:"dont_rsync"`
	BaseIamge      string     `mapstructure:"base_image"`

	ctx interpolate.Context
}

// Cleaner is an interface with a function for cleanup.
type Cleaner interface {
	CleanupFunc(multistep.StateBag) error
}

// Builder represents a builder plugin for Packer.
type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

// Prepare validates given configuration.
func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.BaseIamge == "" {
		b.config.NewImage = true
	}

	if b.config.ExportBuild {
		if b.config.ExportFiles == nil {
			b.config.ExportBuild = false
		}
	}

	if b.config.ImageName == "" {
		b.config.ImageName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-plugin-gdata/{{.ImageName}}"
	}

	if b.config.ExportFolder == "" {
		b.config.ExportFolder = "export"
	}

	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	if b.config.CopyFiles == nil {
		b.config.CopyFiles = []string{"/etc/resolv.conf"}
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-builder-qemu-chroot/{{.ImageName}}"
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warns, errs
	}

	// Fix me
	return nil, warns, nil
}

// Run runs each step of the plugin in order.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("the rhel-chroot builder only works on Linux environments")
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("command_wrapper", NewCommandWrapper(b.config))

	steps := []multistep.Step{
		&StepPrepareOutputDir{},
		&StepPrepareImage{},
		&StepMountExtra{},
		&StepCopyFiles{},
		&StepChrootProvision{},
		&StepEarlyCleanup{},
		&StepCompressImage{},
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("build was halted")
	}

	artifact := &Artifact{
		dir: b.config.OutputDir,
		files: []string{
			state.Get("image_path").(string),
		},
	}

	return artifact, nil
}
