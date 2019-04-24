.PHONY: test check

default: vendor check test

test:
	GO111MODULE=on go test -v ./...

check:
	golangci-lint run

vendor:
	GO111MODULE=on go mod vendor