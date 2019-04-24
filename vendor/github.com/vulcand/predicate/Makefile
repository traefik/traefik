
test: clean
	go test -v ./... -cover

clean:
	find . -name flymake_* -delete

cover: clean
	go test -v .  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l
