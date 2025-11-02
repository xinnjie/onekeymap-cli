.PHONY: all build lint format test completion docs generate-base update-intellij-keymaps update-zed-keymaps image release-image

.DEFAULT_GOAL := all

all: build format lint test completion docs generate-base

GO_LINT_CONFIG := $(CURDIR)/.golangci.yaml
GO_ENV := GO111MODULE=on GOFLAGS=-mod=mod
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_DIRTY := $(shell if test -n "$$(git status --porcelain 2>/dev/null)"; then echo "true"; else echo "false"; fi)
GO_LDFLAGS := -X github.com/xinnjie/onekeymap-cli/internal/cmd.commit=$(GIT_COMMIT) -X github.com/xinnjie/onekeymap-cli/internal/cmd.dirty=$(GIT_DIRTY)

# ko configuration
KO_DOCKER_REPO ?= ghcr.io/xinnjie/onekeymap-cli
KO_PLATFORMS ?= linux/amd64,linux/arm64

build:
	@mkdir -p .bin
	CGO_ENABLED=0 go build -ldflags "$(GO_LDFLAGS)" -o .bin/onekeymap-cli ./cmd/onekeymap-cli

format:
	@$(GO_ENV) go fmt ./...

lint:
	@$(GO_ENV) golangci-lint run --fix --config $(GO_LINT_CONFIG)

test:
	@$(GO_ENV) go test ./...

completion:
	mkdir -p completions
	@$(GO_ENV) go run -ldflags "$(GO_LDFLAGS)" ./cmd/onekeymap-cli completion bash > completions/onekeymap-cli.bash
	@$(GO_ENV) go run -ldflags "$(GO_LDFLAGS)" ./cmd/onekeymap-cli completion zsh > completions/_onekeymap-cli
	@$(GO_ENV) go run -ldflags "$(GO_LDFLAGS)" ./cmd/onekeymap-cli completion fish > completions/onekeymap-cli.fish

docs:
	@$(GO_ENV) go run -ldflags "$(GO_LDFLAGS)" ./cmd/onekeymap-cli dev docSupportActions &> ./action-support-matrix.md

generate-base:
	@$(GO_ENV) go run -ldflags "$(GO_LDFLAGS)" ./cmd/onekeymap-cli dev generateBase

update-intellij-keymaps:
	@python3 scripts/update_keymaps.py --preset intellij

update-zed-keymaps:
	@python3 scripts/update_keymaps.py --preset zed

update-vscode-keymaps:
	@python3 scripts/update_keymaps.py --preset vscode

# release-image: Build and push release container images with version tags
# Usage: VERSION=0.5.1 make release-image
# For local testing: KO_DOCKER_REPO=ko.local VERSION=0.0.1-test make release-image
release-image:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: VERSION=0.5.1 make release-image"; \
		exit 1; \
	fi
	@echo "Building and pushing multi-platform images to $(KO_DOCKER_REPO)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Platforms: $(KO_PLATFORMS)"
	VERSION=$(VERSION) \
	GIT_COMMIT=$(GIT_COMMIT) \
	GIT_DIRTY=false \
	KO_DOCKER_REPO=$(KO_DOCKER_REPO) \
	ko build \
		--bare \
		--platform=$(KO_PLATFORMS) \
		--sbom=none \
		--tags=$(VERSION) \
		--tags=$(GIT_COMMIT) \
		--tags=latest \
		github.com/xinnjie/onekeymap-cli/cmd/onekeymap-cli
