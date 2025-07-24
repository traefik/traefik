SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

TAG_NAME := $(shell git describe --abbrev=0 --tags --exact-match)
SHA := $(shell git rev-parse HEAD)
VERSION_GIT := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))
VERSION := $(if $(VERSION),$(VERSION),$(VERSION_GIT))

BIN_NAME := traefik
CODENAME ?= cheddar

DATE := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

# Default build target
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOGC ?=

LINT_EXECUTABLES = misspell shellcheck

DOCKER_BUILD_PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: default
#? default: Run `make generate` and `make binary`
default: generate binary

#? dist: Create the "dist" directory
dist:
	mkdir -p dist

.PHONY: build-webui-image
#? build-webui-image: Build WebUI Docker image
build-webui-image:
	docker build -t traefik-webui -f webui/buildx.Dockerfile webui

.PHONY: clean-webui
#? clean-webui: Clean WebUI static generated assets
clean-webui:
	rm -rf webui/static

webui/static/index.html:
	$(MAKE) build-webui-image
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui yarn build:prod
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui chown -R $(shell id -u):$(shell id -g) ./static

.PHONY: generate-webui
#? generate-webui: Generate WebUI
generate-webui: webui/static/index.html

.PHONY: generate
#? generate: Generate code (Dynamic and Static configuration documentation reference files)
generate:
	go generate

.PHONY: binary
#? binary: Build the binary
binary: generate-webui dist
	@echo SHA: $(VERSION) $(CODENAME) $(DATE)
	CGO_ENABLED=0 GOGC=${GOGC} GOOS=${GOOS} GOARCH=${GOARCH} go build ${FLAGS[*]} -ldflags "-s -w \
    -X github.com/traefik/traefik/v3/pkg/version.Version=$(VERSION) \
    -X github.com/traefik/traefik/v3/pkg/version.Codename=$(CODENAME) \
    -X github.com/traefik/traefik/v3/pkg/version.BuildDate=$(DATE)" \
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

.PHONY: crossbinary-default
#? crossbinary-default: Build the binary for the standard platforms (linux, darwin, windows)
crossbinary-default: generate generate-webui
	$(CURDIR)/script/crossbinary-default.sh

.PHONY: test
#? test: Run the unit and integration tests
test: test-ui-unit test-unit test-integration

.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover "-coverprofile=cover.out" -v $(TESTFLAGS) ./pkg/... ./cmd/...

.PHONY: test-integration
#? test-integration: Run the integration tests
test-integration:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test ./integration -test.timeout=20m -failfast -v $(TESTFLAGS)

.PHONY: test-gateway-api-conformance
#? test-gateway-api-conformance: Run the conformance tests
test-gateway-api-conformance: build-image-dirty
	# In case of a new Minor/Major version, the k8sConformanceTraefikVersion needs to be updated.
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test ./integration -v -test.run K8sConformanceSuite -k8sConformance -k8sConformanceTraefikVersion="v3.5" $(TESTFLAGS)

.PHONY: test-ui-unit
#? test-ui-unit: Run the unit tests for the webui
test-ui-unit:
	$(MAKE) build-webui-image
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui yarn --cwd webui install
	docker run --rm -v "$(PWD)/webui/static":'/src/webui/static' traefik-webui yarn --cwd webui test:unit:ci

.PHONY: pull-images
#? pull-images: Pull all Docker images to avoid timeout during integration tests
pull-images:
	grep --no-filename -E '^\s+image:' ./integration/resources/compose/*.yml \
		| awk '{print $$2}' \
		| sort \
		| uniq \
		| xargs -P 6 -n 1 docker pull

.PHONY: lint
#? lint: Run golangci-lint
lint:
	golangci-lint run

.PHONY: validate-files
#? validate-files: Validate code and docs
validate-files:
	$(foreach exec,$(LINT_EXECUTABLES),\
            $(if $(shell which $(exec)),,$(error "No $(exec) in PATH")))
	$(CURDIR)/script/validate-vendor.sh
	$(CURDIR)/script/validate-misspell.sh
	$(CURDIR)/script/validate-shell-script.sh

.PHONY: validate
#? validate: Validate code, docs, and vendor
validate: lint validate-files

# Target for building images for multiple architectures.
.PHONY: multi-arch-image-%
multi-arch-image-%: binary-linux-amd64 binary-linux-arm64
	docker buildx build $(DOCKER_BUILDX_ARGS) -t traefik/traefik:$* --platform=$(DOCKER_BUILD_PLATFORMS) -f Dockerfile .


.PHONY: build-image
#? build-image: Clean up static directory and build a Docker Traefik image
build-image: export DOCKER_BUILDX_ARGS := --load
build-image: export DOCKER_BUILD_PLATFORMS := linux/$(GOARCH)
build-image: clean-webui
	@$(MAKE) multi-arch-image-latest

.PHONY: build-image-dirty
#? build-image-dirty: Build a Docker Traefik image without re-building the webui when it's already built
build-image-dirty: export DOCKER_BUILDX_ARGS := --load
build-image-dirty: export DOCKER_BUILD_PLATFORMS := linux/$(GOARCH)
build-image-dirty:
	@$(MAKE) multi-arch-image-latest

.PHONY: docs
#? docs: Build documentation site
docs:
	make -C ./docs docs

.PHONY: docs-serve
#? docs-serve: Serve the documentation site locally
docs-serve:
	make -C ./docs docs-serve

.PHONY: docs-pull-images
#? docs-pull-images: Pull image for doc building
docs-pull-images:
	make -C ./docs docs-pull-images

.PHONY: generate-crd
#? generate-crd: Generate CRD clientset and CRD manifests
generate-crd:
	@$(CURDIR)/script/code-gen.sh

.PHONY: generate-genconf
#? generate-genconf: Generate code from dynamic configuration github.com/traefik/genconf
generate-genconf:
	go run ./cmd/internal/gen/

.PHONY: release-packages
#? release-packages: Create packages for the release
release-packages: generate-webui
	$(CURDIR)/script/release-packages.sh

.PHONY: fmt
#? fmt: Format the Code
fmt:
	gofmt -s -l -w $(SRCS)

.PHONY: help
#? help: Get more info on make commands
help: Makefile
	@echo " Choose a command run in traefik:"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
