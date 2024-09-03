package namespaces

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	CPU_PERIOD = 1_000_000
	/*
		Common Permission Usages

		0755 Commonly used on web servers. The owner can read, write, execute. Everyone else can read and execute but not modify the file.

		0777 Everyone can read write and execute. On a web server, it is not advisable to use ‘777’ permission for your files and folders, as it allows anyone to add malicious code to your server.

		0644 Only the owner can read and write. Everyone else can only read. No one can execute the file.

		0655 Only the owner can read and write, but not execute the file. Everyone else can read and execute, but cannot modify the file.
	*/
	FILE_MODE_EVERYONE = 0777
	FILE_MODE_WEB      = 0755 //0o500 = 0o500
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
	// create a directory structure like /sys/fs/cgroup/<uuid>
	log.Printf("create cgroup:%s", cgroupDir)
	if err := os.Mkdir(cgroupDir, FILE_MODE_WEB); err != nil {
		return fmt.Errorf("error creating new control group: %w", err)
	}

	// create a directory structure like /sys/fs/cgroup/<uuid>/tasks
	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")
	log.Printf("create cgroup/tasks:%s", cgroupTasksDir)
	if err := os.MkdirAll(cgroupTasksDir, FILE_MODE_WEB); err != nil {
		return fmt.Errorf("error creating new control group tasjs: %w", err)
	}

	// instruct the cgroup subtree to enable cpu, io, and memory controllers
	log.Printf("write into:%s", filepath.Join(cgroupDir, "cgroup.subtree_control"))
	if err := os.WriteFile(filepath.Join(cgroupDir, "cgroup.subtree_control"), []byte("+cpu +io +memory"), FILE_MODE_WEB); err != nil {
		return fmt.Errorf("error writing cgroup.subtree_control: %w", err)
	}

	cpuQuota := int(cpu * float64(CPU_PERIOD))
	cpuMaxContent := fmt.Sprintf("%d %d", cpuQuota, CPU_PERIOD)

	log.Printf("write into:%s", filepath.Join(cgroupTasksDir, "cpu.max"))
	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "cpu.max"), []byte(cpuMaxContent), FILE_MODE_WEB); err != nil {
		return fmt.Errorf("error writing cpu.max: %w", err)
	}

	log.Printf("write into:%s", filepath.Join(cgroupTasksDir, "memory.max"))
	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "memory.max"), []byte(strconv.FormatInt(memoryInBytes, 10)), FILE_MODE_WEB); err != nil {
		return fmt.Errorf("error writing memory.max: %w", err)
	}

	// TODO/Future Consideration: add support for specifying rbps, wbps, riops, and wiops for a list of devices
	formattedIOInBytes := strconv.FormatInt(ioInBytes, 10)
	ioMaxContent := fmt.Sprintf("%s rbps=%s wbps=%s riops=max wiops=max", rootDeviceMajMin, formattedIOInBytes, formattedIOInBytes)

	log.Printf("write into:%s", filepath.Join(cgroupTasksDir, "io.max"))
	if err := os.WriteFile(filepath.Join(cgroupTasksDir, "io.max"), []byte(ioMaxContent), FILE_MODE_WEB); err != nil {
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

// CleanupCGroup removes the cgroup directory and all of its contents.
func CleanupCGroup(cgroupDir string) error {
	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")

	log.Printf("remove cgroup/tasks:%s", cgroupTasksDir)
	if err := os.RemoveAll(cgroupTasksDir); err != nil {
		return fmt.Errorf("error removing cgroup tasks directory: %w", err)
	}

	log.Printf("remove cgroup:%s", cgroupDir)
	if err := os.RemoveAll(cgroupDir); err != nil {
		return fmt.Errorf("error removing cgroup directory: %w", err)
	}

	return nil
}
