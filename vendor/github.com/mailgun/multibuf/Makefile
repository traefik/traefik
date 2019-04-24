clean:
	find . -name flymake_* -delete

test: clean
	go test -v .

test-grep: clean
	go test -v . -check.f=$(e)

cover: clean
	go test -v .  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out
