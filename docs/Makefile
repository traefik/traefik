
#######
# This Makefile contains all targets related to the documentation
#######

DOCS_VERIFY_SKIP ?= false
DOCS_LINT_SKIP ?= false

TRAEFIK_DOCS_BUILD_IMAGE ?= traefik-docs
TRAEFIK_DOCS_CHECK_IMAGE ?= $(TRAEFIK_DOCS_BUILD_IMAGE)-check
DOCKER_RUN_DOC_PORT := 8000
DOCKER_RUN_DOC_MOUNT := -v $(CURDIR):/mkdocs
DOCKER_RUN_DOC_OPTS := --rm $(DOCKER_RUN_DOC_MOUNT) -e GIT_REF_NAME -e AWS_S3_BUCKET_ID -e DOCKER_IMAGE_NAME

DOCKER_RUN_DOC_PORT := 8000
DOCKER_RUN_DOC_MOUNT := -v $(CURDIR):/mkdocs
DOCKER_RUN_DOC_OPTS := --rm $(DOCKER_RUN_DOC_MOUNT) -p $(DOCKER_RUN_DOC_PORT):8000

# Default: generates the documentation into $(CURDIR)/site
docs: docs-clean docs-image docs-build docs-verify

# Writer Mode: build and serve docs on http://localhost:8000 with livereload
docs-serve: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOC_IMAGE) mkdocs serve

# Utilities Targets for each step
docs-image:
	docker build -t $(TRAEFIK_DOCS_BUILD_IMAGE) -f docs.Dockerfile ./

docs-build: docs-image
	docker run  $(DOCKER_RUN_DOC_OPTS) $(TRAEFIK_DOCS_BUILD_IMAGE) mkdocs build

docs-verify: site
	@if [ "$(DOCS_VERIFY_SKIP)" != "true" ]; then \
		docker build -t $(TRAEFIK_DOCS_CHECK_IMAGE) -f check.Dockerfile ./; \
		docker run --rm -v $(CURDIR):/app $(TRAEFIK_DOCS_CHECK_IMAGE) /verify.sh; \
	else \
		@echo "DOCS_VERIFY_SKIP is true: no verification done."; \
	fi

docs-lint:
	@if [ "$(DOCS_LINT_SKIP)" != "true" ]; then \
		docker build -t $(TRAEFIK_DOCS_CHECK_IMAGE) -f check.Dockerfile ./ && \
		docker run --rm -v $(CURDIR):/app $(TRAEFIK_DOCS_CHECK_IMAGE) /lint.sh; \
	else \
		@echo "DOCS_LINT_SKIP is true: no linting done."; \
	fi

docs-clean:
	rm -rf $(CURDIR)/site

.PHONY: all docs-verify docs docs-clean docs-build
