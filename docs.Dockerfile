FROM alpine

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

COPY requirements.txt /mkdocs/
WORKDIR /mkdocs

RUN apk --update upgrade \
&& apk --no-cache --no-progress add py-pip \
&& rm -rf /var/cache/apk/* \
&& pip install --user -r requirements.txt
