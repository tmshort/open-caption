// Package web embeds the browser UI assets into the binary (REQ-002).
package web

import "embed"

// Audience holds the audience page (index.html, audience.js, audience.css).
//
//go:embed audience
var Audience embed.FS
