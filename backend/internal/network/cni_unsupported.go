//go:build !linux
// +build !linux

package network

import (
	"fmt"
	"net"
)

// SetupNodeNetwork is a stub for non-Linux OS
func SetupNodeNetwork(nodeIP net.IP, gatewayIP net.IP, subnet *net.IPNet) error {
	return fmt.Errorf("Network overlay (VXLAN/Bridge) is only supported on Linux. Please run AuraDeploy backend inside WSL2 or a Linux VM.")
}

// SetupContainerNetwork is a stub for non-Linux OS
func SetupContainerNetwork(containerID string, netnsPath string, containerIP net.IP, subnet *net.IPNet, gatewayIP net.IP) error {
	return fmt.Errorf("Container networking is only supported on Linux.")
}

// TeardownContainerNetwork is a stub for non-Linux OS
func TeardownContainerNetwork(containerID string) error {
	return nil
}
