.PHONY: build clean install demo test help

# Application name
APP_NAME = hacktober

# Build directory
BUILD_DIR = build

# Default target
all: build

# Build the application
build:
	@echo "🔨 Building $(APP_NAME)..."
	@go build -o $(APP_NAME) ./cmd/hacktober
	@echo "✅ Build complete: ./$(APP_NAME)"

# Build with verbose output
build-verbose:
	@echo "🔨 Building $(APP_NAME) with verbose output..."
	@go build -v -o $(APP_NAME) ./cmd/hacktober

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

# Run the demo
demo:
	@./demo.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f $(APP_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

# Create example config
config:
	@echo "⚙️  Creating example configuration..."
	@cp .hacktober-config.example.json ~/.hacktober-config.json
	@echo "✅ Created ~/.hacktober-config.json"
	@echo "📝 Please edit it with your GitHub token"

# Run tests (when we have them)
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Show help
help:
	@echo "🎃 Hacktoberfest Repository & Issue Explorer"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-verbose - Build with verbose output"
	@echo "  deps          - Install dependencies"
	@echo "  demo          - Show demo information"
	@echo "  config        - Create example config file"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Usage:"
	@echo "  make build                    # Build the app"
	@echo "  GITHUB_TOKEN=xxx ./hacktober  # Run with token"
	@echo ""