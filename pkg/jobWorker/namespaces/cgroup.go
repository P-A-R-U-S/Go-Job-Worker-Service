package namespaces

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	CPU_WEIGHT_File  = "cpu.weight"
	MEMORY_HIGH_File = "memory.high"
	IO_WEIGHT_File   = "io.weight"

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

// AddProcess mutates the given cmd to instruct GO to add the PID of the started process to a given cgroup
func AddProcess(cgroupName string, cmd *exec.Cmd) error {
	// Add job's process to cgroup
	cgroupDir := GetCGroupPath(cgroupName)

	fd, err := syscall.Open(cgroupDir, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	// This is where clone args and namespaces for user, PID and fs can be set
	cmd.SysProcAttr.CgroupFD = fd

	return nil
}

// CreateCGroup creates a directory in the cgroup root path to signal cgroup to create a group
// TODO in production we could check here the cgroup was created correctly, such as checking cgroup.controllers file for supported controllers
func CreateCGroup(cgroupName string) (err error) {
	cgroupDir := GetCGroupPath(cgroupName)

	// create a directory structure like /sys/fs/cgroup/<uuid>
	log.Printf("create cgroup:%s", cgroupDir)
	if err := os.Mkdir(cgroupDir, FILE_MODE_EVERYONE); err != nil {
		log.Printf("error creating new control group: %s", err)
		return fmt.Errorf("error creating new control group: %w", err)
	}
	// create a directory structure like /sys/fs/cgroup/<uuid>/tasks
	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")
	log.Printf("create cgroup/tasks:%s", cgroupTasksDir)
	if err := os.MkdirAll(cgroupTasksDir, FILE_MODE_EVERYONE); err != nil {
		log.Printf("error creating new control group tasjs: %s", err)
		return fmt.Errorf("error creating new control group tasjs: %w", err)
	}
	return nil
}

// DeleteCGroup deletes a cgroup's directory signalling cgroup to delete the group
// TODO in production before deleting a group we could check cgroup.events to ensure no processes are still running in their cgroup
func DeleteCGroup(cgroupName string) error {
	cgroupDir := GetCGroupPath(cgroupName)

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

// AddResourceControl updates the resource control interface file for a given cgroup using JobOpts. The
// three currently supported are CPU, memory and IO
func AddResourceControl(cgroupName string, controller string, value string) (err error) {
	if err = os.WriteFile(filepath.Join(GetCGroupPath(cgroupName), controller), []byte(value), FILE_MODE_EVERYONE); err != nil {
		return err
	}
	return nil
}

// GetCGroupPath returns a given cgroup's directory path identified by name
func GetCGroupPath(cgroupName string) string {
	return filepath.Join(rootCgroupPath, cgroupName)
}

const (
	KB int64 = 1024
	MB       = KB * 1024
	GB       = MB * 1024
)
