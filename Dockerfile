# Use the official Go image to build the binary.
FROM golang:1.23-bookworm AS builder

# Install the ARM cross compiler (for ARM hard float)
RUN apt-get update && apt-get install -y gcc-arm-linux-gnueabihf

# Set the working directory in the container
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Metadata
LABEL project="ModulaCMS"


# Copy the source code.
COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=arm
ENV CC=arm-linux-gnueabihf-gcc

# Build the Go binary statically for Linux.
# RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o modulacms ./cmd
RUN  GO111MODULE=on go build -mod vendor -o modulacms ./cmd	

