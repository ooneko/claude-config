#!/usr/bin/env python3
"""
Claude 配置管理统一工具 - Python版本

整合了配置管理和文件复制功能，提供统一的命令行接口
"""

import sys
import argparse
from pathlib import Path
from typing import Optional

from utils.common import Color, ProxyManager
from utils.config_manager import ConfigManager
from utils.file_operations import FileOperations


class ClaudeConfigTool:
    """Claude 配置管理工具主类"""
    
    def __init__(self):
        self.source_dir = Path(__file__).parent.absolute()
        self.target_dir = Path.home() / '.claude'
        self.config_manager = ConfigManager(self.target_dir)
        self.proxy_manager = ProxyManager(self.target_dir)
    
    def copy_files(self, agents: bool = False, commands: bool = False) -> bool:
        """复制配置文件"""
        # 根据标志决定复制哪些项目
        if agents or commands:
            selected_items = []
            if agents:
                selected_items.append("agents")
            if commands:
                selected_items.append("commands")
        else:
            selected_items = [
                "agents",
                "commands", 
                "hooks",
                "output-styles",
                "CLAUDE.md.to.copy",
                "settings.json"
            ]
        
        file_ops = FileOperations(self.source_dir, self.target_dir, selected_items)
        return file_ops.run_copy_operation()
    
    def handle_proxy_command(self, action: Optional[str] = None) -> None:
        """处理代理相关命令"""
        if action is None or action == "toggle":
            # 切换代理
            if self.config_manager.check_proxy_status():
                result = self.config_manager.disable_proxy()
            else:
                result = self.config_manager.enable_proxy()
        elif action in ["on", "enable"]:
            if self.config_manager.check_proxy_status():
                Color.print_colored("ℹ️  代理已经启用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.enable_proxy()
        elif action in ["off", "disable"]:
            if not self.config_manager.check_proxy_status():
                Color.print_colored("ℹ️  代理已经禁用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.disable_proxy()
        else:
            Color.print_colored(f"❌ 错误：未知的代理操作 '{action}'", Color.RED)
            print("   使用 'claude-config.py help' 查看帮助")
            return
        
        # 显示结果
        if result.success:
            Color.print_colored(f"✅ {result.message}", Color.GREEN)
        else:
            Color.print_colored(f"❌ {result.message}", Color.RED)
        
        self.config_manager.show_status()
    
    def handle_hooks_command(self, language: Optional[str] = None, action: Optional[str] = None) -> None:
        """处理 hooks 相关命令"""
        # 如果指定了语言，则处理语言级别的 hooks 控制
        if language and language in self.config_manager.SUPPORTED_LANGUAGES:
            return self.handle_language_hooks_command(language, action)
        
        # 特殊处理 status 命令
        if language == "status":
            self.show_hooks_status()
            return
        
        # 如果 language 不是支持的语言且不是 None，则视为 action
        if language and language not in self.config_manager.SUPPORTED_LANGUAGES:
            action = language
            language = None
        
        # 全局 hooks 控制
        if action is None or action == "toggle":
            # 切换 hooks
            if self.config_manager.check_hooks_status():
                result = self.config_manager.disable_hooks()
            else:
                result = self.config_manager.enable_hooks()
        elif action in ["on", "enable"]:
            if self.config_manager.check_hooks_status():
                Color.print_colored("ℹ️  Hooks 已经启用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.enable_hooks()
        elif action in ["off", "disable"]:
            if not self.config_manager.check_hooks_status():
                Color.print_colored("ℹ️  Hooks 已经禁用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.disable_hooks()
        elif action == "status":
            self.show_hooks_status()
            return
        else:
            if language:
                Color.print_colored(f"❌ 错误：不支持的语言 '{language}'", Color.RED)
                Color.print_colored(f"   支持的语言: {', '.join(self.config_manager.SUPPORTED_LANGUAGES)}", Color.YELLOW)
            else:
                Color.print_colored(f"❌ 错误：未知的 hooks 操作 '{action}'", Color.RED)
            print("   使用 'claude-config.py help' 查看帮助")
            return
        
        # 显示结果
        if result.success:
            Color.print_colored(f"✅ {result.message}", Color.GREEN)
        else:
            Color.print_colored(f"❌ {result.message}", Color.RED)
        
        self.config_manager.show_status()
    
    def handle_language_hooks_command(self, language: str, action: Optional[str] = None) -> None:
        """处理语言级别的 hooks 命令"""
        if action is None or action == "toggle":
            # 切换语言 hooks
            current_status = self.config_manager.check_language_hook_status(language)
            result = self.config_manager.set_language_hook_status(language, not current_status)
        elif action in ["on", "enable"]:
            if self.config_manager.check_language_hook_status(language):
                Color.print_colored(f"ℹ️  {language} hooks 已经启用", Color.YELLOW)
                self.show_hooks_status()
                return
            else:
                result = self.config_manager.set_language_hook_status(language, True)
        elif action in ["off", "disable"]:
            if not self.config_manager.check_language_hook_status(language):
                Color.print_colored(f"ℹ️  {language} hooks 已经禁用", Color.YELLOW)
                self.show_hooks_status()
                return
            else:
                result = self.config_manager.set_language_hook_status(language, False)
        elif action == "status":
            self.show_hooks_status()
            return
        else:
            Color.print_colored(f"❌ 错误：未知的 {language} hooks 操作 '{action}'", Color.RED)
            print("   使用 'claude-config.py help' 查看帮助")
            return
        
        # 显示结果
        if result.success:
            Color.print_colored(f"✅ {result.message}", Color.GREEN)
        else:
            Color.print_colored(f"❌ {result.message}", Color.RED)
        
        self.show_hooks_status()
    
    def show_hooks_status(self) -> None:
        """显示详细的 hooks 状态"""
        print(f"\\n{Color.BLUE}🪝 Hooks 详细状态：{Color.NC}")
        print("========================")
        
        # 全局 hooks 状态
        if self.config_manager.check_hooks_status():
            print(f"\\n全局状态：{Color.GREEN}✅ 已启用{Color.NC}")
        else:
            print(f"\\n全局状态：{Color.RED}❌ 已禁用{Color.NC}")
            print("\\n要启用 hooks，请运行：claude-config.py hooks on")
            return
        
        # 语言级别的 hooks 状态
        print(f"\\n{Color.YELLOW}语言级别控制：{Color.NC}")
        lang_status = self.config_manager.get_all_language_hook_status()
        for lang in self.config_manager.SUPPORTED_LANGUAGES:
            status = lang_status[lang]
            if status:
                print(f"   {Color.GREEN}{lang:<12}: ✅ 启用{Color.NC}")
            else:
                print(f"   {Color.RED}{lang:<12}: ❌ 禁用{Color.NC}")
        
        print(f"\\n{Color.YELLOW}使用示例：{Color.NC}")
        print("   claude-config.py hooks go off      # 禁用 Go hooks")
        print("   claude-config.py hooks python on   # 启用 Python hooks")
        print("")
    
    def handle_deepseek_command(self, action: Optional[str] = None) -> None:
        """处理 DeepSeek 相关命令"""
        if action is None or action == "toggle":
            # 切换 DeepSeek
            if self.config_manager.check_deepseek_status():
                result = self.config_manager.disable_deepseek()
            else:
                result = self.config_manager.enable_deepseek()
        elif action in ["on", "enable"]:
            if self.config_manager.check_deepseek_status():
                Color.print_colored("ℹ️  DeepSeek 配置已经启用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.enable_deepseek()
        elif action in ["off", "disable"]:
            if not self.config_manager.check_deepseek_status():
                Color.print_colored("ℹ️  DeepSeek 配置已经禁用", Color.YELLOW)
                self.config_manager.show_status()
                return
            else:
                result = self.config_manager.disable_deepseek()
        elif action in ["reset", "clear-key"]:
            result = self.config_manager.clear_api_key()
        else:
            Color.print_colored(f"❌ 错误：未知的 deepseek 操作 '{action}'", Color.RED)
            print("   使用 'claude-config.py help' 查看帮助")
            return
        
        # 显示结果
        if result.success:
            if result.skipped:
                Color.print_colored(f"ℹ️  {result.message}", Color.YELLOW)
            else:
                Color.print_colored(f"✅ {result.message}", Color.GREEN)
        else:
            Color.print_colored(f"❌ {result.message}", Color.RED)
        
        self.config_manager.show_status()
    
    def show_help(self) -> None:
        """显示帮助信息"""
        Color.print_colored("Claude 配置管理工具", Color.BLUE)
        print("====================")
        print("")
        print("用法：")
        Color.print_colored("  claude-config.py                    # 显示当前状态", Color.GREEN)
        Color.print_colored("  claude-config.py status             # 显示当前状态", Color.GREEN)
        print("")
        print("文件复制：")
        Color.print_colored("  claude-config.py copy               # 复制所有配置文件", Color.GREEN)
        Color.print_colored("  claude-config.py copy --agents      # 仅复制agents目录", Color.GREEN)
        Color.print_colored("  claude-config.py copy --commands    # 仅复制commands目录", Color.GREEN)
        Color.print_colored("  claude-config.py copy --agents --commands  # 复制agents和commands", Color.GREEN)
        print("")
        print("代理管理：")
        Color.print_colored("  claude-config.py proxy              # 切换代理（开/关）", Color.GREEN)
        Color.print_colored("  claude-config.py proxy on           # 启用代理", Color.GREEN)
        Color.print_colored("  claude-config.py proxy off          # 禁用代理", Color.GREEN)
        print("")
        print("Hooks 管理：")
        Color.print_colored("  claude-config.py hooks              # 切换 hooks（开/关）", Color.GREEN)
        Color.print_colored("  claude-config.py hooks on           # 启用 hooks", Color.GREEN)
        Color.print_colored("  claude-config.py hooks off          # 禁用 hooks", Color.GREEN)
        Color.print_colored("  claude-config.py hooks status       # 显示 hooks 详细状态", Color.GREEN)
        print("")
        print("语言级别 Hooks 控制：")
        Color.print_colored("  claude-config.py hooks go off       # 禁用 Go hooks", Color.GREEN)
        Color.print_colored("  claude-config.py hooks python on    # 启用 Python hooks", Color.GREEN)
        Color.print_colored("  claude-config.py hooks javascript off # 禁用 JavaScript hooks", Color.GREEN)
        Color.print_colored("  claude-config.py hooks rust on      # 启用 Rust hooks", Color.GREEN)
        Color.print_colored("  支持语言: go, python, javascript, rust, nix, tilt", Color.YELLOW)
        print("")
        print("DeepSeek 配置管理：")
        Color.print_colored("  claude-config.py deepseek           # 切换 DeepSeek 配置（开/关）", Color.GREEN)
        Color.print_colored("  claude-config.py deepseek on        # 启用 DeepSeek 配置", Color.GREEN)
        Color.print_colored("  claude-config.py deepseek off       # 禁用 DeepSeek 配置", Color.GREEN)
        Color.print_colored("  claude-config.py deepseek reset     # 清除保存的 API 密钥", Color.GREEN)
        print("")
        print("其他：")
        Color.print_colored("  claude-config.py backup             # 备份当前配置", Color.GREEN)
        Color.print_colored("  claude-config.py help               # 显示此帮助", Color.GREEN)
        print("")
        print(f"配置文件：{self.config_manager.settings_file}")
        print(f"代理地址：{self.config_manager.proxy_host}")


def parse_args():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description='Claude 配置管理统一工具',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
使用示例:
  claude-config.py                         # 显示当前状态
  claude-config.py copy                    # 复制所有配置文件
  claude-config.py copy --agents           # 仅复制agents目录
  claude-config.py copy --commands         # 仅复制commands目录
  claude-config.py copy --agents --commands # 复制agents和commands目录
  claude-config.py proxy on                # 启用代理
  claude-config.py hooks off               # 禁用hooks
  claude-config.py hooks go off            # 禁用Go hooks
  claude-config.py hooks python on         # 启用Python hooks
  claude-config.py hooks status            # 显示hooks详细状态
  claude-config.py deepseek reset          # 清除API密钥
        '''
    )
    
    # 主命令
    parser.add_argument(
        'command',
        nargs='?',
        choices=['copy', 'proxy', 'hooks', 'deepseek', 'status', 'backup', 'help'],
        default='status',
        help='要执行的命令'
    )
    
    # 对于 hooks 命令，第一个参数可能是语言名或操作
    parser.add_argument(
        'language_or_action',
        nargs='?',
        help='语言名称或操作 (go/python/javascript/rust/nix/tilt/on/off/status等)'
    )
    
    # 第二个参数是具体操作
    parser.add_argument(
        'action',
        nargs='?',
        help='命令的具体操作 (on/off/toggle/reset等)'
    )
    
    # copy 命令的选项
    parser.add_argument(
        '--agents',
        action='store_true',
        help='仅复制agents目录（可与--commands同时使用）'
    )
    
    parser.add_argument(
        '--commands',
        action='store_true',
        help='仅复制commands目录（可与--agents同时使用）'
    )
    
    return parser.parse_args()


def main():
    """主函数"""
    try:
        args = parse_args()
        tool = ClaudeConfigTool()
        
        if args.command == 'copy':
            success = tool.copy_files(agents=args.agents, commands=args.commands)
            sys.exit(0 if success else 1)
        
        elif args.command == 'proxy':
            # 对于非 hooks 命令，language_or_action 就是 action
            tool.handle_proxy_command(args.language_or_action)
        
        elif args.command == 'hooks':
            # 处理 hooks 命令，可能包含语言参数
            tool.handle_hooks_command(args.language_or_action, args.action)
        
        elif args.command == 'deepseek':
            tool.handle_deepseek_command(args.language_or_action)
        
        elif args.command == 'backup':
            result = tool.config_manager.backup_config()
            if result.success:
                Color.print_colored(f"✅ {result.message}", Color.GREEN)
            else:
                Color.print_colored(f"❌ {result.message}", Color.RED)
        
        elif args.command == 'help':
            tool.show_help()
        
        elif args.command == 'status' or args.command is None:
            tool.config_manager.show_status()
        
        else:
            Color.print_colored(f"❌ 错误：未知命令 '{args.command}'", Color.RED)
            print("   使用 'claude-config.py help' 查看帮助")
            sys.exit(1)
        
    except KeyboardInterrupt:
        Color.print_colored("\\n\\n用户中断操作", Color.YELLOW)
        sys.exit(1)
    except Exception as e:
        Color.print_colored(f"运行失败: {e}", Color.RED)
        sys.exit(1)


if __name__ == "__main__":
    main()