# ─── Build stage ─────────────────────────────────────────
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o teachhub .

# ─── Runtime stage ───────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

# Copy binary
COPY --from=builder /app/teachhub .

# Copy templates, schema, i18n, static assets
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/db ./db
COPY --from=builder /app/i18n ./i18n
COPY --from=builder /app/static ./static

# Create upload directories
RUN mkdir -p uploads/resources uploads/submissions uploads/quiz_files uploads/live

# Non-root user
RUN addgroup -S app && adduser -S app -G app
RUN chown -R app:app /app
USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -sf http://localhost:8080/ || exit 1

ENTRYPOINT ["./teachhub"]
