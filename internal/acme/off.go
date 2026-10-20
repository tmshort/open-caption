// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build acme_off

package acme

import (
	"errors"

	"github.com/tmshort/open-caption/internal/config"
)

// This program was built without ACME support (ACME_PROVIDERS=none).

func Available() bool         { return false }
func Providers() []string     { return nil }
func HasProvider(string) bool { return false }

// NewIssuer always fails in a build without ACME.
func NewIssuer(config.ACME, string, func(string, ...any)) (Issuer, error) {
	return nil, errors.New("this program was built without automatic certificates (ACME_PROVIDERS=none)")
}
