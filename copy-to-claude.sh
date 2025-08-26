#!/bin/bash

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 源目录和目标目录
SOURCE_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="$HOME/.claude"

echo "🐠 开始将配置文件从 $SOURCE_DIR 复制到 $TARGET_DIR"

# 创建目标目录（如果不存在）
if [ ! -d "$TARGET_DIR" ]; then
    echo -e "${GREEN}创建目录: $TARGET_DIR${NC}"
    mkdir -p "$TARGET_DIR"
fi

# CLAUDE.md文件特殊处理函数
handle_claude_md() {
    local src="$1"
    local dest="$2"
    
    if [ -f "$dest" ]; then
        if ! cmp -s "$src" "$dest"; then
            echo -e "${YELLOW}⚠️  发现 CLAUDE.md 文件内容不同！${NC}"
            echo "源文件: $src"
            echo "目标文件: $dest"
            echo ""
            echo -e "${YELLOW}是否要覆盖目标文件？${NC}"
            echo "  [y/Y] 是，覆盖"
            echo "  [n/N] 否，跳过"
            echo "  [d/D] 查看差异"
            
            while true; do
                read -p "请选择 (y/n/d): " choice
                case "$choice" in
                    [Yy]* )
                        echo -e "${GREEN}覆盖文件: CLAUDE.md${NC}"
                        cp "$src" "$dest"
                        return 0
                        ;;
                    [Nn]* )
                        echo -e "${YELLOW}跳过文件: CLAUDE.md${NC}"
                        return 0
                        ;;
                    [Dd]* )
                        echo -e "${YELLOW}文件差异:${NC}"
                        diff -u "$dest" "$src" || true
                        echo ""
                        ;;
                    * )
                        echo "请输入 y、n 或 d"
                        ;;
                esac
            done
        else
            echo -e "跳过相同文件: CLAUDE.md"
        fi
    else
        echo -e "${GREEN}复制文件: CLAUDE.md${NC}"
        cp "$src" "$dest"
    fi
}

# 复制函数
copy_item() {
    local src="$1"
    local dest="$2"
    local item_name="$(basename "$src")"
    
    # CLAUDE.md文件特殊处理
    if [ "$item_name" = "CLAUDE.md" ] && [ -f "$src" ]; then
        handle_claude_md "$src" "$dest"
        return
    fi
    
    if [ -d "$src" ]; then
        # 处理目录
        if [ -d "$dest" ]; then
            echo -e "${YELLOW}合并目录: $item_name${NC}"
            # 确保目标目录存在
            mkdir -p "$dest"
            # 递归复制目录内容
            for item in "$src"/*; do
                if [ -e "$item" ]; then
                    copy_item "$item" "$dest/$(basename "$item")"
                fi
            done
            # 复制隐藏文件
            for item in "$src"/.*; do
                base_name="$(basename "$item")"
                if [ "$base_name" != "." ] && [ "$base_name" != ".." ] && [ "$base_name" != ".git" ]; then
                    copy_item "$item" "$dest/$base_name"
                fi
            done
        else
            echo -e "${GREEN}复制目录: $item_name${NC}"
            # 确保父目录存在
            mkdir -p "$(dirname "$dest")"
            cp -r "$src" "$dest"
        fi
    else
        # 处理文件
        if [ -f "$dest" ]; then
            # 检查文件是否相同
            if cmp -s "$src" "$dest"; then
                echo -e "跳过相同文件: $item_name"
            else
                echo -e "${YELLOW}覆盖文件: $item_name${NC}"
                cp "$src" "$dest"
            fi
        else
            echo -e "${GREEN}复制文件: $item_name${NC}"
            cp "$src" "$dest"
        fi
    fi
}

# 定义要复制的 Claude 配置文件和目录
CLAUDE_ITEMS=(
    "agents"
    "commands"
    "hooks"
    "output-styles"
    "CLAUDE.md"
    "claude-config.sh"
    "settings.json"
)

# 开始复制
echo "----------------------------------------"

# 只复制指定的 Claude 配置文件和目录
for item_name in "${CLAUDE_ITEMS[@]}"; do
    src_path="$SOURCE_DIR/$item_name"
    if [ -e "$src_path" ]; then
        copy_item "$src_path" "$TARGET_DIR/$item_name"
    else
        echo "跳过不存在的项目: $item_name"
    fi
done

echo "----------------------------------------"
echo -e "${GREEN}✅ 复制完成！${NC}"
echo "配置文件已成功复制到 $TARGET_DIR"

# 列出目标目录内容
echo ""
echo "目标目录内容："
ls -la "$TARGET_DIR"