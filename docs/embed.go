// Package docs embeds the user-facing guides so the program can show them in its Help pages.
// The Markdown files in this folder are the single source of truth.
package docs

import "embed"

// Guides holds the guides shown in the app's Help.
//
//go:embed usage-guide.md venue-network-guide.md
var Guides embed.FS
