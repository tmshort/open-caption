// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// Log messages follow one style (CLAUDE.md, "Logging"): a lowercase sentence, with a prefix only when the message is
// about one thing the operator knows by name: a feature (HTTPS:, mDNS:), the settings file (config:), or a setting or
// option (https.acme:, --dev, --log-level). This keeps new messages from drifting.
func TestLogMessagesFollowTheStyle(t *testing.T) {
	logxFuncs := map[string]bool{"Debugf": true, "Infof": true, "Warnf": true, "Errorf": true, "Fatalf": true}
	callbacks := map[string]bool{"infof": true, "warnf": true, "Infof": true, "Warnf": true} // the log functions handed to the certificate code
	features := map[string]bool{"HTTPS": true, "mDNS": true}
	checked := 0
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "bin", "dist", ".tools", "web", "test", "tools", "docs":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "loadtest") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			isLog := false
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				if x, ok := fn.X.(*ast.Ident); ok && x.Name == "logx" {
					isLog = logxFuncs[fn.Sel.Name]
				} else {
					isLog = callbacks[fn.Sel.Name]
				}
			case *ast.Ident:
				isLog = callbacks[fn.Name]
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || !isLog {
				return true
			}
			msg, _ := strconv.Unquote(lit.Value)
			if msg == "" || msg == "%s" || msg == "%v" {
				return true // a message made elsewhere (and checked where it is made)
			}
			checked++
			first, _ := firstRune(msg)
			prefix := ""
			if i := strings.Index(msg, ": "); i > 0 && !strings.ContainsAny(msg[:i], " %") {
				prefix = msg[:i]
			}
			switch {
			case prefix != "" && (features[prefix] || prefix == "config" || strings.Contains(prefix, ".") || strings.HasPrefix(prefix, "--")):
			case prefix != "":
				t.Errorf("%s: %q: a prefix must be a feature (HTTPS, mDNS), config, or a setting or option name", path, msg)
			case unicode.IsLower(first) || first == '%' || strings.HasPrefix(msg, "--") || features[strings.Fields(msg)[0]]:
			case strings.HasPrefix(msg, "HTTPS "):
			default:
				t.Errorf("%s: %q: start with a lowercase letter (an acronym such as HTTPS may be capitalised)", path, msg)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 60 {
		t.Errorf("only %d log messages were found: has the way messages are written changed?", checked)
	}
}

func firstRune(s string) (rune, int) {
	for i, r := range s {
		return r, i
	}
	return 0, 0
}
