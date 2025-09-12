FROM golang:1.25-alpine3.22 AS builder

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o seonaut cmd/server/main.go

FROM node:18-alpine3.18 AS front

WORKDIR /home/node
COPY ./web ./app/web

RUN npm install --save-exact esbuild && ./node_modules/esbuild/bin/esbuild ./app/web/css/style.css \
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

ENV WAIT_VERSION=2.12.1
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /bin/wait
RUN chmod +x /bin/wait

WORKDIR /app