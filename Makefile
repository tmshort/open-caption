VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.DEFAULT_GOAL := help
.PHONY: help build test race fuzz vet e2e a11y i18n dist clean

help: ## Show this list of targets
	@echo "open-caption: make <target>"; echo
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-8s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/open-caption for this computer
	go build -ldflags "$(LDFLAGS)" -o bin/open-caption ./cmd/open-caption
test: ## Run the Go tests
	go test ./...
race: ## Run the Go tests with the race detector
	go test -race ./...
fuzz: ## Fuzz the show-file parser for 30 seconds
	go test ./internal/showfile -run xxx -fuzz FuzzParse -fuzztime 30s
vet: ## Run go vet
	go vet ./...
e2e: build ## Browser end-to-end test (needs Node and Chrome; set CHROME if not in the default place)
	cd test/e2e && npm install --no-audit --no-fund && node run.js
a11y: build ## Accessibility audit with axe-core in real Chrome (needs Node and Chrome)
	cd test/e2e && npm install --no-audit --no-fund && node a11y.js
i18n: ## Check every sentence is translated into all six languages (needs Node)
	cd test/e2e && npm install --no-audit --no-fund && node check-i18n.js
dist: ## Cross-compile release binaries into dist/ (no CGO)
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-arm64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-windows-amd64.exe ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-arm64 ./cmd/open-caption
clean: ## Remove bin/ and dist/
	rm -rf bin dist
