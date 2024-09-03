package namespaces

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	CPU_PERIOD     = 1_000_000
	cgroupFileMode = 0o500
)

var (
	rootCgroupPath = "/sys/fs/cgroup"
)

// groupPath returns a given cgroup's directory path identified by name
func GetCGroupPath(cgroup string) string {
	return filepath.Join(rootCgroupPath, cgroup)
}

// createCGroup creates a new cgroup with the given cpu, io, and memory limits.
// No validation on the limits is done since it's expected that the caller has already validated the input.
func CreateCGroup(cgroupDir string, rootDeviceMajMin string, cpu float64, ioInBytes int64, memoryInBytes int64) error {
	if err := os.MkdirAll(cgroupDir, cgroupFileMode); err != nil {
		return fmt.Errorf("error creating new control group: %w", err)
	}

	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")

	// create a directory structure like /sys/fs/cgroup/<uuid>/tasks
	if err := os.MkdirAll(cgroupTasksDir, cgroupFileMode); err != nil {
		return fmt.Errorf("error creating new control group tasjs: %w", err)
	}

	// instruct the cgroup subtree to enable cpu, io, and memory controllers
	if err := os.WriteFile(filepath.Join(cgroupDir, "cgroup.subtree_control"), []byte("+cpu +io +memory"), cgroupFileMode); err != nil {
		return fmt.Errorf("error writing cgroup.subtree_control: %w", err)
	}

	cpuQuota := int(cpu * float64(CPU_PERIOD))
	cpuMaxContent := fmt.Sprintf("%d %d", cpuQuota, CPU_PERIOD)

	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "cpu.max"), []byte(cpuMaxContent), cgroupFileMode); err != nil {
		return fmt.Errorf("error writing cpu.max: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "memory.max"), []byte(strconv.FormatInt(memoryInBytes, 10)), cgroupFileMode); err != nil {
		return fmt.Errorf("error writing memory.max: %w", err)
	}

	// TODO/Future Consideration: add support for specifying rbps, wbps, riops, and wiops for a list of devices
	formattedIOInBytes := strconv.FormatInt(ioInBytes, 10)
	ioMaxContent := fmt.Sprintf("%s rbps=%s wbps=%s riops=max wiops=max", rootDeviceMajMin, formattedIOInBytes, formattedIOInBytes)

	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "io.max"), []byte(ioMaxContent), cgroupFileMode); err != nil {
		return fmt.Errorf("error writing io.max: %w", err)
	}

	return nil
}

// AddProcess mutates the given cmd to instruct GO to add the PID of the started process to a given cgroup
func AddProcess(cgroupName string, cmd *exec.Cmd) error {
	// Add job's process to cgroup
	f, err := syscall.Open(GetCGroupPath(cgroupName), 0x200000, 0)
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

// cleanupCGroup removes the cgroup directory and all of its contents.
func CleanupCGroup(cgroupDir string) error {
	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")

	if err := os.RemoveAll(cgroupTasksDir); err != nil {
		return fmt.Errorf("error removing cgroup tasks directory: %w", err)
	}

	if err := os.RemoveAll(cgroupDir); err != nil {
		return fmt.Errorf("error removing cgroup directory: %w", err)
	}

	return nil
}

//// cgroups v2 interface files for supporting controllers

//
//const FILE_MODE = 0666 //0o500
//

//
//// CreateGroup creates a directory in the cgroup root path to signal cgroup to create a group
//// TODO: In PROD we could check here the cgroup was created correctly,
////
////	such as checking cgroup.controllers file for supported controllers
//func CreateGroup(cgroupName string) (string, error) {
//	return groupPath(cgroupName), os.Mkdir(groupPath(cgroupName), 0755)
//}
//
//// DeleteGroup deletes a cgroup's directory signalling cgroup to delete the group
//// TODO in production before deleting a group we could check cgroup.events to ensure no processes are still running in their cgroup
//func DeleteGroup(cgroupName string) error {
//	return os.RemoveAll(groupPath(cgroupName))
//}
//

//
//// AddResourceControl updates the resource control interface file for a given cgroup using JobOpts.
//func AddResourceControl(cgroupName string, controller string, value string) (err error) {
//	if err = updateController(cgroupName, controller, value); err != nil {
//		return fmt.Errorf("not able to add resources:%s into cgroup controller:%s", value, controller)
//	}
//	return nil
//}
//

//
//// updateController sets the content of the controller interface file for a
//// given resource controller within a CGroup (e.g. "memory.high", etc.)
//func updateController(cgroupName string, file, value string) error {
//	return os.WriteFile(filepath.Join(groupPath(cgroupName), file), []byte(value), FILE_MODE)
//}
