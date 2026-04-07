BINARY_NAME      = kvmcli
GO               = /usr/local/go/bin/go
QEMU_IMG_BINARY  = $(shell which qemu-img)

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
BUILT   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

PROJECT     := github.com/zakariakebairia/kvmcli
LDFLAGS := -X $(PROJECT)/cmd.Version=$(VERSION) \
           -X $(PROJECT)/cmd.Commit=$(COMMIT) \
           -X $(PROJECT)/cmd.Built=$(BUILT) \
           -X $(PROJECT)/vm.QemuImgBinary=$(QEMU_IMG_BINARY)

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	CGO_CFLAGS="-Wno-discarded-qualifiers" $(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

publish: build
	cp $(BINARY_NAME) ~/.local/bin/

run: build
	./$(BINARY_NAME)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY_NAME)

.PHONY: all build publish run fmt vet test clean
