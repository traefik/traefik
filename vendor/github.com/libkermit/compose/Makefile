.PHONY: all

LIBKERMIT_ENVS := \
	-e DOCKER_TEST_HOST \
	-e TESTFLAGS \
	-e CODECOV

BIND_DIR := "dist"
LIBKERMIT_MOUNT := -v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/libkermit/compose/$(BIND_DIR)"

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
LIBKERMIT_DEV_IMAGE := libkermit-compose-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))
REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')

DAEMON_VERSION := $(if $(DAEMON_VERSION),$(DAEMON_VERSION),"default")
DOCKER_RUN_LIBKERMIT := docker run --rm --privileged -it -e DAEMON_VERSION="$(DAEMON_VERSION)" $(LIBKERMIT_ENVS) $(LIBKERMIT_MOUNT) "$(LIBKERMIT_DEV_IMAGE)"

default: all

all: build ## validate all checks, run all tests
	$(DOCKER_RUN_LIBKERMIT) ./hack/make.sh

test: build ## run the unit and integration tests
	$(DOCKER_RUN_LIBKERMIT) ./hack/make.sh test-unit test-integration

test-integration: build ## run the integration tests
	$(DOCKER_RUN_LIBKERMIT) ./hack/make.sh test-integration

test-unit: build ## run the unit tests
	$(DOCKER_RUN_LIBKERMIT) ./hack/make.sh test-unit

validate: build ## validate gofmt, golint and go vet
	$(DOCKER_RUN_LIBKERMIT) ./hack/make.sh validate-gofmt validate-govet validate-golint

lint:
	./hack/make.sh validate-golint

fmt:
	./hack/make.sh validate-gofmt

build: dist
	docker build -t "$(LIBKERMIT_DEV_IMAGE)" .

shell: build ## start a shell inside the build env
	$(DOCKER_RUN_LIBKERMIT) /bin/bash

dist:
	mkdir dist

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
