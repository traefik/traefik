SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
VERSION_GIT := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))
VERSION := $(if $(VERSION),$(VERSION),$(VERSION_GIT))

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
TRAEFIK_DEV_IMAGE := traefik-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))

REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"traefik/traefik")

INTEGRATION_OPTS := $(if $(MAKE_DOCKER_HOST),-e "DOCKER_HOST=$(MAKE_DOCKER_HOST)",-v "/var/run/docker.sock:/var/run/docker.sock")
DOCKER_BUILD_ARGS := $(if $(DOCKER_VERSION), "--build-arg=DOCKER_VERSION=$(DOCKER_VERSION)",)

# only used when running in docker
TRAEFIK_ENVS := \
	-e OS_ARCH_ARG \
	-e OS_PLATFORM_ARG \
	-e TESTFLAGS \
	-e VERBOSE \
	-e VERSION \
	-e CODENAME \
	-e TESTDIRS \
	-e CI \
	-e IN_DOCKER=true		# Indicator for integration tests that we are running inside a container.

TRAEFIK_MOUNT := -v "$(CURDIR)/dist:/go/src/github.com/traefik/traefik/dist"
DOCKER_RUN_OPTS := $(TRAEFIK_ENVS) $(TRAEFIK_MOUNT) "$(TRAEFIK_DEV_IMAGE)"
DOCKER_NON_INTERACTIVE ?= false
DOCKER_RUN_TRAEFIK := docker run $(INTEGRATION_OPTS) $(if $(DOCKER_NON_INTERACTIVE), , -it) $(DOCKER_RUN_OPTS)
DOCKER_RUN_TRAEFIK_TEST := docker run --add-host=host.docker.internal:127.0.0.1 --rm --name=traefik --network traefik-test-network -v $(PWD):$(PWD) -w $(PWD) $(INTEGRATION_OPTS) $(if $(DOCKER_NON_INTERACTIVE), , -it) $(DOCKER_RUN_OPTS)
DOCKER_RUN_TRAEFIK_NOTTY := docker run $(INTEGRATION_OPTS) $(if $(DOCKER_NON_INTERACTIVE), , -i) $(DOCKER_RUN_OPTS)

IN_DOCKER ?= true

.PHONY: default
default: binary

## Create the "dist" directory
dist:
	mkdir -p dist

## Build Dev Docker image
.PHONY: build-dev-image
build-dev-image: dist
ifneq ("$(IN_DOCKER)", "")
	docker build $(DOCKER_BUILD_ARGS) -t "$(TRAEFIK_DEV_IMAGE)" --build-arg HOST_PWD="$(PWD)" -f build.Dockerfile .
endif

## Build Dev Docker image without cache
.PHONY: build-dev-image-no-cache
build-dev-image-no-cache: dist
ifneq ("$(IN_DOCKER)", "")
	docker build $(DOCKER_BUILD_ARGS) --no-cache -t "$(TRAEFIK_DEV_IMAGE)" --build-arg HOST_PWD="$(PWD)" -f build.Dockerfile .
endif

## Build WebUI Docker image
.PHONY: build-webui-image
build-webui-image:
	docker build -t traefik-webui -f webui/Dockerfile webui

## Clean WebUI static generated assets
.PHONY: clean-webui
clean-webui:
	rm -r webui/static
	mkdir -p webui/static
	printf 'For more information see `webui/readme.md`' > webui/static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md

## Generate WebUI
webui/static/index.html:
	$(MAKE) build-webui-image
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui npm run build:nc
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui chown -R $(shell id -u):$(shell id -g) ./static

.PHONY: generate-webui
generate-webui: webui/static/index.html

## Build the binary
.PHONY: binary
binary: generate-webui build-dev-image
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK)) ./script/make.sh generate binary

## Build the linux binary locally
.PHONY: binary-debug
binary-debug: generate-webui
	GOOS=linux ./script/make.sh binary

## Build the binary for the standard platforms (linux, darwin, windows)
.PHONY: crossbinary-default
crossbinary-default: generate-webui build-dev-image
	$(DOCKER_RUN_TRAEFIK_NOTTY) ./script/make.sh generate crossbinary-default

## Build the binary for the standard platforms (linux, darwin, windows) in parallel
.PHONY: crossbinary-default-parallel
crossbinary-default-parallel:
	$(MAKE) generate-webui
	$(MAKE) build-dev-image crossbinary-default

## Run the unit and integration tests
.PHONY: test
test: build-dev-image
	-docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	trap 'docker network rm traefik-test-network' EXIT; \
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_TEST)) ./script/make.sh generate test-unit binary test-integration

## Run the unit tests
.PHONY: test-unit
test-unit: build-dev-image
	-docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	trap 'docker network rm traefik-test-network' EXIT; \
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_TEST)) ./script/make.sh generate test-unit

## Run the integration tests
.PHONY: test-integration
test-integration: build-dev-image
	-docker network create traefik-test-network --driver bridge --subnet 172.31.42.0/24
	trap 'docker network rm traefik-test-network' EXIT; \
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_TEST)) ./script/make.sh generate binary test-integration

## Pull all images for integration tests
.PHONY: pull-images
pull-images:
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml \
		| awk '{print $$2}' \
		| sort \
		| uniq \
		| xargs -P 6 -n 1 docker pull

## Validate code and docs
.PHONY: validate-files
validate-files: build-dev-image
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK)) ./script/make.sh generate validate-lint validate-misspell
	bash $(CURDIR)/script/validate-shell-script.sh

## Validate code, docs, and vendor
.PHONY: validate
validate: build-dev-image
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK)) ./script/make.sh generate validate-lint validate-misspell validate-vendor
	bash $(CURDIR)/script/validate-shell-script.sh

## Clean up static directory and build a Docker Traefik image
.PHONY: build-image
build-image: clean-webui binary
	docker build -t $(TRAEFIK_IMAGE) .

## Build a Docker Traefik image without re-building the webui
.PHONY: build-image-dirty
build-image-dirty: binary
	docker build -t $(TRAEFIK_IMAGE) .

## Locally build traefik for linux, then shove it an alpine image, with basic tools.
.PHONY: build-image-debug
build-image-debug: binary-debug
	docker build -t $(TRAEFIK_IMAGE) -f debug.Dockerfile .

## Start a shell inside the build env
.PHONY: shell
shell: build-dev-image
	$(DOCKER_RUN_TRAEFIK) /bin/bash

## Build documentation site
.PHONY: docs
docs:
	make -C ./docs docs

## Serve the documentation site locally
.PHONY: docs-serve
docs-serve:
	make -C ./docs docs-serve

## Pull image for doc building
.PHONY: docs-pull-images
docs-pull-images:
	make -C ./docs docs-pull-images

## Generate CRD clientset and CRD manifests
.PHONY: generate-crd
generate-crd:
	@$(CURDIR)/script/code-gen.sh

## Generate code from dynamic configuration https://github.com/traefik/genconf
.PHONY: generate-genconf
generate-genconf:
	go run ./cmd/internal/gen/

## Create packages for the release
.PHONY: release-packages
release-packages: generate-webui build-dev-image
	rm -rf dist
	@- $(foreach os, linux darwin windows freebsd openbsd, \
        $(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) goreleaser release --skip-publish -p 2 --timeout="90m" --config $(shell go run ./internal/release $(os)); \
        $(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) go clean -cache; \
    )

	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) cat dist/**/*_checksums.txt >> dist/traefik_${VERSION}_checksums.txt
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) rm dist/**/*_checksums.txt
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) tar cfz dist/traefik-${VERSION}.src.tar.gz \
		--exclude-vcs \
		--exclude .idea \
		--exclude .travis \
		--exclude .semaphoreci \
		--exclude .github \
		--exclude dist .
	$(if $(IN_DOCKER),$(DOCKER_RUN_TRAEFIK_NOTTY)) chown -R $(shell id -u):$(shell id -g) dist/

## Format the Code
.PHONY: fmt
fmt:
	gofmt -s -l -w $(SRCS)

.PHONY: run-dev
run-dev:
	go generate
	GO111MODULE=on go build ./cmd/traefik
	./traefik
