.PHONY: all build lint format test completion docs generate-base update-intellij-keymaps update-zed-keymaps

.DEFAULT_GOAL := all

all: build format lint test completion docs generate-base

GO_LINT_CONFIG := $(CURDIR)/.golangci.yaml
GO_ENV := GO111MODULE=on GOFLAGS=-mod=mod
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_DIRTY := $(shell if test -n "$$(git status --porcelain 2>/dev/null)"; then echo "true"; else echo "false"; fi)
GO_LDFLAGS := -X github.com/xinnjie/onekeymap-cli/internal/cmd.commit=$(GIT_COMMIT) -X github.com/xinnjie/onekeymap-cli/internal/cmd.dirty=$(GIT_DIRTY)

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
