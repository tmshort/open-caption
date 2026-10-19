package netinfo

import (
	"net"
	"strings"
)

// interfaceKind says whether an adapter is Ethernet, Wi-Fi or something else. The operating system is asked where
// it can tell (kind_*.go); otherwise the adapter's name decides.
func interfaceKind(ifc net.Interface) string {
	if k := platformKind(ifc); k != "" {
		return k
	}
	return nameKind(ifc.Name)
}

// nameKind guesses from the adapter's name ("Wi-Fi", "wlan0", "wlp3s0", "Ethernet", "eth0", "enp0s31f6").
func nameKind(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "wi-fi"), strings.Contains(n, "wifi"), strings.Contains(n, "wlan"), strings.Contains(n, "wireless"), strings.HasPrefix(n, "wl"):
		return KindWiFi
	case strings.Contains(n, "ethernet"), strings.HasPrefix(n, "eth"), strings.HasPrefix(n, "enp"), strings.HasPrefix(n, "eno"), strings.HasPrefix(n, "ens"), strings.HasPrefix(n, "enx"):
		return KindEthernet
	}
	return KindOther
}
