// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package mdns

import "syscall"

// shareAddress lets this socket use UDP port 5353 alongside the operating system's own mDNS responder.
func shareAddress(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		if serr == nil {
			if e := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, soReusePort, 1); e != nil {
				serr = nil // not every system has it: SO_REUSEADDR alone is then what we have
			}
		}
	})
	if err != nil {
		return err
	}
	return serr
}
