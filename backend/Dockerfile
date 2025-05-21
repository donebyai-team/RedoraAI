# syntax=docker/dockerfile:1
FROM --platform=$TARGETPLATFORM golang:1.23-bullseye AS builder

WORKDIR /src

# Declare build arguments for railway
ARG GOOGLE_APPLICATION_CREDENTIALS_JSON
RUN echo "$GOOGLE_APPLICATION_CREDENTIALS_JSON" | base64 --decode > /app/firebase-config.json


ENV GOPRIVATE=github.com/shank318/doota
ENV CGO_ENABLED=1
ARG TARGETOS TARGETARCH
ARG VERSION v1.0.12


RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,id=s/173fddb5-d411-4982-b2fa-9d04f96c13e1-/root/.cache/go-build,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,id=s/173fddb5-d411-4982-b2fa-9d04f96c13e1-/root/.cache/go-build,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X main.version=${VERSION}" -o /app/doota ./cmd/doota

FROM debian:bullseye-slim as bullseye-base

# Install dependencies + Node.js + Playwright
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    wget \
    openssl \
    git \
    nodejs \
    npm && \
    npm install -g playwright && \
    playwright install chromium && \
    rm -rf /var/lib/apt/lists/*


RUN mkdir /app

# Write the firebase config file
echo "$GOOGLE_APPLICATION_CREDENTIALS_JSON" | base64 --decode > /app/firebase-config.json


COPY --from=builder /app/doota /app/doota

WORKDIR /app

EXPOSE 8787

CMD [ "/app/doota", "start" ]
