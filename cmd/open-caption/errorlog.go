// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/tmshort/open-caption/internal/logx"
)

// httpErrorLog is the logger for http.Server's own complaints. Left alone the server writes them straight to the
// standard logger, with no level (so log.level cannot quiet them), no place in Setup's recent log, and wording that
// reads like a fault when it is usually a browser trying https first and giving up on the certificate.
func httpErrorLog() *log.Logger { return log.New(errorWriter{}, "", 0) }

type errorWriter struct{}

func (errorWriter) Write(p []byte) (int, error) {
	level, msg := classifyServerError(string(p))
	switch level {
	case logx.Debug:
		logx.Debugf("%s", msg)
	case logx.Info:
		logx.Infof("%s", msg)
	default:
		logx.Warnf("%s", msg)
	}
	return len(p), nil
}

var handshakeRe = regexp.MustCompile(`^http: TLS handshake error from (\S+?): (.*)$`)

// classifyServerError chooses a level and a clearer message for one line of http.Server's error log.
func classifyServerError(line string) (logx.Level, string) {
	line = strings.TrimSpace(line)
	m := handshakeRe.FindStringSubmatch(line)
	if m == nil {
		return logx.Warn, line
	}
	who, why := m[1], m[2]
	switch {
	case strings.Contains(why, "unknown certificate"), strings.Contains(why, "bad certificate"), strings.Contains(why, "certificate unknown"),
		strings.Contains(why, "unknown certificate authority"), strings.Contains(why, "certificate required"), strings.Contains(why, "certificate expired"):
		return logx.Debug, "HTTPS: " + who + " did not accept the certificate (" + why + "). Browsers often try https first and then use http, so this is usually harmless."
	case why == "EOF", strings.Contains(why, "connection reset"), strings.Contains(why, "broken pipe"), strings.Contains(why, "i/o timeout"):
		return logx.Debug, "HTTPS: " + who + " closed the connection during the handshake (" + why + ")"
	case strings.Contains(why, "client sent an HTTP request to an HTTPS server"):
		return logx.Info, "HTTPS: " + who + " used http:// on the https port"
	}
	return logx.Warn, "HTTPS: handshake with " + who + " failed: " + why
}
