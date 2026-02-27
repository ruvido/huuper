FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy prebuilt binary and migrations
COPY bin/huuper ./huuper
COPY migrations ./migrations

# Create directory for data persistence
RUN chmod +x ./huuper && mkdir -p /app/pb_data /app/pb_public

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/health || exit 1

# Run the application
CMD ["./huuper", "serve", "--http=0.0.0.0:8090"]
