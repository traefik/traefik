# this is an example of how to build a deb package for traefik
# for automatic build we should use CI/CD tools like Github Actions and docker containers

# how to use deb package building:
# 1) edit vars bellow first
# 2) create build_dir like
# sudo mkdir /build/traefik-deb
# sudo chown $(whoami): /build/traefik-deb
# 3) run build
# 3.1) local build
# make build_local
# 3.2) docker build
# make build_docker


PKG_NAME=traefik
PKG_VERSION=3.0.0
PKG_ARCH=amd64
PKG_MAINTAINER="Traefik contrib <traefik.dev@example.com>"
PKG_DOWNLOAD=https://github.com/traefik/traefik/releases/download/v3.0.0/traefik_v3.0.0_linux_amd64.tar.gz
BUILD_DIR=/build/traefik-deb


.PHONY: help
help:
	@echo "Usage: make <target>"

.PHONY: dependencies
dependencies:
	@echo "Installing dependencies..."
	sudo apt-get update
	sudo apt-get install -y dpkg-dev debhelper

.PHONY: envprepare
envprepare:
	@echo "Preparing environment..."
	sudo mkdir -p $(BUILD_DIR)/src
	sudo mkdir -p $(BUILD_DIR)/deb
	sudo mkdir -p $(BUILD_DIR)/build/DEBIAN
	sudo mkdir -p $(BUILD_DIR)/build/usr/bin
	sudo mkdir -p $(BUILD_DIR)/build/etc/logrotate.d/
	sudo mkdir -p $(BUILD_DIR)/build/etc/systemd/system
	sudo mkdir -p $(BUILD_DIR)/build/etc/traefik
	sudo mkdir -p $(BUILD_DIR)/build/etc/traefik/conf.d
	sudo cp postinst $(BUILD_DIR)/build/DEBIAN/
	sudo cp postrm $(BUILD_DIR)/build/DEBIAN/
	sudo cp control_template $(BUILD_DIR)/build/DEBIAN/control
	sudo cp systemd.conf $(BUILD_DIR)/build/etc/systemd/system/traefik.service
	sudo cp traefik.yml $(BUILD_DIR)/build/etc/traefik/
	sudo cp default.yml $(BUILD_DIR)/build/etc/traefik/conf.d/
	sudo chmod +x $(BUILD_DIR)/build/DEBIAN/postinst
	sudo chmod +x $(BUILD_DIR)/build/DEBIAN/postrm
	sudo touch $(BUILD_DIR)/build/etc/traefik/acme.json
	sudo chmod 0600 $(BUILD_DIR)/build/etc/traefik/acme.json
	sudo mkdir -p $(BUILD_DIR)/build/var/log/traefik
	sudo chmod 0750 $(BUILD_DIR)/build/var/log/traefik
	sudo rm $(BUILD_DIR)/build/etc/systemd/system/traefik.service
	sudo cp logrotate.conf $(BUILD_DIR)/build/etc/logrotate.d/traefik

.PHONY: envdownload
envdownload: 
	@echo "Downloading sources..."
	wget -c $(PKG_DOWNLOAD) -O $(BUILD_DIR)/src/traefik.tar.gz
	sudo tar -zxvf $(BUILD_DIR)/src/traefik.tar.gz -C $(BUILD_DIR)/src/
	sudo cp $(BUILD_DIR)/src/traefik $(BUILD_DIR)/build/usr/bin/
	sudo cp $(BUILD_DIR)/src/LICENSE.md $(BUILD_DIR)/build/usr/share/doc/traefik/
	sudo gzip -f $(BUILD_DIR)/src/CHANGELOG.md
	sudo cp $(BUILD_DIR)/src/CHANGELOG.md.gz $(BUILD_DIR)/build/usr/share/doc/traefik/

.PHONY: build_prepare
build_prepare:
	@echo "Preparing build..."
	sudo sed -i 's/\$$(PKG_NAME)/$(PKG_NAME)/g; s/\$$(PKG_VERSION)/$(PKG_VERSION)/g; s/\$$(PKG_ARCH)/$(PKG_ARCH)/g; s/\$$(PKG_MAINTAINER)/$(PKG_MAINTAINER)/g' $(BUILD_DIR)/build/DEBIAN/control
# this is need if not running as root
	sudo chown root:root -R $(BUILD_DIR)/build 
	sudo chown www-data:adm -R $(BUILD_DIR)/build/var/log/traefik
	sudo chown www-data:adm -R $(BUILD_DIR)/build/etc/traefik/acme.json

.PHONY: build_dpkg
build_dpkg:
	@echo "Building deb package..."
	sudo dpkg-deb --build $(BUILD_DIR)/build $(BUILD_DIR)/deb/$(PKG_NAME)_$(PKG_VERSION)_$(PKG_ARCH).deb

.PHONY: build_local
build_local: dependencies envprepare envdownload build_prepare build_dpkg
	@echo "Build completed successfully"
	@echo "Deb package is located in $(BUILD_DIR)/deb/$(PKG_NAME)_$(PKG_VERSION)_$(PKG_ARCH).deb"

.PHONY: build_docker
build_docker:
	@echo "Building deb package in docker..."
	docker build -t traefik-deb:$(PKG_VERSION) .
	

	@if docker ps -a | grep -q 'traefik-builder'; then \
		docker stop traefik-builder; \
		docker rm traefik-builder; \
	fi

	docker run -d --name traefik-builder traefik-deb:$(PKG_VERSION)
	docker cp traefik-builder:/build/traefik-deb/deb/traefik_$(PKG_VERSION)_$(PKG_ARCH).deb build/traefik_$(PKG_VERSION)_$(PKG_ARCH).deb

.PHONY: build
build: build_docker
	ls -la build/

.PHONY: clean
clean:
	rm -rf build/*
	docker rmi traefik-builder
	docker rmi traefik-deb:$(PKG_VERSION)