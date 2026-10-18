package acme

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

// Credentials gathers the provider's settings (for example EASYDNS_TOKEN) from three places and puts the missing
// ones into the process environment, where the DNS providers read them. Precedence, highest first:
//
//  1. the environment already there, including NAME_FILE variables that name a file holding the value;
//  2. the credentials_file (NAME=value lines);
//  3. the credentials given in config.yaml.
//
// The file is read again at every attempt, so a rotated token needs no restart. Values are never logged.
type Credentials struct {
	Inline map[string]string
	File   string
	Warnf  func(string, ...any)
}

// Apply returns the names it set (never their values).
func (c Credentials) Apply() ([]string, error) {
	vals := map[string]string{}
	for k, v := range c.Inline {
		vals[k] = v
	}
	if c.File != "" {
		fromFile, err := ParseEnvFile(c.File, c.Warnf)
		if err != nil {
			return nil, err
		}
		for k, v := range fromFile {
			vals[k] = v
		}
	}
	var set []string
	for k, v := range vals {
		if os.Getenv(k) != "" || os.Getenv(k+"_FILE") != "" {
			continue // the environment wins
		}
		if err := os.Setenv(k, v); err != nil {
			return nil, err
		}
		set = append(set, k)
	}
	sort.Strings(set)
	return set, nil
}

// ParseEnvFile reads NAME=value lines: blank lines and lines starting with # are ignored, an optional leading
// "export " is allowed, and a value may be wrapped in single or double quotes. It warns when other users can read the file.
func ParseEnvFile(path string, warnf func(string, ...any)) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the DNS credentials file: %w", err)
	}
	defer f.Close()
	if st, err := f.Stat(); err == nil && runtime.GOOS != "windows" && st.Mode().Perm()&0o077 != 0 && warnf != nil {
		warnf("the DNS credentials file %s can be read by other users; run: chmod 600 %s", path, path)
	}
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if !ok || !validName(k) {
			return nil, fmt.Errorf("%s line %d: expected NAME=value (for example EASYDNS_TOKEN=abc123)", path, n)
		}
		if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
			v = v[1 : len(v)-1]
		}
		out[k] = v
	}
	return out, sc.Err()
}

func validName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}
