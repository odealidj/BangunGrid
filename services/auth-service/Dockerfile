# syntax=docker/dockerfile:1

# ==========================================
# Stage 1: Build stage
# ==========================================
FROM golang:1.23-alpine AS builder

# Pasang dependensi build dan sertifikat CA
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Salin modul pkg dan auth-service
COPY pkg/ ./pkg/
COPY services/auth-service/ ./services/auth-service/

ENV GOWORK=off
WORKDIR /workspace/services/auth-service

# Unduh dependensi dan kompilasi binary statis
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/auth-service ./cmd/server

# ==========================================
# Stage 2: Runtime stage (Minimal & Non-Root)
# ==========================================
FROM alpine:3.20

# Pasang sertifikat CA, tzdata, dan curl untuk healthcheck
RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -S -g 10001 appgroup && \
    adduser -S -u 10001 -G appgroup appuser

WORKDIR /app

# Salin binary hasil kompilasi
COPY --from=builder /app/auth-service /app/auth-service

# Gunakan non-root user
USER 10001:10001

# Port HTTP default auth-service
EXPOSE 8081

# Healthcheck internal
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/healthz || exit 1

ENTRYPOINT ["/app/auth-service"]
