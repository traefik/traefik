.PHONY: all build binary clean cross-binary help test test-unit test-integration test-acceptance validate

LIBCOMPOSE_ENVS := \
	-e OS_PLATFORM_ARG \
	-e OS_ARCH_ARG \
	-e DOCKER_TEST_HOST \
	-e TESTDIRS \
	-e TESTFLAGS \
	-e SHOWWARNING \
	-e TESTVERBOSE

# (default to no bind mount if DOCKER_HOST is set)
BIND_DIR := $(if $(DOCKER_HOST),,bundles)
LIBCOMPOSE_MOUNT := $(if $(BIND_DIR),-v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/docker/libcompose/$(BIND_DIR)")

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
LIBCOMPOSE_IMAGE := libcompose-dev$(if $(GIT_BRANCH_CLEAN),:$(GIT_BRANCH_CLEAN))

DAEMON_VERSION := $(if $(DAEMON_VERSION),$(DAEMON_VERSION),"default")
TTY := $(shell [ -t 0 ] && echo "-t")
DOCKER_RUN_LIBCOMPOSE := docker run --rm -i $(TTY) --privileged -e DAEMON_VERSION="$(DAEMON_VERSION)" $(LIBCOMPOSE_ENVS) $(LIBCOMPOSE_MOUNT) "$(LIBCOMPOSE_IMAGE)"

default: binary

all: build ## validate all checks, build linux binary, run all tests\ncross build non-linux binaries
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh

binary: build ## build the linux binary
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh binary

cross-binary: build ## cross build the non linux binaries (windows, darwin)
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh cross-binary

test: build ## run the unit, integration and acceptance tests
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh binary test-unit test-integration test-acceptance

test-unit: build ## run the unit tests
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh test-unit

test-integration: build ## run the integration tests
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh binary test-integration

test-acceptance: build ## run the acceptance tests
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh binary test-acceptance

validate: build ## validate DCO, git conflicts marks, gofmt, golint and go vet
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh validate-dco validate-git-marks validate-gofmt validate-lint validate-vet

shell: build ## start a shell inside the build env
	$(DOCKER_RUN_LIBCOMPOSE) bash

# Build the docker image, should be prior almost any other goals
build: bundles
	docker build -t "$(LIBCOMPOSE_IMAGE)" .

bundles:
	mkdir bundles

clean: 
	$(DOCKER_RUN_LIBCOMPOSE) ./hack/make.sh clean

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
