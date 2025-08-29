# Go Makefile for sync-tools

BINARY_NAME=sync-tools
VERSION=0.2.0
BUILD_DIR=build
MAIN_PATH=cmd/sync-tools/main.go

# Default target
.PHONY: help
help: ## Show this help message
	@echo "sync-tools Go Build System"
	@echo ""
	@echo "Usage:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(BINARY_NAME) dist/
	@go clean

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

.PHONY: tidy
tidy: ## Tidy up go.mod and go.sum
	@echo "Tidying up dependencies..."
	@go mod tidy

.PHONY: build
build: deps ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-all
build-all: clean deps ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Built binaries:"
	@ls -la $(BUILD_DIR)/

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $$(go env GOPATH)/bin/..."
	@go install $(MAIN_PATH)

.PHONY: test
test: build ## Run all tests including BDD
	@echo "Running unit tests..."
	@go test -v $$(go list ./... | grep -v '/test/bdd') -short
	@echo "Running BDD tests..."
	@cd test/bdd && go test -v

.PHONY: test-with-bdd
test-with-bdd: build ## Run all tests including BDD
	@echo "Running unit tests..."
	@go test -v $$(go list ./... | grep -v '/test/bdd') -short
	@echo "Running BDD tests..."
	@cd test/bdd && go test -v

.PHONY: test-bdd
test-bdd: build ## Run BDD tests with Godog
	@echo "Running BDD tests..."
	@cd test/bdd && go test -v

.PHONY: test-all
test-all: test test-bdd ## Run all tests (unit + BDD)

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all checks including BDD tests

.PHONY: dev
dev: clean deps check build ## Full development build

.PHONY: release
release: clean check build-all ## Prepare release build

.PHONY: run
run: build ## Build and run with arguments (use: make run ARGS="--help")
	@echo "Running $(BINARY_NAME) $(ARGS)..."
	@./$(BINARY_NAME) $(ARGS)

.PHONY: demo
demo: build ## Run a demo sync operation
	@echo "Running demo sync..."
	@mkdir -p demo_source demo_dest
	@echo "Hello from Go sync-tools!" > demo_source/demo.txt
	@./$(BINARY_NAME) sync --source demo_source --dest demo_dest --dry-run -v

.PHONY: demo-interactive
demo-interactive: build ## Run interactive demo
	@echo "Starting interactive demo..."
	@mkdir -p demo_source demo_dest
	@echo "Hello from Go sync-tools!" > demo_source/demo.txt
	@./$(BINARY_NAME) sync --source demo_source --dest demo_dest --interactive

.PHONY: demo-syncfile
demo-syncfile: build ## Run SyncFile demo
	@echo "Creating demo SyncFile..."
	@mkdir -p demo_source demo_dest
	@echo "Hello from SyncFile!" > demo_source/syncfile_demo.txt
	@echo "# Demo SyncFile" > DemoSyncFile
	@echo "VAR SOURCE=demo_source" >> DemoSyncFile
	@echo "VAR DEST=demo_dest" >> DemoSyncFile
	@echo "SYNC \$${SOURCE} \$${DEST}" >> DemoSyncFile
	@echo "DRYRUN true" >> DemoSyncFile
	@./$(BINARY_NAME) syncfile DemoSyncFile -v