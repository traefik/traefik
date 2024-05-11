# Use an official Ubuntu base image
FROM ubuntu:24.04

# Set environment variables
ENV PKG_NAME=traefik \
    PKG_VERSION=3.0.0 \
    PKG_ARCH=amd64 \
    PKG_MAINTAINER="Traefik contrib <traefik.dev@example.com>" \
    PKG_DOWNLOAD=https://github.com/traefik/traefik/releases/download/v3.0.0/traefik_v3.0.0_linux_amd64.tar.gz \
    BUILD_DIR=/build/traefik-deb

#VOLUME ["/build/traefik-deb"]

# Install necessary packages
RUN apt-get update && \
    apt-get install -y dpkg-dev debhelper wget

# Prepare environment
RUN mkdir -p $BUILD_DIR/src \
    $BUILD_DIR/deb \
    $BUILD_DIR/build/DEBIAN \
    $BUILD_DIR/build/usr/bin \
    $BUILD_DIR/build/etc/logrotate.d/ \
    $BUILD_DIR/build/etc/systemd/system \
    $BUILD_DIR/build/etc/traefik \
    $BUILD_DIR/build/etc/traefik/conf.d \
    $BUILD_DIR/build/var/log/traefik \
    $BUILD_DIR/build/usr/share/doc/traefik

# Set file permissions and copy configurations
COPY postinst $BUILD_DIR/build/DEBIAN/
COPY postrm $BUILD_DIR/build/DEBIAN/
COPY control_template $BUILD_DIR/build/DEBIAN/control
COPY systemd.conf $BUILD_DIR/build/etc/systemd/system/traefik.service
COPY traefik.yml $BUILD_DIR/build/etc/traefik/
COPY default.yml $BUILD_DIR/build/etc/traefik/conf.d/
COPY logrotate.conf $BUILD_DIR/build/etc/logrotate.d/traefik

RUN chmod +x $BUILD_DIR/build/DEBIAN/postinst
RUN chmod +x $BUILD_DIR/build/DEBIAN/postrm
RUN touch $BUILD_DIR/build/etc/traefik/acme.json
RUN chmod 0600 $BUILD_DIR/build/etc/traefik/acme.json
RUN chmod 0750 $BUILD_DIR/build/var/log/traefik

# Download and prepare the source
RUN wget -c $PKG_DOWNLOAD -O $BUILD_DIR/src/traefik.tar.gz
RUN tar -zxvf $BUILD_DIR/src/traefik.tar.gz -C $BUILD_DIR/src/
RUN cp $BUILD_DIR/src/traefik $BUILD_DIR/build/usr/bin/
RUN cp $BUILD_DIR/src/LICENSE.md $BUILD_DIR/build/usr/share/doc/traefik/ 
RUN gzip -f $BUILD_DIR/src/CHANGELOG.md 
RUN cp $BUILD_DIR/src/CHANGELOG.md.gz $BUILD_DIR/build/usr/share/doc/traefik/

# Template adjustments with sed
RUN sed -i \
    -e "s|__PKG_NAME__|${PKG_NAME}|g" \
    -e "s|__PKG_VERSION__|${PKG_VERSION}|g" \
    -e "s|__PKG_ARCH__|${PKG_ARCH}|g" \
    -e "s|__PKG_MAINTAINER__|${PKG_MAINTAINER}|g" \
    $BUILD_DIR/build/DEBIAN/control

RUN chown root:root -R $BUILD_DIR/build
RUN chown www-data:adm -R $BUILD_DIR/build/var/log/traefik
RUN chown www-data:adm -R $BUILD_DIR/build/etc/traefik/acme.json
RUN dpkg-deb --build $BUILD_DIR/build $BUILD_DIR/deb/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb

# Output the location of the deb package
CMD echo "Deb package is located in ${BUILD_DIR}/deb/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"
