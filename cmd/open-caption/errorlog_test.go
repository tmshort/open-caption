// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/logx"
)

func TestServerErrorsAreGivenALevelAndClearWords(t *testing.T) {
	for _, c := range []struct {
		line  string
		level logx.Level
		has   string
	}{
		{"http: TLS handshake error from 192.168.1.137:55910: remote error: tls: unknown certificate\n", logx.Debug, "did not accept the certificate"},
		{"http: TLS handshake error from 192.168.1.137:55921: remote error: tls: bad certificate", logx.Debug, "usually harmless"},
		{"http: TLS handshake error from [fe80::1]:5000: EOF", logx.Debug, "closed the connection"},
		{"http: TLS handshake error from 10.0.0.2:4000: read tcp 10.0.0.1:443->10.0.0.2:4000: read: connection reset by peer", logx.Debug, "closed the connection"},
		{"http: TLS handshake error from 10.0.0.2:4000: tls: client sent an HTTP request to an HTTPS server", logx.Info, "used http:// on the https port"},
		{"http: TLS handshake error from 10.0.0.2:4000: tls: first record does not look like a TLS handshake", logx.Warn, "handshake with 10.0.0.2:4000 failed"},
		{"http: panic serving 10.0.0.2:4000: boom", logx.Warn, "panic serving"},
	} {
		level, msg := classifyServerError(c.line)
		if level != c.level || !strings.Contains(msg, c.has) || strings.Contains(msg, "\n") {
			t.Errorf("%q -> %v %q, want %v containing %q", c.line, level, msg, c.level, c.has)
		}
	}
	if level, msg := classifyServerError("http: TLS handshake error from 192.168.1.137:55910: remote error: tls: unknown certificate"); level != logx.Debug || !strings.Contains(msg, "192.168.1.137:55910") {
		t.Errorf("the device's address is kept: %q", msg)
	}
}

func TestHTTPErrorLogGoesThroughTheLevels(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	defer logx.SetLevel(logx.Current())
	refused := "http: TLS handshake error from 1.2.3.4:5: remote error: tls: unknown certificate\n"
	logx.SetLevel(logx.Info)
	httpErrorLog().Print(strings.TrimSpace(refused))
	if buf.Len() != 0 {
		t.Errorf("a refused certificate was logged at the default level: %q", buf.String())
	}
	logx.SetLevel(logx.Debug)
	httpErrorLog().Print(strings.TrimSpace(refused))
	if !strings.Contains(buf.String(), "DEBUG") || !strings.Contains(buf.String(), "did not accept the certificate") {
		t.Errorf("not logged at debug: %q", buf.String())
	}
	buf.Reset()
	logx.SetLevel(logx.Warn)
	httpErrorLog().Print("http: panic serving 1.2.3.4:5: boom")
	if !strings.Contains(buf.String(), "WARN") || !strings.Contains(buf.String(), "panic serving") {
		t.Errorf("a real problem was not logged as a warning: %q", buf.String())
	}
}
