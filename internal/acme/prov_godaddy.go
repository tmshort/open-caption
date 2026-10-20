// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build !acme_off && !acme_custom

package acme

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
)

func init() {
	register("godaddy", func() (challenge.Provider, error) { return godaddy.NewDNSProvider() })
}
