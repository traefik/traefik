.PHONY: all

TRAEFIK_ENVS := \
	-e OS_ARCH_ARG \
	-e OS_PLATFORM_ARG \
	-e TESTFLAGS \
	-e VERBOSE \
	-e VERSION \
	-e CODENAME \
	-e TESTDIRS

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/' | grep -v '^integration/vendor/')

BIND_DIR := "dist"
TRAEFIK_MOUNT := -v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/containous/traefik/$(BIND_DIR)"

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
TRAEFIK_DEV_IMAGE := traefik-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"containous/traefik")
INTEGRATION_OPTS := $(if $(MAKE_DOCKER_HOST),-e "DOCKER_HOST=$(MAKE_DOCKER_HOST)", -v "/var/run/docker.sock:/var/run/docker.sock")

DOCKER_BUILD_ARGS := $(if $(DOCKER_VERSION), "--build-arg=DOCKER_VERSION=$(DOCKER_VERSION)",)
DOCKER_RUN_TRAEFIK := docker run $(INTEGRATION_OPTS) -it $(TRAEFIK_ENVS) $(TRAEFIK_MOUNT) "$(TRAEFIK_DEV_IMAGE)"

print-%: ; @echo $*=$($*)

default: binary

all: generate-webui build ## validate all checks, build linux binary, run all tests\ncross non-linux binaries
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh

binary: generate-webui build ## build the linux binary
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary

crossbinary: generate-webui build ## cross build the non-linux binaries
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate crossbinary

test: build ## run the unit and integration tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit binary test-integration

test-unit: build ## run the unit tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit

test-integration: build ## run the integration tests
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary test-integration

validate: build  ## validate gofmt, golint and go vet
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh  validate-glide validate-gofmt validate-govet validate-golint validate-misspell validate-vendor

build: dist
	docker build $(DOCKER_BUILD_ARGS) -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

build-webui:
	docker build -t traefik-webui -f webui/Dockerfile webui

build-no-cache: dist
	docker build --no-cache -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

shell: build ## start a shell inside the build env
	$(DOCKER_RUN_TRAEFIK) /bin/bash

image: binary ## build a docker traefik image
	docker build -t $(TRAEFIK_IMAGE) .

dist:
	mkdir dist

run-dev:
	go generate
	go build
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
	for f in $(shell find ./integration/resources/compose/ -type f); do \
		docker-compose -f $$f pull; \
	done

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
