#!/usr/bin/env bash
# smart-test.sh - Intelligent test runner for Go projects
#
# SYNOPSIS
#   PostToolUse hook that runs tests when files are edited
#
# DESCRIPTION
#   When Claude edits Go files, this hook:
#   - Extracts directories from $CLAUDE_FILE_PATHS
#   - Runs Go tests only for affected directories
#   - Only outputs FAILED test results to reduce noise
#
# EXIT CODES
#   0 - Success (all checks passed - everything is ✅ GREEN)
#   1 - General error (missing dependencies, etc.)
#   2 - ANY issues found - ALL must be fixed
#
# CONFIGURATION
#   CLAUDE_HOOKS_TEST_ON_EDIT - Enable/disable (default: true)
#   CLAUDE_HOOKS_ENABLE_RACE - Enable race detection (default: true)
#   CLAUDE_HOOKS_TEST_VERBOSE - Show all test output (default: false)

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

# Extract directories from $CLAUDE_FILE_PATHS
get_test_directories() {
    local dirs=()

    # Check if CLAUDE_FILE_PATHS is set (from Claude Code environment)
    if [[ -n "$CLAUDE_FILE_PATHS" ]]; then
        # Parse JSON array and extract directories for Go files
        while IFS= read -r file_path; do
            # Only process Go files
            if [[ "$file_path" =~ \.go$ ]]; then
                local dir=$(dirname "$file_path")
                # Skip if directory already in list
                if [[ ! " ${dirs[@]} " =~ " ${dir} " ]]; then
                    dirs+=("$dir")
                fi
            fi
        done < <(echo "$CLAUDE_FILE_PATHS" | jq -r '.[]' 2>/dev/null)
    fi

    # If no directories found, test current directory
    if [[ ${#dirs[@]} -eq 0 ]]; then
        dirs=(".")
    fi

    printf '%s\n' "${dirs[@]}"
}

# Load configuration
load_config

# ============================================================================
# GO TEST EXECUTION
# ============================================================================

run_go_tests_for_directories() {
    echo -e "${BLUE}🧪 Running Go tests for affected directories...${NC}" >&2

    local directories=("$@")
    local test_output
    local exit_code=0
    local failed_tests=""

    # Build test command flags
    local test_flags="-json"
    if [[ "${CLAUDE_HOOKS_ENABLE_RACE:-true}" == "true" ]]; then
        test_flags="$test_flags -race"
    fi

    # Test each directory
    for dir in "${directories[@]}"; do
        echo -e "${CYAN}  Testing: $dir${NC}" >&2

        # Run go test and capture output
        if ! test_output=$(cd "$dir" && go test $test_flags ./... 2>&1); then
            exit_code=$?
        fi

        # Parse JSON output to extract only failed tests
        local failures=$(echo "$test_output" | jq -r 'select(.Action == "fail" and .Test != null) | "    ❌ \(.Package).\(.Test)"' 2>/dev/null)

        if [[ -n "$failures" ]]; then
            failed_tests+="${RED}Failed tests in $dir:${NC}\n"
            failed_tests+="$failures\n\n"
            add_error "Tests failed in directory: $dir"
        fi

        # If verbose mode, show all output
        if [[ "${CLAUDE_HOOKS_TEST_VERBOSE:-false}" == "true" ]]; then
            echo -e "\n${BLUE}Full test output for $dir:${NC}" >&2
            echo "$test_output" >&2
        fi
    done

    # Show failed tests summary
    if [[ -n "$failed_tests" ]]; then
        echo -e "\n${RED}═══ Failed Tests ═══${NC}" >&2
        echo -e "$failed_tests" >&2
        return $exit_code
    fi

    echo -e "${GREEN}✅ All Go tests passed${NC}" >&2
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
    echo -e "${CYAN}🚀 Smart Test Hook - Testing affected directories...${NC}" >&2

    # Get directories to test from CLAUDE_FILE_PATHS
    local test_dirs=()
    while IFS= read -r dir; do
        test_dirs+=("$dir")
    done < <(get_test_directories)

    # Skip if Go is not enabled
    if [[ "${CLAUDE_HOOKS_GO_ENABLED:-true}" != "true" ]]; then
        echo -e "${YELLOW}ℹ️  Go testing disabled${NC}" >&2
        return 0
    fi

    # Check if we have any Go directories to test
    if [[ ${#test_dirs[@]} -eq 0 ]]; then
        echo -e "${YELLOW}ℹ️  No Go files to test${NC}" >&2
        return 0
    fi

    # Run tests for the directories
    run_go_tests_for_directories "${test_dirs[@]}"

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