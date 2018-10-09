#######
# This Makefile contains all targets related to the documentation
#######

BUILD_DOC_DOCKER_IMAGE ?= traefik-docs
VERIFY_DOC_DOCKER_IMAGE ?= $(BUILD_DOC_DOCKER_IMAGE)-verify
DOCKER_RUN_DOC_PORT := 8000
DOCKER_RUN_DOC_MOUNT := -v $(CURDIR):/mkdocs
DOCKER_RUN_DOC_OPTS := --rm $(DOCKER_RUN_DOC_MOUNT)
DOCS_VERIFY_SKIP ?= false

default: docs

docs: docs-clean docs-image docs-build docs-verify

docs-image:
	docker build -t $(BUILD_DOC_DOCKER_IMAGE) -f build.Dockerfile ./

docs-serve: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) -p $(DOCKER_RUN_DOC_PORT):8000 $(BUILD_DOC_DOCKER_IMAGE) mkdocs serve

docs-build: docs-image site

site:
	docker run  $(DOCKER_RUN_DOC_OPTS) $(BUILD_DOC_DOCKER_IMAGE) mkdocs build

docs-verify: site
	@if [ "$(DOCS_VERIFY_SKIP)" != "true" ]; then \
		docker build -t $(VERIFY_DOC_DOCKER_IMAGE) -f verify.Dockerfile ./; \
		docker run --rm -v $(CURDIR):/app $(VERIFY_DOC_DOCKER_IMAGE); \
	else \
		echo "DOCS_VERIFY_SKIP is true: no verification done."; \
	fi

docs-clean:
	rm -rf $(CURDIR)/site

.PHONY: default docs docs-image docs-serve docs-build docs-verify docs-clean
