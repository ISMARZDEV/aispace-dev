BINARY     := ai-setup
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION)
BUILD_DIR  := dist

.PHONY: build build-all test lint clean install release-dry

## build: compile for current platform
build:
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/ai-setup

## build-all: cross-compile darwin + linux
build-all:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/ai-setup
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/ai-setup
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64  ./cmd/ai-setup

## test: run all tests
test:
	go test ./...

## install: install binary to $GOPATH/bin
install:
	go install -ldflags="$(LDFLAGS)" ./cmd/ai-setup

## release-dry: dry-run GoReleaser (requires goreleaser installed)
release-dry:
	goreleaser release --snapshot --clean

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
