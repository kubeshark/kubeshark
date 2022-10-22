C_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help ui agent agent-debug cli tap docker bpf clean-bpf

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Variables and lists
TS_SUFFIX="$(shell date '+%s')"
GIT_BRANCH="$(shell git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]' | tr '/' '_')"
BUCKET_PATH=static.up9.io/kubeshark/$(GIT_BRANCH)
export VER?=0.0
ARCH=$(shell uname -m)
ifeq ($(ARCH),$(filter $(ARCH),aarch64 arm64))
	BPF_O_ARCH_LABEL=arm64
else
	BPF_O_ARCH_LABEL=x86
endif
BPF_O_FILES = tap/tlstapper/tlstapper46_bpfel_$(BPF_O_ARCH_LABEL).o tap/tlstapper/tlstapper_bpfel_$(BPF_O_ARCH_LABEL).o

ui: ## Build UI.
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build

cli: ## Build CLI.
	@echo "building cli"; cd cli && $(MAKE) build

cli-debug: ## Build CLI.
	@echo "building cli"; cd cli && $(MAKE) build-debug

agent: bpf ## Build agent.
	@(echo "building kubeshark agent .." )
	@(cd agent; go build -o build/kubesharkagent main.go)
	@ls -l agent/build

bpf: $(BPF_O_FILES)

$(BPF_O_FILES): $(wildcard tap/tlstapper/bpf/**/*.[ch])
	@(echo "building tlstapper bpf")
	@(./tap/tlstapper/bpf-builder/build.sh)

agent-debug: ## Build agent for debug.
	@(echo "building kubeshark agent for debug.." )
	@(cd agent; go build -gcflags="all=-N -l" -o build/kubesharkagent main.go)
	@ls -l agent/build

docker: ## Build and publish agent docker image.
	$(MAKE) push-docker

agent-docker: ## Build agent docker image.
	@echo "Building agent docker image"
	@docker build -t kubeshark/kubeshark:devlatest .

push: push-docker push-cli ## Build and publish agent docker image & CLI.

push-docker: ## Build and publish agent docker image.
	@echo "publishing Docker image .. "
	devops/build-push-featurebranch.sh

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

clean-docker:  ## Run clean docker
	@(echo "DOCKER cleanup - NOT IMPLEMENTED YET " )

clean-bpf:
	@(rm $(BPF_O_FILES) ; echo "bpf cleanup done" )

test-lint:  ## Run lint on all modules
	cd agent && golangci-lint run
	cd shared && golangci-lint run
	cd tap && golangci-lint run
	cd cli && golangci-lint run
	cd acceptanceTests && golangci-lint run
	cd tap/api && golangci-lint run
	cd tap/dbgctl && golangci-lint run
	cd tap/extensions/ && for D in */; do cd $$D && golangci-lint run && cd ..; done

test-cli:  ## Run cli tests
	@echo "running cli tests"; cd cli && $(MAKE) test

test-agent:  ## Run agent tests
	@echo "running agent tests"; cd agent && $(MAKE) test

test-shared:  ## Run shared tests
	@echo "running shared tests"; cd shared && $(MAKE) test

test-extensions:  ## Run extensions tests
	@echo "running http tests"; cd tap/extensions/http && $(MAKE) test
	@echo "running redis tests"; cd tap/extensions/redis && $(MAKE) test
	@echo "running kafka tests"; cd tap/extensions/kafka && $(MAKE) test
	@echo "running amqp tests"; cd tap/extensions/amqp && $(MAKE) test

acceptance-test:  ## Run acceptance tests
	@echo "running acceptance tests"; cd acceptanceTests && $(MAKE) test
