// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package logx adds severity levels to the standard logger: lines are written as "DEBUG …", "INFO …",
// "WARN …" or "ERROR …", and lines below the chosen level are dropped (REQ-061).
package logx

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

// Level is a log severity.
type Level int32

// The levels, least to most severe.
const (
	Debug Level = iota
	Info
	Warn
	Error
)

var names = [...]string{"debug", "info", "warn", "error"}

func (l Level) String() string {
	if l < Debug || l > Error {
		return "info"
	}
	return names[l]
}

var exit = os.Exit

var current atomic.Int32

func init() { current.Store(int32(Info)) }

// Parse reads a level name (case-insensitive; "warning" is accepted for "warn"). Empty means info.
func Parse(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return Info, nil
	case "debug":
		return Debug, nil
	case "warn", "warning":
		return Warn, nil
	case "error":
		return Error, nil
	}
	return Info, fmt.Errorf("%q is not a log level: use debug, info, warn or error", s)
}

// SetLevel sets the lowest severity that is written.
func SetLevel(l Level) { current.Store(int32(l)) }

// Current returns the lowest severity that is written.
func Current() Level { return Level(current.Load()) }

// Enabled reports whether lines at l are written.
func Enabled(l Level) bool { return l >= Current() }

func out(l Level, format string, args []any) {
	if !Enabled(l) {
		return
	}
	log.Output(3, strings.ToUpper(l.String())+" "+fmt.Sprintf(format, args...))
}

// Debugf logs detail that is only useful when diagnosing a problem.
func Debugf(format string, args ...any) { out(Debug, format, args) }

// Infof logs normal events: starting, signing in, changing show.
func Infof(format string, args ...any) { out(Info, format, args) }

// Warnf logs something unexpected that the program carried on from.
func Warnf(format string, args ...any) { out(Warn, format, args) }

// Errorf logs a failure.
func Errorf(format string, args ...any) { out(Error, format, args) }

// Fatalf logs a failure and exits (always written, whatever the level).
func Fatalf(format string, args ...any) {
	log.Output(2, "ERROR "+fmt.Sprintf(format, args...))
	exit(1)
}
