BINARY := kvmcli

GO       := /usr/local/go/bin/go
QEMU_IMG := $(shell which qemu-img)

CONFIG := ./configs/admins.hcl

VERSION := $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

PROJECT := github.com/zakariakebairia/kvmcli

LDFLAGS := \
	-X $(PROJECT)/cmd.Version=$(VERSION) \
	-X $(PROJECT)/cmd.Commit=$(COMMIT) \
	-X $(PROJECT)/cmd.Built=$(BUILD_DATE) \
	-X $(PROJECT)/vm.QemuImgBinary=$(QEMU_IMG)

.PHONY: \
	build \
	install \
	run \
	fmt \
	vet \
	test \
	clean \
	recreate-admins

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------

build:
	@echo "Building $(BINARY)..."
	@CGO_CFLAGS="-Wno-discarded-qualifiers" \
	$(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY) .

install: build
	@cp $(BINARY) ~/.local/bin/

run: build
	@./$(BINARY)

clean:
	@rm -f $(BINARY)

# -----------------------------------------------------------------------------
# Quality
# -----------------------------------------------------------------------------

fmt:
	@$(GO) fmt ./...

vet:
	@$(GO) vet ./...

test:
	@$(GO) test ./...

# -----------------------------------------------------------------------------
# Development
# -----------------------------------------------------------------------------

create: build
	@./$(BINARY) delete -f $(CONFIG)
	@./$(BINARY) create -f $(CONFIG)
