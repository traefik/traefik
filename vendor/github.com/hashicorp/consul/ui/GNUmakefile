ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

server:
	python -m SimpleHTTPServer

watch:
	sass styles:static --watch

dist:
	@sh -c "'$(ROOT)/scripts/dist.sh'"

.PHONY: server watch dist
