#!/usr/bin/env bash
# test-go.sh - Go-specific testing functions for Claude Code smart-test
#
# This file is sourced by smart-test.sh to provide Go testing capabilities.
# It follows the same pattern as other language-specific testers.

# ============================================================================
# GO TEST CONFIGURATION
# ============================================================================

# Initialize GO_TEST_CMD to empty to avoid unbound variable errors
GO_TEST_CMD=""

# Set up the Go test command based on available tools
setup_go_test_command() {
    local base_cmd=""
    local race_flag=""
    
    # Set up base command
    if command -v gotestsum >/dev/null 2>&1; then
        # Use gotestsum with dots format for clean output
        base_cmd="gotestsum --format dots --"
        if [[ "${CLAUDE_HOOKS_DEBUG:-0}" == "1" ]]; then
            echo "DEBUG: Found gotestsum at $(command -v gotestsum)" >&2
        fi
    else
        # Fall back to standard go test
        base_cmd="go test -v"
        if [[ "${CLAUDE_HOOKS_DEBUG:-0}" == "1" ]]; then
            echo "DEBUG: gotestsum not found, using go test" >&2
        fi
    fi
    
    # Add race detection if enabled
    if [[ "${CLAUDE_HOOKS_DEBUG:-0}" == "1" ]]; then
        echo "DEBUG: CLAUDE_HOOKS_ENABLE_RACE='${CLAUDE_HOOKS_ENABLE_RACE}'" >&2
    fi
    
    if [[ "${CLAUDE_HOOKS_ENABLE_RACE}" == "true" ]]; then
        race_flag=" -race"
        GO_TEST_CMD="$base_cmd$race_flag"
    else
        GO_TEST_CMD="$base_cmd"
    fi
    
    if [[ "${CLAUDE_HOOKS_DEBUG:-0}" == "1" ]]; then
        echo "DEBUG: GO_TEST_CMD='$GO_TEST_CMD'" >&2
    fi
}

# ============================================================================
# GO TEST RUNNERS
# ============================================================================

run_go_tests() {
    local target="$1"
    
    # Initialize test command if not already done
    if [[ -z "$GO_TEST_CMD" ]]; then
        setup_go_test_command
    fi
    
    # Determine if target is a package path or a file
    local is_package_path=false
    local dir=""
    local base=""
    local test_file=""
    
    if [[ "$target" == "./..." ]] || [[ "$target" =~ ^\.(/|$) ]] || [[ ! "$target" =~ \.go$ ]]; then
        # It's a package path (like ./..., ., ./pkg, etc.)
        is_package_path=true
        dir="$target"
    else
        # It's a Go file
        dir=$(dirname "$target")
        base=$(basename "$target" .go)
        test_file="${dir}/${base}_test.go"
        
        # Check if the file should be skipped
        if should_skip_file "$target"; then
            log_debug "Skipping tests for $target due to .claude-hooks-ignore"
            return 0
        fi
        
        # If this IS a test file, run it directly
        if [[ "$target" =~ _test\.go$ ]]; then
            echo -e "${BLUE}🧪 Running test file directly: $target${NC}" >&2
            local test_output
            if ! test_output=$($GO_TEST_CMD "$target" 2>&1); then
                echo -e "${RED}❌ Tests failed in $target${NC}" >&2
                echo -e "\n${RED}Failed test output:${NC}" >&2
                format_test_output "$test_output" "go" >&2
                return 1
            fi
            echo -e "${GREEN}✅ Tests passed in $target${NC}" >&2
            return 0
        fi
    fi
    
    # Check if we should require tests (only for specific files, not package paths)
    local require_tests=false
    if [[ "$is_package_path" == "false" ]] && ! should_skip_go_test_requirement "$target"; then
        require_tests=true
    fi
    
    # Parse test modes
    IFS=',' read -ra TEST_MODES <<< "$CLAUDE_HOOKS_TEST_MODES"
    
    local failed=0
    local tests_run=0
    local test_file_exists=false
    
    [[ -f "$test_file" ]] && test_file_exists=true
    
    for mode in "${TEST_MODES[@]}"; do
        mode=$(echo "$mode" | xargs)  # Trim whitespace
        
        case "$mode" in
            "focused")
                # Focused tests only make sense for specific files
                if [[ "$is_package_path" == "false" ]]; then
                    if [[ "$test_file_exists" == "true" ]]; then
                        echo -e "${BLUE}🧪 Running focused tests for $base...${NC}" >&2
                        tests_run=$((tests_run + 1))
                        
                        local test_output
                        if ! test_output=$($GO_TEST_CMD -run "Test.*${base}" "$dir" 2>&1); then
                            failed=1
                            echo -e "${RED}❌ Focused tests failed for $base${NC}" >&2
                            echo -e "\n${RED}Failed test output:${NC}" >&2
                            format_test_output "$test_output" "go" >&2
                            add_error "Focused tests failed for $base"
                        fi
                    elif [[ "$require_tests" == "true" ]]; then
                        echo -e "${RED}❌ Missing required test file: $test_file${NC}" >&2
                        echo -e "${YELLOW}📝 This file should have tests!${NC}" >&2
                        add_error "Missing required test file: $test_file"
                        return 2
                    fi
                fi
                ;;
            
            "package")
                local race_msg=""
                if [[ "${CLAUDE_HOOKS_ENABLE_RACE}" == "true" ]]; then
                    race_msg=" (with race detection)"
                fi
                echo -e "${BLUE}📦 Running package tests${race_msg} in $dir...${NC}" >&2
                tests_run=$((tests_run + 1))
                
                # Debug: show the actual command being run
                if [[ "${CLAUDE_HOOKS_DEBUG:-0}" == "1" ]]; then
                    echo "DEBUG: Running command: $GO_TEST_CMD -short \"$dir\"" >&2
                fi
                
                local test_output
                if ! test_output=$($GO_TEST_CMD -short "$dir" 2>&1); then
                    failed=1
                    echo -e "${RED}❌ Package tests failed in $dir${NC}" >&2
                    echo -e "\n${RED}Failed test output:${NC}" >&2
                    format_test_output "$test_output" "go" >&2
                    add_error "Package tests failed in $dir"
                fi
                ;;
            
            "all")
                # Run all tests in the project
                local race_msg=""
                if [[ "${CLAUDE_HOOKS_ENABLE_RACE}" == "true" ]]; then
                    race_msg=" (with race detection)"
                fi
                echo -e "${BLUE}🌍 Running all project tests${race_msg}...${NC}" >&2
                tests_run=$((tests_run + 1))
                
                local test_output
                if ! test_output=$($GO_TEST_CMD -short "./..." 2>&1); then
                    failed=1
                    echo -e "${RED}❌ Project tests failed${NC}" >&2
                    echo -e "\n${RED}Failed test output:${NC}" >&2
                    format_test_output "$test_output" "go" >&2
                    add_error "Project tests failed"
                fi
                ;;
                
            "integration")
                # Check if integration tests exist
                if go test -tags=integration -list . "$dir" >/dev/null 2>&1; then
                    echo -e "${BLUE}🔗 Running integration tests...${NC}" >&2
                    tests_run=$((tests_run + 1))
                    
                    local test_output
                    if ! test_output=$($GO_TEST_CMD -tags=integration "$dir" 2>&1); then
                        failed=1
                        echo -e "${RED}❌ Integration tests failed${NC}" >&2
                        echo -e "\n${RED}Failed test output:${NC}" >&2
                        format_test_output "$test_output" "go" >&2
                        add_error "Integration tests failed"
                    fi
                fi
                ;;
        esac
    done
    
    # Summary
    if [[ $tests_run -eq 0 ]]; then
        if [[ "$require_tests" == "true" && "$test_file_exists" == "false" ]]; then
            echo -e "${RED}❌ No tests found for $target (tests required)${NC}" >&2
            add_error "No tests found for $target (tests required)"
            return 2
        elif [[ "$CLAUDE_HOOKS_TEST_VERBOSE" == "true" ]]; then
            echo -e "${YELLOW}⚠️  No tests run for $target${NC}" >&2
        fi
    elif [[ $failed -eq 0 ]]; then
        log_success "All tests passed for $target"
    fi
    
    return $failed
}

# ============================================================================
# GO-SPECIFIC TEST HELPERS
# ============================================================================

# Check if we should skip test requirement for this Go file
should_skip_go_test_requirement() {
    local file="$1"
    local base=$(basename "$file")
    local dir=$(dirname "$file")
    
    # Files that typically don't have tests
    local skip_patterns=(
        "main.go"           # Entry points
        "doc.go"            # Package documentation
        "*_generated.go"    # Generated code
        "*_string.go"       # Stringer generated
        "*.pb.go"           # Protocol buffer generated
        "*.pb.gw.go"        # gRPC gateway generated
        "bindata.go"        # Embedded assets
        "migrations/*.go"   # Database migrations
    )
    
    # Check patterns
    for pattern in "${skip_patterns[@]}"; do
        if [[ "$base" == $pattern ]]; then
            return 0
        fi
    done
    
    # Skip if in specific directories
    if [[ "$dir" =~ /(vendor|testdata|examples|cmd/[^/]+|gen|generated|.gen)(/|$) ]]; then
        return 0
    fi
    
    # Skip if it's a test file itself
    if [[ "$file" =~ _test\.go$ ]]; then
        return 0
    fi
    
    return 1
}