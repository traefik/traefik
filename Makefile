SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
VERSION_GIT := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))
VERSION := $(if $(VERSION),$(VERSION),$(VERSION_GIT))

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))

REPONAME := $(shell echo $(REPO) | tr '[:upper:]' '[:lower:]')
BIN_NAME := traefik
CODENAME := cheddar

DATE := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

# Default build target
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

LINT_EXECUTABLES = misspell shellcheck

DOCKER_BUILD_PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: default
#? default: Run `make generate` and `make binary`
default: generate binary

#? dist: Create the "dist" directory
dist:
	mkdir -p dist

## Build WebUI Docker image
.PHONY: build-webui-image
#? build-webui-image: Build WebUI Docker image
build-webui-image:
	docker build -t traefik-webui -f webui/Dockerfile webui

## Clean WebUI static generated assets
.PHONY: clean-webui
#? clean-webui: Clean WebUI static generated assets
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
#? generate-webui: Generate WebUI
generate-webui: webui/static/index.html

## Generate code
.PHONY: generate
#? generate: Generate code (Dynamic and Static configuration documentation reference files)
generate:
	go generate

## Build the binary
.PHONY: binary
#? binary: Build the binary
binary: generate-webui dist
	@echo SHA: $(VERSION) $(CODENAME) $(DATE)
	CGO_ENABLED=0 GOGC=off GOOS=${GOOS} GOARCH=${GOARCH} go build ${FLAGS[*]} -ldflags "-s -w \
    -X github.com/traefik/traefik/v2/pkg/version.Version=$(VERSION) \
    -X github.com/traefik/traefik/v2/pkg/version.Codename=$(CODENAME) \
    -X github.com/traefik/traefik/v2/pkg/version.BuildDate=$(DATE)" \
    -installsuffix nocgo -o "./dist/${GOOS}/${GOARCH}/$(BIN_NAME)" ./cmd/traefik

binary-linux-arm64: export GOOS := linux
binary-linux-arm64: export GOARCH := arm64
binary-linux-arm64:
	@$(MAKE) binary

binary-linux-amd64: export GOOS := linux
binary-linux-amd64: export GOARCH := amd64
binary-linux-amd64:
	@$(MAKE) binary

binary-windows-amd64: export GOOS := windows
binary-windows-amd64: export GOARCH := amd64
binary-windows-amd64: export BIN_NAME := traefik.exe
binary-windows-amd64:
	@$(MAKE) binary

## Build the binary for the standard platforms (linux, darwin, windows)
.PHONY: crossbinary-default
#? crossbinary-default: Build the binary for the standard platforms (linux, darwin, windows)
crossbinary-default: generate generate-webui
	$(CURDIR)/script/crossbinary-default.sh

## Run the unit and integration tests
.PHONY: test
#? test: Run the unit and integration tests
test: test-unit test-integration

## Run the unit tests
.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover "-coverprofile=cover.out" -v $(TESTFLAGS) ./pkg/... ./cmd/...

## Run the integration tests
.PHONY: test-integration
#? test-integration: Run the integration tests
test-integration: binary
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test ./integration -test.timeout=20m -failfast -v $(TESTFLAGS)

## Pull all Docker images to avoid timeout during integration tests
.PHONY: pull-images
#? pull-images: Pull all Docker images to avoid timeout during integration tests
pull-images:
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml \
		| awk '{print $$2}' \
		| sort \
		| uniq \
		| xargs -P 6 -n 1 docker pull

## Lint run golangci-lint
.PHONY: lint
#? lint: Run golangci-lint
lint:
	golangci-lint run

## Validate code and docs
.PHONY: validate-files
#? validate-files: Validate code and docs
validate-files: lint
	$(foreach exec,$(LINT_EXECUTABLES),\
            $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))
	$(CURDIR)/script/validate-misspell.sh
	$(CURDIR)/script/validate-shell-script.sh

## Validate code, docs, and vendor
.PHONY: validate
#? validate: Validate code, docs, and vendor
validate: lint
	$(foreach exec,$(EXECUTABLES),\
            $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))
	$(CURDIR)/script/validate-vendor.sh
	$(CURDIR)/script/validate-misspell.sh
	$(CURDIR)/script/validate-shell-script.sh

# Target for building images for multiple architectures.
.PHONY: multi-arch-image-%
multi-arch-image-%: binary-linux-amd64 binary-linux-arm64
	docker buildx build $(DOCKER_BUILDX_ARGS) -t traefik/traefik:$* --platform=$(DOCKER_BUILD_PLATFORMS) -f Dockerfile .


## Clean up static directory and build a Docker Traefik image
.PHONY: build-image
#? build-image: Clean up static directory and build a Docker Traefik image
build-image: export DOCKER_BUILDX_ARGS := --load
build-image: export DOCKER_BUILD_PLATFORMS := linux/$(GOARCH)
build-image: clean-webui
	@$(MAKE) multi-arch-image-latest

## Build a Docker Traefik image without re-building the webui when it's already built
.PHONY: build-image-dirty
#? build-image-dirty: Build a Docker Traefik image without re-building the webui when it's already built
build-image-dirty: export DOCKER_BUILDX_ARGS := --load
build-image-dirty: export DOCKER_BUILD_PLATFORMS := linux/$(GOARCH)
build-image-dirty:
	@$(MAKE) multi-arch-image-latest

## Build documentation site
.PHONY: docs
#? docs: Build documentation site
docs:
	make -C ./docs docs

## Serve the documentation site locally
.PHONY: docs-serve
#? docs-serve: Serve the documentation site locally
docs-serve:
	make -C ./docs docs-serve

## Pull image for doc building
.PHONY: docs-pull-images
#? docs-pull-images: Pull image for doc building
docs-pull-images:
	make -C ./docs docs-pull-images

## Generate CRD clientset and CRD manifests
.PHONY: generate-crd
#? generate-crd: Generate CRD clientset and CRD manifests
generate-crd:
	@$(CURDIR)/script/code-gen-docker.sh

## Generate code from dynamic configuration https://github.com/traefik/genconf
.PHONY: generate-genconf
#? generate-genconf: Generate code from dynamic configuration github.com/traefik/genconf
generate-genconf:
	go run ./cmd/internal/gen/

## Create packages for the release
.PHONY: release-packages
#? release-packages: Format the Code
release-packages: generate-webui
	$(CURDIR)/script/release-packages.sh

## Format the Code
.PHONY: fmt
#? fmt: Format the Code
fmt:
	gofmt -s -l -w $(SRCS)

#? help: Get more info on make commands
help: Makefile
	@echo " Choose a command run in traefik:"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'

.PHONY: help