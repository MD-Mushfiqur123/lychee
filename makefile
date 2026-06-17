.PHONY: build test clean install release lint

# Binary name
BINARY=lychee
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X github.com/MD-Mushfiqur123/lychee/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o dist/$(BINARY) .

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .

test:
	go test ./...

vet:
	go vet ./...

lint:
	go vet ./...
	gofmt -s -l .

clean:
	rm -rf dist/

install:
	go install $(LDFLAGS) .

release: build-all
	@echo "Release $(VERSION) built in dist/"

smoke:
	./scripts/smoke-test.sh

help:
	@echo "Lychee Makefile"
	@echo ""
	@echo "  make build      Build binary"
	@echo "  make build-all  Cross-compile for all platforms"
	@echo "  make test       Run tests"
	@echo "  make vet        Run go vet"
	@echo "  make clean      Remove build artifacts"
	@echo "  make install    Install via go install"
	@echo "  make release    Build all platforms for release"
	@echo "  make smoke      Run API smoke test"
