# Multi-stage build for Referee arbitrage simulation bot
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for building
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o referee ./cmd/referee

# Final stage: minimal runtime image
FROM gcr.io/distroless/static-debian11:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/referee /app/referee

# Copy config file
COPY --from=builder /app/config.example.yaml /app/config.yaml

# Set working directory
WORKDIR /app

# Expose any necessary ports (if needed in the future)
# EXPOSE 8080

# Run the application
ENTRYPOINT ["./referee"] 