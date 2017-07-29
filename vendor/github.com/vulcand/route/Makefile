.PHONY: prof

clean:
	find . -name flymake_* -delete

test: clean
	go test -v .

test-grep:
	go test -v . -check.f=$(e)

cover: clean
	go test -v .  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

bench: clean
	go test . -check.bmem  -check.b -test.bench=.

prof: clean
	go test . -check.bmem  -check.b -test.bench=-. -test.memprofile=/tmp/memprof.prof -test.cpuprofile=/tmp/cpuprof.prof

sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l
