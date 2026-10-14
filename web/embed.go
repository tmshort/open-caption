// Package web embeds the browser UI assets into the binary (REQ-002).
package web

import "embed"

// Audience holds the audience page (index.html, audience.js, audience.css).
//
//go:embed audience
var Audience embed.FS

// Operator holds the operator page (/operator).
//
//go:embed operator
var Operator embed.FS

// Admin holds the setup page (/admin).
//
//go:embed admin
var Admin embed.FS

// Sample holds the example show (script + placeholder images).
//
//go:embed sample/sample-script.md sample/assets
var Sample embed.FS

// Help holds the help topics (Markdown), named admin-*.md and operator-*.md.
//
//go:embed help
var Help embed.FS

// HelpUI holds the help viewer page.
//
//go:embed helpui
var HelpUI embed.FS
