BINARY  := cx
GOOS    := darwin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all fmt vet staticcheck gosec lint build clean

all: lint build

fmt:
	@echo "==> gofmt"
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet:
	@echo "==> go vet"
	go vet ./...

staticcheck:
	@echo "==> staticcheck"
	@which staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	$$(go env GOPATH)/bin/staticcheck ./...

gosec:
	@echo "==> gosec"
	@which gosec >/dev/null 2>&1 || go install github.com/securego/gosec/v2/cmd/gosec@latest
	$$(go env GOPATH)/bin/gosec ./...

lint: fmt vet staticcheck gosec

build:
	@echo "==> build"
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cx

clean:
	rm -f $(BINARY) $(BINARY)-darwin-*
