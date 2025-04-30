.PHONY: build test run clean migrate-up migrate-down migrate-to lint integration-test

# Build variables
BINARY_NAME=pollapp
SERVER_BINARY=server
CMD_DIR=./cmd

# Go-specific variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Default config file location
CONFIG=.env

# Default migration version for migrate-to
VERSION=0

all: build

# Build all binaries
build:
	$(GOBUILD) -o bin/$(SERVER_BINARY) $(CMD_DIR)/server

# Build and run the server
run: build
	bin/$(SERVER_BINARY)

test:
	@echo "Running only working integration tests for Poll Handler API..."
	@echo "Make sure containers are running with: docker-compose up -d"
	$(GOTEST) -v ./internal/delivery/http/pollhandler -run "TestPollEndpoints/(Create_Poll|Get_Polls|Invalid_Option_Index|Invalid_Poll_ID)"

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Migrate database up
migrate-up: build
	bin/$(SERVER_BINARY) -migrate-up

# Migrate database down
migrate-down: build
	bin/$(SERVER_BINARY) -migrate-down

# Run linter
lint:
	$(GOLINT) run

# Tidy up Go modules
tidy:
	$(GOMOD) tidy

# Initialize a new development environment
init: tidy build migrate-up
	@echo "Development environment initialized successfully!"

# Help command
help:
	@echo "Available commands:"
	@echo "  make build                  - Build all binaries"
	@echo "  make run                    - Build and run the server"
	@echo "  make test                   - Run all tests"
	@echo "  make integration-test       - Run the poll handler integration tests"
	@echo "  make integration-test-working - Run only the working integration tests"
	@echo "  make test-coverage          - Run tests with coverage"
	@echo "  make clean                  - Clean build artifacts"
	@echo "  make migrate-up             - Migrate database up"
	@echo "  make migrate-down           - Migrate database down"
	@echo "  make lint                   - Run linter"
	@echo "  make tidy                   - Tidy up Go modules"
	@echo "  make init                   - Initialize development environment"
	@echo "  make help                   - Show this help message" 