test: clean
	go test -v ./ -cover

cover: clean
	go test -v ./ -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

clean:
	find . -name flymake_* -delete

sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l

install: clean
	go install github.com/vulcand/vulcand/vbundle
