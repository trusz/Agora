.PHONY: build run clean dev test

# Build variables
BINARY_NAME=agora
MAIN_FILE=main.go
BUILD_DIR=bin

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting $(BINARY_NAME)..."
	$(GORUN) $(MAIN_FILE)

# Dev mode: run with automatic restart on file changes
# Requires air to be installed: go install github.com/cosmtrek/air@latest
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Error: air is not installed. Install it with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# Clean build files
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) ./...

# Initialize project structure (if needed)
init:
	@echo "Initializing project structure..."
	@mkdir -p src/static/css
	@mkdir -p src/static/js
	@mkdir -p src/templates
	@touch src/static/css/style.css
	@touch src/static/js/main.js
	@echo "Project structure initialized"

# Help command
help:
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run the application"
	@echo "  make dev      - Run with live reload (requires air)"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make init     - Initialize project structure"

# Default target
default: run