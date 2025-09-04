#!/usr/bin/env python3
"""
Claude 配置管理模块
"""

import json
import sys
from pathlib import Path
from typing import Dict, Any, Optional

from .common import Color, OperationResult


class ConfigManager:
    """Claude 配置管理器"""
    
    def __init__(self, claude_dir: Path):
        self.claude_dir = claude_dir
        self.settings_file = claude_dir / "settings.json"
        self.api_key_file = claude_dir / ".deepseek_api_key"
        
        # DeepSeek 默认配置
        self.anthropic_base_url = "https://api.deepseek.com/anthropic"
        self.anthropic_model = "deepseek-chat"
        self.anthropic_small_fast_model = "deepseek-chat"
        
        # 代理设置
        self.proxy_host = "http://127.0.0.1:7890"
    
    def _ensure_settings_exists(self) -> bool:
        """确保设置文件存在"""
        if not self.settings_file.exists():
            Color.print_colored(f"❌ 错误：找不到 {self.settings_file}", Color.RED)
            print("   请先创建 Claude 设置文件")
            return False
        return True
    
    def _backup_settings(self) -> None:
        """备份设置文件"""
        from datetime import datetime
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_file = f"{self.settings_file}.backup.{timestamp}"
        try:
            import shutil
            shutil.copy2(self.settings_file, backup_file)
            Color.print_colored(f"✅ 已备份配置文件到: {backup_file}", Color.GREEN)
        except Exception as e:
            Color.print_colored(f"备份失败: {e}", Color.RED)
    
    def _load_settings(self) -> Dict[str, Any]:
        """加载设置文件"""
        try:
            with open(self.settings_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            Color.print_colored(f"加载设置文件失败: {e}", Color.RED)
            return {}
    
    def _save_settings(self, settings: Dict[str, Any]) -> bool:
        """保存设置文件"""
        try:
            with open(self.settings_file, 'w', encoding='utf-8') as f:
                json.dump(settings, f, indent=2, ensure_ascii=False)
            return True
        except Exception as e:
            Color.print_colored(f"保存设置文件失败: {e}", Color.RED)
            return False
    
    # ===== 代理相关方法 =====
    def check_proxy_status(self) -> bool:
        """检查代理状态"""
        if not self._ensure_settings_exists():
            return False
        
        settings = self._load_settings()
        env = settings.get('env', {})
        return 'http_proxy' in env
    
    def enable_proxy(self) -> OperationResult:
        """启用代理"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        settings = self._load_settings()
        hooks_config = settings.get('hooks', {})
        
        # 更新配置，保留 hooks
        settings['env'] = {
            'http_proxy': self.proxy_host,
            'https_proxy': self.proxy_host
        }
        settings['hooks'] = hooks_config
        
        if self._save_settings(settings):
            return OperationResult(True, f"已启用代理模式 ({self.proxy_host})")
        else:
            return OperationResult(False, "启用代理失败")
    
    def disable_proxy(self) -> OperationResult:
        """禁用代理"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        settings = self._load_settings()
        if 'env' in settings:
            del settings['env']
        
        if self._save_settings(settings):
            return OperationResult(True, "已禁用代理模式")
        else:
            return OperationResult(False, "禁用代理失败")
    
    # ===== DeepSeek 相关方法 =====
    def check_deepseek_status(self) -> bool:
        """检查 DeepSeek 状态"""
        if not self._ensure_settings_exists():
            return False
        
        settings = self._load_settings()
        env = settings.get('env', {})
        return 'ANTHROPIC_AUTH_TOKEN' in env
    
    def _get_api_key(self) -> Optional[str]:
        """获取 API 密钥"""
        if self.api_key_file.exists():
            try:
                with open(self.api_key_file, 'r', encoding='utf-8') as f:
                    return f.read().strip()
            except Exception:
                return None
        
        # 首次使用，提示输入 API 密钥
        if sys.stdin.isatty():
            Color.print_colored("首次使用 DeepSeek 配置，请输入 API 密钥：", Color.YELLOW)
            api_key = input().strip()
            if api_key:
                try:
                    self.claude_dir.mkdir(parents=True, exist_ok=True)
                    with open(self.api_key_file, 'w', encoding='utf-8') as f:
                        f.write(api_key)
                    import os
                    os.chmod(self.api_key_file, 0o600)
                    return api_key
                except Exception as e:
                    Color.print_colored(f"保存 API 密钥失败: {e}", Color.RED)
        
        return None
    
    def enable_deepseek(self) -> OperationResult:
        """启用 DeepSeek 配置"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        api_token = self._get_api_key()
        if not api_token:
            return OperationResult(False, "获取 API 密钥失败")
        
        self._backup_settings()
        
        settings = self._load_settings()
        current_env = settings.get('env', {})
        
        # 更新配置，添加 DeepSeek 相关环境变量
        current_env.update({
            'ANTHROPIC_AUTH_TOKEN': api_token,
            'ANTHROPIC_BASE_URL': self.anthropic_base_url,
            'ANTHROPIC_MODEL': self.anthropic_model,
            'ANTHROPIC_SMALL_FAST_MODEL': self.anthropic_small_fast_model
        })
        settings['env'] = current_env
        
        if self._save_settings(settings):
            return OperationResult(True, f"已启用 DeepSeek 配置\\n   ANTHROPIC_AUTH_TOKEN: {api_token[:10]}...\\n   ANTHROPIC_BASE_URL: {self.anthropic_base_url}")
        else:
            return OperationResult(False, "启用 DeepSeek 配置失败")
    
    def disable_deepseek(self) -> OperationResult:
        """禁用 DeepSeek 配置"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        self._backup_settings()
        
        settings = self._load_settings()
        env = settings.get('env', {})
        
        # 删除 DeepSeek 相关环境变量
        deepseek_keys = ['ANTHROPIC_AUTH_TOKEN', 'ANTHROPIC_BASE_URL', 'ANTHROPIC_MODEL', 'ANTHROPIC_SMALL_FAST_MODEL']
        for key in deepseek_keys:
            env.pop(key, None)
        
        if env:
            settings['env'] = env
        elif 'env' in settings:
            del settings['env']
        
        if self._save_settings(settings):
            return OperationResult(True, "已禁用 DeepSeek 配置\\n   (API 密钥已保留，重新启用时无需再次输入)")
        else:
            return OperationResult(False, "禁用 DeepSeek 配置失败")
    
    def clear_api_key(self) -> OperationResult:
        """清除保存的 API 密钥"""
        try:
            if self.api_key_file.exists():
                self.api_key_file.unlink()
                # 如果当前启用了 DeepSeek，也禁用它
                if self.check_deepseek_status():
                    self.disable_deepseek()
                return OperationResult(True, "已清除保存的 API 密钥\\n   下次启用时需重新输入")
            else:
                return OperationResult(True, "没有找到保存的 API 密钥", skipped=True)
        except Exception as e:
            return OperationResult(False, f"清除 API 密钥失败: {e}")
    
    # ===== Hooks 相关方法 =====
    
    # 支持的语言列表
    SUPPORTED_LANGUAGES = ['go', 'python', 'javascript', 'rust', 'nix', 'tilt']
    
    def check_hooks_status(self) -> bool:
        """检查 hooks 状态"""
        if not self._ensure_settings_exists():
            return False
        
        settings = self._load_settings()
        hooks = settings.get('hooks', {})
        return len(hooks) > 0
    
    def enable_hooks(self) -> OperationResult:
        """启用 hooks"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        # 默认的 hooks 配置
        default_hooks = {
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
        }
        
        settings = self._load_settings()
        backup_file = self.settings_file.with_suffix('.json.hooks_backup')
        
        # 检查是否有备份的 hooks 配置
        if backup_file.exists():
            Color.print_colored("   发现备份的 hooks 配置，正在恢复...", Color.YELLOW)
            try:
                with open(backup_file, 'r', encoding='utf-8') as f:
                    hooks_config = json.load(f)
            except Exception:
                hooks_config = default_hooks
        else:
            Color.print_colored("   使用默认 hooks 配置...", Color.YELLOW)
            hooks_config = default_hooks
        
        settings['hooks'] = hooks_config
        
        if self._save_settings(settings):
            return OperationResult(True, "已启用 hooks")
        else:
            return OperationResult(False, "启用 hooks 失败")
    
    def disable_hooks(self) -> OperationResult:
        """禁用 hooks"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        settings = self._load_settings()
        hooks = settings.get('hooks', {})
        
        # 先备份当前的 hooks 配置
        if hooks:
            backup_file = self.settings_file.with_suffix('.json.hooks_backup')
            try:
                with open(backup_file, 'w', encoding='utf-8') as f:
                    json.dump(hooks, f, indent=2, ensure_ascii=False)
            except Exception as e:
                Color.print_colored(f"备份 hooks 配置失败: {e}", Color.RED)
        
        # 删除 hooks 配置
        if 'hooks' in settings:
            del settings['hooks']
        
        if self._save_settings(settings):
            return OperationResult(True, "已禁用 hooks\\n   (hooks 配置已备份)")
        else:
            return OperationResult(False, "禁用 hooks 失败")
    
    # ===== 语言级别 Hooks 控制 =====
    
    def check_language_hook_status(self, language: str) -> bool:
        """检查特定语言的 hook 状态"""
        if not self._ensure_settings_exists():
            return True  # 默认启用
        
        if language not in self.SUPPORTED_LANGUAGES:
            return False
        
        settings = self._load_settings()
        env = settings.get('env', {})
        env_key = f"CLAUDE_HOOKS_{language.upper()}_ENABLED"
        return env.get(env_key, "true").lower() == "true"
    
    def set_language_hook_status(self, language: str, enabled: bool) -> OperationResult:
        """设置特定语言的 hook 状态"""
        if not self._ensure_settings_exists():
            return OperationResult(False, "设置文件不存在")
        
        if language not in self.SUPPORTED_LANGUAGES:
            return OperationResult(False, f"不支持的语言: {language}\\n   支持的语言: {', '.join(self.SUPPORTED_LANGUAGES)}")
        
        settings = self._load_settings()
        env = settings.get('env', {})
        env_key = f"CLAUDE_HOOKS_{language.upper()}_ENABLED"
        
        # 设置状态
        env[env_key] = "true" if enabled else "false"
        settings['env'] = env
        
        if self._save_settings(settings):
            status_text = "启用" if enabled else "禁用"
            return OperationResult(True, f"已{status_text} {language} hooks")
        else:
            status_text = "启用" if enabled else "禁用"
            return OperationResult(False, f"{status_text} {language} hooks 失败")
    
    def get_all_language_hook_status(self) -> Dict[str, bool]:
        """获取所有语言的 hook 状态"""
        result = {}
        for lang in self.SUPPORTED_LANGUAGES:
            result[lang] = self.check_language_hook_status(lang)
        return result
    
    def show_status(self) -> None:
        """显示配置状态"""
        print(f"\\n{Color.BLUE}📊 Claude 配置状态：{Color.NC}")
        print("====================")
        
        # 代理状态
        print(f"\\n{Color.YELLOW}🌐 代理状态：{Color.NC}")
        if self.check_proxy_status():
            print(f"   {Color.GREEN}✅ 已启用{Color.NC}")
            print(f"   代理地址：{self.proxy_host}")
        else:
            print("   ⚫ 已禁用")
        
        # DeepSeek 状态
        print(f"\\n{Color.YELLOW}🤖 DeepSeek 状态：{Color.NC}")
        if self.check_deepseek_status():
            print(f"   {Color.GREEN}✅ 已启用{Color.NC}")
            settings = self._load_settings()
            env = settings.get('env', {})
            token = env.get('ANTHROPIC_AUTH_TOKEN', '未设置')
            if token != '未设置':
                token = token[:10] + "..."
            print(f"   ANTHROPIC_AUTH_TOKEN: {token}")
            print(f"   ANTHROPIC_BASE_URL: {env.get('ANTHROPIC_BASE_URL', '未设置')}")
            print(f"   ANTHROPIC_MODEL: {env.get('ANTHROPIC_MODEL', '未设置')}")
        else:
            print("   ⚫ 已禁用")
        
        # Hooks 状态
        print(f"\\n{Color.YELLOW}🪝 Hooks 状态：{Color.NC}")
        if self.check_hooks_status():
            print(f"   {Color.GREEN}✅ 已启用{Color.NC}")
            settings = self._load_settings()
            hooks = settings.get('hooks', {})
            post_tool_count = len(hooks.get('PostToolUse', [{}])[0].get('hooks', []))
            stop_count = len(hooks.get('Stop', [{}])[0].get('hooks', []))
            print(f"   PostToolUse hooks: {post_tool_count} 个")
            print(f"   Stop hooks: {stop_count} 个")
            
            # 显示语言级别的 hooks 状态
            print(f"\\n   语言级别控制：")
            env = settings.get('env', {})
            for lang in self.SUPPORTED_LANGUAGES:
                env_key = f"CLAUDE_HOOKS_{lang.upper()}_ENABLED"
                status = env.get(env_key, "true")  # 默认启用
                if status.lower() == "true":
                    print(f"      {Color.GREEN}{lang}: ✅{Color.NC}")
                else:
                    print(f"      {Color.RED}{lang}: ❌{Color.NC}")
        else:
            print("   ⚫ 已禁用")
        
        print("")
    
    def backup_config(self) -> OperationResult:
        """备份配置"""
        try:
            self._backup_settings()
            return OperationResult(True, "已备份配置文件")
        except Exception as e:
            return OperationResult(False, f"备份失败: {e}")