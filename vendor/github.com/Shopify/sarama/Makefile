default: fmt vet errcheck test

# Taken from https://github.com/codecov/example-go#caveat-multiple-files
test:
	echo "" > coverage.txt
	for d in `go list ./... | grep -v vendor`; do \
		go test -v -timeout 60s -race -coverprofile=profile.out -covermode=atomic $$d; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done

vet:
	go vet ./...

errcheck:
	errcheck github.com/Shopify/sarama/...

fmt:
	@if [ -n "$$(go fmt ./...)" ]; then echo 'Please run go fmt on your code.' && exit 1; fi

install_dependencies: install_errcheck get

install_errcheck:
	go get github.com/kisielk/errcheck

get:
	go get -t
