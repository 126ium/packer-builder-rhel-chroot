# RHEL chroot builder for Packer

A builder plugin of Packer to support building RHEL (or other Yum based distros) within chroot.

## Prerequirements

This plugin depends on the following tools:

- Packer
- Release RPMS required to allow install of yum/dnf
- Build node running the same distro as will be running in the Chroot

## Install

Download the binary from the [Releases](https://github.com/summerwind/packer-builder-qemu-chroot/releases) page and place it in one of the following places:


## Building plugin

To build the binary you need to install [Go](https://golang.org/), [dep](https://github.com/golang/dep) and [task](https://github.com/go-task/task).

```
$ task vendor
$ task build
```
Copy the resulting binary:
- The directory where packer is, or the executable directory
- `~/.packer.d/plugins` on Unix systems or `%APPDATA%/packer.d/plugins` on Windows
- The current working directory

## How does it work?

This plugin creates a chroot and either compresses it with squashfs or extracts files from it. Making it usable in automated build processes

## Quick Start

To use this plugin, you need to download the approprate rpms to kick off the chroot build.

eg. with Rocky 8.4
- rocky-release-8.4-32.el8.noarch.rpm
- rocky-gpg-keys-8.4-32.el8.noarch.rpm
- rocky-repos-8.4-32.el8.noarch.rpm
 
Prepare the following template file.

```
$ vim template.json
```

```
{
  "builders": [
    {
      "type": "rhel-chroot",
      "base_rpms": ["rocky-gpg-keys-8.4-32.el8.noarch.rpm ", "rocky-release-8.4-32.el8.noarch.rpm", "rocky-repos-8.4-32.el8.noarch.rpm"],
      "mount_path": "/tmp/rocky_chroot",
      "image_name": "/root/build.img"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "yum -y update"
      ]
    }
  ] 
}
```

Once you have the template, build it using Packer.

```
$ sudo packer build template.json
```

## Configuration Reference

### Required

- `base_rpms` (string) - A List of rpms in the source root to use to start the chroot preperation.
- `mount_path` (string) - Path to mount the chroot at. This will remain after the squashfs/build products are created or not created.

### Optional

- `output_directory` (string) - This is the path to the directory where the resulting image file will be created. By default this is "output-BUILDNAME" where "BUILDNAME" is the name of the builder.
- `image_name` (string) - The name of the resulting image file.
- `compression` (boolean) - Apply compression to the QCOW2 disk file using `qemu-img` convert. Defaults to false.
- `device_path` (string) - The path to the device where the volume of the source image will be attached.
- `mount_path` (string) - The path where the volume will be mounted. This is where the chroot environment will be. This defaults to /mnt/packer-builder-qemu-chroot/{{.Device}}. This is a configuration template where the .Device variable is replaced with the name of the device where the volume is attached.
- `mount_partition` (integer) - The partition number containing the / partition. By default this is the first partition of the volume.
- `mount_options` (array of string) - Options to supply the mount command when mounting devices. Each option will be prefixed with `-o` and supplied to the mount command ran by this plugin.
- `chroot_mounts` (array of array of string) - This is a list of devices to mount into the chroot environment. This configuration parameter requires some additional documentation which is in the "Chroot Mounts" section below. Please read that section for more information on how to use this.
- `copy_files` (array of string) - Paths to files on the running EC2 instance that will be copied into the chroot environment prior to provisioning. Defaults to /etc/resolv.conf so that DNS lookups work. Pass an empty list to skip copying /etc/resolv.conf. You may need to do this if you're building an image that uses systemd.
- `command_wrapper` (string) - How to run shell commands. This defaults to {{.Command}}. This may be useful to set if you want to set environmental variables or perhaps run it with sudo or so on. This is a configuration template where the .Command variable is replaced with the command to be run. Defaults to "{{.Command}}".

### Chroot Mounts

The `chroot_mounts` configuration can be used to mount specific devices within the chroot. By default, the following additional mounts are added into the chroot by this plugin:

- `/proc` (proc)
- `/sys` (sysfs)
- `/dev` (bind to real `/dev`)
- `/dev/pts` (devpts)
- `/proc/sys/fs/binfmt_misc` (binfmt_misc)

These default mounts are usually good enough for anyone and are sane defaults. However, if you want to change or add the mount points, you may using the chroot_mounts configuration. Here is an example configuration which only mounts /prod and /dev:

```
{
  "chroot_mounts": [
    ["proc", "proc", "/proc"],
    ["bind", "/dev", "/dev"]
  ]
}
```

`chroot_mounts` is a list of string arrays with more than three elements. The meaning of each component is as follows in order:

- The filesystem type. If this is "bind", then Packer will properly bind the filesystem to another mount point.
- The source device.
- The mount directory.
- The mount option (This element can be specified multiple times).

## License

Mozilla Public License 2.0

Note that this plugin is implemented by forking [QEMU chroot builder for Packer](https://github.com/summerwind/packer-builder-qemu-chroot) of Packer.
Which was implemented by forking [AMI Builder (chroot)](https://www.packer.io/docs/builders/amazon-chroot.html) of Packer.

