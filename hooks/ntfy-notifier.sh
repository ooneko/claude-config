#!/bin/zsh
# ntfy-notifier.sh - Send notifications to ntfy service for Claude Code events
#
# SYNOPSIS
#   ntfy-notifier.sh <event_type>
#
# DESCRIPTION
#   Sends push notifications via ntfy service when Claude Code events occur.
#   Supports notification and stop events. Automatically detects terminal
#   context and includes it in the notification for better identification.
#
# ARGUMENTS
#   event_type    Either "notification" or "stop"
#
# CONFIGURATION
#   Requires ~/.config/claude-code-ntfy/config.yaml with:
#     ntfy_topic: your-topic-name
#     ntfy_server: https://ntfy.sh (optional, defaults to public server)
#
# ENVIRONMENT
#   CLAUDE_HOOK_PAYLOAD   JSON payload from Claude Code (for notifications)
#   CLAUDE_HOOKS_NTFY_ENABLED   Set to "false" to disable notifications
#
# TERMINAL DETECTION
#   Attempts to detect terminal context from:
#   - tmux window name
#   - macOS Terminal window title
#   - X11 window title (Linux)
#
# EXAMPLES
#   # Send notification
#   ./ntfy-notifier.sh notification
#
#   # Send stop notification
#   ./ntfy-notifier.sh stop
#
# ERROR HANDLING
#   - Validates configuration file exists
#   - Retries failed notifications
#   - Rate limits to prevent spam

set -euo pipefail

# Get the event type from the first argument
EVENT_TYPE="${1:-notification}"

# Check if notifications are enabled (allow easy disable)
if [[ "${CLAUDE_HOOKS_NTFY_ENABLED:-true}" != "true" ]]; then
    exit 0
fi

# Configuration file location
CONFIG_FILE="${CLAUDE_HOOKS_NTFY_CONFIG:-$HOME/.config/claude-code-ntfy/config.yaml}"

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Warning: Ntfy config not found at $CONFIG_FILE" >&2
    echo "Create it with:" >&2
    echo "  ntfy_topic: your-topic-name" >&2
    echo "  ntfy_server: https://ntfy.sh" >&2
    exit 0  # Exit gracefully to not block Claude
fi

# Check if yq is available
if ! command -v yq >/dev/null 2>&1; then
    echo "Warning: yq not found, cannot parse ntfy config" >&2
    exit 0
fi

# Extract configuration with error handling
NTFY_TOPIC=$(yq '.ntfy_topic' "$CONFIG_FILE" 2>/dev/null | grep -v '^null$' || echo "")
NTFY_SERVER=$(yq '.ntfy_server // "https://ntfy.sh"' "$CONFIG_FILE" 2>/dev/null || echo "https://ntfy.sh")

# Validate required configuration
if [[ -z "$NTFY_TOPIC" ]]; then
    echo "Warning: ntfy_topic not configured in $CONFIG_FILE" >&2
    exit 0
fi

# Rate limiting - prevent notification spam
RATE_LIMIT_FILE="/tmp/.claude-ntfy-rate-limit"
if [[ -f "$RATE_LIMIT_FILE" ]]; then
    LAST_NOTIFICATION=$(cat "$RATE_LIMIT_FILE" 2>/dev/null || echo "0")
    CURRENT_TIME=$(date +%s)
    TIME_DIFF=$((CURRENT_TIME - LAST_NOTIFICATION))
    
    # Limit to one notification per 2 seconds
    if [[ $TIME_DIFF -lt 2 ]]; then
        exit 0
    fi
fi
echo "$(date +%s)" > "$RATE_LIMIT_FILE"

# Get context information
CWD=$(pwd)
CWD_BASENAME=$(basename "$CWD")

# Function to clean terminal title
clean_terminal_title() {
    local title="$1"
    # Remove Claude icons and control characters
    echo "$title" | sed -E 's/[✅🤖⚡✨🔮💫☁️🌟🚀🎯🔍🛡️📝🧠🖨️🔐📤⏳❌⚠️]//g' | sed 's/[[:cntrl:]]//g' | xargs
}

# Get terminal title with improved detection
get_terminal_title() {
    local title=""
    
    if [[ "${TERM_PROGRAM:-}" == "tmux" ]] && command -v tmux >/dev/null 2>&1; then
        # In tmux, we can get the pane's environment variables
        # The hook runs in the same pane as claude, so we can get the current pane's info
        # Check if we're in a tmux session
        if [[ -n "${TMUX:-}" ]]; then
            # Get the current pane's window name
            local window_name=$(tmux display-message -p '#W' 2>/dev/null || echo "")
            local pane_title=$(tmux display-message -p '#{pane_title}' 2>/dev/null || echo "")
            
            if [[ -n "$window_name" ]]; then
                title="$window_name"
                [[ -n "$pane_title" && "$pane_title" != "$window_name" ]] && title="$title - $pane_title"
            fi
        else
            # Not in a tmux session, just get the shell's tty
            title="tty: $(tty 2>/dev/null | xargs basename)"
        fi
    elif [[ "${TERM_PROGRAM:-}" == "kitty" ]] && command -v kitty >/dev/null 2>&1; then
        # Kitty: Get window title using kitty remote control
        # This requires allow_remote_control to be enabled in kitty.conf
        title=$(kitty @ ls | jq -r '.[] | select(.is_focused) | .tabs[] | select(.is_focused) | .title' 2>/dev/null || echo "")
        if [[ -z "$title" ]]; then
            # Fallback: get from environment if remote control is disabled
            title="${KITTY_WINDOW_TITLE:-Kitty}"
        fi
    elif [[ "$(uname)" == "Darwin" ]] && command -v osascript >/dev/null 2>&1; then
        # macOS: Get Terminal or iTerm2 window title (only if they're the active terminal)
        if [[ "${TERM_PROGRAM:-}" == "iTerm.app" ]]; then
            title=$(osascript -e 'tell application "iTerm2" to name of current window' 2>/dev/null || echo "")
        elif [[ "${TERM_PROGRAM:-}" == "Apple_Terminal" ]]; then
            # Only try to get Terminal.app title if we're actually running in Terminal.app
            title=$(osascript -e 'tell application "Terminal" to name of front window' 2>/dev/null || echo "")
        fi
    elif [[ -n "${DISPLAY:-}" ]] && command -v xprop >/dev/null 2>&1; then
        # Linux with X11: Get window title
        local window_id=$(xprop -root _NET_ACTIVE_WINDOW 2>/dev/null | awk '{print $5}')
        if [[ -n "$window_id" && "$window_id" != "0x0" ]]; then
            title=$(xprop -id "$window_id" WM_NAME 2>/dev/null | cut -d'"' -f2 || echo "")
        fi
    elif [[ -n "${WAYLAND_DISPLAY:-}" ]] && command -v swaymsg >/dev/null 2>&1; then
        # Wayland with Sway: Get focused window title
        title=$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | .name' 2>/dev/null || echo "")
    fi
    
    clean_terminal_title "$title"
}

TERM_TITLE=$(get_terminal_title)

# Build context string
CONTEXT="Claude Code: $CWD_BASENAME"
if [[ -n "$TERM_TITLE" ]]; then
    CONTEXT="$CONTEXT - $TERM_TITLE"
fi

# Function to send notification with retry
send_notification() {
    local title="$1"
    local message="$2"
    local tags="$3"
    local priority="${4:-default}"
    
    local max_retries=2
    local retry_count=0
    
    while [[ $retry_count -lt $max_retries ]]; do
        if curl -s \
            --max-time 5 \
            -H "Title: $title" \
            -H "Tags: $tags" \
            -H "Priority: $priority" \
            -d "$message" \
            "$NTFY_SERVER/$NTFY_TOPIC" >/dev/null 2>&1; then
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        [[ $retry_count -lt $max_retries ]] && sleep 1
    done
    
    echo "Warning: Failed to send notification after $max_retries attempts" >&2
    return 1
}

# Prepare notification based on event type
case "$EVENT_TYPE" in
    "notification")
        # Claude sent a notification - parse the payload if available
        if [[ -n "${CLAUDE_HOOK_PAYLOAD:-}" ]]; then
            # Extract message from JSON payload
            # 注意：如果 jq 解析失败或没有 message 字段，使用默认消息
            MESSAGE=$(echo "$CLAUDE_HOOK_PAYLOAD" | jq -r '.message // empty' 2>/dev/null || echo "")
            
            # 如果消息为空，使用默认消息
            if [[ -z "$MESSAGE" ]]; then
                MESSAGE="Claude notification"
            fi
            
            # Check for error or warning indicators
            PRIORITY="default"
            if echo "$MESSAGE" | grep -qiE '(error|fail|problem|issue)'; then
                PRIORITY="high"
            elif echo "$MESSAGE" | grep -qiE '(warn|warning|attention)'; then
                PRIORITY="default"
            fi
        else
            MESSAGE="Claude notification"
            PRIORITY="default"
        fi
        
        TITLE="$CONTEXT"
        TAGS="claude-code,notification"
        ;;
    
    "stop")
        TITLE="$CONTEXT"
        MESSAGE="Claude finished responding"
        TAGS="claude-code,stop,checkmark"
        PRIORITY="low"
        ;;
    
    *)
        echo "Error: Unknown event type: $EVENT_TYPE" >&2
        echo "Usage: $0 {notification|stop}" >&2
        exit 1
        ;;
esac

# Send notification
send_notification "$TITLE" "$MESSAGE" "$TAGS" "$PRIORITY"

# Clean up old rate limit files (older than 1 hour)
find /tmp -name ".claude-ntfy-rate-limit" -mmin +60 -delete 2>/dev/null || true
