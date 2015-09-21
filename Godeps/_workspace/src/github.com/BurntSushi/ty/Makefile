all: install

install:
	go install ./...

test: install
	go test ./...

benchcmp: install
	cd fun \
	&& echo "Running reflection benchmarks..." \
	&& go test -cpu 12 -run NONE -benchmem -bench . > reflect.bench \
	&& echo "Running built in benchmarks..." \
	&& go test -cpu 12 -run NONE -benchmem -bench . -builtin > builtin.bench  \
	&& benchcmp builtin.bench reflect.bench > ../perf/cmp.bench \
	&& rm builtin.bench reflect.bench

fmt:
	gofmt -w *.go */*.go
	colcheck *.go */*.go

tags:
	find ./ -name '*.go' -print0 | xargs -0 gotags > TAGS

push:
	git push origin master
	git push github master

