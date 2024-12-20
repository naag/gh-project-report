# Basic variables
BINARY_NAME=gh-project-report
BINARY_DIR=bin
GO=go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

.PHONY: all build clean test

all: clean build

build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BINARY_DIR}
	@${GO} build ${LDFLAGS} -o ${BINARY_DIR}/${BINARY_NAME}

clean:
	@echo "Cleaning up..."
	@rm -rf ${BINARY_DIR}

test:
	@echo "Running tests..."
	@${GO} test -v ./... 