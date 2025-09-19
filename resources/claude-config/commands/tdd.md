---
allowed-tools: all
description: Test-Driven Development workflow for writing tests first
---

# TDD - Test-Driven Development

Strictly follow TDD workflow: write tests first, then implement.

## TDD Three-Step Process

### RED - Write Failing Test
1. Write test cases that define expected functionality
2. Run tests to confirm they fail

### GREEN - Minimal Implementation
1. Write minimal code to make tests pass
2. Verify all tests are passing

### REFACTOR - Improve Code
1. Improve code quality and design
2. Keep tests green

## Usage

```bash
# Start TDD workflow
/tdd [feature-name]

# Example: develop user authentication
/tdd user-auth
```

## Core Principles

- **Test First**: Tests must exist before feature code
- **Minimal Implementation**: Write only code needed to pass tests
- **Continuous Refactor**: Improve design while keeping tests green
- **Fast Cycles**: Maintain short red-green-refactor cycles

**Start TDD Workflow**: Let's build high-quality code with test-driven development!