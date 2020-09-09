SHELL ?= /bin/bash

.DEFAULT_GOAL := test-unit

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	# https://github.com/krancour/go-tools
	# https://hub.docker.com/repository/docker/krancour/go-tools
	DEV_IMAGE := krancour/go-tools:v0.3.0

	DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e GOCACHE=/workspaces/brigade-sdg-for-go/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade-sdk-for-go \
		-w /workspaces/brigade-sdk-for-go \
		$(DEV_IMAGE)
endif

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: lint
lint:
	$(DOCKER_CMD) golangci-lint run --config ./golangci.yaml ./...

.PHONY: test-unit
test-unit:
	$(DOCKER_CMD) go test -v \
		./...
