.PHONY: build run clean test

# Binary name and directory
BINARY_NAME=vsearch
BINARY_DIR=bin

# Ensure bin directory exists
$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Build the application
build: $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) .

# Run the application
run: build
	./$(BINARY_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -rf $(BINARY_DIR)
	go clean

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download

# Build and run
all: clean build run