VERSION?="0.3.12"

website:
	@echo "==> Starting website in Docker..."
	@docker run \
	  --interactive \
	  --rm \
		--tty \
		--publish "4567:4567" \
		--publish "35729:35729" \
		--volume "$(shell pwd):/website" \
		hashicorp/middleman-hashicorp:${VERSION}

.PHONY: website
