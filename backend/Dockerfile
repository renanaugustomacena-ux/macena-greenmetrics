# syntax=docker/dockerfile:1.7
# --- Stage 1: builder ---------------------------------------------------------
FROM golang:1.26-alpine AS builder

WORKDIR /src

# hadolint ignore=DL3018 — Alpine apk uses rolling versions; pinning here would
# force a per-release Dockerfile update with no real benefit.
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download || true

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /out/greenmetrics-backend \
    ./cmd/server

# --- Stage 2: runtime (distroless, non-root) ---------------------------------
# NOTE: digest pinning (v2.0 §10.4). Resolve via: crane digest gcr.io/distroless/static-debian12:nonroot
# Pinned digest is refreshed monthly by Renovate; keep the :nonroot tag for humans.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

WORKDIR /app

COPY --from=builder /out/greenmetrics-backend /app/greenmetrics-backend
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nonroot:nonroot

EXPOSE 8080

ENV APP_ENV=production \
    APP_PORT=8080 \
    TZ=Europe/Rome

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/greenmetrics-backend", "--healthcheck"]

ENTRYPOINT ["/app/greenmetrics-backend"]
