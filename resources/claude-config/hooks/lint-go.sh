#!/usr/bin/env bash
# lint-go.sh - Go-specific linting functions for Claude Code smart-lint
#
# This file is sourced by smart-lint.sh to provide Go linting capabilities.
# It follows the same pattern as other language-specific linters.

# ============================================================================
# GO LINTING
# ============================================================================

lint_go() {
    if [[ "${CLAUDE_HOOKS_GO_ENABLED:-true}" != "true" ]]; then
        log_debug "Go linting disabled"
        return 0
    fi
    
    log_info "Running Go linting..."
    
    # Check if Makefile exists with fmt, lint, and vet targets
    if [[ -f "Makefile" ]]; then
        local has_fmt=$(grep -E "^fmt:" Makefile 2>/dev/null || echo "")
        local has_lint=$(grep -E "^lint:" Makefile 2>/dev/null || echo "")
        local has_vet=$(grep -E "^vet:" Makefile 2>/dev/null || echo "")

        if [[ -n "$has_fmt" && -n "$has_lint" ]]; then
            log_info "Using Makefile targets"

            local fmt_output
            if ! fmt_output=$(make fmt 2>&1); then
                add_error "Go formatting failed (make fmt)"
                echo "$fmt_output" >&2
            fi

            local lint_output
            if ! lint_output=$(make lint 2>&1); then
                add_error "Go linting failed (make lint)"
                echo "$lint_output" >&2
            fi

            # Run make vet if target exists
            if [[ -n "$has_vet" ]]; then
                local vet_output
                if ! vet_output=$(make vet 2>&1); then
                    add_error "Go vet failed (make vet)"
                    echo "$vet_output" >&2
                fi
            fi
        else
            # Fallback to direct commands
            run_go_direct_lint
        fi
    else
        # No Makefile, use direct commands
        run_go_direct_lint
    fi
}

# Get changed Go files from git (including untracked files)
get_changed_go_files() {
    local changed_files=""
    
    # Get staged, working directory changes, and untracked files
    if command_exists git && git rev-parse --git-dir >/dev/null 2>&1; then
        changed_files=$(
            {
                git diff --name-only HEAD 2>/dev/null || true
                git diff --cached --name-only 2>/dev/null || true
                git ls-files --others --exclude-standard 2>/dev/null || true
            } | sort -u | grep '\.go$' 2>/dev/null | while read -r file; do
                [[ -f "$file" ]] && ! should_skip_file "$file" && echo "$file"
            done || true
        )
    fi
    
    echo "$changed_files"
}

# Run Go linting tools directly (when no Makefile targets)
run_go_direct_lint() {
    log_info "Using direct Go tools"
    
    # Get changed Go files
    local changed_go_files
    changed_go_files=$(get_changed_go_files)
    
    if [[ -z "$changed_go_files" ]]; then
        log_debug "No changed Go files to process, falling back to full project scan"
        # Fallback to original behavior for full project scan
        run_go_full_project_lint
        return
    fi
    
    log_debug "Processing changed Go files: $(echo "$changed_go_files" | tr '\n' ' ')"
    
    # Format check - only on changed files (check only, no formatting)
    local unformatted_files
    unformatted_files=$(echo "$changed_go_files" | xargs gofmt -l 2>/dev/null || true)
    
    if [[ -n "$unformatted_files" ]]; then
        add_error "Go files need formatting (run gofmt -w on changed files)"
        echo "Unformatted files:" >&2
        echo "$unformatted_files" >&2
    fi
    
    # Linting - only on changed files
    if command_exists golangci-lint; then
        local exclude_args=""
        if [[ -f ".claude-hooks-ignore" ]]; then
            # Convert ignore patterns to golangci-lint skip-files patterns
            while IFS= read -r pattern; do
                [[ -z "$pattern" || "$pattern" =~ ^[[:space:]]*# ]] && continue
                # Remove quotes and adjust pattern for golangci-lint
                pattern="${pattern//\'/}"
                pattern="${pattern//\"/}"
                exclude_args="${exclude_args} --skip-files=${pattern}"
            done < ".claude-hooks-ignore"
        fi

        # Group files by directory to avoid "named files must all be in one directory" error
        local lint_output=""
        local has_errors=false

        # Get unique directories containing changed Go files
        local directories=$(echo "$changed_go_files" | xargs -I {} dirname {} | sort -u)

        for dir in $directories; do
            # Get files in this directory
            local dir_files=$(echo "$changed_go_files" | grep "^${dir}/" | tr '\n' ' ')
            if [[ -n "$dir_files" ]]; then
                local lint_cmd="golangci-lint run --timeout=2m${exclude_args} ${dir_files}"
                log_debug "Running: $lint_cmd"
                local dir_output
                if ! dir_output=$($lint_cmd 2>&1); then
                    has_errors=true
                    lint_output="${lint_output}${dir_output}\n"
                fi
            fi
        done

        if [[ "$has_errors" == true ]]; then
            add_error "golangci-lint found issues in changed files"
            echo -e "$lint_output" >&2
        fi
    elif command_exists go; then
        # For go vet, we need to check if we can run it on specific files
        # go vet works on packages, so we need to get the packages from files
        local packages
        packages=$(echo "$changed_go_files" | xargs -I {} dirname {} | sort -u | while read -r pkg; do
            # Only include directories that contain Go files and are valid packages
            if [[ -n "$(find "$pkg" -maxdepth 1 -name "*.go" -not -path "*/vendor/*" 2>/dev/null)" ]]; then
                echo "$pkg"
            fi
        done | tr '\n' ' ')
        
        if [[ -n "$packages" ]]; then
            local vet_output
            if ! vet_output=$(go vet $packages 2>&1); then
                add_error "go vet found issues in changed packages"
                echo "$vet_output" >&2
            fi
        fi
    else
        log_error "No Go linting tools available - install golangci-lint or go"
    fi
}

# Fallback function for full project linting (original behavior)
run_go_full_project_lint() {
    log_debug "Running full project Go lint"
    
    # Format check - filter files through should_skip_file (check only, no formatting)
    local unformatted_files=$(gofmt -l . 2>/dev/null | grep -v vendor/ | while read -r file; do
        if ! should_skip_file "$file"; then
            echo "$file"
        fi
    done || true)
    
    if [[ -n "$unformatted_files" ]]; then
        add_error "Go files need formatting (run gofmt -w .)"
        echo "Unformatted files:" >&2
        echo "$unformatted_files" >&2
    fi
    
    # Linting - build exclude args from .claude-hooks-ignore
    if command_exists golangci-lint; then
        local exclude_args=""
        if [[ -f ".claude-hooks-ignore" ]]; then
            # Convert ignore patterns to golangci-lint skip-files patterns
            while IFS= read -r pattern; do
                [[ -z "$pattern" || "$pattern" =~ ^[[:space:]]*# ]] && continue
                # Remove quotes and adjust pattern for golangci-lint
                pattern="${pattern//\'/}"
                pattern="${pattern//\"/}"
                exclude_args="${exclude_args} --skip-files=${pattern}"
            done < ".claude-hooks-ignore"
        fi
        
        local lint_output
        local lint_cmd="golangci-lint run --timeout=2m${exclude_args}"
        log_debug "Running: $lint_cmd"
        if ! lint_output=$($lint_cmd 2>&1); then
            add_error "golangci-lint found issues"
            echo "$lint_output" >&2
        fi
    elif command_exists go; then
        local vet_output
        if ! vet_output=$(go vet ./... 2>&1); then
            add_error "go vet found issues"
            echo "$vet_output" >&2
        fi
    else
        log_error "No Go linting tools available - install golangci-lint or go"
    fi
}

