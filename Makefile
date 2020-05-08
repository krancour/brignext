SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Go build details                                                             #
################################################################################

CLIENT_PLATFORM ?= $(shell go env GOOS)
CLIENT_ARCH ?= $(shell go env GOARCH)

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	# https://github.com/krancour/go-tools
	# https://hub.docker.com/repository/docker/krancour/go-tools
	GO_DEV_IMAGE := krancour/go-tools:v0.1.0
	JS_DEV_IMAGE := node:12.3.1-stretch

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/src \
		-w /src \
		$(GO_DEV_IMAGE)

	JS_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/src \
		-w /src \
		$(JS_DEV_IMAGE)
endif

# Allow for users to supply a different helm cli name,
# for instance, if one has helm v3 as `helm3` and helm v2 as `helm`
HELM ?= helm

################################################################################
# Binaries and Docker images we build and publish                              #
################################################################################

ifdef DOCKER_REGISTRY
	DOCKER_REGISTRY := $(DOCKER_REGISTRY)/
endif

ifdef DOCKER_ORG
	DOCKER_ORG := $(DOCKER_ORG)/
endif

DOCKER_IMAGE_PREFIX := $(DOCKER_REGISTRY)$(DOCKER_ORG)

ifdef VERSION
	MUTABLE_DOCKER_TAG := latest
else
	VERSION            := $(GIT_VERSION)
	MUTABLE_DOCKER_TAG := edge
endif

IMMUTABLE_DOCKER_TAG := $(VERSION)

################################################################################
# Utility targets                                                              #
################################################################################

.PHONY: resolve-go-dependencies
resolve-go-dependencies:
	$(GO_DOCKER_CMD) sh -c 'go mod tidy && go vendor'

.PHONY: resolve-js-dependencies
resolve-js-dependencies:
	$(JS_DOCKER_CMD) sh -c "cd worker && yarn install"

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: verify-vendored-go-code
verify-vendored-go-code:
	$(GO_DOCKER_CMD) go mod verify

.PHONY: lint-go
lint-go:
	$(GO_DOCKER_CMD) golangci-lint run --config ./golangci.yaml

.PHONY: test-unit-go
test-unit-go:
	$(GO_DOCKER_CMD) go test -v ./...

.PHONY: verify-vendored-js-code
verify-vendored-js-code:
	$(JS_DOCKER_CMD) sh -c "cd brigade-worker && yarn check --integrity && yarn check --verify-tree"

.PHONY: test-unit-js
test-unit-js:
	$(JS_DOCKER_CMD) sh -c "cd brigade-worker && yarn build && yarn test"

################################################################################
# Build / Publish                                                              #
################################################################################

.PHONY: build
build: build-images build-cli

.PHONY: build-images
build-images: build-apiserver build-controller build-worker build-logger-linux

.PHONY: build-apiserver
build-apiserver:
	docker build \
		-f apiserver/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brignext-apiserver:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brignext-apiserver:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brignext-apiserver:$(MUTABLE_DOCKER_TAG)

.PHONY: build-controller
build-controller:
	docker build \
		-f controller/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brignext-controller:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brignext-controller:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brignext-controller:$(MUTABLE_DOCKER_TAG)

.PHONY: build-worker
build-worker:
	docker build \
		-f worker/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brignext-worker:$(IMMUTABLE_DOCKER_TAG) \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brignext-worker:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brignext-worker:$(MUTABLE_DOCKER_TAG)

.PHONY: build-logger-linux
build-logger-linux:
	docker build \
		-f logger/Dockerfile.linux \
		-t $(DOCKER_IMAGE_PREFIX)brignext-logger-linux:$(IMMUTABLE_DOCKER_TAG) \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brignext-logger-linux:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brignext-logger-linux:$(MUTABLE_DOCKER_TAG)

.PHONY: build-logger-windows
build-logger-windows:
	docker build \
		-f logger/Dockerfile.winserv-2019 \
		-t $(DOCKER_IMAGE_PREFIX)brignext-logger-windows:$(IMMUTABLE_DOCKER_TAG) \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brignext-logger-windows:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brignext-logger-windows:$(MUTABLE_DOCKER_TAG)

.PHONTY: build-cli
build-cli:
	$(GO_DOCKER_CMD) bash -c "VERSION=\"$(VERSION)\" COMMIT=\"$(GIT_VERSION)\" scripts/build-cli.sh"

.PHONY: push-images
push-images: push-apiserver push-controller push-worker push-logger-linux

.PHONY: push-apiserver
push-apiserver: build-apiserver
	docker push $(DOCKER_IMAGE_PREFIX)brignext-apiserver:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brignext-apiserver:$(MUTABLE_DOCKER_TAG)

.PHONY: push-controller
push-controller: build-controller
	docker push $(DOCKER_IMAGE_PREFIX)brignext-controller:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brignext-controller:$(MUTABLE_DOCKER_TAG)

.PHONY: push-worker
push-worker: build-worker
	docker push $(DOCKER_IMAGE_PREFIX)brignext-worker:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brignext-worker:$(MUTABLE_DOCKER_TAG)

.PHONY: push-logger-linux
push-logger-linux: build-logger-linux
	docker push $(DOCKER_IMAGE_PREFIX)brignext-logger-linux:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brignext-logger-linux:$(MUTABLE_DOCKER_TAG)

.PHONY: push-logger-windows
push-logger-windows: build-logger-windows
	docker push $(DOCKER_IMAGE_PREFIX)brignext-logger-windows:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brignext-logger-windows:$(MUTABLE_DOCKER_TAG)

################################################################################
# Let's hack!!!                                                                #
################################################################################

.PHONY: hack-new-kind-cluster
hack-new-kind-cluster:
	./scripts/new-kind-cluster.sh

.PHONY: hack-install-nfs
hack-install-nfs:
	kubectl get namespace nfs || kubectl create namespace nfs
	helm upgrade nfs stable/nfs-server-provisioner \
		--install \
		--namespace nfs

.PHONY: hack-namespace
hack-namespace:
	kubectl get namespace brignext || kubectl create namespace brignext

.PHONY: hack
hack: push-images build-cli hack-namespace
	kubectl get namespace brignext || kubectl create namespace brignext
	helm upgrade brignext charts/brignext \
		--install \
		--namespace brignext \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=Always \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600 \
		--set controller.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-controller \
		--set controller.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set controller.image.pullPolicy=Always \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=Always \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always

.PHONY: hack-apiserver
hack-apiserver: push-apiserver hack-namespace
	helm upgrade brignext charts/brignext \
		--install \
		--namespace brignext \
		--reuse-values \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=Always \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600

.PHONY: hack-controller
hack-controller: push-controller hack-namespace
	helm upgrade brignext charts/brignext \
		--install \
		--namespace brignext \
		--reuse-values \
		--set controller.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-controller \
		--set controller.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set controller.image.pullPolicy=Always

.PHONY: hack-worker
hack-worker: push-worker hack-namespace
	helm upgrade brignext charts/brignext \
		--install \
		--namespace brignext \
		--reuse-values \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=Always

.PHONY: hack-logger-linux
hack-logger-linux: push-logger-linux hack-namespace
	helm upgrade brignext charts/brignext \
		--install \
		--namespace brignext \
		--reuse-values \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)brignext-logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always
