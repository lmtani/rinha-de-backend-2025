# Stage 1: Build the Go application with a smaller builder image
FROM golang:1.24-alpine AS builder

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=0: Disables CGO for a static binary
# -ldflags="-s -w": Strips debug information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /go/bin/app ./cmd/api

# Stage 2: Create the minimal runtime image
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/app /app

# Set environment variables for the application
ENV PROCESSOR_DEFAULT_URL=http://payment-processor-default:8080
ENV PROCESSOR_FALLBACK_URL=http://payment-processor-fallback:8080

# Expose the application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]
