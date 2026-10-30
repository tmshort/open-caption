// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package config

import "strings"

// CheckDefaults checks the venue-wide defaults and audience limits that Setup edits and returns them in their normal form
// (colours as #rrggbb, for example), or everything that is wrong with them. Nothing else in the configuration is looked at.
func CheckDefaults(d Defaults, a Audience) (Defaults, Audience, []string) {
	c := Default()
	c.Defaults, c.Audience = d, a
	var probs []string
	for _, p := range c.normalize() {
		if strings.HasPrefix(p, "defaults.") || strings.HasPrefix(p, "audience.") {
			probs = append(probs, p)
		}
	}
	return c.Defaults, c.Audience, probs
}
