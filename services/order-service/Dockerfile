# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy everything first
COPY . .

# Build - Go will auto-fetch missing modules and update go.sum (-mod=mod)
# GONOSUMDB=* skips the checksum database so no go.sum is required upfront
RUN GONOSUMDB="*" GONOSUMCHECK="*" GOFLAGS="-mod=mod"     CGO_ENABLED=0 GOOS=linux     go build -ldflags="-w -s" -o /service ./services/order-service

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /service /app/service
ENTRYPOINT ["/app/service"]
