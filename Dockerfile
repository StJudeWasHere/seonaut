FROM golang:1.18-alpine3.16 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o seonaut cmd/server/main.go

FROM alpine:latest AS production
COPY --from=builder /app .

ENV WAIT_VERSION 2.7.2
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /wait
RUN chmod +x /wait

CMD ["./seonaut"]