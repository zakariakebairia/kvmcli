# Makefile for kvmcli project

# BINARY_NAME sets the name of the output executable.
BINARY_NAME = kvmcli
BINARY_PATH = /usr/local/go/bin/go
VERSION ?= $(shell git describe --tags --always --dirty)

# The default target: when you run "make" without arguments, it will run the "build" target.
all: build

# build: Compiles the Go project into a binary executable.
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_CFLAGS="-Wno-discarded-qualifiers" $(BINARY_PATH)  build -ldflags "-X github.com/zakariakebairia/kvmcli/cmd.Version=$(VERSION)" -o $(BINARY_NAME) .
	cp $(BINARY_NAME) ~/.local/bin/

# run: Builds the project (if necessary) and runs the executable.
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# fmt: Formats the source code using the built-in Go formatter.
fmt:
	@echo "Formatting code..."
	go fmt ./...

# vet: Runs static analysis to catch potential issues.
vet:
	@echo "Checking code with go vet..."
	go vet ./...

# test: Executes the unit tests in the project.
test:
	@echo "Running tests..."
	go test ./...

# clean: Removes the built binary to clean up the project directory.
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
