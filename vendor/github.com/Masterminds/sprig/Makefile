
HAS_GLIDE := $(shell command -v glide;)

.PHONY: test
test:
	go test -v .

.PHONY: setup
setup:
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
	glide install
