// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/netip"
	"strings"
)

// Access limits which computers may open the Setup and operator pages. An empty list means anyone who
// can reach the server (the audience page is always open to everyone).
type Access struct {
	AdminAllow    []string `yaml:"admin_allow"`
	OperatorAllow []string `yaml:"operator_allow"`
	// ViewerAllow is the access list of the Follow page (/follow), like AdminAllow for Setup and OperatorAllow for the operator
	// page: the computers on it, and the computers allowed for the operator page or Setup, may open it. Empty means anywhere.
	ViewerAllow []string `yaml:"viewer_allow"`
	// LocalAutoLogin signs in, as admin and without a password, anyone using Setup or the operator page on the
	// computer that runs the program (a browser opened at localhost, 127.0.0.1 or [::1]). Other computers
	// still sign in with a password. Turn it off on a computer that other people can log in to.
	LocalAutoLogin bool `yaml:"local_auto_login"`
}

// AllowList is a parsed list of addresses. The zero value allows everyone.
type AllowList struct {
	prefixes []netip.Prefix
	deny     bool // an unusable list denies everything rather than silently allowing it
}

var (
	localhostPrefixes = []string{"127.0.0.0/8", "::1/128"}
	privatePrefixes   = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "169.254.0.0/16", "fc00::/7", "fe80::/10"}
)

// ParseAllowList understands: "localhost", "private" (home and office networks: 10.x, 172.16-31.x,
// 192.168.x, and link-local), a single address ("192.168.1.50"), or a network ("192.168.1.0/24").
func ParseAllowList(entries []string) (AllowList, error) {
	var l AllowList
	for _, raw := range entries {
		e := strings.ToLower(strings.TrimSpace(raw))
		var pfx []string
		switch e {
		case "":
			continue
		case "localhost":
			pfx = localhostPrefixes
		case "private":
			pfx = privatePrefixes
		default:
			pfx = []string{e}
		}
		for _, p := range pfx {
			if !strings.Contains(p, "/") {
				a, err := netip.ParseAddr(p)
				if err != nil {
					return AllowList{deny: true}, fmt.Errorf("%q is not an address, a network like 192.168.1.0/24, \"localhost\" or \"private\"", raw)
				}
				p = netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()).String()
			}
			pre, err := netip.ParsePrefix(p)
			if err != nil {
				return AllowList{deny: true}, fmt.Errorf("%q is not an address, a network like 192.168.1.0/24, \"localhost\" or \"private\"", raw)
			}
			l.prefixes = append(l.prefixes, pre.Masked())
		}
	}
	return l, nil
}

// Restricted reports whether the list limits anyone.
func (l AllowList) Restricted() bool { return l.deny || len(l.prefixes) > 0 }

// Allows reports whether ip may use the pages the list protects.
func (l AllowList) Allows(ip netip.Addr) bool {
	if l.deny {
		return false
	}
	if len(l.prefixes) == 0 {
		return true
	}
	ip = ip.Unmap().WithZone("") // Prefix.Contains is false for an address with a zone (fe80::1%en0)
	for _, p := range l.prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
