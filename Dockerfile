FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/pollapp ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Add necessary runtime dependencies
RUN apk add --no-cache tzdata ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/pollapp .

# Copy migrations directory
COPY --from=builder /app/internal/repository/postgresql/migrations /app/internal/repository/postgresql/migrations

# Copy entry script
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Copy .env file if needed, or it will be mounted as a volume
COPY .env.docker .env

# Expose the application port
EXPOSE 8080

# Run the entrypoint script
CMD ["./docker-entrypoint.sh"] 