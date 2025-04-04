# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app
RUN apt-get update && apt-get install -y git curl bash git 
# Install Grype
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin
# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o sbom-app main.go

# Stage 2: Use Debian-based runtime image for production
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y git bash curl ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/sbom-app .
COPY --from=builder /app/static ./static
COPY --from=builder /usr/local/bin/grype /usr/local/bin/grype

EXPOSE 3000

ENTRYPOINT ["/app/sbom-app"]