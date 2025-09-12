#!/usr/bin/env bash
# smarter-test.sh - Intelligent test runner with make test-changed support
#
# SYNOPSIS
#   PostToolUse hook that runs tests when files are edited
#
# DESCRIPTION
#   When Claude edits a file, this hook intelligently runs tests:
#   - First checks for 'make test-changed' target and uses it if available
#   - Falls back to original smart-test.sh logic if make test-changed not found
#   - Provides same configuration options as smart-test.sh
#
# CONFIGURATION
#   CLAUDE_HOOKS_TEST_ON_EDIT - Enable/disable (default: true)
#   CLAUDE_HOOKS_TEST_MODES - Comma-separated: focused,package,all,integration
#   CLAUDE_HOOKS_ENABLE_RACE - Enable race detection (default: true)
#   CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS - Fail if test file missing (default: false)

set -euo pipefail

# Source common helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common-helpers.sh"

# ============================================================================
# CONFIGURATION LOADING
# ============================================================================

load_config() {
    # Global defaults
    export CLAUDE_HOOKS_TEST_ON_EDIT="${CLAUDE_HOOKS_TEST_ON_EDIT:-true}"
    export CLAUDE_HOOKS_TEST_MODES="${CLAUDE_HOOKS_TEST_MODES:-focused,package}"
    export CLAUDE_HOOKS_ENABLE_RACE="${CLAUDE_HOOKS_ENABLE_RACE:-true}"
    export CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS="${CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS:-false}"
    export CLAUDE_HOOKS_TEST_VERBOSE="${CLAUDE_HOOKS_TEST_VERBOSE:-false}"
    
    # Load project config if available
    if type -t load_project_config &>/dev/null; then
        load_project_config
    fi
    
    # Quick exit if disabled
    if [[ "$CLAUDE_HOOKS_TEST_ON_EDIT" != "true" ]]; then
        echo "DEBUG: Test on edit disabled, exiting" >&2
        exit 0
    fi
}

# ============================================================================
# HOOK INPUT PARSING
# ============================================================================

# Check if we have input (hook mode) or running standalone (CLI mode)
if [ -t 0 ]; then
    # No input on stdin - CLI mode
    FILE_PATH="./..."
else
    # Read JSON input from stdin
    INPUT=$(cat)
    
    # Check if input is valid JSON
    if echo "$INPUT" | jq . >/dev/null 2>&1; then
        # Extract relevant fields
        TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty')
        TOOL_INPUT=$(echo "$INPUT" | jq -r '.tool_input // empty')
        
        # Only process edit-related tools
        if [[ ! "$TOOL_NAME" =~ ^(Edit|Write|MultiEdit)$ ]]; then
            exit 0
        fi
        
        # Extract file path(s)
        if [[ "$TOOL_NAME" == "MultiEdit" ]]; then
            # MultiEdit has a different structure
            FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // empty')
        else
            FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // empty')
        fi
        
        # Skip if no file path
        [[ -z "$FILE_PATH" ]] && exit 0
    else
        # Not valid JSON - treat as CLI mode
        FILE_PATH="./..."
    fi
fi

# Load configuration
load_config

# ============================================================================
# MAKE TEST-CHANGED DETECTION
# ============================================================================

check_make_test_changed() {
    # Check if we're in a directory with a Makefile
    if [[ ! -f "Makefile" ]]; then
        return 1
    fi
    
    # Check if Makefile contains test-changed target
    if grep -q "^\.PHONY: test-changed" Makefile && grep -q "^test-changed:" Makefile; then
        return 0
    fi
    
    return 1
}

run_make_test_changed() {
    echo -e "${BLUE}🔧 Found 'make test-changed' target, using it for optimized testing...${NC}" >&2
    
    local test_output
    local exit_code=0
    
    # Run make test-changed and capture output
    if ! test_output=$(make test-changed 2>&1); then
        exit_code=$?
        echo -e "${RED}❌ 'make test-changed' failed${NC}" >&2
        echo -e "\n${RED}Failed test output:${NC}" >&2
        echo "$test_output" >&2
        return $exit_code
    fi
    
    echo -e "${GREEN}✅ 'make test-changed' completed successfully${NC}" >&2
    if [[ "${CLAUDE_HOOKS_TEST_VERBOSE:-false}" == "true" ]]; then
        echo -e "\n${BLUE}Test output:${NC}" >&2
        echo "$test_output" >&2
    fi
    
    return 0
}

# ============================================================================
# FALLBACK TO SMART-TEST LOGIC
# ============================================================================

run_fallback_tests() {
    echo -e "${YELLOW}📋 'make test-changed' not available, using fallback testing logic...${NC}" >&2
    
    # Source the original smart-test.sh if available
    if [[ -f "${SCRIPT_DIR}/smart-test.sh" ]]; then
        # We need to be careful here - we're already running from a hook context
        # So we'll use the main function from smart-test.sh if it exists
        source "${SCRIPT_DIR}/smart-test.sh" 2>/dev/null || {
            # If sourcing fails, run basic Go tests
            run_basic_go_tests
        }
        
        # If main function is available from smart-test.sh, call it
        if type -t main &>/dev/null; then
            main
        else
            run_basic_go_tests
        fi
    else
        run_basic_go_tests
    fi
}

run_basic_go_tests() {
    echo -e "${BLUE}🧪 Running basic Go tests...${NC}" >&2
    
    local test_output
    local exit_code=0
    
    if [[ "$FILE_PATH" =~ \.go$ ]] || [[ "$FILE_PATH" == "./..." ]]; then
        # Run Go tests
        if ! test_output=$(go test -v ./... 2>&1); then
            exit_code=$?
            echo -e "${RED}❌ Go tests failed${NC}" >&2
            echo -e "\n${RED}Failed test output:${NC}" >&2
            echo "$test_output" >&2
            return $exit_code
        fi
        
        echo -e "${GREEN}✅ Go tests passed${NC}" >&2
        if [[ "${CLAUDE_HOOKS_TEST_VERBOSE:-false}" == "true" ]]; then
            echo -e "\n${BLUE}Test output:${NC}" >&2
            echo "$test_output" >&2
        fi
    else
        echo -e "${YELLOW}ℹ️  No Go files to test${NC}" >&2
    fi
    
    return 0
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

main() {
    echo -e "${CYAN}🚀 Enhanced Test Hook - Checking for optimized test commands...${NC}" >&2
    
    local failed=0
    
    # First, try to use make test-changed if available
    if check_make_test_changed; then
        run_make_test_changed || failed=1
    else
        # Fall back to original smart-test logic
        run_fallback_tests || failed=1
    fi
    
    if [[ $failed -ne 0 ]]; then
        echo -e "${RED}❌ Tests failed. Please fix the issues before continuing.${NC}" >&2
        exit 2
    else
        echo -e "${GREEN}✅ All tests passed. Continue with your task.${NC}" >&2
        exit 0
    fi
}

# Only run main if we're being executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
fi