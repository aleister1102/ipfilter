SHELL := /bin/bash

BIN_NAME := ipfilter
SRC_DIR := ./cmd/ipfilter
BUILD_DIR := bin
DIST_DIR := dist
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
LEVEL ?= patch

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
		if [[ "$$GOOS" == "windows" ]]; then OUT="$$OUT.exe"; fi; \
		echo "Building $$OUT"; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} CGO_ENABLED=0 go build -o $$OUT $(SRC_DIR); \
	done

clean:
	rm -rf $(BUILD_DIR)

clean-release:
	rm -rf $(DIST_DIR)

test:
	go test ./...

.PHONY: version bump push-tags gh-release bump-patch bump-minor bump-major

version:
	@git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0


publish:
	@set -euo pipefail; \
	cur=$$(git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0); \
	v=$${cur#v}; IFS='.' read -r -a ver <<< "$$v"; \
	MA=$${ver[0]:-0}; MI=$${ver[1]:-0}; PA=$${ver[2]:-0}; \
	case "$(LEVEL)" in major) MA=$$((MA+1)); MI=0; PA=0 ;; minor) MI=$$((MI+1)); PA=0 ;; *) PA=$$((PA+1)) ;; esac; \
	new=v$$(printf "%d.%d.%d" $$MA $$MI $$PA); \
	echo $$new; \
	git tag -a $$new -m "Release $$new"; \
	$(MAKE) release; \
	$(MAKE) push-tags; \
	$(MAKE) gh-release TAG=$$new

bump-patch:
	@$(MAKE) publish LEVEL=patch

bump-minor:
	@$(MAKE) publish LEVEL=minor

bump-major:
	@$(MAKE) publish LEVEL=major

push-tags:
	@git push --tags

gh-release:
	@set -euo pipefail; tag=$${TAG:-$$(git describe --tags --abbrev=0 2>/dev/null)}; gh release create $$tag --title "ipfilter $$tag" --notes "Release $$tag" $(DIST_DIR)/*

publish-patch:
	@$(MAKE) bump-patch
	@$(MAKE) release
	@$(MAKE) push-tags
	@$(MAKE) gh-release

publish-minor:
	@$(MAKE) bump-minor
	@$(MAKE) release
	@$(MAKE) push-tags
	@$(MAKE) gh-release

publish-major:
	@$(MAKE) bump-major
	@$(MAKE) release
	@$(MAKE) push-tags
	@$(MAKE) gh-release
