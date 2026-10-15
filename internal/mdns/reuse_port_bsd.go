//go:build darwin || freebsd || netbsd || openbsd || dragonfly || linux

package mdns

import "syscall"

const soReusePort = syscall.SO_REUSEPORT
