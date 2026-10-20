//go:build darwin

package netinfo

import (
	"net"
	"os/exec"
	"strings"
)

var ports = &portCache{load: hardwarePorts}

// platformKind asks `networksetup -listallhardwareports`, which names each device's hardware port.
func platformKind(ifc net.Interface) string {
	port, ok := ports.port(ifc.Name)
	if !ok {
		return ""
	}
	return portKind(port)
}

func hardwarePorts() map[string]string {
	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		return map[string]string{}
	}
	return parseHardwarePorts(string(out))
}

func parseHardwarePorts(text string) map[string]string {
	m := map[string]string{}
	port := ""
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "Hardware Port:"):
			port = strings.TrimSpace(strings.TrimPrefix(line, "Hardware Port:"))
		case strings.HasPrefix(line, "Device:") && port != "":
			m[strings.TrimSpace(strings.TrimPrefix(line, "Device:"))] = port
		}
	}
	return m
}

// portKind: Wi-Fi is Wi-Fi; Bluetooth, the Thunderbolt bridge and phones are something else; any other port
// (Ethernet, a USB or Thunderbolt adapter named after its chip) is wired.
func portKind(port string) string {
	p := strings.ToLower(port)
	switch {
	case strings.Contains(p, "wi-fi"), strings.Contains(p, "airport"):
		return KindWiFi
	case strings.Contains(p, "bluetooth"), strings.Contains(p, "bridge"), strings.Contains(p, "iphone"), strings.Contains(p, "ipad"), strings.Contains(p, "vpn"):
		return KindOther
	}
	return KindEthernet
}
