# Project variables
APP_NAME = moh-sso-dashboard
BIN_DIR = bin
MAIN_FILE = ./cmd/server/main.go
GO_FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Docker variables (optional)
DOCKER_COMPOSE = docker-compose.yml

# Go commands
GO = go
GOTEST = $(GO) test
GOTIDY = $(GO) mod tidy
GOINSTALL = $(GO) install
GOBUILD = $(GO) build
GORUN = $(GO) run
GOFMT = gofmt
GOLINT = golangci-lint run

# Default target
.DEFAULT_GOAL := help

## —— 🛠️ General Commands ———————————————————————————————————————————————

help: ## Show available commands
	@echo ""
	@echo "📘 Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

tidy: ## Clean and tidy up go modules
	$(GOTIDY)

fmt: ## Format Go files
	@echo "Formatting code..."
	@$(GOFMT) -s -w $(GO_FILES)

lint: ## Run static analysis
	@echo "Running linter..."
	@$(GOLINT) || true

test: ## Run all tests
	@echo "Running tests..."
	@$(GOTEST) ./... -v

## —— 🚀 Build & Run ———————————————————————————————————————————————

build: ## Build the Go binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

run: ## Run the Go application
	@echo "🚀 Starting $(APP_NAME)..."
	$(GORUN) $(MAIN_FILE)

clean: ## Remove binaries and temp files
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)

## —— 🧩 Docker Commands (Optional) ————————————————————————————————————

docker-up: ## Start docker-compose services
	@echo "🟢 Starting Docker containers..."
	docker compose -f $(DOCKER_COMPOSE) up -d

docker-down: ## Stop docker-compose services
	@echo "🔴 Stopping Docker containers..."
	docker compose -f $(DOCKER_COMPOSE) down

docker-logs: ## View logs from docker containers
	docker compose logs -f

docker-restart: docker-down docker-up ## Restart docker services

## —— 🧰 Utility ———————————————————————————————————————————————

deps: ## Install go dependencies
	@echo "📦 Installing dependencies..."
	$(GO) mod download

verify: fmt lint test ## Run code checks before commit
	@echo "✅ All checks passed!"

.PHONY: help tidy fmt lint test build run clean docker-up docker-down docker-logs docker-restart deps verify
