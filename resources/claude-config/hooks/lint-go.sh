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

    # Check if CLAUDE_FILE_PATHS is set and contains Go files
    local target_files=""
    if [[ -n "$CLAUDE_FILE_PATHS" ]]; then
        target_files=$(get_target_go_files)
    fi

    if [[ -n "$target_files" ]]; then
        # Lint only the files specified in CLAUDE_FILE_PATHS
        log_info "Running golangci-lint on specified files"
        run_go_target_files_lint "$target_files"
    elif command_exists git && git rev-parse --git-dir >/dev/null 2>&1; then
        # Use golangci-lint with new-from-rev to only check changed files
        if command_exists golangci-lint; then
            log_info "Running golangci-lint on changes since HEAD~1"

            local lint_output
            if ! lint_output=$(golangci-lint run --new-from-rev=HEAD~1 --timeout=2m 2>&1); then
                add_error "golangci-lint found issues in changed files"
                echo "$lint_output" >&2
            fi
        else
            log_error "golangci-lint not found - please install it"
            add_error "golangci-lint not installed"
        fi
    else
        # Not a git repository, fall back to full project scan
        log_info "Not in a git repository, running full project scan"
        run_go_full_project_lint
    fi
}

# Extract Go files from CLAUDE_FILE_PATHS environment variable
get_target_go_files() {
    local go_files=""

    # Check if CLAUDE_FILE_PATHS is set (from Claude Code environment)
    if [[ -n "$CLAUDE_FILE_PATHS" ]]; then
        # Parse JSON array and extract Go files
        while IFS= read -r file_path; do
            # Only process Go files that exist
            if [[ "$file_path" =~ \.go$ ]] && [[ -f "$file_path" ]]; then
                # Skip if file should be skipped
                if ! should_skip_file "$file_path"; then
                    go_files="${go_files}${file_path} "
                fi
            fi
        done < <(echo "$CLAUDE_FILE_PATHS" | jq -r '.[]' 2>/dev/null)
    fi

    # Trim trailing space and return
    echo "${go_files% }"
}

# Run golangci-lint on specific target files from CLAUDE_FILE_PATHS
run_go_target_files_lint() {
    local files="$1"

    log_debug "Processing target Go files: $(echo "$files" | tr '\n' ' ')"

    if [[ -z "$files" ]]; then
        log_debug "No target Go files to process"
        return 0
    fi

    if ! command_exists golangci-lint; then
        log_error "golangci-lint not found - please install it"
        add_error "golangci-lint not installed"
        return 1
    fi

    local exclude_args
    exclude_args=$(build_exclude_args)

    local has_errors=false

    # Process files grouped by directory to avoid golangci-lint directory restriction
    while IFS=':' read -r dir file_list; do
        [[ -z "$file_list" ]] && continue

        # Trim leading space
        file_list=$(echo "$file_list" | sed 's/^ *//')

        if ! run_golangci_lint "$file_list" "$exclude_args" "target files in $dir"; then
            has_errors=true
        fi
    done < <(group_files_by_directory "$files")

    return $((has_errors == true))
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
    local temp_file=$(mktemp)

    # Group files by directory using temporary file
    # Handle both space-separated and newline-separated files
    echo "$files" | tr ' ' '\n' | while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local dir=$(dirname "$file")
        echo "$dir $file" >> "$temp_file"
    done

    # Output grouped files
    if [[ -f "$temp_file" ]]; then
        sort "$temp_file" | awk '
        BEGIN { prev_dir = ""; files = "" }
        {
            if ($1 != prev_dir) {
                if (prev_dir != "") print prev_dir ":" files;
                prev_dir = $1;
                files = $2;
            } else {
                files = files " " $2;
            }
        }
        END {
            if (prev_dir != "") print prev_dir ":" files;
        }'
        rm -f "$temp_file"
    fi
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

