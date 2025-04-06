FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o sbom-app .

FROM alpine:latest

WORKDIR /app

# Install required dependencies
RUN apk add --no-cache curl bash syft grype wget tar findutils

# Install sbomqs directly from GitHub releases
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        export SBOMQS_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        export SBOMQS_ARCH="arm64"; \
    else \
        echo "Unsupported architecture: $ARCH"; \
        export SBOMQS_ARCH="amd64"; \
    fi && \
    # Download binary directly (releases use direct binaries)
    echo "Downloading sbomqs for architecture: ${SBOMQS_ARCH}" && \
    wget -q "https://github.com/interlynk-io/sbomqs/releases/download/v1.0.3/sbomqs-linux-${SBOMQS_ARCH}" -O /usr/local/bin/sbomqs && \
    chmod +x /usr/local/bin/sbomqs && \
    # Verify installation with fallback for errors
    { sbomqs version && echo "sbomqs installed successfully"; } || \
    { echo "WARNING: sbomqs is not installed correctly. Quality scoring will be unavailable but other functionality will continue to work."; }

# Copy the built application
COPY --from=build /app/sbom-app .

# Copy static files and templates
COPY static/ ./static/

# Create a directory for git cloning
RUN mkdir -p git-repos

# Set environment variables
ENV LLAMA_INDEX_ENDPOINT=http://llama-index-service:8000
ENV OLLAMA_HOST=http://host.docker.internal:11434
ENV DEFAULT_MODEL=mistral
ENV SBOM_OUTPUT_FILE=sbom.cyclonedx.json
ENV LOG_FILE=static/output.log

EXPOSE 3000

# Command to run the application
CMD ["./sbom-app"]