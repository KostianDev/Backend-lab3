# syntax=docker/dockerfile:1

FROM golang:1.25.3 AS builder
WORKDIR /workspace

COPY go.mod go.sum ./
COPY src ./src

RUN go build -o /workspace/bin/app ./src/cmd/app

FROM debian:bookworm-slim AS runtime

ENV GIN_MODE=release
WORKDIR /srv/app

COPY --from=builder /workspace/bin/app ./app

EXPOSE 8080
CMD ["./app"]
