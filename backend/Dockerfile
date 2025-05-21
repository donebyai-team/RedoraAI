# syntax=docker/dockerfile:1
FROM --platform=$TARGETPLATFORM golang:1.23-bullseye AS builder

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
RUN --mount=type=cache,id=s/173fddb5-d411-4982-b2fa-9d04f96c13e1-/root/.cache/go-build,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,id=s/173fddb5-d411-4982-b2fa-9d04f96c13e1-/root/.cache/go-build,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X main.version=${VERSION}" -o /app/doota ./cmd/doota

FROM debian:bullseye-slim as bullseye-base

ARG GOOGLE_APPLICATION_CREDENTIALS_JSON

# Install dependencies + Node.js + Playwright
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    wget \
    openssl \
    git \
    && curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g playwright \
    && npx playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/*


RUN mkdir -p /app

# Write the firebase config file
RUN echo "$GOOGLE_APPLICATION_CREDENTIALS_JSON" | base64 --decode > /app/firebase-config.json

COPY --from=builder /app/doota /app/doota

WORKDIR /app

EXPOSE 8787

CMD [ "/app/doota", "start" ]
