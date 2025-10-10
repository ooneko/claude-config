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
    
    # Check if Makefile exists with lint target
    if [[ -f "Makefile" ]]; then
        local has_lint=$(grep -E "^lint:" Makefile 2>/dev/null || echo "")

        if [[ -n "$has_lint" ]]; then
            log_info "Using Makefile lint target"

            local lint_output
            if ! lint_output=$(make lint 2>&1); then
                add_error "Go linting failed (make lint)"
                echo "$lint_output" >&2
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

# Build exclude arguments from .claude-hooks-ignore file
build_exclude_args() {
    local exclude_args=""
    if [[ -f ".claude-hooks-ignore" ]]; then
        while IFS= read -r pattern; do
            [[ -z "$pattern" || "$pattern" =~ ^[[:space:]]*# ]] && continue
            # Remove quotes and adjust pattern for golangci-lint
            pattern="${pattern//\'/}"
            pattern="${pattern//\"/}"
            exclude_args="${exclude_args} --skip-files=${pattern}"
        done < ".claude-hooks-ignore"
    fi
    echo "$exclude_args"
}

# Run golangci-lint with proper error handling
run_golangci_lint() {
    local files="$1"
    local exclude_args="$2"
    local scope="$3"

    if [[ -z "$files" ]]; then
        log_debug "No files to lint for $scope"
        return 0
    fi

    # Build command array to avoid injection
    local cmd=(golangci-lint run --timeout=2m)
    [[ -n "$exclude_args" ]] && cmd+=(${exclude_args})
    cmd+=(${files})

    log_debug "Running golangci-lint for $scope: ${cmd[*]}"

    local output
    if ! output=$("${cmd[@]}" 2>&1); then
        add_error "golangci-lint found issues in $scope"
        echo "$output" >&2
        return 1
    fi

    return 0
}

# Group files by directory to avoid "named files must all be in one directory" error
group_files_by_directory() {
    local files="$1"
    local -A dir_groups

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local dir=$(dirname "$file")
        dir_groups["$dir"]="${dir_groups["$dir"]} $file"
    done <<< "$files"

    for dir in "${!dir_groups[@]}"; do
        echo "$dir:${dir_groups["$dir"]}"
    done
}

# Run Go linting on changed files only
run_go_changed_files_lint() {
    local changed_go_files="$1"

    log_debug "Processing changed Go files: $(echo "$changed_go_files" | tr '\n' ' ')"

    local exclude_args
    exclude_args=$(build_exclude_args)

    local has_errors=false

    # Process files grouped by directory
    while IFS=':' read -r dir files; do
        [[ -z "$files" ]] && continue

        # Trim leading space
        files=$(echo "$files" | sed 's/^ *//')

        if ! run_golangci_lint "$files" "$exclude_args" "directory $dir"; then
            has_errors=true
        fi
    done < <(group_files_by_directory "$changed_go_files")

    return $((has_errors == true))
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

    if command_exists golangci-lint; then
        run_go_changed_files_lint "$changed_go_files"
    else
        log_error "No Go linting tools available - install golangci-lint"
    fi
}

# Fallback function for full project linting (original behavior)
run_go_full_project_lint() {
    log_debug "Running full project Go lint"

    if command_exists golangci-lint; then
        local exclude_args
        exclude_args=$(build_exclude_args)

        run_golangci_lint "./..." "$exclude_args" "full project"
    else
        log_error "No Go linting tools available - install golangci-lint"
    fi
}

