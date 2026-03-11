# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install required system packages
# (None needed for basic Go build, but git sometimes needed for private modules)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install CA certificates for HTTPS and ffmpeg for future video processing
RUN apk --no-cache add ca-certificates ffmpeg

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
