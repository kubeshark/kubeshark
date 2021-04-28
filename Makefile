C_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help ui api cli docker

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Variables and lists
TS_SUFFIX="$(shell date '+%s')"
DOCKER_IMG="up9inc/mizu"
DOCKER_TAG="latest"

ui: ## build UI
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build  

cli: # build CLI
	@(cd cli; echo "building cli" )

api: ## build API server
	@(echo "building API server .." )
	@(cd api; go build -o build/apiserver main.go)
	@ls -l api/build

docker: ## build Docker image 
	@(echo "building docker image" )
	docker build -t ${DOCKER_IMG}:${DOCKER_TAG} api
	docker images ${DOCKER_IMG}

publish: ## build and publish Mizu docker image & CLI
	@echo "publishing Docker image .. "
	@echo "publishing CLI .. "


clean: clean-api clean-cli clean-ui clean-docker ## Clean all build artifacts

clean-ui: 
	@(cd ui; rm -rf build ; echo "ui cleanup done" )

clean-api: 
	@(cd api; rm -rf build ; echo "api cleanup done" )

clean-cli: 
	@(echo "CLI cleanup - NOT IMPLEMENTED YET " )

clean-docker: 
	@(echo "DOCKER cleanup - NOT IMPLEMENTED YET " )

