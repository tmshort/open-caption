// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/tmshort/open-caption/internal/config"
)

// hashPassword is --hash-password: it reads one line (the password) and prints it as the sha256: digest that
// admin_password or operator_password in config.yaml may hold instead of the password itself. The password comes from
// standard input, never from an argument, so it does not end up in the shell's history. It returns the exit status.
func hashPassword(in io.Reader, out, errOut io.Writer, prompt bool) int {
	if prompt {
		fmt.Fprint(errOut, "Password to hash (it is shown as you type): ")
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintln(errOut, "open-caption: could not read the password:", err)
		return 2
	}
	pw := strings.TrimRight(line, "\r\n")
	if pw == "" {
		fmt.Fprintln(errOut, "open-caption: the password is empty")
		return 1
	}
	fmt.Fprintln(out, config.HashPassword(pw))
	return 0
}
