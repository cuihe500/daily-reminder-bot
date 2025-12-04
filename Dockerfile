# Daily Reminder Bot Dockerfile
# Multi-stage build for optimal image size

# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-s -w -X main.Version=docker -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o /app/daily-reminder-bot \
    ./cmd/bot/main.go

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.19

# Install runtime dependencies (for SQLite and timezone support)
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    openssl

# Create non-root user for security
RUN addgroup -S botuser && adduser -S botuser -G botuser

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/configs /app/data && \
    chown -R botuser:botuser /app

# Copy binary from builder stage
COPY --from=builder /app/daily-reminder-bot /app/

# Copy entrypoint script
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# Set ownership
RUN chown -R botuser:botuser /app

# Switch to non-root user
USER botuser

# Environment variables with defaults
# Telegram Configuration
ENV TELEGRAM_TOKEN=""
ENV TELEGRAM_API_ENDPOINT="https://api.telegram.org"

# QWeather Configuration
ENV QWEATHER_AUTH_MODE="jwt"
ENV QWEATHER_PRIVATE_KEY=""
ENV QWEATHER_KEY_ID=""
ENV QWEATHER_PROJECT_ID=""
ENV QWEATHER_API_KEY=""
ENV QWEATHER_BASE_URL=""

# OpenAI Configuration (optional)
ENV OPENAI_ENABLED="false"
ENV OPENAI_API_KEY=""
ENV OPENAI_BASE_URL="https://api.openai.com/v1"
ENV OPENAI_MODEL="gpt-4o-mini"
ENV OPENAI_MAX_TOKENS="800"
ENV OPENAI_TEMPERATURE="0.7"
ENV OPENAI_TIMEOUT="30"
ENV OPENAI_MAX_RETRIES="3"

# Holiday API Configuration (optional)
ENV HOLIDAY_API_URL=""
ENV HOLIDAY_CACHE_TTL="86400"

# Database Configuration
ENV DATABASE_TYPE="sqlite"
ENV DATABASE_PATH="/app/data/bot.db"
ENV DATABASE_HOST="localhost"
ENV DATABASE_PORT="3306"
ENV DATABASE_USER="root"
ENV DATABASE_PASSWORD=""
ENV DATABASE_NAME="daily_reminder_bot"
ENV DATABASE_CHARSET="utf8mb4"

# Scheduler Configuration
ENV SCHEDULER_TIMEZONE="Asia/Shanghai"

# Logger Configuration
ENV LOGGER_LEVEL="info"
ENV LOGGER_FORMAT="json"

# Data volume for persistence (SQLite database)
VOLUME ["/app/data"]

# Expose no ports (this is a bot, not a web server)

# Health check (check if process is running)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep -f daily-reminder-bot || exit 1

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]
