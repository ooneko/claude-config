#!/bin/bash

# Claude 配置管理工具
# 管理 Claude settings.json 中的代理和 hooks 设置

CLAUDE_DIR="$HOME/.claude"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

# 代理设置
PROXY_HOST="http://127.0.0.1:7890"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查 jq 是否安装
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ 错误：需要安装 jq 工具${NC}"
    echo "   请运行：brew install jq"
    exit 1
fi

# 确保设置文件存在
if [ ! -f "$SETTINGS_FILE" ]; then
    echo -e "${RED}❌ 错误：找不到 $SETTINGS_FILE${NC}"
    echo "   请先创建 Claude 设置文件"
    exit 1
fi

# 备份设置文件
backup_settings() {
    cp "$SETTINGS_FILE" "$SETTINGS_FILE.backup.$(date +%Y%m%d_%H%M%S)"
}

# ===== 代理相关函数 =====
check_proxy_status() {
    if jq -e '.env.http_proxy' "$SETTINGS_FILE" >/dev/null 2>&1; then
        return 0  # 已启用代理
    else
        return 1  # 未启用代理
    fi
}

enable_proxy() {
    echo "🔄 启用代理模式..."
    
    # 保存当前的 hooks 配置
    local hooks_config=$(jq '.hooks // {}' "$SETTINGS_FILE")
    
    # 更新配置，保留 hooks
    jq --argjson hooks "$hooks_config" '.env = {
        "http_proxy": "'"$PROXY_HOST"'",
        "https_proxy": "'"$PROXY_HOST"'"
    } | .hooks = $hooks' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp" && \
    mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 已启用代理模式${NC} ($PROXY_HOST)"
        return 0
    else
        echo -e "${RED}❌ 启用代理失败${NC}"
        return 1
    fi
}

disable_proxy() {
    echo "🔄 禁用代理模式..."
    
    # 使用 jq 删除 env 对象
    jq 'del(.env)' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp" && \
    mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 已禁用代理模式${NC}"
        return 0
    else
        echo -e "${RED}❌ 禁用代理失败${NC}"
        return 1
    fi
}

# ===== Hooks 相关函数 =====
check_hooks_status() {
    if jq -e '.hooks | length > 0' "$SETTINGS_FILE" >/dev/null 2>&1; then
        return 0  # hooks 已启用
    else
        return 1  # hooks 未启用
    fi
}

enable_hooks() {
    echo "🔄 启用 hooks..."
    
    # 默认的 hooks 配置
    local default_hooks='{
      "PostToolUse": [
        {
          "matcher": "Write|Edit|MultiEdit",
          "hooks": [
            {
              "type": "command",
              "command": "~/.claude/hooks/smart-lint.sh"
            },
            {
              "type": "command",
              "command": "~/.claude/hooks/smart-test.sh"
            }
          ]
        }
      ],
      "Stop": [
        {
          "matcher": "",
          "hooks": [
            {
              "type": "command",
              "command": "~/.claude/hooks/ntfy-notifier.sh"
            }
          ]
        }
      ]
    }'
    
    # 检查是否有备份的 hooks 配置
    if [ -f "$SETTINGS_FILE.hooks_backup" ]; then
        echo "   发现备份的 hooks 配置，正在恢复..."
        local hooks_config=$(cat "$SETTINGS_FILE.hooks_backup")
    else
        echo "   使用默认 hooks 配置..."
        local hooks_config="$default_hooks"
    fi
    
    # 更新配置
    jq --argjson hooks "$hooks_config" '.hooks = $hooks' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp" && \
    mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 已启用 hooks${NC}"
        return 0
    else
        echo -e "${RED}❌ 启用 hooks 失败${NC}"
        return 1
    fi
}

disable_hooks() {
    echo "🔄 禁用 hooks..."
    
    # 先备份当前的 hooks 配置
    jq '.hooks // {}' "$SETTINGS_FILE" > "$SETTINGS_FILE.hooks_backup"
    
    # 删除 hooks 配置
    jq 'del(.hooks)' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp" && \
    mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 已禁用 hooks${NC}"
        echo "   (hooks 配置已备份)"
        return 0
    else
        echo -e "${RED}❌ 禁用 hooks 失败${NC}"
        return 1
    fi
}

# ===== 显示状态函数 =====
show_status() {
    echo -e "\n${BLUE}📊 Claude 配置状态：${NC}"
    echo "===================="
    
    # 代理状态
    echo -e "\n${YELLOW}🌐 代理状态：${NC}"
    if check_proxy_status; then
        echo -e "   ${GREEN}✅ 已启用${NC}"
        echo "   代理地址：$PROXY_HOST"
    else
        echo -e "   ⚫ 已禁用"
    fi
    
    # Hooks 状态
    echo -e "\n${YELLOW}🪝 Hooks 状态：${NC}"
    if check_hooks_status; then
        echo -e "   ${GREEN}✅ 已启用${NC}"
        # 显示 hooks 数量
        local post_tool_count=$(jq '.hooks.PostToolUse[0].hooks | length' "$SETTINGS_FILE" 2>/dev/null || echo 0)
        local stop_count=$(jq '.hooks.Stop[0].hooks | length' "$SETTINGS_FILE" 2>/dev/null || echo 0)
        echo "   PostToolUse hooks: $post_tool_count 个"
        echo "   Stop hooks: $stop_count 个"
    else
        echo -e "   ⚫ 已禁用"
    fi
    
    echo ""
}

# ===== 主程序逻辑 =====
case "$1" in
    "proxy")
        # 代理相关操作
        case "$2" in
            ""|"toggle")
                # 切换代理
                if check_proxy_status; then
                    disable_proxy
                else
                    enable_proxy
                fi
                show_status
                ;;
            "on"|"enable")
                if check_proxy_status; then
                    echo "ℹ️  代理已经启用"
                else
                    enable_proxy
                fi
                show_status
                ;;
            "off"|"disable")
                if ! check_proxy_status; then
                    echo "ℹ️  代理已经禁用"
                else
                    disable_proxy
                fi
                show_status
                ;;
            *)
                echo -e "${RED}❌ 错误：未知的代理操作 '$2'${NC}"
                echo "   使用 'claude-config.sh help' 查看帮助"
                exit 1
                ;;
        esac
        ;;
    
    "hooks")
        # Hooks 相关操作
        case "$2" in
            ""|"toggle")
                # 切换 hooks
                if check_hooks_status; then
                    disable_hooks
                else
                    enable_hooks
                fi
                show_status
                ;;
            "on"|"enable")
                if check_hooks_status; then
                    echo "ℹ️  Hooks 已经启用"
                else
                    enable_hooks
                fi
                show_status
                ;;
            "off"|"disable")
                if ! check_hooks_status; then
                    echo "ℹ️  Hooks 已经禁用"
                else
                    disable_hooks
                fi
                show_status
                ;;
            *)
                echo -e "${RED}❌ 错误：未知的 hooks 操作 '$2'${NC}"
                echo "   使用 'claude-config.sh help' 查看帮助"
                exit 1
                ;;
        esac
        ;;
    
    "status"|"")
        # 显示状态
        show_status
        ;;
    
    "backup")
        # 备份当前配置
        backup_settings
        echo -e "${GREEN}✅ 已备份配置文件${NC}"
        echo "   备份位置：$SETTINGS_FILE.backup.$(date +%Y%m%d_%H%M%S)"
        ;;
    
    "help"|"-h"|"--help")
        # 显示帮助
        echo -e "${BLUE}Claude 配置管理工具${NC}"
        echo "===================="
        echo ""
        echo "用法："
        echo -e "  ${GREEN}claude-config.sh${NC}                    # 显示当前状态"
        echo -e "  ${GREEN}claude-config.sh status${NC}             # 显示当前状态"
        echo ""
        echo "代理管理："
        echo -e "  ${GREEN}claude-config.sh proxy${NC}              # 切换代理（开/关）"
        echo -e "  ${GREEN}claude-config.sh proxy on${NC}           # 启用代理"
        echo -e "  ${GREEN}claude-config.sh proxy off${NC}          # 禁用代理"
        echo ""
        echo "Hooks 管理："
        echo -e "  ${GREEN}claude-config.sh hooks${NC}              # 切换 hooks（开/关）"
        echo -e "  ${GREEN}claude-config.sh hooks on${NC}           # 启用 hooks"
        echo -e "  ${GREEN}claude-config.sh hooks off${NC}          # 禁用 hooks"
        echo ""
        echo "其他："
        echo -e "  ${GREEN}claude-config.sh backup${NC}             # 备份当前配置"
        echo -e "  ${GREEN}claude-config.sh help${NC}               # 显示此帮助"
        echo ""
        echo "配置文件：$SETTINGS_FILE"
        echo "代理地址：$PROXY_HOST"
        ;;
    
    *)
        # 未知参数
        echo -e "${RED}❌ 错误：未知命令 '$1'${NC}"
        echo "   使用 'claude-config.sh help' 查看帮助"
        exit 1
        ;;
esac