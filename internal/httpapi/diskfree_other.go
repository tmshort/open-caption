// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build !unix && !windows

package httpapi

// freeBytes is not known on this system.
func freeBytes(string) (uint64, bool) { return 0, false }
