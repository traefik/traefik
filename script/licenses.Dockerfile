# syntax=docker/dockerfile:1
# Image used by `make generate-licenses` / `make validate-licenses` so the SBOM
# is generated on a fixed linux/amd64 toolchain regardless of the host.

ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-bookworm
COPY --from=ghcr.io/astral-sh/uv@sha256:2381d6aa60c326b71fd40023f921a0a3b8f91b14d5db6b90402e65a635053709 /uv /uvx /usr/local/bin/

ARG NODE_MAJOR=24
ARG YARN_VERSION=4.13.0
ARG CYCLONEDX_GOMOD_VERSION=v1.10.0
ARG CYCLONEDX_PY_VERSION=7.3.0
ARG ASSIMILIS_VERSION=81e7c914faaed8e2fe52e39cd02cc946416bf22a

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        gnupg \
        jq \
        python3 \
        python3-pip \
        python3-venv \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && corepack enable \
    && corepack prepare "yarn@${YARN_VERSION}" --activate

RUN pip install --break-system-packages "cyclonedx-bom==${CYCLONEDX_PY_VERSION}"

RUN go install "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@${CYCLONEDX_GOMOD_VERSION}" \
    && go install "github.com/traefik/assimilis/cmd/assimilis@${ASSIMILIS_VERSION}"

# Bind-mounted /src is owned by the host UID; allow git to operate on it.
RUN git config --system --add safe.directory '*'

# Tools that resolve module / package caches must be able to write under HOME.
ENV HOME=/tmp \
    GOPATH=/tmp/go \
    GOCACHE=/tmp/.cache/go-build \
    XDG_CACHE_HOME=/tmp/.cache

WORKDIR /src
USER nobody:nogroup
