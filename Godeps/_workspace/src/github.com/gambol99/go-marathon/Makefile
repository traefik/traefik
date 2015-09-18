#
#   Author: Rohith (gambol99@gmail.com)
#   Date: 2015-02-10 15:35:14 +0000 (Tue, 10 Feb 2015)
#
#  vim:ts=2:sw=2:et
#
NAME="go-marathon"
AUTHOR=gambol99
HARDWARE=$(shell uname -m)
VERSION=$(shell awk '/const Version/ { print $$4 }' version.go | sed 's/"//g')

.PHONY: test examples authors changelog

build:
	go build

authors:
	git log --format='%aN <%aE>' | sort -u > AUTHORS

test:
	go get github.com/stretchr/testify/assert
	go get gopkg.in/yaml.v2
	go test -v

changelog: release
	git log $(shell git tag | tail -n1)..HEAD --no-merges --format=%B > changelog

