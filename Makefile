.PHONY: all test fmt clean

all: fmt test

# Format all Go files
fmt:
	go fmt ./pkg/...

# Run all tests
test:
	go test -v ./pkg/tools
	go test -v ./pkg/tests

# Clean build artifacts
clean:
	go clean

# Build examples
build:
	mkdir -p bin
	go build -o bin/multiple_tools examples/multiple_tools.go
	go build -o bin/multiple_tools_generic examples/multiple_tools_generic.go
	go build -o bin/rag examples/rag_example/main.go

# Print help
help:
	@echo "Available targets:"
	@echo "  make all           - Format code and run tests"
	@echo "  make fmt           - Format code"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make build         - Build examples"
	@echo "  make help          - Print this help" 