// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

//go:build !acme_off && !acme_custom && !acme_all

package acme

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
)

func TestDefaultBuildHasCloudflareGoDaddyAndEasyDNS(t *testing.T) {
	if got := Providers(); !reflect.DeepEqual(got, []string{"cloudflare", "easydns", "godaddy"}) {
		t.Errorf("providers = %v", got)
	}
	if !Available() || !HasProvider("easydns") || HasProvider("route53") {
		t.Error("HasProvider")
	}
}

func TestAProviderThatIsNotBuiltInIsRefusedWithTheListOfWhatIs(t *testing.T) {
	_, err := NewIssuer(config.ACME{DNS: config.ACMEDNS{Provider: "route53"}}, t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "not built into this program") || !strings.Contains(err.Error(), "cloudflare, easydns, godaddy") {
		t.Errorf("%v", err)
	}
	if _, err := NewIssuer(config.ACME{DNS: config.ACMEDNS{Provider: "easydns"}}, t.TempDir(), nil); err != nil {
		t.Errorf("easydns is built in: %v", err)
	}
}
