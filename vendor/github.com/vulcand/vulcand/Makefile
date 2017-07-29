ETCD_NODE1 := http://127.0.0.1:4001
ETCD_NODE2 := http://127.0.0.1:4002
ETCD_NODE3 := http://127.0.0.1:4003
ETCD_NODES := ${ETCD_NODE1},${ETCD_NODE2},${ETCD_NODE3}
API_URL := http://localhost:8182
SERVICE_URL := http://localhost:8181
PREFIX := /vulcandtest
SEAL_KEY := 1b727a055500edd9ab826840ce9428dc8bace1c04addc67bbac6b096e25ede4b
GO15VENDOREXPERIMENT := 1

ETCD_FLAGS := VULCAND_TEST_ETCD_NODES=${ETCD_NODES}
VULCAN_FLAGS := VULCAND_TEST_ETCD_NODES=${ETCD_NODES} VULCAND_TEST_ETCD_PREFIX=${PREFIX} VULCAND_TEST_API_URL=${API_URL} VULCAND_TEST_SERVICE_URL=${SERVICE_URL} VULCAND_TEST_SEAL_KEY=${SEAL_KEY}

test: clean
	go test -v ./... -cover

test-with-etcd: clean
	${ETCD_FLAGS} go test -v ./... -cover

test-with-vulcan: clean
	${VULCAN_FLAGS} go test -v ./... -cover

clean:
	find . -name flymake_* -delete

test-package: clean
	go test -v ./$(p)

test-package-with-etcd: clean
	${ETCD_FLAGS} go test -v ./$(p)

update:
	rm -rf Godeps/
	find . -iregex .*go | xargs sed -i 's:".*Godeps/_workspace/src/:":g'
	godep save -r ./...

test-grep-etcdng: clean
	${ETCD_FLAGS} go test -v ./engine/etcdng -check.f=$(e)

test-grep-package: clean
	go test -v ./$(p) -check.f=$(e)

cover-package: clean
	go test -v ./$(p)  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

cover-package-with-etcd: clean
	${ETCD_FLAGS} go test -v ./$(p)  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

systest: clean install
	${VULCAN_FLAGS} go test -v ./systest

systest-grep: clean install
	${VULCAN_FLAGS} go test -v ./systest -check.f=$(e)

sloccount:
	 find . -path ./Godeps -prune -o -name "*.go" -print0 | xargs -0 wc -l

install: clean
	go install github.com/vulcand/vulcand
	cd vctl && $(MAKE) install && cd ..
	cd vbundle && $(MAKE) install && cd ..

run: install
	vulcand -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=/vulcand -sealKey=${SEAL_KEY} -statsdAddr=localhost:8125 -statsdPrefix=vulcan -logSeverity=INFO

run-fast: install
	vulcand -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=/vulcand -sealKey=${SEAL_KEY}

run-test-mode: install
	vulcand -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=${PREFIX} -sealKey=${SEAL_KEY} -logSeverity=INFO

profile:
	go tool pprof http://localhost:6060/debug/pprof/profile

docker-clean:
	docker rm -f vulcand

docker-build:
	GOOS=linux go build -a -tags netgo -installsuffix cgo -ldflags '-w' -o ./vulcand .
	GOOS=linux go build -a -tags netgo -installsuffix cgo -ldflags '-w' -o ./vctl/vctl ./vctl
	GOOS=linux go build -a -tags netgo -installsuffix cgo -ldflags '-w' -o ./vbundle/vbundle ./vbundle
	docker build -t mailgun/vulcand:latest -f ./Dockerfile-scratch .

docker-minimal-linux:
	bash scripts/build-minimal-linux.sh ${SEAL_KEY}

docker-run-fast: docker-build
	docker run -d --net=host --name vulcand mailgun/vulcand -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=/vulcand -sealKey=${SEAL_KEY}

docker-run-test-mode: docker-build
	docker run -d --net=host --name vulcand mailgun/vulcand -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=${PREFIX} -sealKey=${SEAL_KEY} -logSeverity=INFO

.PHONY: test test-with-etcd test-with-vulcan clean test-package test-package-with-etcd update test-grep-etcdng test-grep-package cover-package cover-package-with-etcd systest systest-grep sloccount install run run-fast run-test-mode profile docker-clean docker-build docker-minimal-linux docker-run-fast docker-run-test-mode
