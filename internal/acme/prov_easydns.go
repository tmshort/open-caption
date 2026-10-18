//go:build !acme_off && !acme_custom

package acme

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/easydns"
)

func init() {
	register("easydns", func() (challenge.Provider, error) { return easydns.NewDNSProvider() })
}
