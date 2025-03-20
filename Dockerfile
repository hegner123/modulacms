# Use the official Go image to build the binary.
FROM golang:1.20-alpine AS builder
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Metadata
LABEL project="ModulaCMS"

# Copy the source code.
COPY . .

# Build the Go binary statically for Linux.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o modulacms .

# Use a minimal base image.
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage.
WORKDIR /
COPY --from=builder /app/modulacms .

# Expose port
EXPOSE 4000

# Run the Go binary.
ENTRYPOINT ["./modulacms"]

# Stop signal
STOPSIGNAL SIGTERM


