# Stage 1: Build Go binary
FROM golang:1.24-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Enable Go toolchain auto-download
ENV GOTOOLCHAIN=auto

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (pb_public will be mounted as volume)
COPY . .

# Ensure go.mod is tidy
RUN go mod tidy

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o huuper .

# Stage 2: Final runtime image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=go-builder /app/huuper .

# Copy migrations
COPY --from=go-builder /app/migrations ./migrations

# Create directory for data persistence
RUN mkdir -p /app/pb_data

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/health || exit 1

# Run the application
CMD ["./huuper", "serve", "--http=0.0.0.0:8090"]
