SHELL=/bin/bash

.PHONY: help
.DEFAULT_GOAL := build
.ONESHELL:

SUFFIX=$(GOOS)_$(GOARCH)
COMMIT_HASH=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
GIT_VERSION=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
BUILD_TIMESTAMP=$(shell date +%s)
export VER?=0.0

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-debug:  ## Build for debuging.
	export GCLFAGS='-gcflags="all=-N -l"'
	${MAKE} build-base

build: ## Build.
	export LDFLAGS_EXT='-s -w'
	${MAKE} build-base

build-base: ## Build binary (select the platform via GOOS / GOARCH env variables).
	go build ${GCLFAGS} -ldflags="${LDFLAGS_EXT} \
					-X 'github.com/kubeshark/kubeshark/misc.GitCommitHash=$(COMMIT_HASH)' \
					-X 'github.com/kubeshark/kubeshark/misc.Branch=$(GIT_BRANCH)' \
					-X 'github.com/kubeshark/kubeshark/misc.BuildTimestamp=$(BUILD_TIMESTAMP)' \
					-X 'github.com/kubeshark/kubeshark/misc.Platform=$(SUFFIX)' \
					-X 'github.com/kubeshark/kubeshark/misc.Ver=$(VER)'" \
					-o bin/kubeshark_$(SUFFIX) kubeshark.go && \
	cd bin && shasum -a 256 kubeshark_${SUFFIX} > kubeshark_${SUFFIX}.sha256

build-all: ## Build for all supported platforms.
	echo "Compiling for every OS and Platform" && \
	mkdir -p bin && sed s/_VER_/$(VER)/g RELEASE.md.TEMPLATE >  bin/README.md && \
	$(MAKE) build GOOS=linux GOARCH=amd64 && \
	$(MAKE) build GOOS=linux GOARCH=arm64 && \
	$(MAKE) build GOOS=darwin GOARCH=amd64 && \
	$(MAKE) build GOOS=darwin GOARCH=arm64 && \
	$(MAKE) build GOOS=windows GOARCH=amd64 && \
	mv ./bin/kubeshark_windows_amd64 ./bin/kubeshark.exe && \
	echo "---------" && \
	find ./bin -ls

clean: ## Clean all build artifacts.
	go clean
	rm -rf ./bin/*

test: ## Run cli tests.
	@go test ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

lint: ## Lint the source code.
	golangci-lint run
