package namespaces

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func PivotRoot(rootfs string) error {
	putold := filepath.Join(rootfs, "/.pivot_root")

	// bind mount rootfs to itself - this is a slight hack needed to satisfy the
	// pivot_root requirement that rootfs and putold must not be on the same
	// filesystem as the current root
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("error (syscall.Mount) %s", err)
	}

	// create rootfs/.pivot_root as path for old_root
	pivotDir := filepath.Join(rootfs, ".pivot_root")
	if err := os.Mkdir(pivotDir, FileModeEveryone); err != nil {
		return fmt.Errorf("error (syscall.MkdirAll) %s", err)
	}

	// call pivot_root
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return fmt.Errorf("error (syscall.PivotRoot(%s, %s)) - %s", rootfs, putold, err)
	}

	// ensure current working directory is set to new root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("error (syscall.Chdir) %s", err)

	}
	// path to pivot root now changed, update
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root(which is now /.pivot_root) with all submounts
	// now we have only mounts that we mounted ourselves in `mount`
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
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

// MountProc - mount proc filesystem at /proc.
func MountProc() error {
	err := syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return fmt.Errorf("error mounting proc: %w", err)
	}
	return nil
}

// UnmountProc - unmount proc filesystem at /proc.
func UnmountProc() error {
	err := syscall.Unmount("/proc", 0)
	if err != nil {
		return fmt.Errorf("error unmounting proc: %w", err)
	}

	return nil
}
