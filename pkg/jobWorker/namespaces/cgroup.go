package namespaces

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// cgroups v2 interface files for supporting controllers
const (
	CPU_WEIGHT_FILE  = "cpu.weight"
	MEMORY_HIGH_FILE = "memory.high"
	// Note: looks like Block IO controller not supported in some Ubuntu Kernels
	//		 try to enable it: https://docs.kernel.org/admin-guide/cgroup-v1/blkio-controller.html
	// 		 if you have any issue to set this value
	IO_WEIGHT_FILE = "io.weight"
)

const FILE_MODE = 0666 //0o500

var (
	rootCgroupPath = "/sys/fs/cgroup"
)

// CreateGroup creates a directory in the cgroup root path to signal cgroup to create a group
// TODO: In PROD we could check here the cgroup was created correctly,
//
//	such as checking cgroup.controllers file for supported controllers
func CreateGroup(cgroupName string) error {
	return os.Mkdir(groupPath(cgroupName), 0755)
}

// DeleteGroup deletes a cgroup's directory signalling cgroup to delete the group
// TODO in production before deleting a group we could check cgroup.events to ensure no processes are still running in their cgroup
func DeleteGroup(cgroupName string) error {
	return os.RemoveAll(groupPath(cgroupName))
}

// AddProcess mutates the given cmd to instruct GO to add the PID of the started process to a given cgroup
func AddProcess(cgroupName string, cmd *exec.Cmd) error {
	// Add job's process to cgroup
	f, err := syscall.Open(groupPath(cgroupName), 0x200000, 0)
	if err != nil {
		return err
	}
	// This is where clone args and namespaces for user, PID and fs can be set
	cmd.SysProcAttr = &syscall.SysProcAttr{
		UseCgroupFD: true,
		CgroupFD:    f,
	}
	return nil
}

// AddResourceControl updates the resource control interface file for a given cgroup using JobOpts. The
// three currently supported are CPU, memory and IO
func AddResourceControl(cgroupName string, controller string, value string) (err error) {
	if err = updateController(cgroupName, controller, value); err != nil {
		return err
	}
	return nil
}

// groupPath returns a given cgroup's directory path identified by name
func groupPath(cgroup string) string {
	return filepath.Join(rootCgroupPath, cgroup)
}

// updateController sets the content of the controller interface file for a
// given resource controller within a CGroup (e.g. "memory.high", etc.)
func updateController(cgroupName string, file, value string) error {
	controller := filepath.Join(groupPath(cgroupName), file)
	log.Printf("update constoller:%s, value:%s", controller, value)
	return os.WriteFile(controller, []byte(value), FILE_MODE)
}
