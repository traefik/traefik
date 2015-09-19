.PHONY: all

TRAEFIK_ENVS := \
	-e OS_ARCH_ARG \
	-e OS_PLATFORM_ARG \
	-e TESTFLAGS

BIND_DIR := $(if $(DOCKER_HOST),,dist)
TRAEFIK_MOUNT := $(if $(BIND_DIR),-v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/emilevauge/traefik/$(BIND_DIR)")

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
TRAEFIK_DEV_IMAGE := traefik-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
TRAEFIK_IMAGE := $(if $(REPONAME),$(REPONAME),"emilevauge/traefik")

DOCKER_RUN_TRAEFIK := docker run $(if $(CIRCLECI),,--rm) -it $(TRAEFIK_ENVS) $(TRAEFIK_MOUNT) "$(TRAEFIK_DEV_IMAGE)"

print-%: ; @echo $*=$($*)

default: binary

binary: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary

test: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit test-integration

test-unit: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate test-unit

test-integration: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary test-integration

validate: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-gofmt

validate-gofmt: build
	$(DOCKER_RUN_TRAEFIK) ./script/make.sh validate-gofmt

build: dist
	docker build -t "$(TRAEFIK_DEV_IMAGE)" -f build.Dockerfile .

image: build
	if ! [ -a dist/traefik_linux-386 ] ; \
	then \
		$(DOCKER_RUN_TRAEFIK) ./script/make.sh generate binary; \
	fi;
	docker build -t $(TRAEFIK_IMAGE) .

dist:
	mkdir dist

run-dev:
	go build
	./traefik
