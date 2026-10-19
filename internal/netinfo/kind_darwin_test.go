//go:build darwin

package netinfo

import "testing"

func TestMacHardwarePortsGiveTheKind(t *testing.T) {
	m := parseHardwarePorts("Hardware Port: Wi-Fi\nDevice: en0\nEthernet Address: aa\n\nHardware Port: AX88179B\nDevice: en15\nEthernet Address: bb\n\nHardware Port: Thunderbolt Bridge\nDevice: bridge0\n\nHardware Port: iPhone USB\nDevice: en7\n")
	for dev, want := range map[string]string{"en0": KindWiFi, "en15": KindEthernet, "bridge0": KindOther, "en7": KindOther} {
		if got := portKind(m[dev]); got != want {
			t.Errorf("%s (%s): %s, want %s", dev, m[dev], got, want)
		}
	}
}
