test: clean
	go test -v ./...

coverage: clean
	gocov test -v ./... | gocov report

annotate: clean
	FILENAME=$(shell uuidgen)
	gocov test -v ./... > /tmp/--go-test-server-coverage.json
	gocov annotate /tmp/--go-test-server-coverage.json $(fn)

deps:
	go get -v -u github.com/axw/gocov
	go install github.com/axw/gocov/gocov
	go get -v -u launchpad.net/gocheck

clean:
	find . -name flymake_* -delete

msloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l
