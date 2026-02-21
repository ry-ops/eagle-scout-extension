# eagle-scout-extension — Docker Desktop Extension

# Stage 1: build backend
FROM golang:1.25.7-alpine AS builder
WORKDIR /app
COPY backend/ .
RUN go mod init github.com/ry-ops/eagle-scout-extension && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /backend .

# Stage 2: extension image
FROM scratch

LABEL org.opencontainers.image.title="Eagle Scout" \
      org.opencontainers.image.description="Docker Scout security scanning dashboard for Docker Desktop" \
      org.opencontainers.image.vendor="ry-ops" \
      org.opencontainers.image.source="https://github.com/ry-ops/eagle-scout-extension" \
      com.docker.desktop.extension.api.version=">= 0.3.3" \
      com.docker.desktop.extension.icon="https://raw.githubusercontent.com/ry-ops/eagle-scout-extension/main/eagle-scout.svg" \
      com.docker.extension.screenshots="" \
      com.docker.extension.detailed-description="Eagle Scout surfaces Docker Scout CVE scanning, quickview, and base image recommendations directly inside Docker Desktop." \
      com.docker.extension.publisher-url="https://github.com/ry-ops" \
      com.docker.extension.categories="security"

COPY --from=builder /backend /backend
COPY metadata.json /metadata.json
COPY compose.yaml /compose.yaml
COPY eagle-scout.svg /eagle-scout.svg
COPY ui/ /ui
