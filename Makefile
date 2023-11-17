SHELL=/bin/bash

.PHONY: help
.DEFAULT_GOAL := build
.ONESHELL:

SUFFIX=$(GOOS)_$(GOARCH)
COMMIT_HASH=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
GIT_VERSION=$(shell git branch --show-current | tr '[:upper:]' '[:lower:]')
BUILD_TIMESTAMP=$(shell date +%s)
export VER?=0.0.0

help: ## Print this help message.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-debug:  ## Build for debuging.
	export CGO_ENABLED=1
	export GCLFAGS='-gcflags="all=-N -l"'
	${MAKE} build-base

build: ## Build.
	export CGO_ENABLED=0
	export LDFLAGS_EXT='-extldflags=-static -s -w'
	${MAKE} build-base

build-race: ## Build with -race flag.
	export CGO_ENABLED=1
	export GCLFAGS='-race'
	export LDFLAGS_EXT='-extldflags=-static -s -w'
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
	export CGO_ENABLED=0
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

kubectl-view-all-resources: ## This command outputs all Kubernetes resources using YAML format and pipes it to VS Code
	./kubectl.sh view-all-resources

kubectl-view-kubeshark-resources: ## This command outputs all Kubernetes resources in "kubeshark" namespace using YAML format and pipes it to VS Code
	./kubectl.sh view-kubeshark-resources

generate-helm-values: ## Generate the Helm values from config.yaml
	./bin/kubeshark__ config > ./helm-chart/values.yaml

generate-manifests: ## Generate the manifests from the Helm chart using default configuration
	helm template kubeshark -n default ./helm-chart > ./manifests/complete.yaml

logs-worker:
	export LOGS_POD_PREFIX=kubeshark-worker-
	export LOGS_FOLLOW=
	${MAKE} logs

logs-worker-follow:
	export LOGS_POD_PREFIX=kubeshark-worker-
	export LOGS_FOLLOW=--follow
	${MAKE} logs

logs-hub:
	export LOGS_POD_PREFIX=kubeshark-hub
	export LOGS_FOLLOW=
	${MAKE} logs

logs-hub-follow:
	export LOGS_POD_PREFIX=kubeshark-hub
	export LOGS_FOLLOW=--follow
	${MAKE} logs

logs-front:
	export LOGS_POD_PREFIX=kubeshark-front
	export LOGS_FOLLOW=
	${MAKE} logs

logs-front-follow:
	export LOGS_POD_PREFIX=kubeshark-front
	export LOGS_FOLLOW=--follow
	${MAKE} logs

logs:
	kubectl logs $$(kubectl get pods | awk '$$1 ~ /^$(LOGS_POD_PREFIX)/' | awk 'END {print $$1}') $(LOGS_FOLLOW)

ssh-node:
	kubectl ssh node $$(kubectl get nodes | awk 'END {print $$1}')

exec-worker:
	export EXEC_POD_PREFIX=kubeshark-worker-
	${MAKE} exec

exec-hub:
	export EXEC_POD_PREFIX=kubeshark-hub
	${MAKE} exec

exec-front:
	export EXEC_POD_PREFIX=kubeshark-front
	${MAKE} exec

exec:
	kubectl exec --stdin --tty $$(kubectl get pods | awk '$$1 ~ /^$(EXEC_POD_PREFIX)/' | awk 'END {print $$1}') -- /bin/sh

helm-install:
	cd helm-chart && helm install kubeshark . && cd ..

helm-install-canary:
	cd helm-chart && helm install kubeshark . --set tap.docker.tag=canary && cd ..

helm-install-dev:
	cd helm-chart && helm install kubeshark . --set tap.docker.tag=dev && cd ..

helm-install-debug:
	cd helm-chart && helm install kubeshark . --set tap.debug=true && cd ..

helm-install-debug-canary:
	cd helm-chart && helm install kubeshark . --set tap.debug=true --set tap.docker.tag=canary && cd ..

helm-install-debug-dev:
	cd helm-chart && helm install kubeshark . --set tap.debug=true --set tap.docker.tag=dev && cd ..

helm-uninstall:
	helm uninstall kubeshark

proxy:
	kubeshark proxy

port-forward-worker:
	kubectl port-forward $$(kubectl get pods | awk '$$1 ~ /^$(LOGS_POD_PREFIX)/' | awk 'END {print $$1}') $(LOGS_FOLLOW) 8897:8897

release:
	@cd ../worker && git checkout master && git pull && git tag -d v$(VERSION); git tag v$(VERSION) && git push origin --tags
	@cd ../hub && git checkout master && git pull && git tag -d v$(VERSION); git tag v$(VERSION) && git push origin --tags
	@cd ../front && git checkout master && git pull && git tag -d v$(VERSION); git tag v$(VERSION) && git push origin --tags
	@cd ../kubeshark && sed -i 's/^version:.*/version: "$(VERSION)"/' helm-chart/Chart.yaml
	@git add -A . && git commit -m ":bookmark: Bump the Helm chart version to $(VERSION)" && git push
	@git tag v$(VERSION) && git push origin --tags
	@cd helm-chart && cp -r . ../../kubeshark.github.io/charts/chart
	@cd ../../kubeshark.github.io/ && git add -A . && git commit -m ":sparkles: Update the Helm chart" && git push
	@cd ../kubeshark
