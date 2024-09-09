package namespaces

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

//

func newNetworkNamespace() error {
	// Create a new network namespace
	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("Failed to create a new namespace: %v\n", err)
	}
	return nil
}

func main() {
	// Create a new network namespace
	err := unix.Unshare(unix.CLONE_NEWNET)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create a new namespace: %v\n", err)
		return
	}

	// Create a network interface (virtual loopback for demonstration)
	iface := "lo"
	err = unix.IoctlSetIfreq(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0), unix.SIOCGIFFLAGS, &unix.Ifreq{Name: iface})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get interface flags: %v\n", err)
		return
	}

	// Bringing the interface up
	err = unix.IoctlSetIfreq(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0), unix.SIOCSIFFLAGS, &unix.Ifreq{
		Name:  iface,
		Flags: unix.IFF_UP,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bring interface up: %v\n", err)
		return
	}

	// Assign an IP address to the interface
	addr := unix.SockaddrInet4{
		Addr: [4]byte{192, 168, 1, 1},
	}
	err = unix.Bind(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0), &addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind IP address: %v\n", err)
		return
	}

	fmt.Println("Network namespace created and interface configured.")
}
