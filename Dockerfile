# Build stage
FROM golang:1.25.6-alpine3.23 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Install ca-certificates and timezone data for HTTPS and time handling
RUN apk --no-cache add ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /misskey-summarizer ./cmd/misskey-summarizer

# Runtime stage
FROM scratch

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /misskey-summarizer /misskey-summarizer

# Run as non-root (UID 65534 is nobody)
USER 65534:65534

ENTRYPOINT ["/misskey-summarizer"]
