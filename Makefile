SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
VERSION_GIT := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))
VERSION := $(if $(VERSION),$(VERSION),$(VERSION_GIT))

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))

.PHONY: default
default: binary

## Create the "dist" directory
dist:
	mkdir -p dist

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
binary: generate-webui
	./script/make.sh generate binary

## Build the linux binary locally
.PHONY: binary-debug
binary-debug: generate-webui
	GOOS=linux ./script/make.sh binary

## Build the binary for the standard platforms (linux, darwin, windows)
.PHONY: crossbinary-default
crossbinary-default: generate-webui
	./script/make.sh generate crossbinary-default

## Build the binary for the standard platforms (linux, darwin, windows) in parallel
.PHONY: crossbinary-default-parallel
crossbinary-default-parallel:
	$(MAKE) generate-webui
	$(MAKE) crossbinary-default

## Run the unit and integration tests
.PHONY: test
test:
	./script/make.sh generate test-unit binary test-integration

## Run the unit tests
.PHONY: test-unit
test-unit:
	./script/make.sh generate test-unit

## Run the integration tests
.PHONY: test-integration
test-integration:
	./script/make.sh generate binary test-integration

## Pull all images for integration tests
.PHONY: pull-images
pull-images:
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml \
		| awk '{print $$2}' \
		| sort \
		| uniq \
		| xargs -P 6 -n 1 docker pull

EXECUTABLES = misspell shellcheck

## Validate code and docs
.PHONY: validate-files
validate-files:
	$(foreach exec,$(EXECUTABLES),\
            $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))
	./script/make.sh generate validate-lint validate-misspell
	bash $(CURDIR)/script/validate-shell-script.sh

## Validate code, docs, and vendor
.PHONY: validate
validate:
	$(foreach exec,$(EXECUTABLES),\
            $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))
	./script/make.sh generate validate-lint validate-misspell validate-vendor
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
	@$(CURDIR)/script/code-gen-docker.sh

## Generate code from dynamic configuration https://github.com/traefik/genconf
.PHONY: generate-genconf
generate-genconf:
	go run ./cmd/internal/gen/

## Create packages for the release
.PHONY: release-packages
release-packages: generate-webui
	rm -rf dist
	@- $(foreach os, linux darwin windows freebsd openbsd, \
        goreleaser release --skip-publish -p 2 --timeout="90m" --config $(shell go run ./internal/release $(os)); \
        go clean -cache; \
    )

	cat dist/**/*_checksums.txt >> dist/traefik_${VERSION}_checksums.txt
	rm dist/**/*_checksums.txt
	tar cfz dist/traefik-${VERSION}.src.tar.gz \
		--exclude-vcs \
		--exclude .idea \
		--exclude .travis \
		--exclude .semaphoreci \
		--exclude .github \
		--exclude dist .
	chown -R $(shell id -u):$(shell id -g) dist/

## Format the Code
.PHONY: fmt
fmt:
	gofmt -s -l -w $(SRCS)

.PHONY: run-dev
run-dev:
	go generate
	GO111MODULE=on go build ./cmd/traefik
	./traefik
