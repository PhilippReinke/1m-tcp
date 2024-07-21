package shared

import (
	"fmt"
	"net"
)

// LocalIP returns 1st IP it encounters when listing network interfaces that is
// not loopback
func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("error listing addresses of network interfaces: %v", err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP, nil
		}
	}
	return nil, fmt.Errorf("no IP beside loopback found")
}
