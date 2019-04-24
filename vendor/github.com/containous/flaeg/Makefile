.PHONY: all


PKGS := $(shell go list ./... | grep -v '/vendor/')
GOFILES := $(shell go list -f '{{range $$index, $$element := .GoFiles}}{{$$.Dir}}/{{$$element}}{{"\n"}}{{end}}' ./... | grep -v '/vendor/')
SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')
TXT_FILES := $(shell find * -type f -not -path 'vendor/**')

default: clean checks lint test build

test: clean
	go test -v -cover $(PKGS)

dependencies:
	dep ensure -v

clean:
	rm -f cover.out

build:
	go build

lint:
	golint -set_exit_status $(PKGS)

checks: check-fmt
	staticcheck $(PKGS)
	gosimple $(PKGS)

check-fmt: SHELL := /bin/bash
check-fmt:
	diff -u <(echo -n) <(gofmt -d $(GOFILES))

misspell:
	misspell -source=text -error $(TXT_FILES)

fmt:
	gofmt -s -l -w $(SRCS)
