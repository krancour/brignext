SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	# https://github.com/krancour/go-tools
	# https://hub.docker.com/repository/docker/krancour/go-tools
	GO_DEV_IMAGE := krancour/go-tools:v0.4.0
	JS_DEV_IMAGE := node:12.16.2-alpine3.11

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e GOCACHE=/workspaces/brigade/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(GO_DEV_IMAGE)

	JS_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
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

.PHONY: resolve-js-dependencies
resolve-js-dependencies:
	$(JS_DOCKER_CMD) sh -c 'cd v2/worker && yarn install'

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: lint-go
lint-go:
	$(GO_DOCKER_CMD) sh -c 'cd v2 && golangci-lint run --config ../golangci.yaml ./... '

.PHONY: test-unit-go
test-unit-go:
	$(GO_DOCKER_CMD) sh -c 'cd v2 && go test -v -timeout=30s -race -coverprofile=coverage.txt -covermode=atomic ./...'

.PHONY: verify-vendored-js-code
verify-vendored-js-code:
	$(JS_DOCKER_CMD) sh -c "cd v2/worker && yarn check --integrity && yarn check --verify-tree"

.PHONY: test-unit-js
test-unit-js:
	$(JS_DOCKER_CMD) sh -c "cd v2/worker && yarn build && yarn test"

################################################################################
# Build / Publish                                                              #
################################################################################

.PHONY: build
build: build-images xbuild-cli

.PHONY: build-images
build-images: build-apiserver build-observer build-scheduler build-worker build-logger-linux

.PHONY: build-apiserver
build-apiserver:
	docker build \
		-f v2/apiserver/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brigade-apiserver:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-apiserver:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-apiserver:$(MUTABLE_DOCKER_TAG)

.PHONY: build-observer
build-observer:
	docker build \
		-f v2/observer/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brigade-observer:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-observer:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-observer:$(MUTABLE_DOCKER_TAG)

.PHONY: build-scheduler
build-scheduler:
	docker build \
		-f v2/scheduler/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)brigade-scheduler:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-scheduler:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-scheduler:$(MUTABLE_DOCKER_TAG)

.PHONY: build-worker
build-worker:
	docker build \
		-t $(DOCKER_IMAGE_PREFIX)brigade-worker:$(IMMUTABLE_DOCKER_TAG) \
		v2/worker/
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-worker:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-worker:$(MUTABLE_DOCKER_TAG)

.PHONY: build-logger-linux
build-logger-linux:
	docker build \
		-f v2/logger/Dockerfile.linux \
		-t $(DOCKER_IMAGE_PREFIX)brigade-logger-linux:$(IMMUTABLE_DOCKER_TAG) \
		v2/logger
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-logger-linux:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-logger-linux:$(MUTABLE_DOCKER_TAG)

.PHONY: build-logger-windows
build-logger-windows:
	docker build \
		-f v2/logger/Dockerfile.winserv-2019 \
		-t $(DOCKER_IMAGE_PREFIX)brigade-logger-windows:$(IMMUTABLE_DOCKER_TAG) \
		v2/logger
	docker tag $(DOCKER_IMAGE_PREFIX)brigade-logger-windows:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)brigade-logger-windows:$(MUTABLE_DOCKER_TAG)

.PHONTY: build-cli
build-cli:
	$(GO_DOCKER_CMD) bash -c "cd v2 && OSES=$(shell go env GOOS) ARCHS=$(shell go env GOARCH) VERSION=\"$(VERSION)\" COMMIT=\"$(GIT_VERSION)\" ../scripts/build-cli.sh"

.PHONTY: xbuild-cli
xbuild-cli:
	$(GO_DOCKER_CMD) bash -c "cd v2 && VERSION=\"$(VERSION)\" COMMIT=\"$(GIT_VERSION)\" ../scripts/build-cli.sh"

.PHONY: push-images
push-images: push-apiserver push-observer push-scheduler push-worker push-logger-linux

.PHONY: push-apiserver
push-apiserver: build-apiserver
	docker push $(DOCKER_IMAGE_PREFIX)brigade-apiserver:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-apiserver:$(MUTABLE_DOCKER_TAG)

.PHONY: push-observer
push-observer: build-observer
	docker push $(DOCKER_IMAGE_PREFIX)brigade-observer:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-observer:$(MUTABLE_DOCKER_TAG)

.PHONY: push-scheduler
push-scheduler: build-scheduler
	docker push $(DOCKER_IMAGE_PREFIX)brigade-scheduler:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-scheduler:$(MUTABLE_DOCKER_TAG)

.PHONY: push-worker
push-worker: build-worker
	docker push $(DOCKER_IMAGE_PREFIX)brigade-worker:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-worker:$(MUTABLE_DOCKER_TAG)

.PHONY: push-logger-linux
push-logger-linux: build-logger-linux
	docker push $(DOCKER_IMAGE_PREFIX)brigade-logger-linux:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-logger-linux:$(MUTABLE_DOCKER_TAG)

.PHONY: push-logger-windows
push-logger-windows: build-logger-windows
	docker push $(DOCKER_IMAGE_PREFIX)brigade-logger-windows:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)brigade-logger-windows:$(MUTABLE_DOCKER_TAG)

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
	kubectl get namespace brigade || kubectl create namespace brigade

.PHONY: hack
hack: push-images build-cli hack-namespace
	kubectl get namespace brigade || kubectl create namespace brigade
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=Always \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600 \
		--set scheduler.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-scheduler \
		--set scheduler.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set scheduler.image.pullPolicy=Always \
		--set observer.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-observer \
		--set observer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set observer.image.pullPolicy=Always \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=Always \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always

.PHONY: hack-apiserver
hack-apiserver: push-apiserver hack-namespace
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--reuse-values \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=Always \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600

.PHONY: hack-observer
hack-observer: push-observer hack-namespace
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--reuse-values \
		--set observer.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-observer \
		--set observer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set observer.image.pullPolicy=Always

.PHONY: hack-scheduler
hack-scheduler: push-scheduler hack-namespace
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--reuse-values \
		--set scheduler.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-scheduler \
		--set scheduler.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set scheduler.image.pullPolicy=Always

.PHONY: hack-worker
hack-worker: push-worker hack-namespace
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--reuse-values \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=Always

.PHONY: hack-logger-linux
hack-logger-linux: push-logger-linux hack-namespace
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--reuse-values \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)brigade-logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always
