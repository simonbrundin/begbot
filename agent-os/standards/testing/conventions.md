## Testing Conventions

### Test File Location

Place test files in the same package as the code being tested:

```
internal/services/bot.go
internal/services/bot_test.go
```

### Test Function Naming

Use descriptive names that explain what is being tested:

```go
func TestDatabaseValuationMethod_Name(t *testing.T)
func TestValuationInput_Struct(t *testing.T)
func TestValuationCompiler_compileWeightedAverage(t *testing.T)
```

Format: `Test<StructOrMethod>_<WhatIsBeingTested>`

### Unit Testing Patterns

#### Struct Method Testing

Test struct methods directly without external dependencies:

```go
func TestDatabaseValuationMethod_Priority(t *testing.T) {
    method := &DatabaseValuationMethod{}
    priority := method.Priority()

    if priority != 1 {
        t.Errorf("Expected priority 1, got %d", priority)
    }
}
```

#### Struct Initialization Testing

Test that structs can be created with expected values:

```go
func TestValuationInput_Struct(t *testing.T) {
    input := ValuationInput{
        Type:        "Test",
        Value:       1500,
        Confidence:  0.75,
    }

    if input.Type != "Test" {
        t.Errorf("Expected type 'Test', got '%s'", input.Type)
    }
}
```

### Test Organization

1. **Happy path first** - Test the normal case
2. **Edge cases** - Test boundary conditions
3. **Error cases** - Test error handling

### Avoiding External Dependencies

- Mock external services (HTTP, databases)
- Use in-memory solutions when possible
- Test business logic in isolation
