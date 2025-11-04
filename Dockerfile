# syntax=docker/dockerfile:1

FROM golang:1.25.3 AS builder
WORKDIR /workspace

COPY go.work ./
COPY src/go.mod src/
COPY src/go.sum src/
COPY src/ src/

RUN cd src && go build -o /workspace/bin/app ./cmd/app

FROM debian:bookworm-slim AS runtime

ENV GIN_MODE=release
WORKDIR /srv/app

COPY --from=builder /workspace/bin/app ./app

EXPOSE 8080
CMD ["./app"]
