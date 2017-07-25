.PHONY: all prepare build lint vet test check release

all: prepare lint vet test build

prepare:
	go get -v github.com/golang/lint/golint
	go get -v github.com/Masterminds/glide
	glide install

build:
	GOARCH=amd64 GOOS=linux go install

lint:
	for pkg in $$(go list ./... | grep -v /vendor/); do golint $$pkg; done

vet:
	GOARCH=amd64 GOOS=linux go vet $$(go list ./... | grep -v /vendor/)

test:
	GOARCH=amd64 GOOS=linux go test $$(go list ./... | grep -v /vendor/)

check: lint vet test

release:
	goxc -os="linux darwin windows freebsd openbsd" -tasks-=validate
