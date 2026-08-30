MODULE := github.com/seneal/wtrto
VERSION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BUILD_DATE := $(shell date -u +%Y-%m-%d)
BUILD_HASH := $(shell find . -name '*.go' -not -path './bin/*' | LC_ALL=C sort | xargs cat | sha256sum | cut -c1-12)
LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE) -X $(MODULE)/internal/version.BuildHash=$(BUILD_HASH)

.PHONY: build build-linux build-windows release run

build: build-linux

build-linux:
	go build -ldflags "$(LDFLAGS)" -o bin/wtrto ./cmd/wtrto

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -H=windowsgui" -o bin/wtrto.exe ./cmd/wtrto

release: build-linux build-windows

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/wtrto
