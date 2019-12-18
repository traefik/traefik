.PHONY: all docs docs-serve

SRCS         = $(shell git ls-files '*.go' | grep -v '^vendor/')
TAG_NAME    := $(shell git tag -l --contains HEAD)
SHA         := $(shell git rev-parse HEAD)
VERSION_GIT := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))
VERSION     := $(if $(VERSION),$(VERSION),$(VERSION_GIT))
BIN_DIR     := "dist"

GIT_BRANCH          := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
TRAEFIK_DEV_VERSION := $(if $(GIT_BRANCH),$(subst /,-,$(GIT_BRANCH)))
TRAEFIK_DEV_IMAGE   := traefik-dev$(if $(GIT_BRANCH),:$(TRAEFIK_DEV_VERSION))

REPONAME      := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"containous/traefik")

INTEGRATION_OPTS  := $(if $(MAKE_DOCKER_HOST),-e "DOCKER_HOST=$(MAKE_DOCKER_HOST)", -e "TEST_CONTAINER=1" -v "/var/run/docker.sock:/var/run/docker.sock")
DOCKER_BUILD_ARGS := $(if $(DOCKER_VERSION), "--build-arg=DOCKER_VERSION=$(DOCKER_VERSION)",)
DOCKER_BUILD_ARGS += --build-arg="TRAEFIK_IMAGE_VERSION=$(TRAEFIK_DEV_VERSION)"

DOCKER_ENV_VARS := -e CROSSBUILD_ARCH -e CROSSBUILD_OS
DOCKER_ENV_VARS += -e TESTFLAGS
DOCKER_ENV_VARS += -e VERBOSE
DOCKER_ENV_VARS += -e VERSION
DOCKER_ENV_VARS += -e CODENAME
DOCKER_ENV_VARS += -e TESTDIRS
DOCKER_ENV_VARS += -e CI
DOCKER_ENV_VARS += -e CONTAINER=DOCKER		# Indicator for integration tests that we are running inside a container.

TRAEFIK_DIST_MOUNT        := -v "$(CURDIR)/$(BIN_DIR):/go/src/github.com/containous/traefik/$(BIN_DIR)"
DOCKER_RUN_OPTS           := $(TRAEFIK_ENVS) "$(TRAEFIK_DEV_IMAGE)"
DOCKER_RUN_TRAEFIK        := docker run $(INTEGRATION_OPTS) -it $(DOCKER_RUN_OPTS)
DOCKER_RUN_TRAEFIK_NOTTY  := docker run $(INTEGRATION_OPTS) -i $(DOCKER_RUN_OPTS)
DOCKER_NO_CACHE           := $(if $(DOCKER_NO_CACHE),--no-cache)

CROSSBUILD_DEFAULT ?= true

default: build

# -- docker --------------------------------------------------------------------

docker-build-frontend:
	@echo "== docker-build-frontend ==========================================="
	docker build $(DOCKER_NO_CACHE) $(DOCKER_BUILD_ARGS) -t "traefik-frontend:$(TRAEFIK_DEV_VERSION)" -f traefik-frontend.Dockerfile .

docker-build-backend: docker-build-frontend
	@echo "== docker-build-backend ============================================"
	docker build $(DOCKER_NO_CACHE) $(DOCKER_BUILD_ARGS) -t "traefik-backend:$(TRAEFIK_DEV_VERSION)" -f traefik-backend.Dockerfile .

docker-build-test: docker-build-backend
	@echo "== docker-build-test ==============================================="
	docker build $(DOCKER_NO_CACHE) $(DOCKER_BUILD_ARGS) -t "traefik-test:$(TRAEFIK_DEV_VERSION)" -f traefik-test.Dockerfile .

docker-build: docker-build-backend
	@echo "== docker-build ===================================================="
	docker build $(DOCKER_NO_CACHE) $(DOCKER_BUILD_ARGS) -t "traefik:$(TRAEFIK_DEV_VERSION)" -f traefik.Dockerfile .

docker-crossbuild:
	@echo "== docker-crossbuild ==============================================="
	docker run -it $(TRAEFIK_DIST_MOUNT) $(DOCKER_ENV_VARS) -e CROSSBUILD_DEFAULT="" "traefik-backend:$(TRAEFIK_DEV_VERSION)" make crossbuild

# -- build ---------------------------------------------------------------------

build-generate:
	@echo "== generate ========================================================="
	./script/make.sh generate

build-frontend:
	@echo "== build-frontend =================================================="
	npm run --prefix=webui build:nc
	cp -R webui/dist/pwa/* static/
	echo 'For more informations show `webui/readme.md`' > $$PWD/static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md

build-backend: build-frontend build-generate
	@echo "== binary =========================================================="
	./script/make.sh binary

build: build-backend

# -- crossbuild ---------------------------------------------------------------------

crossbuild: build-generate
	@echo "== crossbuild ======================================================"
ifdef (CROSSBUILD_DEFAULT)
	OS= ARCH= ./script/make.sh binary
endif
	for os in linux windows darwin; do \
		for arch in amd64 386; do \
			OS=$$os ARCH=$$arch ./script/make.sh binary; \
		done; \
	done;
	OS=linux ARCH=arm64 ./script/make.sh binary

# -- tests ---------------------------------------------------------------------

test: build
	@echo "== test ============================================================"
	./script/make.sh test-unit test-integration

test-unit: 
	@echo "== test-unit ======================================================="
	./script/make.sh test-unit

test-integration: docker-build-test
	@echo "== test-integration ================================================"
	CI=1 TEST_CONTAINER=1 docker run -it $(DOCKER_ENV_VARS) $(INTEGRATION_OPTS) traefik-test:$(TRAEFIK_DEV_VERSION) ./script/make.sh test-integration
	CI=1 TEST_HOST=1 ./script/make.sh test-integration

# Pull all images for integration tests
pull-images:
	@echo "== pull-images ====================================================="
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml | awk '{print $$2}' | sort | uniq | xargs -P 6 -n 1 docker pull

# -- validation ----------------------------------------------------------------

validate-files:
	@echo "== validate-files =================================================="
	./script/make.sh generate validate-lint validate-misspell
	bash $(CURDIR)/script/validate-shell-script.sh

# Validate code, docs, and vendor
validate:
	@echo "== validate ========================================================"
	./script/make.sh generate validate-lint validate-misspell validate-vendor
	bash $(CURDIR)/script/validate-shell-script.sh

# -- docs ----------------------------------------------------------------------

docs:
	make -C ./docs docs

docs-serve:
	make -C ./docs docs-serve

# -- misc ----------------------------------------------------------------------

dist:
	mkdir dist

shell:
	@echo "== shell ==========================================================="
	docker run -it $(TRAEFIK_DIST_MOUNT) $(DOCKER_ENV_VARS) "traefik-backend:$(TRAEFIK_DEV_VERSION)" /bin/bash

## Generate CRD clientset
generate-crd:
	./script/update-generated-crd-code.sh

## Create packages for the release
release-packages: build
	rm -rf dist
	$(DOCKER_RUN_TRAEFIK_NOTTY) goreleaser release --skip-publish --timeout="60m"
	$(DOCKER_RUN_TRAEFIK_NOTTY) tar cfz dist/traefik-${VERSION}.src.tar.gz \
		--exclude-vcs \
		--exclude .idea \
		--exclude .travis \
		--exclude .semaphoreci \
		--exclude .github \
		--exclude dist .
	$(DOCKER_RUN_TRAEFIK_NOTTY) chown -R $(shell id -u):$(shell id -g) dist/

## Format the Code
fmt:
	gofmt -s -l -w $(SRCS)

run-dev:
	go generate
	GO111MODULE=on go build ./cmd/traefik
	./traefik
