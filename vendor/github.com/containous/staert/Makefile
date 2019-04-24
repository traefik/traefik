.PHONY: all

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')
GOFILES := $(shell go list -f '{{range $$index, $$element := .GoFiles}}{{$$.Dir}}/{{$$element}}{{"\n"}}{{end}}' ./... | grep -v '/vendor/')

default: clean checks test build

test: clean
	go test -v -race -cover ./...

dependencies:
	dep ensure -v

clean:
	rm -f cover.out

build:
	GOOS=darwin go build;
	GOOS=windows go build;
	GOOS=linux go build;

checks: check-fmt
	golangci-lint run

check-fmt: SHELL := /bin/bash
check-fmt:
	diff -u <(echo -n) <(gofmt -d $(GOFILES))

fmt:
	gofmt -s -l -w $(SRCS)
