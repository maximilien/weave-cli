# Test Mocks

This directory contains mock implementations for testing Weave CLI components.

## Available Mocks

### MockVectorDBClient

Mock implementation of `vectordb.VectorDBClient` interface.

**Usage:**

```go
import (
    "testing"
    "github.com/maximilien/weave-cli/tests/mocks"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestYourFunction(t *testing.T) {
    // Create mock
    mockClient := new(mocks.MockVectorDBClient)

    // Set expectations
    mockClient.On("Health", mock.Anything).Return(nil)
    mockClient.On("CreateCollection", mock.Anything, "test", mock.Anything).
        Return(nil)

    // Use in your code
    err := mockClient.Health(context.Background())
    assert.NoError(t, err)

    // Assert all expectations were met
    mockClient.AssertExpectations(t)
}
```

### MockLLMClient

Mock implementation of `llm.Client` interface for testing embedding
generation and LLM interactions.

**Usage:**

```go
func TestEmbeddingGeneration(t *testing.T) {
    mockLLM := new(mocks.MockLLMClient)

    // Mock embedding generation
    embedding := []float64{0.1, 0.2, 0.3}
    mockLLM.On("GenerateEmbedding", mock.Anything, "test text", "").
        Return(embedding, nil)

    result, err := mockLLM.GenerateEmbedding(context.Background(), "test text", "")
    assert.NoError(t, err)
    assert.Equal(t, embedding, result)

    mockLLM.AssertExpectations(t)
}
```

## Mock Patterns

### Basic Mock Setup

```go
mock := new(mocks.MockVectorDBClient)
mock.On("MethodName", arg1, arg2).Return(returnValue, error)
```

### Argument Matchers

```go
// Any argument
mock.On("Method", mock.Anything).Return(nil)

// Specific type
mock.On("Method", mock.AnythingOfType("string")).Return(nil)

// Custom matcher
mock.On("Method", mock.MatchedBy(func(arg string) bool {
    return len(arg) > 0
})).Return(nil)
```

### Return Values

```go
// Simple return
mock.On("Method").Return(value, nil)

// Error return
mock.On("Method").Return(nil, errors.New("test error"))

// Multiple calls with different returns
mock.On("Method").Return(value1, nil).Once()
mock.On("Method").Return(value2, nil).Once()
```

### Verification

```go
// Assert specific method was called
mock.AssertCalled(t, "MethodName", arg1, arg2)

// Assert method was not called
mock.AssertNotCalled(t, "MethodName")

// Assert all expectations
mock.AssertExpectations(t)
```

## Testing Best Practices

1. **Isolate Units**: Use mocks to test components in isolation
2. **Test Error Paths**: Mock error returns to test error handling
3. **Verify Interactions**: Always call `AssertExpectations()` to verify
   mocks were used correctly
4. **Clean Setup**: Use test helpers to reduce mock setup boilerplate
5. **Meaningful Assertions**: Assert on important behaviors, not
   implementation details

## Adding New Mocks

When adding new mocks:

1. Create a new file: `tests/mocks/your_interface.go`
2. Implement all interface methods using `mock.Mock`
3. Use `mock.Called()` and return appropriate values
4. Document usage in this README
5. Add examples to help other developers

## Dependencies

This package requires:

- `github.com/stretchr/testify/mock` - Mock framework
- `github.com/stretchr/testify/assert` - Assertions

Install with:

```bash
go get github.com/stretchr/testify
```
