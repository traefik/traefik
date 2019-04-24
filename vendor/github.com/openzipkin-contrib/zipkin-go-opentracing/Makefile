
.DEFAULT_GOAL := test

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: bench
bench:
	go test -v -run - -bench . -benchmem ./...

.PHONY: lint
lint:
	# Ignore grep's exit code since no match returns 1.
	-golint ./... | grep --invert-match -E '^.*\.pb\.go|^thrift'
	@
	@! (golint ./... | grep --invert-match -E '^.*\.pb\.go|^thrift' | read dummy)

.PHONY: vet
vet:
	go vet ./...

.PHONY: all
all: vet lint test bench

.PHONY: example
