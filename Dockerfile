#syntax=docker/dockerfile:1.18.0
# builder https://docs.docker.com/build/buildkit/dockerfile-release-notes/
FROM golang:1.25-alpine3.22 AS builder

RUN mkdir /app
COPY . /app
WORKDIR /app

# RUN https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -o seonaut cmd/server/main.go

FROM node:18-alpine3.18 AS front

WORKDIR /home/node
COPY ./web ./app/web

RUN --mount=type=cache,target=/root/.npm \
	npm install --save-exact esbuild && \
	./node_modules/esbuild/bin/esbuild ./app/web/css/style.css \
	--bundle \
	--minify \
	--outdir=./app/web/static \
	--public-path=/resources \
	--loader:.woff=file \
	--loader:.woff2=file

FROM alpine:latest AS production

COPY --from=builder /app/seonaut /app/seonaut
COPY --from=front /home/node/app /app/

COPY ./translations /app/translations
COPY ./migrations /app/migrations
COPY ./config /app/config

ARG TARGETARCH
# https://medium.com/@tonistiigi/new-dockerfile-capabilities-in-v1-7-0-be6873650741
# WAIT_ARCH argument string substitution requires Dockerfile 1.7.0 or newer syntax.
ARG WAIT_ARCH=${TARGETARCH/amd64/_x86_64}
ARG WAIT_ARCH=${WAIT_ARCH/arm64/_aarch64}
ARG WAIT_ARCH=${WAIT_ARCH/arm_v7/_armv7}
ARG WAIT_ARCH=${WAIT_ARCH:-}
# WAIT_VERSION https://github.com/ufoscout/docker-compose-wait/releases
ARG WAIT_VERSION=2.12.1
ADD --chmod=755 https://github.com/ufoscout/docker-compose-wait/releases/download/${WAIT_VERSION}/wait${WAIT_ARCH} /bin/wait

WORKDIR /app
