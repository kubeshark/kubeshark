FOLDER=$(GOOS).$(GOARCH)

.PHONY: help
.DEFAULT_GOAL := help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	
install:
	go install mizu.go

build:	## build mizu CLI binary (select platform via GOOS / GOARCH env variables)
	go build -o bin/$(FOLDER)/mizu mizu.go

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
