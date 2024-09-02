package namespaces

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func PivotRoot(newroot string) error {
	putold := filepath.Join(newroot, "/.pivot_root")

	// bind mount newroot to itself - this is a slight hack needed to satisfy the
	// pivot_root requirement that newroot and putold must not be on the same
	// filesystem as the current root
	if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("error (syscall.Mount) %s", err)
	}

	// create putold directory
	if err := os.MkdirAll(putold, 0700); err != nil {
		return fmt.Errorf("error (syscall.MkdirAll) %s", err)
	}

	// call pivot_root
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return fmt.Errorf("error (syscall.PivotRoot) %s", err)
	}

	// ensure current working directory is set to new root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("error (syscall.Chdir) %s", err)

	}

	// umount putold, which now lives at /.pivot_root
	putold = "/.pivot_root"
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("error (syscall.Unmount) %s", err)
	}

	// remove putold
	if err := os.RemoveAll(putold); err != nil {
		return fmt.Errorf("error (syscall.RemoveAll) %s", err)
	}

	return nil
}

func MountProc(newroot string) error {
	source := "proc"
	target := filepath.Join(newroot, "/proc")
	fstype := "proc"
	flags := 0
	data := ""

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}
	if err := syscall.Mount(source, target, fstype, uintptr(flags), data); err != nil {
		return err
	}

	return nil
}

// unmount - mounts a proc filesystem at /proc.
func UnmountProc(newroot string) error {
	err := syscall.Unmount("/proc", 0)
	if err != nil {
		return fmt.Errorf("error unmounting proc: %w", err)
	}

	return nil
}
