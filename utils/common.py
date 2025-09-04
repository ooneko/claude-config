#!/usr/bin/env python3
"""
通用工具类和常量定义
"""

import filecmp
import difflib
from pathlib import Path
from dataclasses import dataclass
from enum import Enum


class Color:
    """命令行颜色输出"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

    @staticmethod
    def print_colored(text: str, color: str) -> None:
        print(f"{color}{text}{Color.NC}")

    @staticmethod
    def input_colored(prompt: str, color: str = None) -> str:
        if color is None:
            color = Color.YELLOW
        return input(f"{color}{prompt}{Color.NC}")


class ConflictResolution(Enum):
    """冲突解决方式"""
    OVERWRITE = "overwrite"
    SKIP = "skip"
    SHOW_DIFF = "diff"
    MERGE = "merge"


@dataclass
class OperationResult:
    """操作结果"""
    success: bool
    message: str
    skipped: bool = False


class FileComparator:
    """文件比较工具"""
    
    @staticmethod
    def files_are_same(file1: Path, file2: Path) -> bool:
        """比较两个文件是否相同"""
        if not file1.exists() or not file2.exists():
            return False
        return filecmp.cmp(file1, file2, shallow=False)
    
    @staticmethod
    def show_file_diff(file1: Path, file2: Path) -> None:
        """显示两个文件的差异"""
        try:
            with open(file1, 'r', encoding='utf-8') as f1, open(file2, 'r', encoding='utf-8') as f2:
                diff = difflib.unified_diff(
                    f1.readlines(),
                    f2.readlines(),
                    fromfile=str(file1),
                    tofile=str(file2),
                    lineterm=''
                )
                Color.print_colored("文件差异:", Color.YELLOW)
                for line in diff:
                    if line.startswith('+++') or line.startswith('---'):
                        Color.print_colored(line, Color.BLUE)
                    elif line.startswith('@@'):
                        Color.print_colored(line, Color.YELLOW)
                    elif line.startswith('+'):
                        Color.print_colored(line, Color.GREEN)
                    elif line.startswith('-'):
                        Color.print_colored(line, Color.RED)
                    else:
                        print(line)
        except Exception as e:
            Color.print_colored(f"显示差异失败: {e}", Color.RED)


class ProxyManager:
    """代理地址管理器"""
    
    def __init__(self, claude_dir: Path):
        self.claude_dir = claude_dir
        self.proxy_file = claude_dir / ".proxy_config"
        self.default_proxy = "http://127.0.0.1:7890"
    
    def get_proxy_address(self) -> str:
        """获取代理地址，如果文件不存在则返回默认值"""
        if self.proxy_file.exists():
            try:
                with open(self.proxy_file, 'r', encoding='utf-8') as f:
                    return f.read().strip()
            except Exception:
                return self.default_proxy
        return self.default_proxy
    
    def save_proxy_address(self, proxy_address: str) -> None:
        """保存代理地址到文件"""
        try:
            self.claude_dir.mkdir(parents=True, exist_ok=True)
            with open(self.proxy_file, 'w', encoding='utf-8') as f:
                f.write(proxy_address)
        except Exception as e:
            Color.print_colored(f"保存代理地址失败: {e}", Color.RED)
    
    def prompt_for_proxy(self) -> str:
        """提示用户输入代理地址"""
        Color.print_colored("🌐 首次配置代理，请输入代理地址", Color.YELLOW)
        print(f"默认代理地址: {self.default_proxy}")
        print("直接按回车使用默认地址，或输入自定义代理地址:")
        
        while True:
            proxy_input = Color.input_colored("代理地址: ").strip()
            
            if not proxy_input:
                proxy_address = self.default_proxy
                break
            elif proxy_input.startswith(('http://', 'https://')):
                proxy_address = proxy_input
                break
            else:
                print("❌ 代理地址必须以 http:// 或 https:// 开头")
                print("   示例: http://127.0.0.1:7890")
        
        self.save_proxy_address(proxy_address)
        Color.print_colored(f"✅ 代理地址已保存: {proxy_address}", Color.GREEN)
        return proxy_address