.PHONY: build clean test run docker-build docker-push

# Binary name
BINARY_NAME=misskey-summarizer

# Docker image settings
DOCKER_REGISTRY?=ghcr.io
DOCKER_IMAGE_NAME?=soli0222/misskey-summarizer
DOCKER_TAG?=latest

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd/misskey-summarizer

# Build with optimizations
build-release:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_NAME) ./cmd/misskey-summarizer

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

# Run tests
test:
	go test -v ./...

# Run the application (requires env vars)
run:
	go run ./cmd/misskey-summarizer

# Run with yesterday's date
run-yesterday:
	go run ./cmd/misskey-summarizer --yesterday

# Build Docker image
docker-build:
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

# Push Docker image
docker-push:
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Build and push Docker image
docker-release: docker-build docker-push

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-release  - Build optimized binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  run            - Run the application"
	@echo "  run-yesterday  - Run with --yesterday flag"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image"
	@echo "  docker-release - Build and push Docker image"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
