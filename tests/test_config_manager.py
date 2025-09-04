#!/usr/bin/env python3
"""
utils.config_manager 模块测试
"""

import unittest
import tempfile
import json
import os
from pathlib import Path
from unittest.mock import patch, mock_open, MagicMock

from utils.config_manager import ConfigManager
from utils.common import OperationResult, Color


class TestConfigManager(unittest.TestCase):
    """ConfigManager 类测试"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.claude_dir = Path(self.temp_dir)
        self.config_manager = ConfigManager(self.claude_dir)
        
        # 创建测试用的 settings.json
        self.test_settings = {
            "some_setting": "value"
        }
    
    def tearDown(self):
        """清理测试环境"""
        import shutil
        shutil.rmtree(self.temp_dir)
    
    def _create_settings_file(self, settings=None):
        """创建设置文件"""
        if settings is None:
            settings = self.test_settings
        with open(self.config_manager.settings_file, 'w', encoding='utf-8') as f:
            json.dump(settings, f, indent=2)
    
    def test_init(self):
        """测试初始化"""
        self.assertEqual(self.config_manager.claude_dir, self.claude_dir)
        self.assertEqual(self.config_manager.settings_file, self.claude_dir / "settings.json")
        self.assertEqual(self.config_manager.api_key_file, self.claude_dir / ".deepseek_api_key")
        self.assertEqual(self.config_manager.anthropic_base_url, "https://api.deepseek.com/anthropic")
        self.assertEqual(self.config_manager.anthropic_model, "deepseek-chat")
        self.assertEqual(self.config_manager.proxy_host, "http://127.0.0.1:7890")
    
    def test_ensure_settings_exists_true(self):
        """测试确保设置文件存在 - 存在"""
        self._create_settings_file()
        self.assertTrue(self.config_manager._ensure_settings_exists())
    
    @patch.object(Color, 'print_colored')
    @patch('builtins.print')
    def test_ensure_settings_exists_false(self, mock_print, mock_color_print):
        """测试确保设置文件存在 - 不存在"""
        self.assertFalse(self.config_manager._ensure_settings_exists())
        mock_color_print.assert_called_with(f"❌ 错误：找不到 {self.config_manager.settings_file}", Color.RED)
    
    @patch('shutil.copy2')
    @patch.object(Color, 'print_colored')
    def test_backup_settings_success(self, mock_color_print, mock_copy):
        """测试备份设置文件 - 成功"""
        self._create_settings_file()
        
        self.config_manager._backup_settings()
        
        mock_copy.assert_called_once()
        mock_color_print.assert_called()
        args = mock_color_print.call_args[0]
        self.assertTrue(args[0].startswith("✅ 已备份配置文件到:"))
    
    @patch('shutil.copy2', side_effect=Exception("备份失败"))
    @patch.object(Color, 'print_colored')
    def test_backup_settings_failure(self, mock_color_print, mock_copy):
        """测试备份设置文件 - 失败"""
        self._create_settings_file()
        
        self.config_manager._backup_settings()
        mock_color_print.assert_called_with("备份失败: 备份失败", Color.RED)
    
    def test_load_settings_success(self):
        """测试加载设置文件 - 成功"""
        self._create_settings_file()
        settings = self.config_manager._load_settings()
        self.assertEqual(settings, self.test_settings)
    
    @patch.object(Color, 'print_colored')
    def test_load_settings_failure(self, mock_color_print):
        """测试加载设置文件 - 失败"""
        # 创建无效JSON文件
        with open(self.config_manager.settings_file, 'w') as f:
            f.write("无效的JSON")
        
        settings = self.config_manager._load_settings()
        self.assertEqual(settings, {})
        mock_color_print.assert_called()
        args = mock_color_print.call_args[0]
        self.assertTrue(args[0].startswith("加载设置文件失败:"))
    
    def test_save_settings_success(self):
        """测试保存设置文件 - 成功"""
        result = self.config_manager._save_settings(self.test_settings)
        self.assertTrue(result)
        
        # 验证文件内容
        with open(self.config_manager.settings_file, 'r', encoding='utf-8') as f:
            saved_settings = json.load(f)
        self.assertEqual(saved_settings, self.test_settings)
    
    @patch('builtins.open', side_effect=Exception("写入失败"))
    @patch.object(Color, 'print_colored')
    def test_save_settings_failure(self, mock_color_print, mock_open):
        """测试保存设置文件 - 失败"""
        result = self.config_manager._save_settings(self.test_settings)
        self.assertFalse(result)
        mock_color_print.assert_called_with("保存设置文件失败: 写入失败", Color.RED)
    
    def test_check_proxy_status_enabled(self):
        """测试检查代理状态 - 已启用"""
        settings_with_proxy = {
            "env": {
                "http_proxy": "http://127.0.0.1:7890"
            }
        }
        self._create_settings_file(settings_with_proxy)
        self.assertTrue(self.config_manager.check_proxy_status())
    
    def test_check_proxy_status_disabled(self):
        """测试检查代理状态 - 已禁用"""
        self._create_settings_file()
        self.assertFalse(self.config_manager.check_proxy_status())
    
    def test_enable_proxy_success(self):
        """测试启用代理 - 成功"""
        self._create_settings_file()
        result = self.config_manager.enable_proxy()
        
        self.assertTrue(result.success)
        self.assertEqual(result.message, f"已启用代理模式 ({self.config_manager.proxy_host})")
        
        # 验证设置文件
        settings = self.config_manager._load_settings()
        self.assertIn('env', settings)
        self.assertEqual(settings['env']['http_proxy'], self.config_manager.proxy_host)
        self.assertEqual(settings['env']['https_proxy'], self.config_manager.proxy_host)
    
    def test_enable_proxy_with_existing_hooks(self):
        """测试启用代理 - 保留现有hooks配置"""
        settings_with_hooks = {
            "hooks": {"some_hook": "value"}
        }
        self._create_settings_file(settings_with_hooks)
        
        result = self.config_manager.enable_proxy()
        self.assertTrue(result.success)
        
        # 验证hooks被保留
        settings = self.config_manager._load_settings()
        self.assertIn('hooks', settings)
        self.assertEqual(settings['hooks'], {"some_hook": "value"})
    
    def test_disable_proxy_success(self):
        """测试禁用代理 - 成功"""
        settings_with_proxy = {
            "env": {
                "http_proxy": "http://127.0.0.1:7890",
                "https_proxy": "http://127.0.0.1:7890"
            }
        }
        self._create_settings_file(settings_with_proxy)
        
        result = self.config_manager.disable_proxy()
        self.assertTrue(result.success)
        self.assertEqual(result.message, "已禁用代理模式")
        
        # 验证设置文件
        settings = self.config_manager._load_settings()
        self.assertNotIn('env', settings)
    
    def test_check_deepseek_status_enabled(self):
        """测试检查DeepSeek状态 - 已启用"""
        settings_with_deepseek = {
            "env": {
                "ANTHROPIC_AUTH_TOKEN": "test_token"
            }
        }
        self._create_settings_file(settings_with_deepseek)
        self.assertTrue(self.config_manager.check_deepseek_status())
    
    def test_check_deepseek_status_disabled(self):
        """测试检查DeepSeek状态 - 已禁用"""
        self._create_settings_file()
        self.assertFalse(self.config_manager.check_deepseek_status())
    
    def test_get_api_key_file_exists(self):
        """测试获取API密钥 - 文件存在"""
        test_key = "test_api_key"
        self.config_manager.api_key_file.write_text(test_key, encoding='utf-8')
        
        key = self.config_manager._get_api_key()
        self.assertEqual(key, test_key)
    
    def test_get_api_key_file_not_exists_not_tty(self):
        """测试获取API密钥 - 文件不存在且非TTY"""
        with patch('sys.stdin.isatty', return_value=False):
            key = self.config_manager._get_api_key()
            self.assertIsNone(key)
    
    @patch('os.chmod')
    @patch.object(Color, 'print_colored')
    @patch('builtins.input', return_value='test_input_key')
    @patch('sys.stdin.isatty', return_value=True)
    def test_get_api_key_input_success(self, mock_tty, mock_input, mock_color_print, mock_chmod):
        """测试获取API密钥 - 用户输入成功"""
        key = self.config_manager._get_api_key()
        self.assertEqual(key, 'test_input_key')
        
        # 验证文件是否保存
        saved_key = self.config_manager.api_key_file.read_text(encoding='utf-8')
        self.assertEqual(saved_key, 'test_input_key')
        
        # 验证文件权限
        mock_chmod.assert_called_once_with(self.config_manager.api_key_file, 0o600)
    
    @patch('builtins.input', return_value='')
    @patch('sys.stdin.isatty', return_value=True)
    def test_get_api_key_empty_input(self, mock_tty, mock_input):
        """测试获取API密钥 - 空输入"""
        key = self.config_manager._get_api_key()
        self.assertIsNone(key)
    
    def test_enable_deepseek_success(self):
        """测试启用DeepSeek - 成功"""
        self._create_settings_file()
        
        with patch.object(self.config_manager, '_get_api_key', return_value='test_key'), \
             patch.object(self.config_manager, '_backup_settings'):
            
            result = self.config_manager.enable_deepseek()
            self.assertTrue(result.success)
            self.assertIn("已启用 DeepSeek 配置", result.message)
            
            # 验证设置文件
            settings = self.config_manager._load_settings()
            env = settings['env']
            self.assertEqual(env['ANTHROPIC_AUTH_TOKEN'], 'test_key')
            self.assertEqual(env['ANTHROPIC_BASE_URL'], self.config_manager.anthropic_base_url)
    
    def test_enable_deepseek_no_api_key(self):
        """测试启用DeepSeek - 无API密钥"""
        self._create_settings_file()
        
        with patch.object(self.config_manager, '_get_api_key', return_value=None):
            result = self.config_manager.enable_deepseek()
            self.assertFalse(result.success)
            self.assertEqual(result.message, "获取 API 密钥失败")
    
    def test_disable_deepseek_success(self):
        """测试禁用DeepSeek - 成功"""
        settings_with_deepseek = {
            "env": {
                "ANTHROPIC_AUTH_TOKEN": "test_key",
                "ANTHROPIC_BASE_URL": "test_url",
                "http_proxy": "http://127.0.0.1:7890"
            }
        }
        self._create_settings_file(settings_with_deepseek)
        
        with patch.object(self.config_manager, '_backup_settings'):
            result = self.config_manager.disable_deepseek()
            self.assertTrue(result.success)
            self.assertIn("已禁用 DeepSeek 配置", result.message)
            
            # 验证DeepSeek相关配置被删除，但代理配置保留
            settings = self.config_manager._load_settings()
            env = settings['env']
            self.assertNotIn('ANTHROPIC_AUTH_TOKEN', env)
            self.assertNotIn('ANTHROPIC_BASE_URL', env)
            self.assertIn('http_proxy', env)
    
    def test_clear_api_key_file_exists(self):
        """测试清除API密钥 - 文件存在"""
        self.config_manager.api_key_file.write_text("test_key", encoding='utf-8')
        
        with patch.object(self.config_manager, 'check_deepseek_status', return_value=True), \
             patch.object(self.config_manager, 'disable_deepseek', return_value=OperationResult(True, "已禁用")):
            
            result = self.config_manager.clear_api_key()
            self.assertTrue(result.success)
            self.assertIn("已清除保存的 API 密钥", result.message)
            self.assertFalse(self.config_manager.api_key_file.exists())
    
    def test_clear_api_key_file_not_exists(self):
        """测试清除API密钥 - 文件不存在"""
        result = self.config_manager.clear_api_key()
        self.assertTrue(result.success)
        self.assertTrue(result.skipped)
        self.assertEqual(result.message, "没有找到保存的 API 密钥")
    
    def test_check_hooks_status_enabled(self):
        """测试检查hooks状态 - 已启用"""
        settings_with_hooks = {
            "hooks": {"PostToolUse": []}
        }
        self._create_settings_file(settings_with_hooks)
        self.assertTrue(self.config_manager.check_hooks_status())
    
    def test_check_hooks_status_disabled(self):
        """测试检查hooks状态 - 已禁用"""
        self._create_settings_file()
        self.assertFalse(self.config_manager.check_hooks_status())
    
    def test_enable_hooks_success(self):
        """测试启用hooks - 成功"""
        self._create_settings_file()
        result = self.config_manager.enable_hooks()
        
        self.assertTrue(result.success)
        self.assertEqual(result.message, "已启用 hooks")
        
        # 验证hooks配置
        settings = self.config_manager._load_settings()
        self.assertIn('hooks', settings)
        self.assertIn('PostToolUse', settings['hooks'])
        self.assertIn('Stop', settings['hooks'])
    
    @patch.object(Color, 'print_colored')
    def test_enable_hooks_with_backup(self, mock_color_print):
        """测试启用hooks - 从备份恢复"""
        self._create_settings_file()
        
        # 创建备份文件
        backup_hooks = {"PostToolUse": [{"matcher": "test"}]}
        backup_file = self.config_manager.settings_file.with_suffix('.json.hooks_backup')
        with open(backup_file, 'w', encoding='utf-8') as f:
            json.dump(backup_hooks, f)
        
        result = self.config_manager.enable_hooks()
        self.assertTrue(result.success)
        
        # 验证使用了备份配置
        settings = self.config_manager._load_settings()
        self.assertEqual(settings['hooks']['PostToolUse'][0]['matcher'], 'test')
    
    def test_disable_hooks_success(self):
        """测试禁用hooks - 成功"""
        settings_with_hooks = {
            "hooks": {"PostToolUse": []}
        }
        self._create_settings_file(settings_with_hooks)
        
        result = self.config_manager.disable_hooks()
        self.assertTrue(result.success)
        self.assertIn("已禁用 hooks", result.message)
        
        # 验证hooks被删除
        settings = self.config_manager._load_settings()
        self.assertNotIn('hooks', settings)
        
        # 验证备份文件被创建
        backup_file = self.config_manager.settings_file.with_suffix('.json.hooks_backup')
        self.assertTrue(backup_file.exists())
    
    def test_show_status(self):
        """测试显示状态"""
        self._create_settings_file()
        
        # 简单验证方法能够运行而不抛出异常
        try:
            self.config_manager.show_status()
        except Exception as e:
            self.fail(f"show_status 方法抛出了异常: {e}")
    
    def test_backup_config_success(self):
        """测试备份配置 - 成功"""
        self._create_settings_file()
        
        with patch.object(self.config_manager, '_backup_settings'):
            result = self.config_manager.backup_config()
            self.assertTrue(result.success)
            self.assertEqual(result.message, "已备份配置文件")
    
    def test_backup_config_failure(self):
        """测试备份配置 - 失败"""
        with patch.object(self.config_manager, '_backup_settings', side_effect=Exception("备份失败")):
            result = self.config_manager.backup_config()
            self.assertFalse(result.success)
            self.assertEqual(result.message, "备份失败: 备份失败")


if __name__ == '__main__':
    unittest.main()