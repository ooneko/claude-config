---
allowed-tools: all
description: Generate unit tests for git-modified Go files
---

# 🧪 TEST - Smart Test Generation

Automatically generate unit tests for Go files modified in git.

## 🎯 Core Features

1. **Detect Git Changes**: Identify new and modified Go files
2. **Analyze Code Structure**: Extract functions, methods, and structs
3. **Generate Test Cases**: Create corresponding _test.go files
4. **Run Test Verification**: Ensure generated tests pass

## 📋 Workflow

### Step 1: Detect Changes
```bash
git status --porcelain | grep '\.go$'
```

### Step 2: Analyze Code
- Extract function signatures and return types
- Identify error handling patterns
- Analyze business logic complexity

### Step 3: Generate Tests
- Basic functionality tests
- Boundary value tests (null, zero values)
- Error path tests
- Table-driven tests (complex logic)

### Step 4: Verify Execution
- Run generated tests
- Check code coverage
- Fix test issues

## 🔨 Test Templates

### Simple Function Test
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    inputType
        expected expectedType
        wantErr  bool
    }{
        {"normal case", validInput, expectedOutput, false},
        {"edge case", edgeInput, edgeOutput, false},
        {"error case", invalidInput, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

## ⚡ Usage

```bash
# Generate tests for all git-modified files
/add-test

# Generate tests for specific file
/add-test path/to/file.go

# Generate and run tests
/add-test --run
```

## ✅ Success Criteria

- All generated tests pass
- Cover main functionality and error cases
- Code format follows Go conventions
- Consistent with existing test structure

**Start Test Generation**: Let me create high-quality unit tests for your code changes!