#!/usr/bin/env python3
"""
Claude配置文件复制工具 - Python版本

智能复制Claude配置文件到~/.claude目录，特别支持settings.json的深度合并
"""

import json
import shutil
import sys
import argparse
from pathlib import Path
from typing import Any, Dict
import filecmp
import difflib
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
    def input_colored(prompt: str, color: str = YELLOW) -> str:
        return input(f"{color}{prompt}{Color.NC}")


class ConflictResolution(Enum):
    """冲突解决方式"""
    OVERWRITE = "overwrite"
    SKIP = "skip"
    SHOW_DIFF = "diff"
    MERGE = "merge"


@dataclass
class CopyResult:
    """复制操作结果"""
    success: bool
    message: str
    skipped: bool = False


class SettingsJsonMerger:
    """settings.json智能合并器"""
    
    @staticmethod
    def deep_merge_dict(target: Dict[str, Any], source: Dict[str, Any]) -> Dict[str, Any]:
        """深度合并字典，source覆盖target"""
        result = target.copy()
        
        for key, value in source.items():
            if key in result:
                if isinstance(result[key], dict) and isinstance(value, dict):
                    # 特殊处理hooks字典
                    if key == "hooks":
                        result[key] = SettingsJsonMerger.merge_hooks(result[key], value)
                    else:
                        result[key] = SettingsJsonMerger.deep_merge_dict(result[key], value)
                elif isinstance(result[key], list) and isinstance(value, list):
                    # 其他数组直接合并，去重（只处理基本类型）
                    combined = result[key] + value
                    # 对于包含字典的列表，不能直接用dict.fromkeys()
                    seen = set()
                    unique_combined = []
                    for item in combined:
                        if isinstance(item, dict):
                            # 字典类型不能hash，直接添加
                            unique_combined.append(item)
                        else:
                            # 基本类型可以去重
                            if item not in seen:
                                seen.add(item)
                                unique_combined.append(item)
                    result[key] = unique_combined
                else:
                    result[key] = value
            else:
                result[key] = value
                
        return result

    @staticmethod
    def merge_hooks(target_hooks: Dict[str, Any], source_hooks: Dict[str, Any]) -> Dict[str, Any]:
        """智能合并hooks配置"""
        result = target_hooks.copy()
        
        for event_type, source_configs in source_hooks.items():
            if event_type not in result:
                result[event_type] = source_configs.copy() if isinstance(source_configs, list) else source_configs
            else:
                # 合并同一事件类型的hooks
                existing_configs = result[event_type]
                if isinstance(existing_configs, list) and isinstance(source_configs, list):
                    # 按matcher合并，避免重复
                    existing_matchers_map = {config.get('matcher', ''): i for i, config in enumerate(existing_configs) if isinstance(config, dict)}
                    
                    for config in source_configs:
                        if not isinstance(config, dict):
                            continue
                            
                        matcher = config.get('matcher', '')
                        if matcher not in existing_matchers_map:
                            # 新的matcher，直接添加
                            existing_configs.append(config.copy() if hasattr(config, 'copy') else config)
                        else:
                            # 相同matcher，合并hooks命令（自动合并，不再询问用户）
                            existing_index = existing_matchers_map[matcher]
                            existing_config = existing_configs[existing_index]
                            existing_hooks = existing_config.get('hooks', [])
                            new_hooks = config.get('hooks', [])
                            
                            # 按command去重合并
                            existing_commands = {h.get('command', '') for h in existing_hooks if isinstance(h, dict)}
                            for hook in new_hooks:
                                if isinstance(hook, dict) and hook.get('command', '') not in existing_commands:
                                    existing_hooks.append(hook.copy() if hasattr(hook, 'copy') else hook)
                            
                            # 更新现有配置
                            existing_configs[existing_index]['hooks'] = existing_hooks
                
        return result

    @staticmethod
    def merge_settings(target_file: Path, source_file: Path) -> CopyResult:
        """合并settings.json文件"""
        try:
            # 读取源文件
            with open(source_file, 'r', encoding='utf-8') as f:
                source_data = json.load(f)
            
            # 读取目标文件（如果存在）
            if target_file.exists():
                with open(target_file, 'r', encoding='utf-8') as f:
                    target_data = json.load(f)
                
                # 深度合并
                merged_data = SettingsJsonMerger.deep_merge_dict(target_data, source_data)
                
                # 检查是否有变化
                if merged_data != target_data:
                    Color.print_colored("🔄 检测到settings.json配置变化", Color.YELLOW)
                    print("将进行智能合并，保留您的个人配置")
                    
                    # 写入合并后的配置
                    with open(target_file, 'w', encoding='utf-8') as f:
                        json.dump(merged_data, f, indent=2, ensure_ascii=False)
                    
                    return CopyResult(True, "智能合并settings.json配置")
                else:
                    return CopyResult(True, "settings.json配置无变化", skipped=True)
            else:
                # 目标文件不存在，直接复制
                shutil.copy2(source_file, target_file)
                return CopyResult(True, "复制settings.json配置")
                
        except json.JSONDecodeError as e:
            return CopyResult(False, f"JSON格式错误: {e}")
        except Exception as e:
            return CopyResult(False, f"合并settings.json失败: {e}")


class ClaudeConfigCopier:
    """Claude配置文件复制器"""
    
    def __init__(self, source_dir: Path, target_dir: Path, agents_only: bool = False):
        self.source_dir = source_dir
        self.target_dir = target_dir
        self.agents_only = agents_only
        
        # 根据agents_only标志决定复制哪些项目
        if agents_only:
            self.claude_items = ["agents"]
        else:
            self.claude_items = [
                "agents",
                "commands", 
                "hooks",
                "output-styles",
                "CLAUDE.md",
                "claude-config.sh",
                "settings.json"
            ]

    def create_target_dir(self) -> bool:
        """创建目标目录"""
        try:
            self.target_dir.mkdir(parents=True, exist_ok=True)
            if not self.target_dir.exists():
                Color.print_colored(f"创建目录: {self.target_dir}", Color.GREEN)
            return True
        except Exception as e:
            Color.print_colored(f"创建目录失败: {e}", Color.RED)
            return False

    def handle_claude_md(self, src_path: Path, dest_path: Path) -> CopyResult:
        """特殊处理CLAUDE.md文件"""
        if not dest_path.exists():
            shutil.copy2(src_path, dest_path)
            return CopyResult(True, "复制CLAUDE.md")
        
        # 检查文件是否相同
        if filecmp.cmp(src_path, dest_path, shallow=False):
            return CopyResult(True, "跳过相同的CLAUDE.md", skipped=True)
        
        # 文件不同，询问用户处理方式
        Color.print_colored("⚠️  发现CLAUDE.md文件内容不同！", Color.YELLOW)
        print(f"源文件: {src_path}")
        print(f"目标文件: {dest_path}")
        print()
        Color.print_colored("请选择处理方式:", Color.YELLOW)
        print("  [y/Y] 覆盖目标文件")
        print("  [n/N] 跳过此文件")
        print("  [d/D] 查看文件差异")
        
        while True:
            choice = Color.input_colored("请选择 (y/n/d): ", Color.YELLOW).strip().lower()
            
            if choice in ['y', 'yes']:
                shutil.copy2(src_path, dest_path)
                return CopyResult(True, "覆盖CLAUDE.md")
            elif choice in ['n', 'no']:
                return CopyResult(True, "跳过CLAUDE.md", skipped=True)
            elif choice in ['d', 'diff']:
                self.show_file_diff(dest_path, src_path)
                print()
            else:
                print("请输入 y、n 或 d")

    def show_file_diff(self, file1: Path, file2: Path) -> None:
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

    def copy_file(self, src_path: Path, dest_path: Path) -> CopyResult:
        """复制单个文件"""
        try:
            # 特殊处理不同类型的文件
            if src_path.name == "CLAUDE.md":
                return self.handle_claude_md(src_path, dest_path)
            elif src_path.name == "settings.json":
                return SettingsJsonMerger.merge_settings(dest_path, src_path)
            
            # 普通文件处理
            if dest_path.exists():
                if filecmp.cmp(src_path, dest_path, shallow=False):
                    return CopyResult(True, f"跳过相同文件: {src_path.name}", skipped=True)
                else:
                    shutil.copy2(src_path, dest_path)
                    return CopyResult(True, f"覆盖文件: {src_path.name}")
            else:
                # 确保目标目录存在
                dest_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src_path, dest_path)
                return CopyResult(True, f"复制文件: {src_path.name}")
                
        except Exception as e:
            return CopyResult(False, f"复制文件{src_path.name}失败: {e}")

    def copy_directory(self, src_path: Path, dest_path: Path) -> CopyResult:
        """递归复制目录"""
        try:
            results = []
            dest_path.mkdir(parents=True, exist_ok=True)
            
            # 复制所有文件和子目录
            for item in src_path.iterdir():
                if item.name.startswith('.') and item.name not in ['.gitkeep']:
                    continue  # 跳过隐藏文件，除了.gitkeep
                
                dest_item = dest_path / item.name
                
                if item.is_file():
                    result = self.copy_file(item, dest_item)
                    results.append(result)
                elif item.is_dir():
                    result = self.copy_directory(item, dest_item)
                    results.append(result)
            
            # 统计结果
            success_count = sum(1 for r in results if r.success and not r.skipped)
            skip_count = sum(1 for r in results if r.skipped)
            
            if success_count > 0:
                return CopyResult(True, f"处理目录: {src_path.name} ({success_count}个文件)")
            else:
                return CopyResult(True, f"跳过目录: {src_path.name} (无变化)", skipped=True)
                
        except Exception as e:
            return CopyResult(False, f"复制目录{src_path.name}失败: {e}")

    def copy_item(self, src_path: Path, dest_path: Path) -> CopyResult:
        """复制文件或目录"""
        if src_path.is_file():
            return self.copy_file(src_path, dest_path)
        elif src_path.is_dir():
            return self.copy_directory(src_path, dest_path)
        else:
            return CopyResult(False, f"未知类型: {src_path.name}")

    def run(self) -> bool:
        """执行复制操作"""
        if self.agents_only:
            print("🐠 开始仅复制agents配置从", str(self.source_dir), "到", str(self.target_dir))
        else:
            print("🐠 开始将配置文件从", str(self.source_dir), "复制到", str(self.target_dir))
        
        # 创建目标目录
        if not self.create_target_dir():
            return False
        
        print("-" * 50)
        
        success_count = 0
        skip_count = 0
        error_count = 0
        
        # 复制每个配置项
        for item_name in self.claude_items:
            src_path = self.source_dir / item_name
            
            if not src_path.exists():
                print(f"跳过不存在的项目: {item_name}")
                continue
            
            dest_path = self.target_dir / item_name
            result = self.copy_item(src_path, dest_path)
            
            if result.success:
                if result.skipped:
                    print(result.message)
                    skip_count += 1
                else:
                    Color.print_colored(result.message, Color.GREEN)
                    success_count += 1
            else:
                Color.print_colored(f"❌ {result.message}", Color.RED)
                error_count += 1
        
        print("-" * 50)
        
        # 显示结果统计
        if error_count == 0:
            Color.print_colored("✅ 复制完成！", Color.GREEN)
            print(f"成功: {success_count}项, 跳过: {skip_count}项")
        else:
            Color.print_colored(f"⚠️  复制完成，但有{error_count}项失败", Color.YELLOW)
            print(f"成功: {success_count}项, 跳过: {skip_count}项, 失败: {error_count}项")
        
        print(f"配置文件位置: {self.target_dir}")
        
        # 显示目标目录内容
        try:
            print("\n目标目录内容:")
            items = list(self.target_dir.iterdir())
            items.sort(key=lambda x: (x.is_file(), x.name))
            
            for item in items:
                if item.is_dir():
                    Color.print_colored(f"📁 {item.name}/", Color.BLUE)
                else:
                    print(f"📄 {item.name}")
        except Exception as e:
            Color.print_colored(f"列出目录内容失败: {e}", Color.RED)
        
        return error_count == 0


def parse_args():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description='Claude配置文件复制工具',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
使用示例:
  python copy_to_claude.py           # 复制所有配置文件
  python copy_to_claude.py --agents  # 仅复制agents目录
        '''
    )
    
    parser.add_argument(
        '--agents',
        action='store_true',
        help='仅复制agents目录（默认复制所有配置文件）'
    )
    
    return parser.parse_args()


def main():
    """主函数"""
    try:
        # 解析命令行参数
        args = parse_args()
        
        # 确定源目录和目标目录
        script_path = Path(__file__).parent.absolute()
        source_dir = script_path
        target_dir = Path.home() / '.claude'
        
        # 创建复制器并运行
        copier = ClaudeConfigCopier(source_dir, target_dir, agents_only=args.agents)
        success = copier.run()
        
        sys.exit(0 if success else 1)
        
    except KeyboardInterrupt:
        Color.print_colored("\n\n用户中断操作", Color.YELLOW)
        sys.exit(1)
    except Exception as e:
        Color.print_colored(f"运行失败: {e}", Color.RED)
        sys.exit(1)


if __name__ == "__main__":
    main()