// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package netinfo

import (
	"testing"
	"time"
)

func TestAnAdapterPluggedInLaterIsFoundWithoutWaitingAMinute(t *testing.T) {
	now := time.Unix(1000, 0)
	loads := 0
	devices := map[string]string{"en0": "Wi-Fi"}
	c := &portCache{now: func() time.Time { return now }, load: func() map[string]string {
		loads++
		cp := map[string]string{}
		for k, v := range devices {
			cp[k] = v
		}
		return cp
	}}
	if p, ok := c.port("en0"); !ok || p != "Wi-Fi" || loads != 1 {
		t.Fatalf("%q %v loads=%d", p, ok, loads)
	}
	// A USB adapter is plugged in: it is not in the list that was read, so it is read again at once ...
	devices["en15"] = "AX88179B"
	now = now.Add(10 * time.Second)
	if p, ok := c.port("en15"); !ok || p != "AX88179B" || loads != 2 {
		t.Errorf("a new device is found at once: %q %v loads=%d", p, ok, loads)
	}
	// ... but a device that has no hardware port (a VPN) does not make the command run for every question.
	for i := 0; i < 5; i++ {
		c.port("utun3")
	}
	if loads != 2 {
		t.Errorf("a device that is not there is not looked up again for a few seconds: loads=%d", loads)
	}
	now = now.Add(5 * time.Second)
	c.port("utun3")
	if loads != 3 {
		t.Errorf("... but is, later: loads=%d", loads)
	}
	// Known devices are served from the list for a minute.
	for i := 0; i < 5; i++ {
		c.port("en0")
	}
	if loads != 3 {
		t.Errorf("known devices do not reload: loads=%d", loads)
	}
	now = now.Add(2 * time.Minute)
	c.port("en0")
	if loads != 4 {
		t.Errorf("the list is reloaded after a minute: loads=%d", loads)
	}
}
