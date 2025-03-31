default: build

API_BASE_URI?=
DOCS_URI?=

SRCS := $(shell find src/ -name '*.tsx' -o -name '*.ts')
STATICS := $(shell find public -type f)

.PHONY: start
start: yarn.lock
	@yarn start

.PHONY: lint
lint: yarn.lock
	@yarn lint

.PHONY: build
build: dist/.build-sentinel

dist/.build-sentinel: $(SRCS) $(STATICS) package.json
	@mkdir -p dist
	docker buildx build -f buildx.Dockerfile --output type=local,dest=. .
	@touch $@

.PHONY: clean
clean:
	rm -rf dist
