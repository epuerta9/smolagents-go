# smolagents-go Test Suite

This directory contains comprehensive tests for the smolagents-go library, organized by package.

## Test Structure

- `pkg/tools/tools_test.go` - Tests for tool creation, execution, and schema generation
- `pkg/models/models_test.go` - Tests for model interfaces and implementations
- `pkg/memory/memory_test.go` - Tests for memory storage and management
- `pkg/agents/agents_test.go` - Tests for agent creation and execution

## Running Tests

You can run all tests with:

```bash
make test
```

Or run tests for a specific package:

```bash
go test -v ./pkg/tools
go test -v ./pkg/models
go test -v ./pkg/memory
go test -v ./pkg/agents
```

## Test Coverage

The tests aim to cover:

1. **Tools Package**:
   - Tool creation with generics
   - Tool execution with various argument types
   - Schema generation
   - Parameter conversion

2. **Models Package**:
   - Message role constants
   - Model option application
   - API request/response handling
   - Error handling

3. **Memory Package**:
   - Memory step creation
   - Tool call recording
   - Message history management
   - Memory state transitions

4. **Agents Package**:
   - Agent creation with options
   - Message construction
   - Tool calling
   - JSON extraction

## Mock Objects

The tests use mock objects to simulate:

- Language models with pre-defined responses
- Tools with simple implementations
- API servers (using httptest) 