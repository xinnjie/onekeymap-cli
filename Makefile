.PHONY: build lint format test

GO_LINT_CONFIG := $(CURDIR)/.golangci.yaml
GO_ENV := GO111MODULE=on GOFLAGS=-mod=mod
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_DIRTY := $(shell if test -n "$$(git status --porcelain 2>/dev/null)"; then echo "true"; else echo "false"; fi)
GO_LDFLAGS := -X github.com/xinnjie/onekeymap-cli/internal/cmd.commit=$(GIT_COMMIT) -X github.com/xinnjie/onekeymap-cli/internal/cmd.dirty=$(GIT_DIRTY)

build:
	@mkdir -p .bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o .bin/onekeymap-cli-arm64 ./cmd/onekeymap-cli
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o .bin/onekeymap-cli-amd64 ./cmd/onekeymap-cli
	lipo -create -output .bin/onekeymap-cli .bin/onekeymap-cli-arm64 .bin/onekeymap-cli-amd64
	rm .bin/onekeymap-cli-arm64 .bin/onekeymap-cli-amd64
	./.bin/onekeymap-cli dev docSupportActions &> ./action-support-matrix.md
	mkdir -p completions
	./.bin/onekeymap-cli completion bash > completions/onekeymap-cli.bash
	./.bin/onekeymap-cli completion zsh > completions/_onekeymap-cli
	./.bin/onekeymap-cli completion fish > completions/onekeymap-cli.fish

format:
	@$(GO_ENV) go fmt ./...

lint:
	@$(GO_ENV) golangci-lint run --fix --config $(GO_LINT_CONFIG)

test:
	@$(GO_ENV) go test ./...
