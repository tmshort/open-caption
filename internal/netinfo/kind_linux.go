//go:build linux

package netinfo

import (
	"net"
	"os"
	"path/filepath"
)

// platformKind reads /sys/class/net: a wireless adapter has a "wireless" (or "phy80211") folder; a physical one
// has a "device" link; a virtual one (docker0, veth..., tun0) has neither.
func platformKind(ifc net.Interface) string {
	base := filepath.Join("/sys/class/net", ifc.Name)
	for _, w := range []string{"wireless", "phy80211"} {
		if _, err := os.Stat(filepath.Join(base, w)); err == nil {
			return KindWiFi
		}
	}
	if _, err := os.Stat(filepath.Join(base, "device")); err == nil {
		return KindEthernet
	}
	if _, err := os.Stat(base); err == nil {
		return KindOther
	}
	return ""
}
