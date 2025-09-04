#!/usr/bin/env python3
"""
文件操作模块 - 处理文件复制、合并和管理
"""

import json
import shutil
from pathlib import Path
from typing import Dict, Any, List

from .common import Color, OperationResult, FileComparator


class SettingsJsonMerger:
    """settings.json智能合并器"""
    
    @staticmethod
    def should_preserve_proxy_config(target_data: Dict[str, Any]) -> bool:
        """检查是否应该保留目标文件中的代理配置"""
        env = target_data.get('env', {})
        return 'http_proxy' in env or 'https_proxy' in env
    
    @staticmethod
    def filter_proxy_from_source(source_data: Dict[str, Any]) -> Dict[str, Any]:
        """从源数据中移除代理配置"""
        result = source_data.copy()
        if 'env' in result and isinstance(result['env'], dict):
            env = result['env'].copy()
            env.pop('http_proxy', None)
            env.pop('https_proxy', None)
            if env:
                result['env'] = env
            else:
                result.pop('env', None)
        return result
    
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
    def merge_settings(target_file: Path, source_file: Path) -> OperationResult:
        """合并settings.json文件"""
        try:
            # 读取源文件
            with open(source_file, 'r', encoding='utf-8') as f:
                source_data = json.load(f)
            
            # 读取目标文件（如果存在）
            if target_file.exists():
                with open(target_file, 'r', encoding='utf-8') as f:
                    target_data = json.load(f)
                
                # 检查是否需要保留代理配置
                preserve_proxy = SettingsJsonMerger.should_preserve_proxy_config(target_data)
                
                # 如果目标文件有代理配置，从源文件中移除代理配置
                if preserve_proxy:
                    Color.print_colored("📡 检测到现有代理配置，将保留用户代理设置", Color.YELLOW)
                    source_data = SettingsJsonMerger.filter_proxy_from_source(source_data)
                
                # 深度合并
                merged_data = SettingsJsonMerger.deep_merge_dict(target_data, source_data)
                
                # 检查是否有变化
                if merged_data != target_data:
                    Color.print_colored("🔄 检测到settings.json配置变化", Color.YELLOW)
                    print("将进行智能合并，保留您的个人配置")
                    if preserve_proxy:
                        print("   - 保留现有代理配置")
                    
                    # 写入合并后的配置
                    with open(target_file, 'w', encoding='utf-8') as f:
                        json.dump(merged_data, f, indent=2, ensure_ascii=False)
                    
                    return OperationResult(True, "智能合并settings.json配置")
                else:
                    return OperationResult(True, "settings.json配置无变化", skipped=True)
            else:
                # 目标文件不存在，检查源文件是否包含代理配置
                if 'env' in source_data and isinstance(source_data['env'], dict):
                    env = source_data['env']
                    if 'http_proxy' in env or 'https_proxy' in env:
                        Color.print_colored("⚠️  源文件包含代理配置，但将被跳过", Color.YELLOW)
                        print("   请使用 claude-config.py proxy on 来配置代理")
                        # 移除代理配置后再复制
                        source_data = SettingsJsonMerger.filter_proxy_from_source(source_data)
                
                # 写入过滤后的配置
                with open(target_file, 'w', encoding='utf-8') as f:
                    json.dump(source_data, f, indent=2, ensure_ascii=False)
                return OperationResult(True, "复制settings.json配置")
                
        except json.JSONDecodeError as e:
            return OperationResult(False, f"JSON格式错误: {e}")
        except Exception as e:
            return OperationResult(False, f"合并settings.json失败: {e}")


class FileOperations:
    """文件操作管理器"""
    
    def __init__(self, source_dir: Path, target_dir: Path, selected_items: List[str] = None):
        self.source_dir = source_dir
        self.target_dir = target_dir
        self.selected_items = selected_items or [
            "agents",
            "commands", 
            "hooks",
            "output-styles",
            "CLAUDE.md.to.copy",
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

    def handle_claude_md(self, src_path: Path, dest_path: Path) -> OperationResult:
        """特殊处理CLAUDE.md文件"""
        if not dest_path.exists():
            shutil.copy2(src_path, dest_path)
            return OperationResult(True, "复制CLAUDE.md")
        
        # 检查文件是否相同
        if FileComparator.files_are_same(src_path, dest_path):
            return OperationResult(True, "跳过相同的CLAUDE.md", skipped=True)
        
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
            choice = Color.input_colored("请选择 (y/n/d): ").strip().lower()
            
            if choice in ['y', 'yes']:
                shutil.copy2(src_path, dest_path)
                return OperationResult(True, "覆盖CLAUDE.md")
            elif choice in ['n', 'no']:
                return OperationResult(True, "跳过CLAUDE.md", skipped=True)
            elif choice in ['d', 'diff']:
                FileComparator.show_file_diff(dest_path, src_path)
                print()
            else:
                print("请输入 y、n 或 d")

    def copy_file(self, src_path: Path, dest_path: Path) -> OperationResult:
        """复制单个文件"""
        try:
            # 特殊处理不同类型的文件
            if src_path.name == "CLAUDE.md.to.copy":
                return self.handle_claude_md(src_path, dest_path)
            elif src_path.name == "settings.json":
                return SettingsJsonMerger.merge_settings(dest_path, src_path)
            
            # 普通文件处理
            if dest_path.exists():
                if FileComparator.files_are_same(src_path, dest_path):
                    return OperationResult(True, f"跳过相同文件: {src_path.name}", skipped=True)
                else:
                    shutil.copy2(src_path, dest_path)
                    return OperationResult(True, f"覆盖文件: {src_path.name}")
            else:
                # 确保目标目录存在
                dest_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src_path, dest_path)
                return OperationResult(True, f"复制文件: {src_path.name}")
                
        except Exception as e:
            return OperationResult(False, f"复制文件{src_path.name}失败: {e}")

    def copy_directory(self, src_path: Path, dest_path: Path) -> OperationResult:
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
            
            if success_count > 0:
                return OperationResult(True, f"处理目录: {src_path.name} ({success_count}个文件)")
            else:
                return OperationResult(True, f"跳过目录: {src_path.name} (无变化)", skipped=True)
                
        except Exception as e:
            return OperationResult(False, f"复制目录{src_path.name}失败: {e}")

    def copy_item(self, src_path: Path, dest_path: Path) -> OperationResult:
        """复制文件或目录"""
        if src_path.is_file():
            return self.copy_file(src_path, dest_path)
        elif src_path.is_dir():
            return self.copy_directory(src_path, dest_path)
        else:
            return OperationResult(False, f"未知类型: {src_path.name}")

    def run_copy_operation(self) -> bool:
        """执行复制操作"""
        if len(self.selected_items) < 7:  # 不是全部项目
            print(f"🐠 开始仅复制{', '.join(self.selected_items)}配置从 {self.source_dir} 到 {self.target_dir}")
        else:
            print(f"🐠 开始将配置文件从 {self.source_dir} 复制到 {self.target_dir}")
        
        # 创建目标目录
        if not self.create_target_dir():
            return False
        
        print("-" * 50)
        
        success_count = 0
        skip_count = 0
        error_count = 0
        
        # 复制每个配置项
        for item_name in self.selected_items:
            src_path = self.source_dir / item_name
            
            if not src_path.exists():
                print(f"跳过不存在的项目: {item_name}")
                continue
            
            # 特殊处理：将 CLAUDE.md.to.copy 复制为 CLAUDE.md
            if item_name == "CLAUDE.md.to.copy":
                dest_path = self.target_dir / "CLAUDE.md"
            else:
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
        
        # 显示代理配置提示
        if error_count == 0:
            print("\\n💡 代理配置提示:")
            print("   - 启用代理: ./claude-config.py proxy on")
            print("   - 禁用代理: ./claude-config.py proxy off")
            print("   - 查看状态: ./claude-config.py status")
        
        # 显示目标目录内容
        try:
            print("\\n目标目录内容:")
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