// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/sha1" // #nosec G505 -- only offered for digests a person already has; the checker warns about it
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

// A password in config.yaml (admin_password, operator_password) is plain text unless a prefix says otherwise:
//
//	hunter22                  plain text (no prefix)
//	plain:sha256:hunter22     plain text, written with a prefix so that it may itself start with "sha256:" and the like
//	sha256:<64 hex digits>    the SHA-256 of the password (also sha512: and, though weak, sha1:)
//
// A hashed password cannot be shown again, only checked.
const PlainPrefix = "plain:"

type hashScheme struct {
	name string
	size int // bytes in the digest
	sum  func([]byte) []byte
}

var hashSchemes = []hashScheme{
	{"sha256", sha256.Size, func(b []byte) []byte { s := sha256.Sum256(b); return s[:] }},
	{"sha512", sha512.Size, func(b []byte) []byte { s := sha512.Sum512(b); return s[:] }},
	{"sha1", sha1.Size, func(b []byte) []byte { s := sha1.Sum(b); return s[:] }}, // #nosec G401
}

// parsedPassword is a stored password taken apart: either the plain text, or a scheme and its digest.
type parsedPassword struct {
	plain  string
	scheme *hashScheme
	digest []byte
	bad    string // the problem with a value that begins like a hash but is not one
}

func parsePassword(stored string) parsedPassword {
	if rest, ok := strings.CutPrefix(stored, PlainPrefix); ok {
		return parsedPassword{plain: rest}
	}
	for i := range hashSchemes {
		sc := &hashSchemes[i]
		hexDigits, ok := strings.CutPrefix(stored, sc.name+":")
		if !ok {
			continue
		}
		d, err := hex.DecodeString(hexDigits)
		if err != nil || len(d) != sc.size {
			return parsedPassword{bad: fmt.Sprintf("starts with %q but is not a %s digest of %d hex characters (to use a password that begins like that as it is, write %q before it)",
				sc.name+":", sc.name, 2*sc.size, PlainPrefix)}
		}
		return parsedPassword{scheme: sc, digest: d}
	}
	return parsedPassword{plain: stored}
}

// IsHashed reports whether a stored password is a digest, so it can be checked but never shown.
func IsHashed(stored string) bool { return parsePassword(stored).scheme != nil }

// PlainPassword is the password itself when the stored value is plain text.
func PlainPassword(stored string) (string, bool) {
	p := parsePassword(stored)
	return p.plain, p.scheme == nil && p.bad == ""
}

// PasswordProblem says what is wrong with a stored password ("" when it is usable).
func PasswordProblem(stored string) string {
	p := parsePassword(stored)
	switch {
	case p.bad != "":
		return p.bad
	case p.scheme == nil && p.plain == "":
		return "must not be empty"
	}
	return ""
}

// CheckPassword reports whether given is the password that stored holds (in constant time).
func CheckPassword(stored, given string) bool {
	p := parsePassword(stored)
	switch {
	case p.bad != "":
		return false
	case p.scheme != nil:
		return subtle.ConstantTimeCompare(p.digest, p.scheme.sum([]byte(given))) == 1
	}
	a, b := sha256.Sum256([]byte(p.plain)), sha256.Sum256([]byte(given))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

// SamePassword reports whether two stored passwords are the same password, as far as can be told: a digest is
// compared with plain text by hashing it, and two digests only when made the same way.
func SamePassword(a, b string) bool {
	pa, pb := parsePassword(a), parsePassword(b)
	switch {
	case pa.bad != "" || pb.bad != "":
		return false
	case pa.scheme == nil:
		return CheckPassword(b, pa.plain)
	case pb.scheme == nil:
		return CheckPassword(a, pb.plain)
	}
	return pa.scheme == pb.scheme && subtle.ConstantTimeCompare(pa.digest, pb.digest) == 1
}

// HashPassword is the stored form (sha256:...) of a password, for putting in config.yaml in place of the plain text.
func HashPassword(password string) string {
	sc := hashSchemes[0]
	return sc.name + ":" + hex.EncodeToString(sc.sum([]byte(password)))
}

// StoredLike is how to store a password set in Setup: as a SHA-256 digest, so that nothing readable is kept in the file. The
// one exception is a password that was written with the plain: prefix on purpose, which stays plain text (with the prefix).
func StoredLike(old, password string) string {
	if strings.HasPrefix(old, PlainPrefix) {
		return PlainPrefix + password
	}
	return HashPassword(password)
}

// PasswordID identifies a stored password without revealing it (for noticing that a password changed); the same
// for a plain password as it has always been, the SHA-256 of the password.
func PasswordID(stored string) [32]byte {
	if p := parsePassword(stored); p.scheme == nil && p.bad == "" {
		return sha256.Sum256([]byte(p.plain))
	}
	return sha256.Sum256([]byte(stored))
}
