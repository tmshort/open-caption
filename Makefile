VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
# Which DNS providers for automatic certificates (ACME DNS-01) are built in: a list of names, "all", or "none" (docs/10).
# The default adds about 0.3 MB; for example: make build ACME_PROVIDERS="easydns route53"
ACME_PROVIDERS ?= cloudflare godaddy easydns

.DEFAULT_GOAL := help
.PHONY: help build test race fuzz vet vulncheck tidy e2e a11y i18n dist man acme-variants clean

help: ## Show this list of targets
	@echo "open-caption: make <target>"; echo
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/open-caption for this computer (ACME_PROVIDERS="..." | all | none picks the DNS providers for automatic certificates)
	tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o bin/open-caption ./cmd/open-caption
test: ## Run the Go tests
	go test ./...
race: ## Run the Go tests with the race detector
	go test -race ./...
fuzz: ## Fuzz the show-file parser for 30 seconds
	go test ./internal/showfile -run xxx -fuzz FuzzParse -fuzztime 30s
vet: ## Run go vet
	go vet ./...
vulncheck: ## Check the dependencies and Go for known vulnerabilities (govulncheck; needs internet)
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
tidy: ## Tidy go.mod and go.sum (go mod tidy)
	go mod tidy
e2e: build ## Browser end-to-end test (needs Node and Chrome; set CHROME if not in the default place)
	cd test/e2e && npm install --no-audit --no-fund && node run.js
a11y: build ## Accessibility audit with axe-core in real Chrome (needs Node and Chrome)
	cd test/e2e && npm install --no-audit --no-fund && node a11y.js
i18n: ## Check every sentence is translated into all six languages (needs Node)
	cd test/e2e && npm install --no-audit --no-fund && node check-i18n.js
dist: ## Cross-compile release binaries into dist/ with the man page (no CGO; same ACME_PROVIDERS choice as build)
	mkdir -p dist
	cp docs/open-caption.1 dist/
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-arm64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o dist/open-caption-darwin-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o dist/open-caption-windows-amd64.exe ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-amd64 ./cmd/open-caption
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 tools/acme-go.sh "$(ACME_PROVIDERS)" build -ldflags "$(LDFLAGS)" -o dist/open-caption-linux-arm64 ./cmd/open-caption
man: ## Show the manual page (docs/open-caption.1)
	man ./docs/open-caption.1
acme-variants: ## Check that the ACME build choices compile: none, the default, one provider, a custom list, and all (needs internet)
	@for v in none "" "easydns" "easydns route53" all; do \
	  echo "== ACME_PROVIDERS=$${v:-default}"; tools/acme-go.sh "$$v" build -o /dev/null ./cmd/open-caption || exit 1; \
	done
clean: ## Remove bin/ and dist/ (and the files tools/acme-go.sh generates)
	rm -rf bin dist internal/acme/zz_providers_gen.go go.acme.mod go.acme.sum
