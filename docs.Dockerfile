FROM alpine:3.7

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

COPY requirements.txt /mkdocs/
WORKDIR /mkdocs
VOLUME /mkdocs

RUN apk --update upgrade \
&& apk --no-cache --no-progress add py-pip \
&& rm -rf /var/cache/apk/* \
&& pip install --upgrade pip \
&& pip install --user -r requirements.txt