#!/bin/bash

VERSION="v0.14.0"

ADDITIONAL_FLAGS="--dangerously-skip-permissions --output-format json"

NOTES_FILE="SHARED_TASK_NOTES.md"

PROMPT_JQ_INSTALL="Please install jq for JSON parsing"

PROMPT_COMMIT_MESSAGE="Please review all uncommitted changes in the git repository (both modified and new files). Write a commit message with: (1) a short one-line summary, (2) two newlines, (3) then a detailed explanation. Do not include any footers or metadata like 'Generated with Claude Code' or 'Co-Authored-By'. Feel free to look at the last few commits to get a sense of the commit message style for consistency. First run 'git add .' to stage all changes including new untracked files, then commit using 'git commit -m \"your message\"' (don't push, just commit, no need to ask for confirmation)."

PROMPT_WORKFLOW_CONTEXT="## CONTINUOUS WORKFLOW CONTEXT

This is part of a continuous development loop where work happens incrementally across multiple iterations. You might run once, then a human developer might make changes, then you run again, and so on. This could happen daily or on any schedule.

**Important**: You don't need to complete the entire goal in one iteration. Just make meaningful progress on one thing, then leave clear notes for the next iteration (human or AI). Think of it as a relay race where you're passing the baton.

**Project Completion Signal**: If you determine that not just your current task but the ENTIRE project goal is fully complete (nothing more to be done on the overall goal), only include the exact phrase \"COMPLETION_SIGNAL_PLACEHOLDER\" in your response. Only use this when absolutely certain that the whole project is finished, not just your individual task. We will stop working on this project when multiple developers independently determine that the project is complete.

## PRIMARY GOAL"

PROMPT_NOTES_UPDATE_EXISTING="Update the \`$NOTES_FILE\` file with relevant context for the next iteration. Add new notes and remove outdated information to keep it current and useful."

PROMPT_NOTES_CREATE_NEW="Create a \`$NOTES_FILE\` file with relevant context and instructions for the next iteration."

PROMPT_NOTES_GUIDELINES="

This file helps coordinate work across iterations (both human and AI developers). It should:

- Contain relevant context and instructions for the next iteration
- Stay concise and actionable (like a notes file, not a detailed report)
- Help the next developer understand what to do next

The file should NOT include:
- Lists of completed work or full reports
- Information that can be discovered by running tests/coverage
- Unnecessary details"

PROMPT=""
MAX_RUNS=""
MAX_COST=""
MAX_DURATION=""
ENABLE_COMMITS=true
GIT_BRANCH_PREFIX="continuous-claude/"
WORKTREE_NAME=""
WORKTREE_BASE_DIR="../continuous-claude-worktrees"
CLEANUP_WORKTREE=false
LIST_WORKTREES=false
DRY_RUN=false
COMPLETION_SIGNAL="CONTINUOUS_CLAUDE_PROJECT_COMPLETE"
COMPLETION_THRESHOLD=3
ERROR_LOG=""
error_count=0
extra_iterations=0
successful_iterations=0
total_cost=0
completion_signal_count=0
i=1
EXTRA_CLAUDE_FLAGS=()
start_time=""

parse_duration() {
    # Parse a duration string like "2h", "30m", "1h30m", "90s" to seconds
    # Returns: number of seconds, or empty string on error
    local duration_str="$1"
    
    # Remove all whitespace
    duration_str=$(echo "$duration_str" | tr -d '[:space:]')
    
    if [ -z "$duration_str" ]; then
        return 1
    fi
    
    local total_seconds=0
    local remaining="$duration_str"
    
    # Parse hours (e.g., "2h" or "2H")
    if [[ "$remaining" =~ ([0-9]+)[hH] ]]; then
        local hours="${BASH_REMATCH[1]}"
        total_seconds=$((total_seconds + hours * 3600))
        remaining="${remaining/${BASH_REMATCH[0]}/}"
    fi
    
    # Parse minutes (e.g., "30m" or "30M")
    if [[ "$remaining" =~ ([0-9]+)[mM] ]]; then
        local minutes="${BASH_REMATCH[1]}"
        total_seconds=$((total_seconds + minutes * 60))
        remaining="${remaining/${BASH_REMATCH[0]}/}"
    fi
    
    # Parse seconds (e.g., "45s" or "45S")
    if [[ "$remaining" =~ ([0-9]+)[sS] ]]; then
        local seconds="${BASH_REMATCH[1]}"
        total_seconds=$((total_seconds + seconds))
        remaining="${remaining/${BASH_REMATCH[0]}/}"
    fi
    
    # Check if anything unparsed remains (invalid format)
    if [ -n "$remaining" ]; then
        return 1
    fi
    
    # Must have parsed at least something
    if [ $total_seconds -eq 0 ]; then
        return 1
    fi
    
    echo "$total_seconds"
    return 0
}

format_duration() {
    # Format seconds into a human-readable duration string
    local seconds="$1"
    
    if [ -z "$seconds" ] || [ "$seconds" -eq 0 ]; then
        echo "0s"
        return
    fi
    
    local hours=$((seconds / 3600))
    local minutes=$(((seconds % 3600) / 60))
    local secs=$((seconds % 60))
    
    local result=""
    if [ $hours -gt 0 ]; then
        result="${hours}h"
    fi
    if [ $minutes -gt 0 ]; then
        result="${result}${minutes}m"
    fi
    if [ $secs -gt 0 ] || [ -z "$result" ]; then
        result="${result}${secs}s"
    fi
    
    echo "$result"
}

show_help() {
    cat << EOF
Continuous Claude - Run Claude Code iteratively with local commits

USAGE:
    continuous-claude -p "prompt" (-m max-runs | --max-cost max-cost | --max-duration duration) [options]

REQUIRED OPTIONS:
    -p, --prompt <text>           The prompt/goal for Claude Code to work on
    -m, --max-runs <number>       Maximum number of successful iterations (use 0 for unlimited with --max-cost or --max-duration)
    --max-cost <dollars>          Maximum cost in USD to spend (alternative to --max-runs)
    --max-duration <duration>     Maximum duration to run (e.g., "2h", "30m", "1h30m") (alternative to --max-runs)

OPTIONAL FLAGS:
    -h, --help                    Show this help message
    -v, --version                 Show version information
    --disable-commits             Disable automatic commits
    --git-branch-prefix <prefix>  Branch prefix for iterations (default: "continuous-claude/")
    --notes-file <file>           Shared notes file for iteration context (default: "SHARED_TASK_NOTES.md")
    --worktree <name>             Run in a git worktree for parallel execution (creates if needed)
    --worktree-base-dir <path>    Base directory for worktrees (default: "../continuous-claude-worktrees")
    --cleanup-worktree            Remove worktree after completion
    --list-worktrees              List all active git worktrees and exit
    --dry-run                     Simulate execution without making changes
    --completion-signal <phrase>  Phrase that agents output when project is complete (default: "CONTINUOUS_CLAUDE_PROJECT_COMPLETE")
    --completion-threshold <num>  Number of consecutive signals to stop early (default: 3)

EXAMPLES:
    # Run 5 iterations to fix bugs
    continuous-claude -p "Fix all linter errors" -m 5

    # Run with cost limit
    continuous-claude -p "Add tests" --max-cost 10.00

    # Run for a maximum duration (time-boxed)
    continuous-claude -p "Add documentation" --max-duration 2h

    # Run for 30 minutes
    continuous-claude -p "Refactor module" --max-duration 30m

    # Run without commits (testing mode)
    continuous-claude -p "Refactor code" -m 3 --disable-commits

    # Combine duration and cost limits (whichever comes first)
    continuous-claude -p "Add tests" --max-duration 1h30m --max-cost 5.00

    # Run in a worktree for parallel execution
    continuous-claude -p "Add unit tests" -m 5 --worktree instance-1

    # Run multiple instances in parallel (in different terminals)
    continuous-claude -p "Task A" -m 5 --worktree task-a
    continuous-claude -p "Task B" -m 5 --worktree task-b

    # List all active worktrees
    continuous-claude --list-worktrees

    # Clean up worktree after completion
    continuous-claude -p "Quick fix" -m 1 --worktree temp --cleanup-worktree

    # Use completion signal to stop early when project is done
    continuous-claude -p "Add unit tests to all files" -m 50 --completion-threshold 3

REQUIREMENTS:
    - Claude Code CLI (https://claude.ai/code)
    - Git repository (unless --disable-commits is used)

For more information, visit: https://github.com/AnandChowdhary/continuous-claude
EOF
}

show_version() {
    echo "continuous-claude version $VERSION"
}

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                show_version
                exit 0
                ;;
            -p|--prompt)
                PROMPT="$2"
                shift 2
                ;;
            -m|--max-runs)
                MAX_RUNS="$2"
                shift 2
                ;;
            --max-cost)
                MAX_COST="$2"
                shift 2
                ;;
            --max-duration)
                MAX_DURATION="$2"
                shift 2
                ;;
            --git-branch-prefix)
                GIT_BRANCH_PREFIX="$2"
                shift 2
                ;;
            --disable-commits)
                ENABLE_COMMITS=false
                shift
                ;;
            --notes-file)
                NOTES_FILE="$2"
                shift 2
                ;;
            --worktree)
                WORKTREE_NAME="$2"
                shift 2
                ;;
            --worktree-base-dir)
                WORKTREE_BASE_DIR="$2"
                shift 2
                ;;
            --cleanup-worktree)
                CLEANUP_WORKTREE=true
                shift
                ;;
            --list-worktrees)
                LIST_WORKTREES=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --completion-signal)
                COMPLETION_SIGNAL="$2"
                shift 2
                ;;
            --completion-threshold)
                COMPLETION_THRESHOLD="$2"
                shift 2
                ;;
            *)
                # Collect unknown flags to forward to claude
                EXTRA_CLAUDE_FLAGS+=("$1")
                shift
                ;;
        esac
    done
}

validate_arguments() {
    if [ -z "$PROMPT" ]; then
        echo "❌ Error: Prompt is required. Use -p to provide a prompt." >&2
        echo "Run '$0 --help' for usage information." >&2
        exit 1
    fi

    if [ -z "$MAX_RUNS" ] && [ -z "$MAX_COST" ] && [ -z "$MAX_DURATION" ]; then
        echo "❌ Error: Either --max-runs, --max-cost, or --max-duration is required." >&2
        echo "Run '$0 --help' for usage information." >&2
        exit 1
    fi

    if [ -n "$MAX_RUNS" ] && ! [[ "$MAX_RUNS" =~ ^[0-9]+$ ]]; then
        echo "❌ Error: --max-runs must be a non-negative integer" >&2
        exit 1
    fi

    if [ -n "$MAX_COST" ]; then
        if ! [[ "$MAX_COST" =~ ^[0-9]+\.?[0-9]*$ ]] || [ "$(awk "BEGIN {print ($MAX_COST <= 0)}")" = "1" ]; then
            echo "❌ Error: --max-cost must be a positive number" >&2
            exit 1
        fi
    fi

    if [ -n "$MAX_DURATION" ]; then
        local duration_seconds
        if ! duration_seconds=$(parse_duration "$MAX_DURATION"); then
            echo "❌ Error: --max-duration must be a valid duration (e.g., '2h', '30m', '1h30m', '90s')" >&2
            exit 1
        fi
        # Store parsed duration in seconds back to MAX_DURATION for later use
        MAX_DURATION="$duration_seconds"
    fi

    if [ -n "$COMPLETION_THRESHOLD" ]; then
        if ! [[ "$COMPLETION_THRESHOLD" =~ ^[0-9]+$ ]] || [ "$COMPLETION_THRESHOLD" -lt 1 ]; then
            echo "❌ Error: --completion-threshold must be a positive integer" >&2
            exit 1
        fi
    fi
}

validate_requirements() {
    if ! command -v claude &> /dev/null; then
        echo "❌ Error: Claude Code is not installed: https://claude.ai/code" >&2
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        echo "⚠️ jq is required for JSON parsing but is not installed. Asking Claude Code to install it..." >&2
        claude -p "$PROMPT_JQ_INSTALL" --allowedTools "Bash,Read"
        if ! command -v jq &> /dev/null; then
            echo "❌ Error: jq is still not installed after Claude Code attempt." >&2
            exit 1
        fi
    fi
}

create_iteration_branch() {
    local iteration_display="$1"
    local iteration_num="$2"
    
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo ""
        return 0
    fi

    local current_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
    
    if [[ "$current_branch" == ${GIT_BRANCH_PREFIX}* ]]; then
        echo "⚠️  $iteration_display Already on iteration branch: $current_branch" >&2
        git checkout main >/dev/null 2>&1 || return 1
        current_branch="main"
    fi
    
    local date_str=$(date +%Y-%m-%d)
    
    local random_hash
    if command -v openssl >/dev/null 2>&1; then
        random_hash=$(openssl rand -hex 4)
    elif [ -r /dev/urandom ]; then
        random_hash=$(LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c 8)
    else
        random_hash=$(printf "%x" $(($(date +%s) % 100000000)))$(printf "%x" $$)
        random_hash=${random_hash:0:8}
    fi
    
    local branch_name="${GIT_BRANCH_PREFIX}iteration-${iteration_num}/${date_str}-${random_hash}"
    
    echo "🌿 $iteration_display Creating branch: $branch_name" >&2
    
    if [ "$DRY_RUN" = "true" ]; then
        echo "   (DRY RUN) Would create branch $branch_name" >&2
        echo "$branch_name"
        return 0
    fi
    
    if ! git checkout -b "$branch_name" >/dev/null 2>&1; then
        echo "⚠️  $iteration_display Failed to create branch" >&2
        echo ""
        return 1
    fi
    
    echo "$branch_name"
    return 0
}

continuous_claude_commit() {
    local iteration_display="$1"
    local branch_name="$2"
    local main_branch="$3"

    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        return 0
    fi

    # Check for any changes
    local has_changes=false
    if ! git diff --quiet || ! git diff --cached --quiet; then
        has_changes=true
    fi

    if [ -z "$(git ls-files --others --exclude-standard)" ]; then
        : # no untracked files
    else
        has_changes=true
    fi

    if [ "$has_changes" = "false" ]; then
        echo "🫙 $iteration_display No changes detected, staying on branch: $branch_name" >&2
        return 0
    fi

    if [ "$DRY_RUN" = "true" ]; then
        echo "💬 $iteration_display (DRY RUN) Would commit changes..." >&2
        echo "📦 $iteration_display (DRY RUN) Changes would be committed on branch: $branch_name" >&2
        return 0
    fi

    echo "💬 $iteration_display Committing changes..." >&2

    if ! claude -p "$PROMPT_COMMIT_MESSAGE" --allowedTools "Bash(git)" --dangerously-skip-permissions >/dev/null 2>&1; then
        echo "⚠️  $iteration_display Failed to commit changes" >&2
        return 1
    fi

    # Verify all changes were committed
    if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
        echo "⚠️  $iteration_display Commit command ran but changes still present" >&2
        return 1
    fi

    echo "📦 $iteration_display Changes committed on branch: $branch_name" >&2

    local commit_message=$(git log -1 --format="%B" "$branch_name")
    local commit_title=$(echo "$commit_message" | head -n 1)

    echo "✅ $iteration_display Commit created: $commit_title" >&2
    echo "🌿 $iteration_display Staying on branch: $branch_name" >&2

    return 0
}

list_worktrees() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo "❌ Error: Not in a git repository" >&2
        exit 1
    fi
    
    echo "📋 Active Git Worktrees:"
    echo ""
    
    if ! git worktree list 2>/dev/null; then
        echo "❌ Error: Failed to list worktrees" >&2
        exit 1
    fi
    
    exit 0
}

setup_worktree() {
    if [ -z "$WORKTREE_NAME" ]; then
        # No worktree specified, work in current directory
        return 0
    fi
    
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo "❌ Error: Not in a git repository. Worktrees require a git repository." >&2
        exit 1
    fi
    
    # Get the main repo directory
    local main_repo_dir=$(git rev-parse --show-toplevel)
    local worktree_path="${WORKTREE_BASE_DIR}/${WORKTREE_NAME}"
    
    # Make worktree path absolute if it's relative
    if [[ "$worktree_path" != /* ]]; then
        worktree_path="${main_repo_dir}/${worktree_path}"
    fi
    
    # Get current branch (usually main or master)
    local current_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
    
    # Check if worktree already exists
    if [ -d "$worktree_path" ]; then
        echo "🌿 Worktree '$WORKTREE_NAME' already exists at: $worktree_path" >&2
        echo "📂 Switching to worktree directory..." >&2
        
        if ! cd "$worktree_path"; then
            echo "❌ Error: Failed to change to worktree directory: $worktree_path" >&2
            exit 1
        fi
    else
        echo "🌿 Creating new worktree '$WORKTREE_NAME' at: $worktree_path" >&2
        
        # Create base directory if it doesn't exist
        local base_dir=$(dirname "$worktree_path")
        if [ ! -d "$base_dir" ]; then
            mkdir -p "$base_dir" || {
                echo "❌ Error: Failed to create worktree base directory: $base_dir" >&2
                exit 1
            }
        fi
        
        # Create the worktree
        if ! git worktree add "$worktree_path" "$current_branch" 2>&1; then
            echo "❌ Error: Failed to create worktree" >&2
            exit 1
        fi
        
        echo "📂 Switching to worktree directory..." >&2
        if ! cd "$worktree_path"; then
            echo "❌ Error: Failed to change to worktree directory: $worktree_path" >&2
            exit 1
        fi
    fi
    
    echo "✅ Worktree '$WORKTREE_NAME' ready at: $worktree_path" >&2
    return 0
}

cleanup_worktree() {
    if [ -z "$WORKTREE_NAME" ] || [ "$CLEANUP_WORKTREE" = "false" ]; then
        return 0
    fi
    
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        return 0
    fi
    
    local worktree_path="${WORKTREE_BASE_DIR}/${WORKTREE_NAME}"
    
    # Get the main repo directory to make path absolute
    local main_repo_dir=$(git rev-parse --show-toplevel 2>/dev/null)
    if [ -n "$main_repo_dir" ]; then
        if [[ "$worktree_path" != /* ]]; then
            worktree_path="${main_repo_dir}/${worktree_path}"
        fi
    fi
    
    echo "" >&2
    echo "🗑️  Cleaning up worktree '$WORKTREE_NAME'..." >&2
    
    # Try to find the main repo
    local current_dir=$(pwd)
    local git_common_dir=$(git rev-parse --git-common-dir 2>/dev/null)
    
    if [ -n "$git_common_dir" ]; then
        local main_repo=$(dirname "$git_common_dir")
        if [ -d "$main_repo" ]; then
            cd "$main_repo" 2>/dev/null || true
        fi
    fi
    
    # Remove the worktree
    if git worktree remove "$worktree_path" --force 2>/dev/null; then
        echo "✅ Worktree removed successfully" >&2
    else
        echo "⚠️  Warning: Failed to remove worktree (may need manual cleanup)" >&2
        echo "   You can manually remove it with: git worktree remove $worktree_path --force" >&2
    fi
}

get_iteration_display() {
    local iteration_num=$1
    local max_runs=$2
    local extra_iters=$3
    
    if [ $max_runs -eq 0 ]; then
        echo "($iteration_num)"
    else
        local total=$((max_runs + extra_iters))
        echo "($iteration_num/$total)"
    fi
}

run_claude_iteration() {
    local prompt="$1"
    local flags="$2"
    local error_log="$3"

    if [ "$DRY_RUN" = "true" ]; then
        echo "🤖 (DRY RUN) Would run Claude Code with prompt: $prompt" >&2
        echo "📝 (DRY RUN) Output: This is a simulated response from Claude Code." > "$error_log"
        return 0
    fi

    # Run claude and capture both stdout and stderr
    # Use temporary files for both to ensure synchronous capture
    local temp_stdout=$(mktemp)
    local temp_stderr=$(mktemp)
    local exit_code=0
    
    # Capture both stdout and stderr to temp files
    claude -p "$prompt" $flags "${EXTRA_CLAUDE_FLAGS[@]}" >"$temp_stdout" 2>"$temp_stderr" || exit_code=$?
    
    # Output stdout (JSON result) so caller can capture it
    if [ -f "$temp_stdout" ] && [ -s "$temp_stdout" ]; then
        cat "$temp_stdout"
    fi
    
    # Display stderr to terminal and save to error log
    if [ -f "$temp_stderr" ] && [ -s "$temp_stderr" ]; then
        cat "$temp_stderr" >&2
        cat "$temp_stderr" > "$error_log"
    fi
    
    # If claude failed, check for error info in both stderr and stdout (JSON)
    if [ $exit_code -ne 0 ]; then
        # If stderr is empty, try to extract error from JSON stdout
        if [ ! -s "$error_log" ] && [ -f "$temp_stdout" ] && [ -s "$temp_stdout" ]; then
            # Check if stdout contains JSON with error info
            local json_error=$(cat "$temp_stdout" | jq -r 'if type == "array" then .[-1] else . end | if .is_error == true then .result // .error // "Unknown error" else empty end' 2>/dev/null || echo "")
            if [ -n "$json_error" ]; then
                echo "$json_error" > "$error_log"
                echo "$json_error" >&2
            fi
        fi
        
        # If still no error info, provide fallback message
        if [ ! -s "$error_log" ]; then
            {
                echo "Claude Code exited with code $exit_code but produced no error output"
                echo ""
                echo "This usually means:"
                echo "  - Claude Code crashed or failed to start"
                echo "  - An authentication or permission issue occurred"
                echo "  - The command arguments are invalid"
                echo ""
                echo "Try running this command directly to see the full error:"
                echo "  claude -p \"$prompt\" $flags ${EXTRA_CLAUDE_FLAGS[*]}"
            } >> "$error_log"
        fi
        
        # Cleanup temp files after error handling
        rm -f "$temp_stdout" "$temp_stderr"
        return $exit_code
    fi
    
    # Cleanup temp files on success
    rm -f "$temp_stdout" "$temp_stderr"
    
    return 0
}

parse_claude_result() {
    local result="$1"
    
    if ! echo "$result" | jq -e . >/dev/null 2>&1; then
        echo "invalid_json"
        return 1
    fi
    
    local is_error=$(echo "$result" | jq -r 'if type == "array" then .[-1].is_error // false else .is_error // false end')
    if [ "$is_error" = "true" ]; then
        echo "claude_error"
        return 1
    fi
    
    echo "success"
    return 0
}

handle_iteration_error() {
    local iteration_display="$1"
    local error_type="$2"
    local error_output="$3"
    
    error_count=$((error_count + 1))
    extra_iterations=$((extra_iterations + 1))
    
    case "$error_type" in
        "exit_code")
            echo "" >&2
            echo "❌ $iteration_display Error occurred ($error_count consecutive errors):" >&2
            echo "" >&2
            if [ -f "$ERROR_LOG" ] && [ -s "$ERROR_LOG" ]; then
                echo "Error details:" >&2
                cat "$ERROR_LOG" >&2
            else
                echo "No error details captured in log file" >&2
                echo "Error log path: $ERROR_LOG" >&2
            fi
            echo "" >&2
            ;;
        "invalid_json")
            echo "" >&2
            echo "❌ $iteration_display Error: Invalid JSON response ($error_count consecutive errors):" >&2
            echo "" >&2
            echo "$error_output" >&2
            echo "" >&2
            ;;
        "claude_error")
            echo "" >&2
            echo "❌ $iteration_display Error in Claude Code response ($error_count consecutive errors):" >&2
            echo "" >&2
            echo "$error_output" | jq -r 'if type == "array" then .[-1].result // .[-1] else .result // . end' >&2
            echo "" >&2
            ;;
    esac
    
    if [ $error_count -ge 3 ]; then
        echo "❌ Fatal: 3 consecutive errors occurred. Exiting." >&2
        exit 1
    fi
    
    return 1
}

handle_iteration_success() {
    local iteration_display="$1"
    local result="$2"
    local branch_name="$3"
    local main_branch="$4"
    
    echo "📝 $iteration_display Output:" >&2
    local result_text=$(echo "$result" | jq -r 'if type == "array" then .[-1].result // empty else .result // empty end')
    if [ -n "$result_text" ]; then
        echo "$result_text"
    else
        echo "(no output)" >&2
    fi

    # Check for completion signal in the output
    if [ -n "$result_text" ] && [[ "$result_text" == *"$COMPLETION_SIGNAL"* ]]; then
        completion_signal_count=$((completion_signal_count + 1))
        echo "" >&2
        echo "🎯 $iteration_display Completion signal detected ($completion_signal_count/$COMPLETION_THRESHOLD)" >&2
    else
        if [ $completion_signal_count -gt 0 ]; then
            echo "" >&2
            echo "🔄 $iteration_display Completion signal not found, resetting counter" >&2
        fi
        completion_signal_count=0
    fi

    local cost=$(echo "$result" | jq -r 'if type == "array" then .[-1].total_cost_usd // empty else .total_cost_usd // empty end')
    if [ -n "$cost" ]; then
        echo "" >&2
        printf "💰 $iteration_display Cost: \$%.3f\n" "$cost" >&2
        total_cost=$(awk "BEGIN {printf \"%.3f\", $total_cost + $cost}")
    fi

    echo "✅ $iteration_display Work completed" >&2
    if [ "$ENABLE_COMMITS" = "true" ]; then
        if ! continuous_claude_commit "$iteration_display" "$branch_name" "$main_branch"; then
            error_count=$((error_count + 1))
            extra_iterations=$((extra_iterations + 1))
            echo "❌ $iteration_display PR merge queue failed ($error_count consecutive errors)" >&2
            if [ $error_count -ge 3 ]; then
                echo "❌ Fatal: 3 consecutive errors occurred. Exiting." >&2
                exit 1
            fi
            return 1
        fi
    else
        echo "⏭️  $iteration_display Skipping commits (--disable-commits flag set)" >&2
        # Clean up branch if commits are disabled
        if [ -n "$branch_name" ] && git rev-parse --git-dir > /dev/null 2>&1; then
            git checkout "$main_branch" >/dev/null 2>&1
            git branch -D "$branch_name" >/dev/null 2>&1 || true
        fi
    fi
    
    error_count=0
    if [ $extra_iterations -gt 0 ]; then
        extra_iterations=$((extra_iterations - 1))
    fi
    successful_iterations=$((successful_iterations + 1))
    return 0
}

execute_single_iteration() {
    local iteration_num=$1
    
    local iteration_display=$(get_iteration_display $iteration_num $MAX_RUNS $extra_iterations)
    echo "🔄 $iteration_display Starting iteration..." >&2

    # Get current branch and create iteration branch
    local main_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
    local branch_name=""
    
    if [ "$ENABLE_COMMITS" = "true" ]; then
        branch_name=$(create_iteration_branch "$iteration_display" "$iteration_num")
        if [ $? -ne 0 ] || [ -z "$branch_name" ]; then
            if git rev-parse --git-dir > /dev/null 2>&1; then
                echo "❌ $iteration_display Failed to create branch" >&2
                handle_iteration_error "$iteration_display" "exit_code" ""
                return 1
            fi
            # Not a git repo, continue without branch
            branch_name=""
        fi
    fi

    local enhanced_prompt="${PROMPT_WORKFLOW_CONTEXT//COMPLETION_SIGNAL_PLACEHOLDER/$COMPLETION_SIGNAL}

$PROMPT

"

    if [ -f "$NOTES_FILE" ]; then
        local notes_content
        notes_content=$(cat "$NOTES_FILE")
        enhanced_prompt+="## CONTEXT FROM PREVIOUS ITERATION

The following is from $NOTES_FILE, maintained by previous iterations to provide context:

$notes_content

"
    fi

    enhanced_prompt+="## ITERATION NOTES

"
    
    if [ -f "$NOTES_FILE" ]; then
        enhanced_prompt+="$PROMPT_NOTES_UPDATE_EXISTING"
    else
        enhanced_prompt+="$PROMPT_NOTES_CREATE_NEW"
    fi
    
    enhanced_prompt+="$PROMPT_NOTES_GUIDELINES"

    echo "🤖 $iteration_display Running Claude Code..." >&2
    
    local result
    local claude_exit_code=0
    result=$(run_claude_iteration "$enhanced_prompt" "$ADDITIONAL_FLAGS" "$ERROR_LOG") || claude_exit_code=$?
    
    if [ $claude_exit_code -ne 0 ]; then
        echo "" >&2
        echo "⚠️  Claude Code command failed with exit code: $claude_exit_code" >&2
        # Clean up branch on error
        if [ -n "$branch_name" ] && git rev-parse --git-dir > /dev/null 2>&1; then
            git checkout "$main_branch" >/dev/null 2>&1
            git branch -D "$branch_name" >/dev/null 2>&1 || true
        fi
        handle_iteration_error "$iteration_display" "exit_code" ""
        return 1
    fi
    
    local parse_result=$(parse_claude_result "$result")
    if [ "$?" != "0" ]; then
        # Clean up branch on error
        if [ -n "$branch_name" ] && git rev-parse --git-dir > /dev/null 2>&1; then
            git checkout "$main_branch" >/dev/null 2>&1
            git branch -D "$branch_name" >/dev/null 2>&1 || true
        fi
        handle_iteration_error "$iteration_display" "$parse_result" "$result"
        return 1
    fi
    
    handle_iteration_success "$iteration_display" "$result" "$branch_name" "$main_branch"
    return 0
}

main_loop() {
    # Initialize start time if MAX_DURATION is set
    if [ -n "$MAX_DURATION" ]; then
        start_time=$(date +%s)
    fi
    
    while true; do
        # Check if we should continue based on limits
        local should_continue=false
        
        # Continue if MAX_RUNS is not set or not reached
        if [ -z "$MAX_RUNS" ] || [ "$MAX_RUNS" -eq 0 ] || [ $successful_iterations -lt $MAX_RUNS ]; then
            should_continue=true
        fi
        
        # Stop if MAX_COST is set and reached/exceeded
        if [ -n "$MAX_COST" ] && [ "$(awk "BEGIN {print ($total_cost >= $MAX_COST)}")" = "1" ]; then
            should_continue=false
        fi
        
        # Stop if MAX_DURATION is set and reached/exceeded
        if [ -n "$MAX_DURATION" ] && [ -n "$start_time" ]; then
            local current_time=$(date +%s)
            local elapsed_time=$((current_time - start_time))
            if [ $elapsed_time -ge $MAX_DURATION ]; then
                echo "" >&2
                echo "⏱️  Maximum duration reached ($(format_duration $elapsed_time))" >&2
                should_continue=false
            fi
        fi
        
        # If both limits are set and both are reached, stop
        if [ -n "$MAX_RUNS" ] && [ "$MAX_RUNS" -ne 0 ] && [ $successful_iterations -ge $MAX_RUNS ]; then
            should_continue=false
        fi
        
        # Stop if completion signal threshold reached
        if [ $completion_signal_count -ge $COMPLETION_THRESHOLD ]; then
            echo "" >&2
            echo "🎉 Project completion signal detected $completion_signal_count times consecutively!" >&2
            should_continue=false
        fi
        
        if [ "$should_continue" = "false" ]; then
            break
        fi
        
        execute_single_iteration $i
        
        sleep 1
        i=$((i + 1))
    done
}

show_completion_summary() {
    # Calculate elapsed time if start_time was set
    local elapsed_msg=""
    if [ -n "$start_time" ]; then
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))
        elapsed_msg=" (elapsed: $(format_duration $elapsed_time))"
    fi
    
    # Show completion signal message if that's why we stopped
    if [ $completion_signal_count -ge $COMPLETION_THRESHOLD ]; then
        if [ -n "$total_cost" ] && [ "$(awk "BEGIN {print ($total_cost > 0)}")" = "1" ]; then
            printf "✨ Project completed! Detected completion signal %d times in a row. Total cost: \$%.3f%s\n" "$completion_signal_count" "$total_cost" "$elapsed_msg"
        else
            printf "✨ Project completed! Detected completion signal %d times in a row.%s\n" "$completion_signal_count" "$elapsed_msg"
        fi
    elif [ -n "$MAX_RUNS" ] && [ $MAX_RUNS -ne 0 ] || [ -n "$MAX_COST" ] || [ -n "$MAX_DURATION" ]; then
        if [ -n "$total_cost" ] && [ "$(awk "BEGIN {print ($total_cost > 0)}")" = "1" ]; then
            printf "🎉 Done with total cost: \$%.3f%s\n" "$total_cost" "$elapsed_msg"
        else 
            printf "🎉 Done%s\n" "$elapsed_msg"
        fi
    fi
}

main() {
    parse_arguments "$@"
    validate_arguments
    validate_requirements

    # Handle --list-worktrees flag
    if [ "$LIST_WORKTREES" = "true" ]; then
        list_worktrees
    fi
    
    # Setup worktree if specified
    setup_worktree
    
    ERROR_LOG=$(mktemp)
    trap "rm -f $ERROR_LOG; cleanup_worktree" EXIT
    
    main_loop
    show_completion_summary
    
    # Cleanup worktree if requested
    cleanup_worktree
}

if [ -z "$TESTING" ]; then
    main "$@"
fi
