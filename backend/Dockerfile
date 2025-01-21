# syntax=docker/dockerfile:1
FROM --platform=$TARGETPLATFORM golang:1.22-bullseye AS builder

WORKDIR /src


ENV GOPRIVATE=github.com/shank318/doota
ENV CGO_ENABLED=1
ARG TARGETOS TARGETARCH
ARG VERSION v1.0.12


RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X main.version=${VERSION}" -o /app/doota ./cmd/doota

FROM debian:bullseye-slim as bullseye-base

RUN apt-get update && apt-get install -y \
    openssl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*


RUN mkdir /app

COPY --from=builder /app/doota /app/doota

WORKDIR /app

EXPOSE 8787

CMD [ "/app/doota" ]
