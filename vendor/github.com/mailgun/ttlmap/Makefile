test: clean
	go test -cover -v ./...

coverage: clean
	go test -coverprofile=/tmp/coverage.out -v ./...
	go tool cover -func=/tmp/coverage.out

htmlcoverage: clean
	go test -covermode=count -coverprofile=/tmp/coverage.out -v ./...
	go tool cover -html=/tmp/coverage.out

deps:
	go get -v -u launchpad.net/gocheck
	go get -v -u github.com/mailgun/minheap
	go get -v -u github.com/mailgun/timetools

clean:
	find . -name flymake_* -delete

sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l
