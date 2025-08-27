#!/usr/bin/env python3
"""
copy-to-claude.py 的单元测试
"""

import json
import shutil
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

# 导入待测试的模块
from copy_to_claude import (
    Color, SettingsJsonMerger, ClaudeConfigCopier, CopyResult
)


class TestColor(unittest.TestCase):
    """测试颜色输出类"""
    
    def test_print_colored(self):
        """测试彩色打印不会抛出异常"""
        try:
            Color.print_colored("test", Color.GREEN)
        except Exception as e:
            self.fail(f"print_colored raised {e} unexpectedly")


class TestSettingsJsonMerger(unittest.TestCase):
    """测试settings.json合并器"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.temp_dir)
    
    def test_deep_merge_dict_simple(self):
        """测试简单字典合并"""
        target = {"a": 1, "b": 2}
        source = {"b": 3, "c": 4}
        
        result = SettingsJsonMerger.deep_merge_dict(target, source)
        expected = {"a": 1, "b": 3, "c": 4}
        
        self.assertEqual(result, expected)
    
    def test_deep_merge_dict_nested(self):
        """测试嵌套字典合并"""
        target = {"env": {"http_proxy": "old"}, "other": "keep"}
        source = {"env": {"https_proxy": "new"}, "new_key": "value"}
        
        result = SettingsJsonMerger.deep_merge_dict(target, source)
        expected = {
            "env": {"http_proxy": "old", "https_proxy": "new"},
            "other": "keep",
            "new_key": "value"
        }
        
        self.assertEqual(result, expected)
    
    def test_merge_hooks_different_matchers(self):
        """测试不同matcher的hooks合并"""
        target = {
            "PostToolUse": [
                {"matcher": "Write", "hooks": [{"type": "command", "command": "lint1"}]}
            ]
        }
        source = {
            "PostToolUse": [
                {"matcher": "Edit", "hooks": [{"type": "command", "command": "lint2"}]}
            ]
        }
        
        result = SettingsJsonMerger.merge_hooks(target, source)
        
        self.assertEqual(len(result["PostToolUse"]), 2)
        matchers = {config["matcher"] for config in result["PostToolUse"]}
        self.assertEqual(matchers, {"Write", "Edit"})
    
    def test_merge_settings_new_file(self):
        """测试合并到新文件"""
        source_file = Path(self.temp_dir) / "source.json"
        target_file = Path(self.temp_dir) / "target.json"
        
        source_data = {"includeCoAuthoredBy": True, "env": {"test": "value"}}
        
        with open(source_file, 'w') as f:
            json.dump(source_data, f)
        
        result = SettingsJsonMerger.merge_settings(target_file, source_file)
        
        self.assertTrue(result.success)
        self.assertTrue(target_file.exists())
        
        with open(target_file, 'r') as f:
            loaded_data = json.load(f)
        
        self.assertEqual(loaded_data, source_data)
    
    def test_merge_settings_existing_file(self):
        """测试合并到已存在文件"""
        source_file = Path(self.temp_dir) / "source.json"
        target_file = Path(self.temp_dir) / "target.json"
        
        source_data = {"includeCoAuthoredBy": True, "env": {"new": "value"}}
        target_data = {"includeCoAuthoredBy": False, "env": {"old": "keep"}, "other": "preserve"}
        
        with open(source_file, 'w') as f:
            json.dump(source_data, f)
        with open(target_file, 'w') as f:
            json.dump(target_data, f)
        
        result = SettingsJsonMerger.merge_settings(target_file, source_file)
        
        self.assertTrue(result.success)
        
        with open(target_file, 'r') as f:
            merged_data = json.load(f)
        
        expected = {
            "includeCoAuthoredBy": True,  # 源文件覆盖
            "env": {"old": "keep", "new": "value"},  # 深度合并
            "other": "preserve"  # 保留目标文件的内容
        }
        
        self.assertEqual(merged_data, expected)
    
    @patch('builtins.input', return_value='y')  # 模拟用户选择合并
    def test_merge_hooks_duplicate_prevention(self, mock_input):
        """测试防止hooks重复合并的核心问题"""
        target_hooks = {
            "PostToolUse": [
                {
                    "matcher": "Write|Edit|MultiEdit",
                    "hooks": [
                        {"type": "command", "command": "~/.claude/hooks/smart-lint.sh"},
                        {"type": "command", "command": "~/.claude/hooks/smart-test.sh"}
                    ]
                }
            ]
        }
        
        # 相同的源配置（模拟重复运行脚本）
        source_hooks = {
            "PostToolUse": [
                {
                    "matcher": "Write|Edit|MultiEdit",
                    "hooks": [
                        {"type": "command", "command": "~/.claude/hooks/smart-lint.sh"},
                        {"type": "command", "command": "~/.claude/hooks/smart-test.sh"}
                    ]
                }
            ]
        }
        
        result = SettingsJsonMerger.merge_hooks(target_hooks, source_hooks)
        
        # 关键断言：不应该产生重复的配置块
        self.assertEqual(len(result["PostToolUse"]), 1, "应该只有一个PostToolUse配置块")
        
        hooks = result["PostToolUse"][0]["hooks"]
        self.assertEqual(len(hooks), 2, "应该只有两个hooks，不应重复")
        
        # 验证命令不重复
        commands = [hook["command"] for hook in hooks]
        self.assertEqual(len(set(commands)), 2, "命令不应重复")
        self.assertIn("~/.claude/hooks/smart-lint.sh", commands)
        self.assertIn("~/.claude/hooks/smart-test.sh", commands)
    
    @patch('builtins.input', return_value='y')
    def test_multiple_script_runs_no_duplication(self, mock_input):
        """测试多次运行脚本不会产生重复配置（真实场景测试）"""
        source_file = Path(self.temp_dir) / "source.json"
        target_file = Path(self.temp_dir) / "target.json"
        
        # 原始源配置
        source_data = {
            "includeCoAuthoredBy": False,
            "hooks": {
                "PostToolUse": [
                    {
                        "matcher": "Write|Edit|MultiEdit",
                        "hooks": [
                            {"type": "command", "command": "~/.claude/hooks/smart-lint.sh"},
                            {"type": "command", "command": "~/.claude/hooks/smart-test.sh"}
                        ]
                    }
                ],
                "Stop": [
                    {
                        "matcher": "",
                        "hooks": [
                            {"type": "command", "command": "~/.claude/hooks/ntfy-notifier.sh"}
                        ]
                    }
                ]
            }
        }
        
        with open(source_file, 'w') as f:
            json.dump(source_data, f)
        
        # 第一次运行脚本
        result1 = SettingsJsonMerger.merge_settings(target_file, source_file)
        self.assertTrue(result1.success)
        
        # 读取第一次合并结果
        with open(target_file, 'r') as f:
            first_merge = json.load(f)
        
        # 第二次运行脚本（模拟重复运行）
        result2 = SettingsJsonMerger.merge_settings(target_file, source_file)
        
        # 读取第二次合并结果
        with open(target_file, 'r') as f:
            second_merge = json.load(f)
        
        # 第三次运行脚本
        result3 = SettingsJsonMerger.merge_settings(target_file, source_file)
        
        # 读取第三次合并结果
        with open(target_file, 'r') as f:
            third_merge = json.load(f)
        
        # 关键断言：多次运行不应产生重复
        self.assertEqual(len(third_merge["hooks"]["PostToolUse"]), 1, 
                        "PostToolUse应该只有一个配置块，不管运行多少次")
        self.assertEqual(len(third_merge["hooks"]["Stop"]), 1,
                        "Stop应该只有一个配置块，不管运行多少次")
        
        post_hooks = third_merge["hooks"]["PostToolUse"][0]["hooks"]
        stop_hooks = third_merge["hooks"]["Stop"][0]["hooks"]
        
        self.assertEqual(len(post_hooks), 2, "PostToolUse hooks应该只有2个")
        self.assertEqual(len(stop_hooks), 1, "Stop hooks应该只有1个")
        
        # 验证所有三次运行的结果应该相同
        self.assertEqual(first_merge, second_merge, "第一次和第二次运行结果应该相同")
        self.assertEqual(second_merge, third_merge, "第二次和第三次运行结果应该相同")


class TestClaudeConfigCopier(unittest.TestCase):
    """测试Claude配置复制器"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.source_dir = Path(self.temp_dir) / "source"
        self.target_dir = Path(self.temp_dir) / "target"
        
        self.source_dir.mkdir()
        self.target_dir.mkdir()
        
        self.copier = ClaudeConfigCopier(self.source_dir, self.target_dir)
        
        self.addCleanup(shutil.rmtree, self.temp_dir)
    
    def test_create_target_dir(self):
        """测试创建目标目录"""
        new_target = Path(self.temp_dir) / "new_target"
        copier = ClaudeConfigCopier(self.source_dir, new_target)
        
        result = copier.create_target_dir()
        
        self.assertTrue(result)
        self.assertTrue(new_target.exists())
        self.assertTrue(new_target.is_dir())
    
    def test_copy_file_new(self):
        """测试复制新文件"""
        src_file = self.source_dir / "test.txt"
        dest_file = self.target_dir / "test.txt"
        
        src_file.write_text("test content")
        
        result = self.copier.copy_file(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertFalse(result.skipped)
        self.assertTrue(dest_file.exists())
        self.assertEqual(dest_file.read_text(), "test content")
    
    def test_copy_file_identical(self):
        """测试复制相同文件会跳过"""
        src_file = self.source_dir / "test.txt"
        dest_file = self.target_dir / "test.txt"
        
        content = "identical content"
        src_file.write_text(content)
        dest_file.write_text(content)
        
        result = self.copier.copy_file(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertTrue(result.skipped)
    
    def test_copy_file_different(self):
        """测试覆盖不同文件"""
        src_file = self.source_dir / "test.txt"
        dest_file = self.target_dir / "test.txt"
        
        src_file.write_text("new content")
        dest_file.write_text("old content")
        
        result = self.copier.copy_file(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertFalse(result.skipped)
        self.assertEqual(dest_file.read_text(), "new content")
    
    @patch('builtins.input')
    def test_handle_claude_md_skip(self, mock_input):
        """测试跳过CLAUDE.md文件"""
        mock_input.return_value = 'n'
        
        src_file = self.source_dir / "CLAUDE.md"
        dest_file = self.target_dir / "CLAUDE.md"
        
        src_file.write_text("new content")
        dest_file.write_text("old content")
        
        result = self.copier.handle_claude_md(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertTrue(result.skipped)
        self.assertEqual(dest_file.read_text(), "old content")  # 内容未变
    
    @patch('builtins.input')
    def test_handle_claude_md_overwrite(self, mock_input):
        """测试覆盖CLAUDE.md文件"""
        mock_input.return_value = 'y'
        
        src_file = self.source_dir / "CLAUDE.md"
        dest_file = self.target_dir / "CLAUDE.md"
        
        src_file.write_text("new content")
        dest_file.write_text("old content")
        
        result = self.copier.handle_claude_md(src_file, dest_file)
        
        self.assertTrue(result.success)
        self.assertFalse(result.skipped)
        self.assertEqual(dest_file.read_text(), "new content")  # 内容已更新
    
    def test_copy_directory(self):
        """测试复制目录"""
        src_dir = self.source_dir / "testdir"
        dest_dir = self.target_dir / "testdir"
        
        src_dir.mkdir()
        (src_dir / "file1.txt").write_text("content1")
        (src_dir / "file2.txt").write_text("content2")
        
        result = self.copier.copy_directory(src_dir, dest_dir)
        
        self.assertTrue(result.success)
        self.assertTrue(dest_dir.exists())
        self.assertTrue((dest_dir / "file1.txt").exists())
        self.assertTrue((dest_dir / "file2.txt").exists())
        self.assertEqual((dest_dir / "file1.txt").read_text(), "content1")
        self.assertEqual((dest_dir / "file2.txt").read_text(), "content2")


class TestCopyResult(unittest.TestCase):
    """测试复制结果类"""
    
    def test_copy_result_creation(self):
        """测试创建复制结果"""
        result = CopyResult(True, "Success message")
        
        self.assertTrue(result.success)
        self.assertEqual(result.message, "Success message")
        self.assertFalse(result.skipped)  # 默认值
        
        result_with_skip = CopyResult(False, "Failed", skipped=True)
        
        self.assertFalse(result_with_skip.success)
        self.assertEqual(result_with_skip.message, "Failed")
        self.assertTrue(result_with_skip.skipped)


if __name__ == '__main__':
    # 添加详细的测试输出
    unittest.main(verbosity=2)