# These env vars have to be set in the CI
# GITHUB_TOKEN

.PHONY: deps test release clean ci-compile ci-release version help

PROJECT := auroradns_client
PLATFORMS := linux
ARCH := amd64

VERSION := $(shell cat VERSION)
SHA := $(shell git rev-parse --short HEAD)

all: help

help:
	@echo "make build - build binary in the current environment"
	@echo "make deps - install build dependencies"
	@echo "make vet - run vet & gofmt checks"
	@echo "make lint - run golint"
	@echo "make test - run tests"
	@echo "make clean - clean"
	@echo "make release - tag with version and trigger CI release build"
	@echo "make version - show app version"

build: build-dir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 godep go build -ldflags "-X main.Version=$(VERSION) -X main.Git=$(SHA)" -o build/$(PROJECT)-linux-amd64

deps:
	go get github.com/tools/godep
	go get github.com/golang/lint/golint

vet:
	scripts/vet

test:
	godep go test -v ./...

lint:
	@find . -type f -name \*.go | grep -v ^./vendor | xargs -n 1 golint

release:
	git tag --force -s `cat VERSION` -m `cat VERSION`
	git push --force origin master --tags

clean:
	go clean
	rm -fr ./build

version:
	@echo $(VERSION) $(SHA)

ci-compile: build-dir $(PLATFORMS)

build-dir:
	@rm -rf build && mkdir build

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$@ GOARCH=$(ARCH) godep go build -ldflags "-X main.Version=$(VERSION) -X main.Git=$(SHA) -w -s" -a -o build/$(PROJECT)-$@-$(ARCH)/$(PROJECT)

ci-release:
	@previous_tag=$$(git describe --abbrev=0 --tags $(VERSION)^); \
	comparison="$$previous_tag..HEAD"; \
	if [ -z "$$previous_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
	github-release $(CIRCLE_PROJECT_USERNAME)/$(CIRCLE_PROJECT_REPONAME) $(VERSION) master "**Changelog**<br/>$$changelog"
