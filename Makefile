.PHONY: dev build setup clean test lint help

# Default target
.DEFAULT_GOAL := help

# Development
dev: ## Start development servers (Vite + Go)
	@./bin/dev

# Production build
build: ## Build for production
	@./bin/build

# Setup
setup: ## Set up development environment
	@./bin/setup

# Clean
clean: ## Clean build artifacts and node_modules
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf static/dist
	@rm -rf ui/node_modules
	@rm -f base-app
	@echo "✨ Clean complete!"

# Frontend only
ui-dev: ## Start only Vite dev server  
	@echo "🚀 Starting Vite dev server..."
	@if command -v bun >/dev/null 2>&1; then \
		cd ui && bun run dev; \
	else \
		cd ui && npm run dev; \
	fi

ui-build: ## Build only frontend assets
	@echo "📦 Building frontend assets..."
	@if command -v bun >/dev/null 2>&1; then \
		cd ui && bun run build; \
	else \
		cd ui && npm run build; \
	fi

ui-install: ## Install frontend dependencies
	@echo "📥 Installing frontend dependencies..."
	@if command -v bun >/dev/null 2>&1; then \
		echo "Using Bun..."; \
		cd ui && bun install; \
	else \
		echo "Using npm..."; \
		cd ui && npm install; \
	fi

# Go only
go-dev: ## Start only Go server
	@echo "🚀 Starting Go server..."
	@export ENVIRONMENT=development && go run main.go

go-build: ## Build only Go binary
	@echo "📦 Building Go binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o base-app main.go

go-test: ## Run Go tests
	@echo "🧪 Running Go tests..."
	@go test ./...

# Linting and formatting
lint: ## Lint and format code
	@echo "🔍 Linting Go code..."
	@go fmt ./...
	@go vet ./...
	@echo "🔍 Linting frontend code..."
	@cd ui && npm run lint 2>/dev/null || echo "No lint script found"

# shadcn/ui components
add-component: ## Add shadcn/ui component (usage: make add-component COMPONENT=button)
	@if [ -z "$(COMPONENT)" ]; then \
		echo "❌ Please specify a component: make add-component COMPONENT=button"; \
		exit 1; \
	fi
	@echo "📦 Adding shadcn/ui component: $(COMPONENT)"
	@if command -v bun >/dev/null 2>&1; then \
		cd ui && bunx shadcn-ui@latest add $(COMPONENT); \
	else \
		cd ui && npx shadcn-ui@latest add $(COMPONENT); \
	fi

# Database (if applicable)
db-migrate: ## Run database migrations
	@echo "🗄️ Running database migrations..."
	@go run main.go migrate 2>/dev/null || echo "No migrations found"

# Docker (optional)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t baseui-app .

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 baseui-app

# Troubleshooting
fix-registry: ## Fix npm/bun registry issues
	@./bin/fix-registry

troubleshoot: ## Show troubleshooting information
	@echo "🔧 BaseUI Troubleshooting"
	@echo ""
	@echo "Common issues and solutions:"
	@echo "• Registry errors: make fix-registry"
	@echo "• Port conflicts: kill processes on 8080/5173"
	@echo "• Cache issues: make clean && make setup"
	@echo "• Bun install: curl -fsSL https://bun.sh/install | bash"
	@echo ""

# Help
help: ## Show this help message
	@echo "🎯 BaseUI Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "📖 Examples:"
	@echo "  make dev                    # Start development environment"
	@echo "  make build                  # Build for production"
	@echo "  make add-component COMPONENT=button  # Add shadcn/ui button"