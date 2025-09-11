#!/usr/bin/env bash
# lint-tilt.sh - Tiltfile/Starlark-specific linting logic for smart-lint.sh
#
# This file is sourced by smart-lint.sh when Tiltfiles are detected.
# It provides the lint_tilt() function and associated helpers.

# ============================================================================
# TILT/STARLARK LINTING
# ============================================================================

lint_tilt() {
    if [[ "${CLAUDE_HOOKS_TILT_ENABLED:-true}" != "true" ]]; then
        log_debug "Tilt linting disabled"
        return 0
    fi
    
    log_info "Running Tiltfile/Starlark linters..."
    
    # Check if we're in a project with Tiltfiles
    local tiltfiles=$(find . -name "Tiltfile" -not -path "./vendor/*" -not -path "./.git/*" -not -path "./node_modules/*" | head -20)
    
    if [[ -z "$tiltfiles" ]]; then
        log_debug "No Tiltfiles found"
        return 0
    fi
    
    # Filter out files that should be skipped
    local filtered_files=""
    for file in $tiltfiles; do
        if ! should_skip_file "$file"; then
            filtered_files="$filtered_files$file "
        fi
    done
    
    tiltfiles="$filtered_files"
    if [[ -z "$tiltfiles" ]]; then
        log_debug "All Tiltfiles were skipped by .claude-hooks-ignore"
        return 0
    fi
    
    # Check for Makefile with lint-tilt target
    if [[ -f "Makefile" ]]; then
        local has_lint_tilt=$(grep -E "^lint-tilt:" Makefile 2>/dev/null || echo "")
        local has_fix_tilt=$(grep -E "^fix-tilt:" Makefile 2>/dev/null || echo "")
        
        if [[ -n "$has_lint_tilt" ]]; then
            log_info "Using Makefile lint-tilt target"
            
            # First try to fix issues
            if [[ -n "$has_fix_tilt" ]]; then
                local fix_output
                if ! fix_output=$(make fix-tilt 2>&1); then
                    log_debug "make fix-tilt output: $fix_output"
                fi
            fi
            
            # Then run lint
            local lint_output
            if ! lint_output=$(make lint-tilt 2>&1); then
                add_error "Tiltfile linting failed (make lint-tilt)"
                echo "$lint_output" >&2
            fi
            return 0
        fi
    fi
    
    # Check for buildifier
    if command_exists buildifier; then
        log_info "Using buildifier for Tiltfile formatting"
        
        # First, try to auto-fix formatting issues
        local fixed_count=0
        for tiltfile in $tiltfiles; do
            log_debug "Checking $tiltfile with buildifier"
            
            # Check if file needs formatting
            if ! buildifier --mode=check --type=default "$tiltfile" &>/dev/null; then
                # Try to fix it
                if buildifier --mode=fix --lint=fix --type=default "$tiltfile" 2>/dev/null; then
                    ((fixed_count++))
                    log_debug "Fixed formatting in $tiltfile"
                fi
            fi
        done
        
        if [[ $fixed_count -gt 0 ]]; then
            log_info "Auto-fixed formatting in $fixed_count Tiltfile(s)"
        fi
        
        # Now check if any issues remain
        local has_issues=false
        for tiltfile in $tiltfiles; do
            local lint_output
            if ! lint_output=$(buildifier --mode=check --lint=warn --type=default "$tiltfile" 2>&1); then
                has_issues=true
                add_error "Buildifier found issues in $tiltfile"
                echo "$lint_output" >&2
            fi
        done
        
        if [[ "$has_issues" == "false" ]]; then
            log_debug "All Tiltfiles passed buildifier checks"
        fi
    else
        log_debug "buildifier not found, checking for basic issues"
        
        # Basic syntax check using Python (since Starlark is Python-like)
        if command_exists python || command_exists python3; then
            local python_cmd=$(command -v python3 || command -v python)
            
            for tiltfile in $tiltfiles; do
                local syntax_output
                if ! syntax_output=$($python_cmd -m py_compile "$tiltfile" 2>&1); then
                    add_error "Syntax error in $tiltfile"
                    echo "$syntax_output" >&2
                fi
            done
        fi
    fi
    
    # Check for custom linter script
    if [[ -f "scripts/lint-tiltfiles.sh" ]] && [[ -x "scripts/lint-tiltfiles.sh" ]]; then
        log_info "Running custom Tiltfile linter"
        local custom_output
        if ! custom_output=$(./scripts/lint-tiltfiles.sh 2>&1); then
            add_error "Custom Tiltfile linter found issues"
            echo "$custom_output" >&2
        fi
    fi
    
    # Check for Python-based custom linter
    if [[ -f "scripts/tiltfile-custom-lint.py" ]] && [[ -x "scripts/tiltfile-custom-lint.py" ]]; then
        log_info "Running Python-based custom Tiltfile linter"
        local custom_output
        if ! custom_output=$(./scripts/tiltfile-custom-lint.py 2>&1); then
            add_error "Custom Python linter found issues"
            echo "$custom_output" >&2
        fi
    fi
    
    return 0
}

