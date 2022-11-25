C_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help cli

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

cli: ## Build CLI.
	@echo "building cli"; cd cli && $(MAKE) build

cli-debug: ## Build CLI.
	@echo "building cli"; cd cli && $(MAKE) build-debug

clean: clean-ui clean-cli ## Clean all build artifacts.

clean-cli:  ## Clean CLI.
	@(cd cli; make clean ; echo "CLI cleanup done" )


lint:  ## Run lint on all modules
	cd shared && golangci-lint run
	cd cli && golangci-lint run

test: test-cli test-shared

test-cli:  ## Run cli tests
	@echo "running cli tests"; cd cli && $(MAKE) test

test-shared:  ## Run shared tests
	@echo "running shared tests"; cd shared && $(MAKE) test
