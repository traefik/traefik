FROM alpine:3.22

ENV PATH="${PATH}:/venv/bin"

COPY requirements.txt /mkdocs/
WORKDIR /mkdocs
VOLUME /mkdocs

RUN apk --no-cache --no-progress add py3-pip gcc musl-dev python3-dev \
  && python3 -m venv /venv \
  && source /venv/bin/activate \
  && pip3 install -r requirements.txt
