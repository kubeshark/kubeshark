FOLDER=$(GOOS)_$(GOARCH)
COMMIT_HASH=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
GIT_VERSION=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
BUILD_TIMESTAMP=$(shell date +%s)
BUILD_TIME_UTC=$(shell date -u)

ifeq ($(GIT_BRANCH), master)
	# SECONDS_SINCE_START_OF_DAY=$(shell echo $(( $(date -u '+%-H * 3600 + %-M * 60 + %-S') ))) # TODO: in case we do more than once a day
	CAL_VER=$(shell date +'%Y.%m.%d')
else
	CAL_VER=$(shell date +'%Y.%m.%d.%H.%M.%S')
endif


.PHONY: help
.DEFAULT_GOAL := help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install:
	go install mizu.go

build:	## build mizu CLI binary (select platform via GOOS / GOARCH env variables)
	go build -ldflags="-X 'github.com/up9inc/mizu/cli/mizu.GitCommitHash=$(COMMIT_HASH)' -X 'github.com/up9inc/mizu/cli/mizu.Branch=$(GIT_BRANCH)' -X 'github.com/up9inc/mizu/cli/mizu.CalVer=$(CAL_VER)' -X 'github.com/up9inc/mizu/cli/mizu.BuildTimestamp=$(BUILD_TIMESTAMP)' -X 'github.com/up9inc/mizu/cli/mizu.BuiltTimeUTC=$(BUILD_TIME_UTC)'" -o bin/mizu_$(FOLDER) mizu.go

build-all:  ## build for all supported platforms
	@echo "Compiling for every OS and Platform"
	@$(MAKE) build GOOS=darwin GOARCH=amd64
	@$(MAKE) build GOOS=linux GOARCH=amd64
	@# $(MAKE) GOOS=windows GOARCH=amd64
	@# $(MAKE) GOOS=linux GOARCH=386
	@# $(MAKE) GOOS=windows GOARCH=386
	@# $(MAKE) GOOS=darwin GOARCH=arm64
	@# $(MAKE) GOOS=linux GOARCH=arm64
	@# $(MAKE) GOOS=windows GOARCH=arm64
	@echo "---------"
	@find ./bin -ls

clean: ## clean all build artifacts
	go clean
	rm -rf ./bin/*
