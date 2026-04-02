BINARY  := cx
GOOS    := darwin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all lint shellcheck vulncheck test build clean

all: lint shellcheck vulncheck test build

lint:
	@echo "==> golangci-lint"
	@which golangci-lint >/dev/null 2>&1 || { echo "Install: brew install golangci-lint"; exit 1; }
	golangci-lint run ./...

shellcheck:
	@echo "==> shellcheck"
	@which shellcheck >/dev/null 2>&1 || { echo "Install: brew install shellcheck"; exit 1; }
	shellcheck cmd/cx/completions/cx.bash

vulncheck:
	@echo "==> govulncheck"
	@which govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	$$(go env GOPATH)/bin/govulncheck ./...

test:
	@echo "==> test"
	go test -v -count=1 ./...

build:
	@echo "==> build"
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cx

clean:
	rm -f $(BINARY) $(BINARY)-darwin-*
