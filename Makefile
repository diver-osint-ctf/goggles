.PHONY: build test clean run deps

# Variables
BINARY_NAME=goggles

# Build the application
build:
	go build -o $(BINARY_NAME) cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	go mod download
	go mod tidy