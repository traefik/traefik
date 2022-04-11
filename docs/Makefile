#######
# This Makefile contains all targets related to the documentation
#######

DOCS_VERIFY_SKIP ?= false
DOCS_LINT_SKIP ?= false

TRAEFIK_DOCS_BUILD_IMAGE ?= traefik-docs
TRAEFIK_DOCS_CHECK_IMAGE ?= $(TRAEFIK_DOCS_BUILD_IMAGE)-check

SITE_DIR := $(CURDIR)/site

DOCKER_RUN_DOC_PORT := 8000
DOCKER_RUN_DOC_MOUNTS := -v $(CURDIR):/mkdocs
DOCKER_RUN_DOC_OPTS := --rm $(DOCKER_RUN_DOC_MOUNTS) -p $(DOCKER_RUN_DOC_PORT):8000

# Default: generates the documentation into $(SITE_DIR)
docs: docs-clean docs-image docs-lint docs-build docs-verify
.PHONY: docs

# Writer Mode: build and serve docs on http://localhost:8000 with livereload
docs-serve: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOCS_BUILD_IMAGE) mkdocs serve
.PHONY: docs-serve

## Pull image for doc building
docs-pull-images:
	grep --no-filename -E '^FROM' ./*.Dockerfile \
		| awk '{print $$2}' \
		| sort \
		| uniq \
		| xargs -P 6 -n 1 docker pull
.PHONY: docs-pull-images

# Utilities Targets for each step
docs-image:
	docker build -t $(TRAEFIK_DOCS_BUILD_IMAGE) -f docs.Dockerfile ./
.PHONY: docs-image

docs-build: docs-image
	docker run $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOCS_BUILD_IMAGE) sh -c "mkdocs build \
		&& chown -R $(shell id -u):$(shell id -g) ./site"
.PHONY: docs-build

docs-verify: docs-build
ifneq ("$(DOCS_VERIFY_SKIP)", "true")
	docker build -t $(TRAEFIK_DOCS_CHECK_IMAGE) -f check.Dockerfile ./
	docker run --rm -v $(CURDIR):/app $(TRAEFIK_DOCS_CHECK_IMAGE) /verify.sh
else
	echo "DOCS_VERIFY_SKIP is true: no verification done."
endif
.PHONY: docs-verify

docs-lint:
ifneq ("$(DOCS_VERIFY_SKIP)", "true")
	docker build -t $(TRAEFIK_DOCS_CHECK_IMAGE) -f check.Dockerfile ./
	docker run --rm -v $(CURDIR):/app $(TRAEFIK_DOCS_CHECK_IMAGE) /lint.sh
else
	echo "DOCS_LINT_SKIP is true: no linting done."
endif
.PHONY: docs-lint

docs-clean:
	rm -rf $(SITE_DIR)
.PHONY: docs-clean
