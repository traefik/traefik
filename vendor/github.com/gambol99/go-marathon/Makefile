#
#   Author: Rohith (gambol99@gmail.com)
#   Date: 2015-02-10 15:35:14 +0000 (Tue, 10 Feb 2015)
#
#  vim:ts=2:sw=2:et
#
HARDWARE=$(shell uname -m)
VERSION=$(shell awk '/const Version/ { print $$4 }' version.go | sed 's/"//g')
DEPS=$(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
PACKAGES=$(shell go list ./...)
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

.PHONY: test examples authors changelog check-format coverage cover

build:
	go build

authors:
	git log --format='%aN <%aE>' | sort -u > AUTHORS

deps:
	@echo "--> Installing build dependencies"
	@go get -d -v ./... $(DEPS)

lint:
	@echo "--> Running golint"
	@which golint 2>/dev/null ; if [ $$? -eq 1 ]; then \
		go get -u github.com/golang/lint/golint; \
	fi
	@golint .

vet:
	@echo "--> Running go tool vet $(VETARGS) ."
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go tool vet $(VETARGS) .

cover:
	@echo "--> Running go test --cover"
	@go test --cover

coverage:
	@echo "--> Running go coverage"
	@go test -covermode=count -coverprofile=coverage

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

check-format:
	@echo "--> Checking format"
	@if gofmt -l . 2>&1 | grep -q '.go'; then \
		echo "found unformatted files:"; \
		echo; \
		gofmt -l .; \
		exit 1; \
	fi

test: deps
	@echo "--> Running go tests"
	@go test -race -v
	@$(MAKE) vet
	@$(MAKE) cover

examples:
	make -C examples all

changelog: release
	git log $(shell git tag | tail -n1)..HEAD --no-merges --format=%B > changelog
