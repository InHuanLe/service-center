# Makefile for service-center multi-binary build
# Inspired by Kubernetes build system

# Project info
PROJECT_NAME := service-center
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directories
BUILD_DIR := _output
BIN_DIR := $(BUILD_DIR)/bin
DOCKER_DIR := $(BUILD_DIR)/docker

# Go build settings
GO := go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# Binaries to build
BINARIES := server client
CMD_DIRS := $(foreach bin,$(BINARIES),cmd/$(bin))

# Docker settings
DOCKER_REGISTRY ?= 
DOCKER_IMAGE_PREFIX ?= $(PROJECT_NAME)
DOCKER_TARGETS := server client all

# Build flags
LDFLAGS := -w -s \
	-X main.Version=$(VERSION) \
	-X main.BuildDate=$(BUILD_DATE) \
	-X main.GitCommit=$(GIT_COMMIT)

# Targets
.PHONY: all
all: build

.PHONY: help
help: ## Display this help
	@echo "Service Center Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: $(BINARIES) ## Build all binaries

.PHONY: $(BINARIES)
$(BINARIES): %: ## Build specific binary (e.g., make server)
	@echo "Building $@ for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BIN_DIR)/$(GOOS)/$(GOARCH)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build -ldflags "$(LDFLAGS)" \
		-o $(BIN_DIR)/$(GOOS)/$(GOARCH)/$@ \
		./cmd/$@

.PHONY: build-linux
build-linux: ## Build all binaries for Linux
	@$(MAKE) build GOOS=linux GOARCH=amd64

.PHONY: build-windows
build-windows: ## Build all binaries for Windows
	@$(MAKE) build GOOS=windows GOARCH=amd64

.PHONY: build-darwin
build-darwin: ## Build all binaries for macOS
	@$(MAKE) build GOOS=darwin GOARCH=amd64

.PHONY: build-all-platforms
build-all-platforms: build-linux build-windows build-darwin ## Build for all platforms

.PHONY: docker-build
docker-build: ## Build all Docker images
	@for target in $(DOCKER_TARGETS); do \
		echo "Building Docker image for $$target..."; \
		docker build --target $$target -t $(DOCKER_IMAGE_PREFIX):$$target-$(VERSION) .; \
		docker tag $(DOCKER_IMAGE_PREFIX):$$target-$(VERSION) $(DOCKER_IMAGE_PREFIX):$$target-latest; \
	done

.PHONY: docker-build-server
docker-build-server: ## Build server Docker image
	docker build --target server -t $(DOCKER_IMAGE_PREFIX):server-$(VERSION) .
	docker tag $(DOCKER_IMAGE_PREFIX):server-$(VERSION) $(DOCKER_IMAGE_PREFIX):server-latest

.PHONY: docker-build-client
docker-build-client: ## Build client Docker image
	docker build --target client -t $(DOCKER_IMAGE_PREFIX):client-$(VERSION) .
	docker tag $(DOCKER_IMAGE_PREFIX):client-$(VERSION) $(DOCKER_IMAGE_PREFIX):client-latest

.PHONY: docker-push
docker-push: ## Push Docker images to registry
	@for target in $(DOCKER_TARGETS); do \
		echo "Pushing $$target image..."; \
		docker push $(DOCKER_REGISTRY)$(DOCKER_IMAGE_PREFIX):$$target-$(VERSION); \
		docker push $(DOCKER_REGISTRY)$(DOCKER_IMAGE_PREFIX):$$target-latest; \
	done

.PHONY: test
test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: lint
lint: ## Run linters
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod
	$(GO) mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: proto
proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

.PHONY: run-server
run-server: server ## Build and run server locally
	$(BIN_DIR)/$(GOOS)/$(GOARCH)/server

.PHONY: run-client
run-client: client ## Build and run client locally
	$(BIN_DIR)/$(GOOS)/$(GOARCH)/client

.PHONY: install
install: ## Install binaries to GOPATH/bin
	@for bin in $(BINARIES); do \
		echo "Installing $$bin..."; \
		$(GO) install -ldflags "$(LDFLAGS)" ./cmd/$$bin; \
	done

# Quick dev workflow
.PHONY: dev
dev: fmt vet test build ## Quick development workflow: format, vet, test, build

.DEFAULT_GOAL := help
