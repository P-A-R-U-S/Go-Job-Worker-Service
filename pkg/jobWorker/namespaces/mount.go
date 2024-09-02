package namespaces

import (
	"fmt"
	"syscall"
)

// mount - mounts a proc filesystem at /proc.
//  		 it should not be used in a host mount namespace, otherwise
// 			 the host's proc filesystem will be messed up and require manual intervention to fix.

// NOTE: The user is expected to unmount the proc filesystem by calling the unmountVDA function.
func mountVDA() error {
	// TODO: validate this process is running in a new mount namespace (e.g. / is mounted as private)
	err := syscall.Mount("proc", "/proc", "proc", syscall.MS_NOEXEC, "")
	if err != nil {
		return fmt.Errorf("error mounting proc: %w", err)
	}
	return nil

}

// unmount - mounts a proc filesystem at /proc.
func unmountVDA() error {
	err := syscall.Unmount("/proc", 0)
	if err != nil {
		return fmt.Errorf("error unmounting proc: %w", err)
	}

	return nil
}
