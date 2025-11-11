# -----------------------------
# Project Metadata
# -----------------------------
PROJECT_NAME = moh-sso-dashboard
GO_CMD = go
GO_MAIN = ./cmd/server/main.go
FRONTEND_DIR = ./frontend
BINARY_NAME = sso-server
DOCKER_IMAGE = $(PROJECT_NAME):latest

# -----------------------------
# Go Backend Commands
# -----------------------------
.PHONY: build
build:
	@echo "🚀 Building Go backend..."
	cd backend && $(GO_CMD) build -o ../bin/$(BINARY_NAME) $(GO_MAIN)
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

.PHONY: run
run:
	@echo "🏃 Running Go backend..."
	cd backend && $(GO_CMD) run $(GO_MAIN)

.PHONY: test
test:
	@echo "🧪 Running tests..."
	cd backend && $(GO_CMD) test ./... -v

.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	cd backend && $(GO_CMD) fmt ./...

.PHONY: tidy
tidy:
	@echo "🧹 Tidying dependencies..."
	cd backend && $(GO_CMD) mod tidy

.PHONY: clean
clean:
	@echo "🧽 Cleaning up..."
	rm -rf bin/
	@echo "✅ Clean complete"

# -----------------------------
# React Frontend Commands
# -----------------------------
.PHONY: frontend-install
frontend-install:
	@echo "📦 Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-build
frontend-build:
	@echo "🏗️ Building React app..."
	cd $(FRONTEND_DIR) && npm run build

.PHONY: frontend-start
frontend-start:
	@echo "🌐 Starting React development server..."
	cd $(FRONTEND_DIR) && npm start

# -----------------------------
# Docker Commands
# -----------------------------
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# -----------------------------
# Docker Compose Commands
# -----------------------------
.PHONY: compose-up
compose-up:
	@echo "📦 Starting services with Docker Compose..."
	docker compose up -d

.PHONY: compose-down
compose-down:
	@echo "🧯 Stopping services..."
	docker compose down

# -----------------------------
# Utility Commands
# -----------------------------
.PHONY: dev
dev:
	@echo "⚡ Starting full-stack dev environment..."
	make -j2 run frontend-start

.PHONY: all
all: tidy build frontend-build
	@echo "✅ All components built successfully!"
