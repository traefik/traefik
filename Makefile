.PHONY: all

TRAEFIK_ENVS := \
	-e OS_ARCH_ARG \
	-e OS_PLATFORM_ARG \
	-e TESTFLAGS \
	-e VERSION

SRCS = $(shell git ls-files '*.go' | grep -v '^external/')

BIND_DIR := "dist"
TRAEFIK_MOUNT := -v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/containous/traefik/$(BIND_DIR)"

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
TRAEFIK_DEV_IMAGE := traefik-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"containous/traefik")
INTEGRATION_OPTS := $(if $(MAKE_DOCKER_HOST),-e "DOCKER_HOST=$(MAKE_DOCKER_HOST)", -v "/var/run/docker.sock:/var/run/docker.sock")

DOCKER_RUN_TRAEFIK := docker run $(INTEGRATION_OPTS) -it $(TRAEFIK_ENVS) $(TRAEFIK_MOUNT) "$(TRAEFIK_DEV_IMAGE)"

print-%: ; @echo $*=$($*)

default: binary

all: generate-webui build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh

binary: generate-webui build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary

crossbinary: generate-webui build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate crossbinary

test: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit binary test-integration

test-unit: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit

test-integration: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-integration

validate: build 
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh  validate-gofmt validate-govet validate-golint 

validate-gofmt: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-gofmt

validate-govet: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-govet

validate-golint: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-golint

build: dist
	docker build -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

build-webui:
	docker build -t traefik-webui -f webui/Dockerfile webui

build-no-cache: dist
	docker build --no-cache -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

shell: build
	$(DOCKER_RUN_TRAEFIK) /bin/bash

image: build
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
		docker run --rm -v "$$PWD/static":'/src/static' traefik-webui gulp; \
		echo 'For more informations show `webui/readme.md`' > $$PWD/static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md; \
	fi

lint:
	script/validate-golint

fmt:
	gofmt -s -l -w $(SRCS)

deploy:
	./script/deploy.sh
