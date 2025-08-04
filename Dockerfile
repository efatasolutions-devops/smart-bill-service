# Multi-stage build untuk optimasi ukuran image
FROM golang:1.21-alpine AS builder

# Install dependencies yang diperlukan
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smart-bill-service main.go

# Final stage - menggunakan image minimal
FROM alpine:latest

# Install ca-certificates untuk HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary dari builder stage
COPY --from=builder /app/smart-bill-service .

# Copy storage directory structure
COPY --from=builder /app/storage ./storage

# Create necessary directories
RUN mkdir -p storage/logs/general_log storage/public/images

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run aplikasi
CMD ["./smart-bill-service"]
