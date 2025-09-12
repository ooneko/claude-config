#!/bin/bash
#
# debug-hook.sh - 调试hook脚本，记录Claude发送的所有内容
# 
# 使用方法：
# 1. 将此脚本设为可执行: chmod +x debug-hook.sh
# 2. 在Claude Code设置中配置为hook脚本
# 3. 运行命令触发hook，然后查看 /tmp/claude-hook-debug.log
#

# 日志文件位置
LOG_FILE="/tmp/claude-hook-debug.log"

# 创建日志文件头部信息
echo "========================================" >> "$LOG_FILE"
echo "Claude Hook Debug Log" >> "$LOG_FILE"
echo "Timestamp: $(date)" >> "$LOG_FILE"
echo "Script: $0" >> "$LOG_FILE"
echo "Process ID: $$" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

# 记录所有环境变量（过滤敏感信息）
echo "" >> "$LOG_FILE"
echo "--- 环境变量 ---" >> "$LOG_FILE"
env | grep -E '^(CLAUDE_|PATH|HOME|USER|PWD)' | sort >> "$LOG_FILE"

echo "" >> "$LOG_FILE"
echo "--- 所有环境变量 ---" >> "$LOG_FILE"
env | sort >> "$LOG_FILE"

# 记录命令行参数
echo "" >> "$LOG_FILE"
echo "--- 命令行参数 ---" >> "$LOG_FILE"
echo "参数个数: $#" >> "$LOG_FILE"
for i in $(seq 1 $#); do
    echo "参数 $i: ${!i}" >> "$LOG_FILE"
done

# 记录当前工作目录
echo "" >> "$LOG_FILE"
echo "--- 当前工作目录 ---" >> "$LOG_FILE"
echo "PWD: $(pwd)" >> "$LOG_FILE"
echo "目录内容:" >> "$LOG_FILE"
ls -la >> "$LOG_FILE"

# 记录标准输入内容（如果有的话）
echo "" >> "$LOG_FILE"
echo "--- 标准输入内容 ---" >> "$LOG_FILE"
if [ -t 0 ]; then
    echo "无标准输入数据（终端输入）" >> "$LOG_FILE"
else
    echo "检测到标准输入数据:" >> "$LOG_FILE"
    # 读取并保存标准输入
    stdin_content=$(cat)
    if [ -n "$stdin_content" ]; then
        echo "$stdin_content" >> "$LOG_FILE"
        echo "" >> "$LOG_FILE"
        echo "标准输入内容长度: ${#stdin_content} 字符" >> "$LOG_FILE"
        
        # 尝试解析JSON格式
        if echo "$stdin_content" | jq . >/dev/null 2>&1; then
            echo "标准输入是有效的JSON格式:" >> "$LOG_FILE"
            echo "$stdin_content" | jq . >> "$LOG_FILE"
        fi
    else
        echo "标准输入为空" >> "$LOG_FILE"
    fi
fi

# 记录文件描述符信息
echo "" >> "$LOG_FILE"
echo "--- 文件描述符信息 ---" >> "$LOG_FILE"
if [ -d "/proc/$$/fd" ]; then
    ls -la "/proc/$$/fd" >> "$LOG_FILE" 2>/dev/null || echo "无法读取/proc/$$/fd" >> "$LOG_FILE"
fi

# 记录进程相关信息
echo "" >> "$LOG_FILE"
echo "--- 进程信息 ---" >> "$LOG_FILE"
echo "父进程ID: $PPID" >> "$LOG_FILE"
echo "进程组ID: $(ps -o pgid= -p $$)" >> "$LOG_FILE"

# 如果有特定的Claude相关文件，记录它们
echo "" >> "$LOG_FILE"
echo "--- Claude相关文件 ---" >> "$LOG_FILE"
if [ -d "$HOME/.claude" ]; then
    echo "~/.claude 目录存在" >> "$LOG_FILE"
    ls -la "$HOME/.claude" >> "$LOG_FILE" 2>/dev/null
else
    echo "~/.claude 目录不存在" >> "$LOG_FILE"
fi

echo "" >> "$LOG_FILE"
echo "--- Hook执行完成 ---" >> "$LOG_FILE"
echo "结束时间: $(date)" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# 输出到终端，告知用户日志位置
echo "Claude hook debug info saved to: $LOG_FILE"
echo "View with: tail -f $LOG_FILE"

# 返回成功状态码
exit 0