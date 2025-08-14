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

# 复制函数
copy_item() {
    local src="$1"
    local dest="$2"
    local item_name="$(basename "$src")"
    
    # 跳过脚本自身和 .git 目录
    if [ "$item_name" = "copy-to-claude.sh" ] || [ "$item_name" = ".git" ]; then
        return
    fi
    
    if [ -d "$src" ]; then
        # 处理目录
        if [ -d "$dest" ]; then
            echo -e "${YELLOW}合并目录: $item_name${NC}"
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

# 开始复制
echo "----------------------------------------"

# 复制所有文件和目录
for item in "$SOURCE_DIR"/*; do
    if [ -e "$item" ]; then
        copy_item "$item" "$TARGET_DIR/$(basename "$item")"
    fi
done

# 复制隐藏文件（除了 .git）
for item in "$SOURCE_DIR"/.*; do
    base_name="$(basename "$item")"
    if [ "$base_name" != "." ] && [ "$base_name" != ".." ] && [ "$base_name" != ".git" ]; then
        copy_item "$item" "$TARGET_DIR/$base_name"
    fi
done

echo "----------------------------------------"
echo -e "${GREEN}✅ 复制完成！${NC}"
echo "配置文件已成功复制到 $TARGET_DIR"

# 列出目标目录内容
echo ""
echo "目标目录内容："
ls -la "$TARGET_DIR"