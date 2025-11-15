SHELL := /bin/bash

BIN_NAME := ipfilter
SRC_DIR := ./cmd/ipfilter
BUILD_DIR := bin
DIST_DIR := dist
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: build release clean clean-release test

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BIN_NAME) $(SRC_DIR)

release: clean-release
	@mkdir -p $(DIST_DIR)
	@set -euo pipefail; \
	for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		OUT="$(DIST_DIR)/$(BIN_NAME)-$${GOOS}-$${GOARCH}"; \
		echo "Building $$OUT"; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} go build -o $$OUT $(SRC_DIR); \
	done

clean:
	rm -rf $(BUILD_DIR)

clean-release:
	rm -rf $(DIST_DIR)

test:
	go test ./...

