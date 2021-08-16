C_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help ui agent cli tap docker

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Variables and lists
TS_SUFFIX="$(shell date '+%s')"
GIT_BRANCH="$(shell git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]' | tr '/' '_')"
BUCKET_PATH=static.up9.io/mizu/$(GIT_BRANCH)
export SEM_VER?=0.0.0

ui: ## Build UI.
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build  

cli: ## Build CLI.
	@echo "building cli"; cd cli && $(MAKE) build

agent: ## Build agent.
	@(echo "building mizu agent .." )
	@(cd agent; go build -o build/mizuagent main.go)
	@ls -l agent/build

docker: ## Build and publish agent docker image.
	$(MAKE) push-docker

push: push-docker push-cli ## Build and publish agent docker image & CLI.

push-docker: ## Build and publish agent docker image.
	@echo "publishing Docker image .. "
	PUSH='true' ./build-push-featurebranch.sh

build-docker: ## Build agent docker image.
	@echo "publishing Docker image .. "
	GIT_BRANCH='integration_tests' ./build-push-featurebranch.sh

push-cli: ## Build and publish CLI.
	@echo "publishing CLI .. "
	@cd cli; $(MAKE) build-all
	@echo "publishing file ${OUTPUT_FILE} .."
	#gsutil mv gs://${BUCKET_PATH}/${OUTPUT_FILE} gs://${BUCKET_PATH}/${OUTPUT_FILE}.${SUFFIX}
	gsutil cp -r ./cli/bin/* gs://${BUCKET_PATH}/
	gsutil setmeta -r -h "Cache-Control:public, max-age=30" gs://${BUCKET_PATH}/\*

clean: clean-ui clean-agent clean-cli clean-docker ## Clean all build artifacts.

clean-ui: ## Clean UI.
	@(rm -rf ui/build ; echo "UI cleanup done" )

clean-agent: ## Clean agent.
	@(rm -rf agent/build ; echo "agent cleanup done" )

clean-cli:  ## Clean CLI.
	@(cd cli; make clean ; echo "CLI cleanup done" )

clean-docker:
	@(echo "DOCKER cleanup - NOT IMPLEMENTED YET " )

unit-test-cli: ## Run cli unit tests.
	@echo "running cli unit tests"; cd cli && $(MAKE) unit-test

unit-test-agent: ## Run agent unit tests.
	@echo "running agent unit tests"; cd agent && $(MAKE) unit-test

integration-test-cli: ## Run cli integration tests.
	@echo "running cli integration tests"; cd cli && $(MAKE) integration-test
