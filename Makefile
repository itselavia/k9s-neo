NAME            := k9s
VERSION         ?= v0.50.18
PACKAGE         := github.com/derailed/$(NAME)
OUTPUT_BIN      ?= execs/${NAME}
GO_FLAGS        ?=
GO_TAGS	        ?= netgo
CGO_ENABLED     ?=0
GIT_REV         ?= $(shell git rev-parse --short HEAD)

IMG_NAME        := derailed/k9s
IMAGE           := ${IMG_NAME}:${VERSION}
BUILD_PLATFORMS ?= linux/amd64,linux/arm64

SOURCE_DATE_EPOCH ?= $(shell date +%s)
ifeq ($(shell uname), Darwin)
DATE            ?= $(shell TZ=UTC /bin/date -j -f "%s" ${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
else
DATE            ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
endif

default: help

test:                    ## Run all tests
	@go clean --testcache && go test ./...

bench-replay-validate:   ## Run replay-only harness validation
	@python3 -m unittest discover -s hack/bench/tests
	@python3 hack/bench/replay.py --fixtures hack/bench/fixtures/replay --label replay-validation

local-lab-install:       ## Install kubectl, minikube, colima, lima, and docker CLI into the user-scoped local lab
	@bash hack/local-lab/install-tools.sh

local-lab-start:         ## Start the disposable local colima + minikube profile
	@bash hack/local-lab/start-cluster.sh

local-lab-start-metrics-small: ## Start the metrics-small local profile with metrics-server enabled
	@bash hack/local-lab/start-metrics-small.sh

local-lab-start-nodes-small: ## Start the nodes-small local profile with two minikube nodes
	@bash hack/local-lab/start-nodes-small.sh

local-lab-seed:          ## Seed the local benchmark namespace and workloads
	@bash hack/local-lab/seed-bench.sh

local-lab-seed-metrics-small: ## Seed the benchmark namespace on the metrics-small profile
	@bash hack/local-lab/seed-metrics-small.sh

local-lab-seed-nodes-small: ## Seed the benchmark namespace on the nodes-small profile
	@bash hack/local-lab/seed-nodes-small.sh

local-lab-write-vars:    ## Write hack/bench/vars.local.json for the disposable local profile
	@bash hack/local-lab/write-vars.sh

local-lab-write-vars-metrics-small: ## Write hack/bench/vars.metrics-small.json for the metrics-small profile
	@bash hack/local-lab/write-vars-metrics-small.sh

local-lab-write-vars-nodes-small: ## Write hack/bench/vars.nodes-small.json for the nodes-small profile
	@bash hack/local-lab/write-vars-nodes-small.sh

local-lab-smoke-required: ## Run the required Step 6 scenarios once cold and once warm
	@bash hack/local-lab/smoke-required.sh

local-lab-smoke-step7a-metrics-small: ## Run the Step 7A control smoke on metrics-small
	@bash hack/local-lab/smoke-step7a-metrics-small.sh

local-lab-smoke-step7a-nodes-small: ## Run the Step 7A control smoke on nodes-small
	@bash hack/local-lab/smoke-step7a-nodes-small.sh

local-lab-smoke-step7e-node-path-characterization-nodes-small: ## Run the Step 7E node-path characterization smoke on nodes-small
	@bash hack/local-lab/smoke-step7e-node-path-characterization-nodes-small.sh

local-lab-capture-baseline: ## Run the required Step 6 scenarios with baseline counts and write the baseline note
	@bash hack/local-lab/capture-baseline.sh

local-lab-capture-step7a-metrics-small: ## Capture the Step 7A control on metrics-small and write the control note
	@bash hack/local-lab/capture-step7a-metrics-small.sh

local-lab-capture-step7a-nodes-small: ## Capture the Step 7A control on nodes-small and write the control note
	@bash hack/local-lab/capture-step7a-nodes-small.sh

local-lab-capture-step7e-node-path-characterization-nodes-small: ## Capture the Step 7E node-path characterization run on nodes-small and write the characterization note
	@bash hack/local-lab/capture-step7e-node-path-characterization-nodes-small.sh

local-lab-delete:        ## Delete the disposable local colima + minikube profile
	@bash hack/local-lab/delete-cluster.sh

local-lab-delete-metrics-small: ## Delete the metrics-small local profile
	@bash hack/local-lab/delete-metrics-small.sh

local-lab-delete-nodes-small: ## Delete the nodes-small local profile
	@bash hack/local-lab/delete-nodes-small.sh

cover:                   ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

build:                   ## Builds the CLI
	@CGO_ENABLED=${CGO_ENABLED} go build ${GO_FLAGS} \
	-ldflags "-w -s -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags=${GO_TAGS} -o ${OUTPUT_BIN} main.go

kubectl-stable-version:  ## Get kubectl latest stable version
	@curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt

imgx:                    ## Build Docker Image
	@docker buildx build --platform ${BUILD_PLATFORMS} --rm -t ${IMAGE} --load .

pushx:                   ## Push Docker image to registry
	@docker buildx build --platform ${BUILD_PLATFORMS} --rm -t ${IMAGE} --push .

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
