//go:build !linux && !darwin && !windows

package netinfo

import "net"

func platformKind(net.Interface) string { return "" }
