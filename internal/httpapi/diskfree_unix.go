// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package httpapi

import "golang.org/x/sys/unix"

// freeBytes is the space a program without special rights can still use on the disk that holds dir.
func freeBytes(dir string) (uint64, bool) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return 0, false
	}
	return uint64(st.Bavail) * uint64(st.Bsize), true //nolint:unconvert // the field types differ between systems
}
