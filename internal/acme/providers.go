//go:build !acme_off

package acme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-acme/lego/v4/challenge"
)

// Which DNS providers are linked in is decided when the program is built (see the Makefile, tools/acme-go.sh and docs/10):
//
//	no tags      cloudflare, godaddy, easydns (prov_*.go: the default)
//	acme_off     ACME left out entirely (lego is not linked)
//	acme_custom  exactly the providers in zz_providers_gen.go, which tools/acme-go.sh writes for the build: a list of
//	             names, or every provider lego has ("all")
//
// A provider registers itself by its lego name. Only the providers named are linked, so the program stays small.

type factory func() (challenge.Provider, error)

var (
	registry  = map[string]factory{}
	lookupAny func(string) (challenge.Provider, error) // set when every provider is built in
)

func register(name string, f factory) { registry[name] = f }

// Available says whether this program can get certificates at all.
func Available() bool { return true }

// Providers lists the DNS providers built in ("all" when every one is).
func Providers() []string {
	if lookupAny != nil {
		return []string{"all"}
	}
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// HasProvider says whether the named provider is built in.
func HasProvider(name string) bool {
	if lookupAny != nil {
		return true
	}
	_, ok := registry[name]
	return ok
}

func newProvider(name string) (challenge.Provider, error) {
	if f, ok := registry[name]; ok {
		return f()
	}
	if lookupAny != nil {
		return lookupAny(name)
	}
	return nil, fmt.Errorf("the DNS provider %q is not built into this program (built with: %s); see docs/10-acme-lets-encrypt.md for how to build with others",
		name, strings.Join(Providers(), ", "))
}

func sum(s string) uint32 { // a short stable name for a CA address
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h
}
