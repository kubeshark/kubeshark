SUFFIX=$(GOOS)_$(GOARCH)
COMMIT_HASH=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
GIT_VERSION=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
BUILD_TIMESTAMP=$(shell date +%s)
export SEM_VER?=0.0.0

.PHONY: help
.DEFAULT_GOAL := help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install:
	go install mizu.go

build: ## Build mizu CLI binary (select platform via GOOS / GOARCH env variables).
	go build -ldflags="-X 'github.com/up9inc/mizu/cli/mizu.GitCommitHash=$(COMMIT_HASH)' \
					   -X 'github.com/up9inc/mizu/cli/mizu.Branch=$(GIT_BRANCH)' \
					   -X 'github.com/up9inc/mizu/cli/mizu.BuildTimestamp=$(BUILD_TIMESTAMP)' \
					   -X 'github.com/up9inc/mizu/cli/mizu.SemVer=$(SEM_VER)'" \
					   -o bin/mizu_$(SUFFIX) mizu.go 
	(cd bin && shasum -a 256 mizu_${SUFFIX} > mizu_${SUFFIX}.sha256)

build-all: ## Build for all supported platforms.
	@echo "Compiling for every OS and Platform"
	@mkdir -p bin && echo "SHA256 checksums available for compiled binaries \n\nRun \`shasum -a 256 -c mizu_OS_ARCH.sha256\` to verify\n\n" >  bin/README.md
	@$(MAKE) build GOOS=darwin GOARCH=amd64
	@$(MAKE) build GOOS=linux GOARCH=amd64
	@# $(MAKE) build GOOS=darwin GOARCH=arm64
	@# $(MAKE) GOOS=windows GOARCH=amd64
	@# $(MAKE) GOOS=linux GOARCH=386
	@# $(MAKE) GOOS=windows GOARCH=386
	@# $(MAKE) GOOS=linux GOARCH=arm64
	@# $(MAKE) GOOS=windows GOARCH=arm64
	@echo "---------"
	@find ./bin -ls

clean: ## Clean all build artifacts.
	go clean
	rm -rf ./bin/*

unit-test: ## Run cli unit tests.
	@go test ./... -short -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

integration-test: ## Run cli integration tests.
	@go test ./... -run Integration -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic
