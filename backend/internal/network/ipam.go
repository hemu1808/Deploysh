package network

import (
	"crypto/sha1"
	"fmt"
	"net"
	"sync"
)

// IPAM manages IP address allocation for containers on this node
type IPAM struct {
	mu           sync.Mutex
	clusterCIDR  *net.IPNet
	nodeSubnet   *net.IPNet
	allocatedIPs map[string]bool
}

// NewIPAM initializes an IPAM manager for the current node.
func NewIPAM(nodeID string, clusterCIDR string) (*IPAM, error) {
	_, parsedCIDR, err := net.ParseCIDR(clusterCIDR)
	if err != nil {
		return nil, err
	}

	parsedIP := parsedCIDR.IP.To4()
	if parsedIP == nil {
		return nil, fmt.Errorf("cluster CIDR must be IPv4")
	}

	// Deterministically generate a /24 subnet for this nodeID based on its hash
	// (In a real system, Raft would distribute subnets to prevent collisions)
	hash := sha1.Sum([]byte(nodeID))
	nodeOctet := hash[0]
	if nodeOctet == 0 || nodeOctet == 255 {
		nodeOctet = 1
	}

	// We assume a IPv4 /16, like 10.244.0.0/16
	// The generated subnet will be 10.244.<nodeOctet>.0/24
	subnetIP := make(net.IP, len(parsedIP))
	copy(subnetIP, parsedIP)
	
	// ipv4 offsets in the 4 byte array representation
	subnetIP[2] = nodeOctet

	nodeSubnet := &net.IPNet{
		IP:   subnetIP,
		Mask: net.CIDRMask(24, 32),
	}

	return &IPAM{
		clusterCIDR:  parsedCIDR,
		nodeSubnet:   nodeSubnet,
		allocatedIPs: make(map[string]bool),
	}, nil
}

// AllocateIP allocates an available IP from the Node's /24 subnet
func (i *IPAM) AllocateIP() (net.IP, *net.IPNet, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Try from .2 to .254 (.1 is the CNI bridge gateway)
	for octet := 2; octet < 255; octet++ {
		ip := make(net.IP, len(i.nodeSubnet.IP))
		copy(ip, i.nodeSubnet.IP)
		ip[3] = byte(octet)

		ipStr := ip.String()
		if !i.allocatedIPs[ipStr] {
			i.allocatedIPs[ipStr] = true
			return ip, i.nodeSubnet, nil
		}
	}

	return nil, nil, fmt.Errorf("no IPs available in node subnet %s", i.nodeSubnet.String())
}

// ReleaseIP marks the IP as available again
func (i *IPAM) ReleaseIP(ip net.IP) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.allocatedIPs, ip.String())
}

// GetGatewayIP returns the .1 IP for this subnet
func (i *IPAM) GetGatewayIP() net.IP {
	ip := make(net.IP, len(i.nodeSubnet.IP))
	copy(ip, i.nodeSubnet.IP)
	ip[3] = 1 
	return ip
}

// GetNodeSubnet returns the IPNet of the allocated subnet for this node
func (i *IPAM) GetNodeSubnet() *net.IPNet {
	return i.nodeSubnet
}
