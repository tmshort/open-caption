#!/bin/sh
# Writes the README that goes into a release download: this repository's README.md without the parts that only
# make sense next to the source code (they are fenced by <!-- source-only:start --> and <!-- source-only:end -->
# comment lines, which GitHub does not show). Used by .goreleaser.yaml; run it by hand to see the result.
#
#   tools/package-readme.sh [output-file]      (default: .tools/README.release.md)
set -e
cd "$(dirname "$0")/.."
out="${1:-.tools/README.release.md}"
mkdir -p "$(dirname "$out")"
awk '
  /<!-- source-only:start -->/ { skip = 1; next }
  /<!-- source-only:end -->/   { skip = 0; next }
  !skip
' README.md > "$out"
echo "wrote $out"
