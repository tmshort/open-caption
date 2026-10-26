#!/bin/sh
# Copyright the open-caption authors
# SPDX-License-Identifier: Apache-2.0

# Go test coverage: runs the tests with every package counted (-coverpkg), writes coverage/go.out and coverage/go.html
# (open it to see which lines no test ran), and prints the share of statements run per package and in total.
# Usage: tools/cover.sh        (or: make cover)
set -e
cd "$(dirname "$0")/.."
mkdir -p coverage
pkgs=$(go list ./... | grep -v /node_modules/)
list=$(echo "$pkgs" | tr '\n' ',' | sed 's/,$//')
# shellcheck disable=SC2086
go test -coverpkg="$list" -coverprofile=coverage/go.out $pkgs >coverage/go-test.log 2>&1 || {
	cat coverage/go-test.log
	echo "cover: the tests failed" >&2
	exit 1
}
go tool cover -html=coverage/go.out -o coverage/go.html
# One line per block: file:from,to statements count. With -coverpkg every test run reports every block, so a block is
# counted once, as run if any test run ran it; then add them up per package.
awk 'NR > 1 {
	key = $1; n = $2; c = $3
	size[key] = n
	if (c > 0) ran[key] = 1
	else if (!(key in ran)) ran[key] = 0
}
END {
	for (k in size) {
		f = k; sub(/:.*$/, "", f); sub(/\/[^\/]*$/, "", f); sub(/^github.com\/tmshort\/open-caption\/?/, "", f); if (f == "") f = "."
		stmts[f] += size[k]; all += size[k]
		if (ran[k]) { hit[f] += size[k]; allhit += size[k] }
	}
	for (p in stmts) printf "%6.1f%%  %5d statements  %s\n", 100 * hit[p] / stmts[p], stmts[p], p | "sort -rn"
	close("sort -rn")
	printf "\n%6.1f%%  %5d statements  total\n", 100 * allhit / all, all
}' coverage/go.out
echo
echo "Lines no test ran: open coverage/go.html (go tool cover -func=coverage/go.out lists every function)."
