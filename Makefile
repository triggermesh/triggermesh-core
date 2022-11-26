# Copyright 2022 TriggerMesh Inc.
# SPDX-License-Identifier: Apache-2.0

KREPO      = triggermesh-core
KREPO_DESC = TriggerMesh Core Components (broker, triggers)

BASE_DIR          ?= $(CURDIR)
OUTPUT_DIR        ?= $(BASE_DIR)/_output

# Dynamically generate the list of commands based on the directory name cited in the cmd directory
COMMANDS          := $(notdir $(wildcard cmd/*))

# Commands and images that require custom build proccess
CUSTOM_BUILD_BINARIES :=
CUSTOM_BUILD_IMAGES   :=

BIN_OUTPUT_DIR    ?= $(OUTPUT_DIR)
DOCS_OUTPUT_DIR   ?= $(OUTPUT_DIR)
TEST_OUTPUT_DIR   ?= $(OUTPUT_DIR)
COVER_OUTPUT_DIR  ?= $(OUTPUT_DIR)
DIST_DIR          ?= $(OUTPUT_DIR)

# Rely on ko for building/publishing images and generating/deploying manifests
KO                ?= ko
KOFLAGS           ?=
IMAGE_TAG         ?= $(shell git rev-parse HEAD)

# Go build variables
GO                ?= go
GOFMT             ?= gofmt
GOLINT            ?= golangci-lint run --timeout 5m
GOTOOL            ?= go tool
GOTEST            ?= gotestsum --junitfile $(TEST_OUTPUT_DIR)/$(KREPO)-unit-tests.xml --format pkgname-and-test-fails --

GOMODULE           = github.com/triggermesh/triggermesh-core

GOPKGS             = ./cmd/... ./pkg/apis/... ./pkg/reconciler/...
GOPKGS_SKIP_TESTS  =

# List of packages that expect the environment to have installed
# the dependencies for running tests:
#
# ...
#
GOPKGS_TESTS_WITH_DEPENDENCIES  =

# This environment variable should be set when dependencies have been installed
# at the running instance.
WITH_DEPENDENCIES          ?=

LDFLAGS            = -w -s
LDFLAGS_STATIC     = $(LDFLAGS) -extldflags=-static

HAS_GOTESTSUM     := $(shell command -v gotestsum;)
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)

SED               := sed -i

OS                := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
ifeq ($(OS),$(filter $(OS),Darwin FreeBSD NetBSD))
    SED           += ''
endif

.PHONY: help all build release vm-images test lint fmt fmt-test images clean install-gotestsum install-golangci-lint deploy undeploy

.DEFAULT_GOAL := build

all: codegen build test lint

# Verify lint and tests
install-gotestsum:
ifndef HAS_GOTESTSUM
	curl -SL https://github.com/gotestyourself/gotestsum/releases/download/v1.8.0/gotestsum_1.8.0_linux_amd64.tar.gz | tar -C $(shell go env GOPATH)/bin -zxf -
endif

install-golangci-lint:
ifndef HAS_GOLANGCI_LINT
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.45.2
endif

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "; printf "\n$(KREPO_DESC)\n\nUsage:\n  make \033[36m<cmd>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: $(COMMANDS)  ## Build all artifacts

$(filter-out $(CUSTOM_BUILD_BINARIES), $(COMMANDS)): ## Build artifact
	$(GO) build -ldflags "$(LDFLAGS_STATIC)" -o $(BIN_OUTPUT_DIR)/$@ ./cmd/$@

deploy: ## Deploy TriggerMesh stack to default Kubernetes cluster
	$(KO) resolve -f $(BASE_DIR)/config > $(BASE_DIR)/triggermesh-core-$(IMAGE_TAG).yaml
	@for component in $(CUSTOM_BUILD_IMAGES); do \
		$(MAKE) -C ./cmd/$$component build CONTEXT=$(BASE_DIR) IMAGE_TAG=$(KO_DOCKER_REPO)/$$component-$(IMAGE_TAG) && \
		$(MAKE) -C ./cmd/$$component push IMAGE_TAG=$(KO_DOCKER_REPO)/$$component-$(IMAGE_TAG) && \
		$(SED) 's|value: .*'$$component'.*|value: $(KO_DOCKER_REPO)/'$$component'-$(IMAGE_TAG)|g' $(BASE_DIR)/triggermesh-core-$(IMAGE_TAG).yaml || exit 1; \
	done

	$(KO) apply -f $(BASE_DIR)/triggermesh-core-$(IMAGE_TAG).yaml
	@rm $(BASE_DIR)/triggermesh-core-$(IMAGE_TAG).yaml

undeploy: ## Remove TriggerMesh stack from default Kubernetes cluster
	$(KO) delete -f $(BASE_DIR)/config

vm-images:
	@$(MAKE) -C packer/

release: ## Publish container images and generate release manifests
	@mkdir -p $(DIST_DIR)
	$(KO) resolve -f config/ -l 'triggermesh.io/crd-install' > $(DIST_DIR)/triggermesh-core-crds.yaml
	@cp config/namespace/100-namespace.yaml $(DIST_DIR)/triggermesh-core.yaml
	$(KO) resolve $(KOFLAGS) -B -t latest -f config/ -l '!triggermesh.io/crd-install' > /dev/null
	$(KO) resolve $(KOFLAGS) -B -t $(IMAGE_TAG) --tag-only -f config/ -l '!triggermesh.io/crd-install' >> $(DIST_DIR)/triggermesh-core.yaml

	@for component in $(CUSTOM_BUILD_IMAGES); do \
		$(MAKE) -C ./cmd/$$component build CONTEXT=$(BASE_DIR) IMAGE_TAG=$(KO_DOCKER_REPO)/$$component:$(IMAGE_TAG) && \
		$(MAKE) -C ./cmd/$$component tag IMAGE_TAG=$(KO_DOCKER_REPO)/$$component:$(IMAGE_TAG) TAGS=$(KO_DOCKER_REPO)/$$component:latest && \
		$(MAKE) -C ./cmd/$$component push IMAGE_TAG=$(KO_DOCKER_REPO)/$$component && \
		$(SED) 's/'$$component':.*/'$$component':$(IMAGE_TAG)/g' $(DIST_DIR)/triggermesh-core.yaml || exit 1; \
	done

gen-apidocs: ## Generate API docs
	GOPATH="" OUTPUT_DIR=$(DOCS_OUTPUT_DIR) ./hack/gen-api-reference-docs.sh


GOPKGS_LIST ?= $(filter-out $(GOPKGS_SKIP_TESTS), $(shell go list $(GOPKGS)))
ifdef WITH_DEPENDENCIES
	GOPKGS_LIST += $(GOPKGS_TESTS_WITH_DEPENDENCIES)
endif
test: install-gotestsum ## Run unit tests
	@mkdir -p $(TEST_OUTPUT_DIR)

	$(GOTEST) -p=1 -race -cover -coverprofile=$(TEST_OUTPUT_DIR)/$(KREPO)-c.out $(GOPKGS_LIST)

	@for component in $(CUSTOM_BUILD_IMAGES); do \
		$(MAKE) -C ./cmd/$$component test || exit 1; \
	done

cover: test ## Generate code coverage
	@mkdir -p $(COVER_OUTPUT_DIR)
	$(GOTOOL) cover -html=$(TEST_OUTPUT_DIR)/$(KREPO)-c.out -o $(COVER_OUTPUT_DIR)/$(KREPO)-coverage.html

lint: install-golangci-lint ## Lint source files
	$(GOLINT) $(GOPKGS) ./test/...

fmt: ## Format source files
	$(GOFMT) -s -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS))

fmt-test: ## Check source formatting
	@test -z $(shell $(GOFMT) -l $(shell $(GO) list -f '{{$$d := .Dir}}{{range .GoFiles}}{{$$d}}/{{.}} {{end}} {{$$d := .Dir}}{{range .TestGoFiles}}{{$$d}}/{{.}} {{end}}' $(GOPKGS)))

KO_PUBLISHABLE 	   = $(filter-out $(CUSTOM_BUILD_IMAGES), $(COMMANDS))
KO_IMAGES 		   = $(foreach cmd,$(KO_PUBLISHABLE),$(cmd).image)
CUSTOM_IMAGES 	   = $(foreach cmd,$(CUSTOM_BUILD_IMAGES),$(cmd).image)

images: $(KO_IMAGES) $(CUSTOM_IMAGES) ## Build container images
$(KO_IMAGES): %.image:
	$(KO) publish --push=false -B --tag-only -t $(IMAGE_TAG) ./cmd/$*

$(CUSTOM_IMAGES): %.image:
	@$(MAKE) -C ./cmd/$* image CONTEXT=$(BASE_DIR) IMAGE_TAG=$(KO_DOCKER_REPO)/$*:$(IMAGE_TAG)

clean: ## Clean build artifacts
	@for bin in $(COMMANDS) ; do \
		$(RM) -v $(BIN_OUTPUT_DIR)/$$bin; \
	done
	@$(RM) -v $(DIST_DIR)/triggermesh-core-crds.yaml $(DIST_DIR)/triggermesh-core.yaml
	@$(RM) -v $(TEST_OUTPUT_DIR)/$(KREPO)-c.out $(TEST_OUTPUT_DIR)/$(KREPO)-unit-tests.xml
	@$(RM) -v $(COVER_OUTPUT_DIR)/$(KREPO)-coverage.html

# Code generation
include $(BASE_DIR)/hack/inc.Codegen.mk
