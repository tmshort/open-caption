// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build openbsd

package mdns

// SO_REUSEPORT, spelled out because Go's syscall package does not define it on every system and version.
const soReusePort = 0x800
