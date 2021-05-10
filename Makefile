C_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help ui api cli tap docker

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Variables and lists
TS_SUFFIX="$(shell date '+%s')"
GIT_BRANCH="$(shell git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]' | tr '/' '_')"
BUCKET_PATH=static.up9.io/mizu/$(GIT_BRANCH)

ui: ## build UI
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build  

cli: # build CLI
	@echo "building cli"; cd cli && $(MAKE) build

api: ## build API server
	@(echo "building API server .." )
	@(cd api; go build -o build/apiserver main.go)
	@ls -l api/build

#tap: ## build tap binary
#	@(cd tap; go build -o build/tap ./src)
#	@ls -l tap/build

docker: ## build Docker image 
	@(echo "building docker image" )
	docker build -t up9inc/mizu:latest .
	#./build-push-featurebranch.sh

push: push-docker push-cli ## build and publish Mizu docker image & CLI

push-docker: 
	@echo "publishing Docker image .. "
	./build-push-featurebranch.sh

push-cli:
	@echo "publishing CLI .. "
	@cd cli; $(MAKE) build-all
	@echo "publishing file ${OUTPUT_FILE} .."
	#gsutil mv gs://${BUCKET_PATH}/${OUTPUT_FILE} gs://${BUCKET_PATH}/${OUTPUT_FILE}.${SUFFIX}
	gsutil cp -r ./cli/bin/* gs://${BUCKET_PATH}/
	gsutil setmeta -r -h "Cache-Control:public, max-age=30" gs://${BUCKET_PATH}/\*


clean: clean-ui clean-api clean-cli clean-docker ## Clean all build artifacts

clean-ui: 
	@(rm -rf ui/build ; echo "UI cleanup done" )

clean-api: 
	@(rm -rf api/build ; echo "api cleanup done" )

clean-cli: 
	@(cd cli; make clean ; echo "CLI cleanup done" )

clean-docker: 
	@(echo "DOCKER cleanup - NOT IMPLEMENTED YET " )

