.PHONY: all docs-verify docs docs-clean docs-build

TRAEFIK_ENVS := \
	-e OS_ARCH_ARG \
	-e OS_PLATFORM_ARG \
	-e TESTFLAGS \
	-e VERBOSE \
	-e VERSION \
	-e CODENAME \
	-e TESTDIRS \
	-e CI \
	-e CONTAINER=DOCKER		# Indicator for integration tests that we are running inside a container.

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

BIND_DIR := "dist"
TRAEFIK_MOUNT := -v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/containous/traefik/$(BIND_DIR)"

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
TRAEFIK_DEV_IMAGE := traefik-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))
REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"containous/traefik")
INTEGRATION_OPTS := $(if $(MAKE_DOCKER_HOST),-e "DOCKER_HOST=$(MAKE_DOCKER_HOST)", -e "TEST_CONTAINER=1" -v "/var/run/docker.sock:/var/run/docker.sock")
TRAEFIK_DOC_IMAGE := traefik-docs
TRAEFIK_DOC_VERIFY_IMAGE := $(TRAEFIK_DOC_IMAGE)-verify
DOCS_VERIFY_SKIP ?= false

DOCKER_BUILD_ARGS := $(if $(DOCKER_VERSION), "--build-arg=DOCKER_VERSION=$(DOCKER_VERSION)",)
DOCKER_RUN_OPTS := $(TRAEFIK_ENVS) $(TRAEFIK_MOUNT) "$(TRAEFIK_DEV_IMAGE)"
DOCKER_RUN_TRAEFIK := docker run $(INTEGRATION_OPTS) -it $(DOCKER_RUN_OPTS)
DOCKER_RUN_TRAEFIK_NOTTY := docker run $(INTEGRATION_OPTS) -i $(DOCKER_RUN_OPTS)
DOCKER_RUN_DOC_PORT := 8000
DOCKER_RUN_DOC_MOUNT := -v $(CURDIR):/mkdocs
DOCKER_RUN_DOC_OPTS := --rm $(DOCKER_RUN_DOC_MOUNT) -p $(DOCKER_RUN_DOC_PORT):8000


print-%: ; @echo $*=$($*)

default: binary

all: generate-webui build ## validate all checks, build linux binary, run all tests\ncross non-linux binaries
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh

binary: generate-webui build ## build the linux binary
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary

crossbinary: generate-webui build ## cross build the non-linux binaries
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate crossbinary

crossbinary-parallel:
	$(MAKE) generate-webui
	$(MAKE) build crossbinary-default crossbinary-others

crossbinary-default: generate-webui build
	$(DOCKER_RUN_TRAEFIK_NOTTY) ./script/make.sh generate crossbinary-default

crossbinary-default-parallel:
	$(MAKE) generate-webui
	$(MAKE) build crossbinary-default

crossbinary-others: generate-webui build
	$(DOCKER_RUN_TRAEFIK_NOTTY) ./script/make.sh generate crossbinary-others

crossbinary-others-parallel:
	$(MAKE) generate-webui
	$(MAKE) build crossbinary-others

test: build ## run the unit and integration tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit binary test-integration

test-unit: build ## run the unit tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit

test-integration: build ## run the integration tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary test-integration
	TEST_HOST=1 ./script/make.sh test-integration

validate: build  ## validate code, vendor and autogen
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-gofmt validate-govet validate-golint validate-misspell validate-vendor validate-autogen

build: dist
	docker build $(DOCKER_BUILD_ARGS) -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

build-webui:
	docker build -t traefik-webui -f webui/Dockerfile webui

build-no-cache: dist
	docker build --no-cache -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

shell: build ## start a shell inside the build env
	$(DOCKER_RUN_TRAEFIK) /bin/bash

image-dirty: binary ## build a docker traefik image
	docker build -t $(TRAEFIK_IMAGE) .

image: clear-static binary ## clean up static directory and build a docker traefik image
	docker build -t $(TRAEFIK_IMAGE) .

docs-image:
	docker build -t $(TRAEFIK_DOC_IMAGE) -f docs.Dockerfile .

docs: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOC_IMAGE) mkdocs serve

docs-build: site

docs-verify: site
ifeq ($(DOCS_VERIFY_SKIP),false)
	docker build -t $(TRAEFIK_DOC_VERIFY_IMAGE) ./script/docs-verify-docker-image
	docker run --rm -v $(CURDIR):/app $(TRAEFIK_DOC_VERIFY_IMAGE)
else
	@echo "DOCS_LINT_SKIP is true: no linting done."
endif

site: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOC_IMAGE) mkdocs build

docs-clean:
	rm -rf $(CURDIR)/site

clear-static:
	rm -rf static

dist:
	mkdir dist

run-dev:
	go generate
	go build ./cmd/traefik
	./traefik

generate-webui: build-webui
	if [ ! -d "static" ]; then \
		mkdir -p static; \
		docker run --rm -v "$$PWD/static":'/src/static' traefik-webui npm run build; \
		echo 'For more informations show `webui/readme.md`' > $$PWD/static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md; \
	fi

lint:
	script/validate-golint

fmt:
	gofmt -s -l -w $(SRCS)

pull-images:
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml | awk '{print $$2}' | sort | uniq  | xargs -P 6 -n 1 docker pull

dep-ensure:
	dep ensure -v
	./script/prune-dep.sh

dep-prune:
	./script/prune-dep.sh

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
