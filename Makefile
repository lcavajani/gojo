#RUNTIME           := $(shell which docker 2>/dev/null || which podman)
RUNTIME           := buildah
REPO_ROOT         := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION           := $(shell cat $(REPO_ROOT)/VERSION)
EFFECTIVE_VERSION := $(VERSION)-$(shell git rev-parse HEAD)

REGISTRY          ?= r.spiarh.fr
FROM_IMAGE_BUILDER := docker.io/library/golang:1.16
#FROM_IMAGE        := $(REGISTRY)/alpine:3.13.2
FROM_IMAGE        := gcr.io/kaniko-project/executor:debug
# IMAGE             := $(REGISTRY)/library/gojo:$(EFFECTIVE_VERSION)
IMAGE             := $(REGISTRY)/library/gojo-kaniko:$(EFFECTIVE_VERSION)

.PHONY: revendor
revendor:
	@$(REPO_ROOT)/hack/revendor.sh

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/...

.PHONY: verify
verify: check


.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/cmd $(REPO_ROOT)/pkg

.PHONY: build
build:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) GO_ACTION="build" ./hack/build-or-install.sh

.PHONY: install
install:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) GO_ACTION="install" ./hack/build-or-install.sh

.PHONY: build-image
build-image:
	@$(RUNTIME) bud --no-cache \
		--build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
		--build-arg FROM_IMAGE_BUILDER=$(FROM_IMAGE_BUILDER) \
		--build-arg FROM_IMAGE=$(FROM_IMAGE) \
		-t $(IMAGE) .

.PHONY: push-image
push-image:
	@$(RUNTIME) push $(IMAGE)

.PHONY: build-push-image
build-push-image: build-image push-image
