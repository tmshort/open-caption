// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package httpapi

import "golang.org/x/sys/windows"

// freeBytes is the space the current user can still use on the disk that holds dir.
func freeBytes(dir string) (uint64, bool) {
	p, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return 0, false
	}
	var free, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &free, &total, &totalFree); err != nil {
		return 0, false
	}
	return free, true
}
