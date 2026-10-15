VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test race fuzz vet e2e dist clean
build:
	go build -ldflags "$(LDFLAGS)" -o bin/open-caption ./cmd/open-caption
test:
	go test ./...
race:
	go test -race ./...
fuzz:
	go test ./internal/showfile -run xxx -fuzz FuzzParse -fuzztime 30s
vet:
	go vet ./...
# Browser end-to-end test: needs Node and Chrome (set CHROME to its path if not in the default place).
e2e: build
	cd test/e2e && npm install --no-audit --no-fund && node run.js
# Cross-compile release binaries (REQ-001): no CGO.
dist:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-arm64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-windows-amd64.exe ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-arm64 ./cmd/open-caption
clean:
	rm -rf bin dist
