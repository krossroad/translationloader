# Variables
BINARY_NAME=sync
BIN_DIR=bin
MAIN_PATH=cmd/sync/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean run test test-unit test-integration test-coverage lint tidy help docker-up docker-down generate-mocks generate

all: lint test build

build: ## Build the binary
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

clean: ## Remove binary and coverage files
	rm -rf $(BIN_DIR)
	rm -f coverage.out

run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)

test: ## Run all tests
	$(GOTEST) -v -tags=integration ./...

test-unit: ## Run unit tests (excluding integration)
	$(GOTEST) -v ./...

test-integration: ## Run integration tests
	$(GOTEST) -v -tags=integration ./test/integration/...

test-coverage: ## Run tests and generate coverage report
	$(GOTEST) -v -tags=integration -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

test-coverage-html: test-coverage ## Open coverage report in browser
	$(GOCMD) tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run ./...

tidy: ## Tidy go modules
	$(GOMOD) tidy

generate-mocks: ## Generate mocks for all interfaces in internal/core/ports
	@echo "Generating mocks..."
	go run github.com/vektra/mockery/v2@v2.53.6 --dir internal/core/ports --all --output test/mocks --outpkg mocks --case underscore

generate: ## Run go generate
	$(GOCMD) generate ./...

docker-up: ## Start local infrastructure (Postgres)
	docker-compose up -d

docker-down: ## Stop local infrastructure
	docker-compose down

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
