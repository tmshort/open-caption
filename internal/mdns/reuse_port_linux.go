//go:build linux

package mdns

// SO_REUSEPORT, spelled out because Go's syscall package does not define it on every system and version.
const soReusePort = 0xf
