// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package netinfo

import "net"

func platformKind(net.Interface) string { return "" }
