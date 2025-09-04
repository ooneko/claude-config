#!/usr/bin/env python3
"""
utils.file_operations 模块核心功能测试
"""

import unittest
import tempfile
import json
import shutil
from pathlib import Path

from utils.file_operations import SettingsJsonMerger, FileOperations
from utils.common import OperationResult


class TestSettingsJsonMerger(unittest.TestCase):
    """SettingsJsonMerger 核心功能测试"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.temp_path = Path(self.temp_dir)
    
    def tearDown(self):
        """清理测试环境"""
        shutil.rmtree(self.temp_dir)
    
    def test_should_preserve_proxy_config(self):
        """测试是否应该保留代理配置"""
        # 包含代理配置
        data_with_proxy = {"env": {"http_proxy": "http://127.0.0.1:7890"}}
        self.assertTrue(SettingsJsonMerger.should_preserve_proxy_config(data_with_proxy))
        
        # 不包含代理配置
        data_without_proxy = {"some_key": "value"}
        self.assertFalse(SettingsJsonMerger.should_preserve_proxy_config(data_without_proxy))
    
    def test_filter_proxy_from_source(self):
        """测试从源数据中过滤代理配置"""
        source_data = {
            "env": {
                "http_proxy": "http://127.0.0.1:7890",
                "https_proxy": "http://127.0.0.1:7890",
                "OTHER_VAR": "value"
            },
            "other_config": "value"
        }
        
        result = SettingsJsonMerger.filter_proxy_from_source(source_data)
        
        # 验证代理配置被移除
        self.assertNotIn("http_proxy", result["env"])
        self.assertNotIn("https_proxy", result["env"])
        # 验证其他配置保留
        self.assertIn("OTHER_VAR", result["env"])
        self.assertEqual(result["other_config"], "value")
    
    def test_deep_merge_dict_basic(self):
        """测试基础字典深度合并"""
        target = {"a": 1, "b": {"c": 2}}
        source = {"b": {"d": 3}, "e": 4}
        
        result = SettingsJsonMerger.deep_merge_dict(target, source)
        
        expected = {"a": 1, "b": {"c": 2, "d": 3}, "e": 4}
        self.assertEqual(result, expected)
    
    def test_merge_hooks_different_matchers(self):
        """测试合并不同matcher的hooks"""
        target_hooks = {
            "PostToolUse": [
                {"matcher": "Write", "hooks": [{"command": "cmd1"}]}
            ]
        }
        source_hooks = {
            "PostToolUse": [
                {"matcher": "Edit", "hooks": [{"command": "cmd2"}]}
            ]
        }
        
        result = SettingsJsonMerger.merge_hooks(target_hooks, source_hooks)
        
        # 应该包含两个matcher
        self.assertEqual(len(result["PostToolUse"]), 2)
    
    def test_merge_settings_preserve_proxy(self):
        """测试合并settings.json并保留代理配置"""
        # 创建源文件（包含代理配置）
        source_file = self.temp_path / "source.json"
        source_data = {
            "env": {"http_proxy": "http://127.0.0.1:7890"},
            "new_config": "value"
        }
        with open(source_file, 'w', encoding='utf-8') as f:
            json.dump(source_data, f)
        
        # 创建目标文件（已有代理配置）
        target_file = self.temp_path / "target.json"
        target_data = {
            "env": {"http_proxy": "http://127.0.0.1:8080"},
            "existing_config": "value"
        }
        with open(target_file, 'w', encoding='utf-8') as f:
            json.dump(target_data, f)
        
        result = SettingsJsonMerger.merge_settings(target_file, source_file)
        
        self.assertTrue(result.success)
        
        # 验证合并后的文件
        with open(target_file, 'r', encoding='utf-8') as f:
            merged_data = json.load(f)
        
        # 原有代理配置应该被保留
        self.assertEqual(merged_data["env"]["http_proxy"], "http://127.0.0.1:8080")
        # 新配置应该被添加
        self.assertEqual(merged_data["new_config"], "value")


class TestFileOperations(unittest.TestCase):
    """FileOperations 核心功能测试"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.source_dir = Path(self.temp_dir) / "source"
        self.target_dir = Path(self.temp_dir) / "target"
        self.source_dir.mkdir()
        self.target_dir.mkdir()
    
    def tearDown(self):
        """清理测试环境"""
        shutil.rmtree(self.temp_dir)
    
    def test_copy_file_basic(self):
        """测试基础文件复制"""
        file_ops = FileOperations(self.source_dir, self.target_dir)
        
        # 创建源文件
        src_file = self.source_dir / "test.txt"
        src_file.write_text("测试内容", encoding='utf-8')
        
        dest_file = self.target_dir / "test.txt"
        
        result = file_ops.copy_file(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertTrue(dest_file.exists())
        self.assertEqual(dest_file.read_text(encoding='utf-8'), "测试内容")
    
    def test_copy_directory_basic(self):
        """测试基础目录复制"""
        file_ops = FileOperations(self.source_dir, self.target_dir)
        
        # 创建源目录结构
        src_dir = self.source_dir / "subdir"
        src_dir.mkdir()
        (src_dir / "file1.txt").write_text("内容1", encoding='utf-8')
        (src_dir / "file2.txt").write_text("内容2", encoding='utf-8')
        
        dest_dir = self.target_dir / "subdir"
        
        result = file_ops.copy_directory(src_dir, dest_dir)
        
        self.assertTrue(result.success)
        self.assertTrue(dest_dir.exists())
        self.assertTrue((dest_dir / "file1.txt").exists())
        self.assertTrue((dest_dir / "file2.txt").exists())
    
    def test_run_copy_operation_selected_items(self):
        """测试运行复制操作 - 选择特定项目"""
        # 创建测试文件
        (self.source_dir / "agents").mkdir()
        (self.source_dir / "agents" / "test-agent.md").write_text("代理配置", encoding='utf-8')
        
        (self.source_dir / "commands").mkdir()
        (self.source_dir / "commands" / "test-cmd.md").write_text("命令配置", encoding='utf-8')
        
        file_ops = FileOperations(self.source_dir, self.target_dir, ["agents", "commands"])
        result = file_ops.run_copy_operation()
        
        self.assertTrue(result)
        self.assertTrue((self.target_dir / "agents" / "test-agent.md").exists())
        self.assertTrue((self.target_dir / "commands" / "test-cmd.md").exists())


if __name__ == '__main__':
    unittest.main()