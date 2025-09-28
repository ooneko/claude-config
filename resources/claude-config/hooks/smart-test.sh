#!/usr/bin/env bash
# smarter-test.sh - Intelligent test runner with make test support
#
# SYNOPSIS
#   PostToolUse hook that runs tests when files are edited
#
# DESCRIPTION
#   When Claude edits a file, this hook intelligently runs tests:
#   - First checks for 'make test' target and uses it if available
#   - Falls back to original smart-test.sh logic if make test not found
#   - Provides same configuration options as smart-test.sh
#
# EXIT CODES
#   0 - Success (all checks passed - everything is ✅ GREEN)
#   1 - General error (missing dependencies, etc.)
#   2 - ANY issues found - ALL must be fixed
#
# CONFIGURATION
#   CLAUDE_HOOKS_TEST_ON_EDIT - Enable/disable (default: true)
#   CLAUDE_HOOKS_TEST_MODES - Comma-separated: focused,package,all,integration
#   CLAUDE_HOOKS_ENABLE_RACE - Enable race detection (default: true)
#   CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS - Fail if test file missing (default: false)

# Don't use set -e - we need to control exit codes carefully like smart-lint.sh
set +e

# Source common helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common-helpers.sh"

# ============================================================================
# ERROR TRACKING (extends common-helpers.sh)
# ============================================================================

# Use the CLAUDE_HOOKS_ERRORS array from common-helpers.sh but with a different name for summary
declare -a CLAUDE_HOOKS_SUMMARY=()

# Override add_error to also add to summary
add_error() {
    local message="$1"
    CLAUDE_HOOKS_ERROR_COUNT+=1
    CLAUDE_HOOKS_ERRORS+=("${RED}❌${NC} $message")
    CLAUDE_HOOKS_SUMMARY+=("${RED}❌${NC} $message")
}

print_summary() {
    if [[ $CLAUDE_HOOKS_ERROR_COUNT -gt 0 ]]; then
        # Only show failures when there are errors
        echo -e "\n${BLUE}═══ Test Summary ═══${NC}" >&2
        for item in "${CLAUDE_HOOKS_SUMMARY[@]}"; do
            echo -e "$item" >&2
        done

        echo -e "\n${RED}Found $CLAUDE_HOOKS_ERROR_COUNT test issue(s) that MUST be fixed!${NC}" >&2
        echo -e "${RED}════════════════════════════════════════════${NC}" >&2
        echo -e "${RED}❌ ALL TEST ISSUES ARE BLOCKING ❌${NC}" >&2
        echo -e "${RED}════════════════════════════════════════════${NC}" >&2
        echo -e "${RED}Fix EVERYTHING above until all tests are ✅ GREEN${NC}" >&2
    fi
}

# ============================================================================
# CONFIGURATION LOADING
# ============================================================================

load_config() {
    # Global defaults
    export CLAUDE_HOOKS_ENABLED="${CLAUDE_HOOKS_ENABLED:-true}"
    export CLAUDE_HOOKS_FAIL_FAST="${CLAUDE_HOOKS_FAIL_FAST:-false}"
    export CLAUDE_HOOKS_SHOW_TIMING="${CLAUDE_HOOKS_SHOW_TIMING:-false}"

    export CLAUDE_HOOKS_TEST_ON_EDIT="${CLAUDE_HOOKS_TEST_ON_EDIT:-true}"
    export CLAUDE_HOOKS_TEST_MODES="${CLAUDE_HOOKS_TEST_MODES:-focused,package}"
    export CLAUDE_HOOKS_ENABLE_RACE="${CLAUDE_HOOKS_ENABLE_RACE:-true}"
    export CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS="${CLAUDE_HOOKS_FAIL_ON_MISSING_TESTS:-false}"
    export CLAUDE_HOOKS_TEST_VERBOSE="${CLAUDE_HOOKS_TEST_VERBOSE:-false}"

    # Language enables
    export CLAUDE_HOOKS_GO_ENABLED="${CLAUDE_HOOKS_GO_ENABLED:-true}"
    export CLAUDE_HOOKS_PYTHON_ENABLED="${CLAUDE_HOOKS_PYTHON_ENABLED:-true}"
    export CLAUDE_HOOKS_JS_ENABLED="${CLAUDE_HOOKS_JS_ENABLED:-true}"
    export CLAUDE_HOOKS_RUST_ENABLED="${CLAUDE_HOOKS_RUST_ENABLED:-true}"
    export CLAUDE_HOOKS_NIX_ENABLED="${CLAUDE_HOOKS_NIX_ENABLED:-true}"

    # Project-specific overrides
    if [[ -f ".claude-hooks-config.sh" ]]; then
        source ".claude-hooks-config.sh" || {
            log_error "Failed to load .claude-hooks-config.sh"
            exit 2
        }
    fi

    # Quick exit if hooks are disabled
    if [[ "$CLAUDE_HOOKS_ENABLED" != "true" ]]; then
        log_info "Claude hooks are disabled"
        exit 0
    fi

    # Quick exit if test on edit is disabled
    if [[ "$CLAUDE_HOOKS_TEST_ON_EDIT" != "true" ]]; then
        log_info "Test on edit disabled, exiting"
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
# MAKE TEST DETECTION
# ============================================================================

check_make_test() {
    # Check if we're in a directory with a Makefile
    if [[ ! -f "Makefile" ]]; then
        return 1
    fi

    # Check if Makefile contains test target
    if grep -q "^\.PHONY: test" Makefile && grep -q "^test:" Makefile; then
        return 0
    fi

    return 1
}

run_make_test() {
    echo -e "${BLUE}🔧 Found 'make test' target, using it for testing...${NC}" >&2

    local test_output
    local exit_code=0

    # Run make test and capture output
    if ! test_output=$(make test 2>&1); then
        exit_code=$?
        add_error "'make test' failed with exit code $exit_code"
        echo -e "\n${RED}Failed test output:${NC}" >&2
        # Always show full output when make test fails - it may contain important debugging info
        echo "$test_output" >&2
        return $exit_code
    fi

    echo -e "${GREEN}✅ 'make test' completed successfully${NC}" >&2
    if [[ "${CLAUDE_HOOKS_TEST_VERBOSE:-false}" == "true" ]]; then
        echo -e "\n${BLUE}Full test output:${NC}" >&2
        echo "$test_output" >&2
    fi

    return 0
}

run_basic_go_tests() {
    echo -e "${BLUE}🧪 Running basic Go tests...${NC}" >&2

    local test_output
    local exit_code=0

    if [[ "$FILE_PATH" =~ \.go$ ]] || [[ "$FILE_PATH" == "./..." ]]; then
        # Run Go tests
        if ! test_output=$(go test -v ./... 2>&1); then
            exit_code=$?
            add_error "Go tests failed with exit code $exit_code"
            echo -e "\n${RED}Failed test output:${NC}" >&2
            # Always show full output when Go tests fail - it may contain important debugging info
            echo "$test_output" >&2
            return $exit_code
        fi

        echo -e "${GREEN}✅ Go tests passed${NC}" >&2
        if [[ "${CLAUDE_HOOKS_TEST_VERBOSE:-false}" == "true" ]]; then
            echo -e "\n${BLUE}Full test output:${NC}" >&2
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

# Parse command line options
FAST_MODE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            export CLAUDE_HOOKS_DEBUG=1
            shift
            ;;
        --fast)
            FAST_MODE=true
            shift
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 2
            ;;
    esac
done

# Print header
echo "" >&2
echo "🧪 Test Check - Running tests..." >&2
echo "────────────────────────────────────────────" >&2

# Load configuration
load_config

# Start timing
START_TIME=$(time_start)

main() {
    echo -e "${CYAN}🚀 Enhanced Test Hook - Checking for optimized test commands...${NC}" >&2

    # First, try to use make test if available
    if check_make_test; then
        run_make_test
    else
        # Fall back to basic Go tests
        echo -e "${YELLOW}📋 'make test' not available, running basic Go tests...${NC}" >&2
        run_basic_go_tests
    fi

    # Show timing if enabled
    time_end "$START_TIME"

    # Print summary
    print_summary

    # Return exit code - any issues mean failure
    if [[ $CLAUDE_HOOKS_ERROR_COUNT -gt 0 ]]; then
        return 2
    else
        return 0
    fi
}

# Run main function
main
exit_code=$?

# Final message and exit - always exit with 2 so Claude sees the continuation message
if [[ $exit_code -eq 2 ]]; then
    echo -e "\n${RED}🛑 FAILED - Fix all test issues above! 🛑${NC}" >&2
    echo -e "${YELLOW}📋 NEXT STEPS:${NC}" >&2
    echo -e "${YELLOW}  1. Fix the test issues listed above${NC}" >&2
    echo -e "${YELLOW}  2. Verify the fix by running the test command again${NC}" >&2
    echo -e "${YELLOW}  3. Continue with your original task${NC}" >&2
    exit 2
else
    echo -e "\n${GREEN}✅ All tests passed. Continue with your task.${NC}" >&2
    exit 2
fi