FROM alpine

RUN apk --update upgrade \
&& apk --no-cache --no-progress add py-pip \
&& rm -rf /var/cache/apk/*

COPY requirements.txt /mkdocs/
WORKDIR /mkdocs


RUN pip install --user -r requirements.txt

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin
