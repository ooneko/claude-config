#!/usr/bin/env python3
"""
utils.common 模块测试
"""

import unittest
import tempfile
import os
from pathlib import Path
from unittest.mock import patch, mock_open

from utils.common import Color, ConflictResolution, OperationResult, FileComparator, ProxyManager


class TestColor(unittest.TestCase):
    """Color 类测试"""
    
    def test_color_constants(self):
        """测试颜色常量"""
        self.assertEqual(Color.RED, '\033[0;31m')
        self.assertEqual(Color.GREEN, '\033[0;32m')
        self.assertEqual(Color.YELLOW, '\033[1;33m')
        self.assertEqual(Color.BLUE, '\033[0;34m')
        self.assertEqual(Color.NC, '\033[0m')
    
    @patch('builtins.print')
    def test_print_colored(self, mock_print):
        """测试彩色打印"""
        Color.print_colored("测试消息", Color.GREEN)
        mock_print.assert_called_once_with(f"{Color.GREEN}测试消息{Color.NC}")
    
    @patch('builtins.input', return_value='用户输入')
    def test_input_colored_default(self, mock_input):
        """测试彩色输入 - 默认颜色"""
        result = Color.input_colored("提示: ")
        mock_input.assert_called_once_with(f"{Color.YELLOW}提示: {Color.NC}")
        self.assertEqual(result, '用户输入')
    
    @patch('builtins.input', return_value='用户输入')
    def test_input_colored_custom(self, mock_input):
        """测试彩色输入 - 自定义颜色"""
        result = Color.input_colored("提示: ", Color.RED)
        mock_input.assert_called_once_with(f"{Color.RED}提示: {Color.NC}")
        self.assertEqual(result, '用户输入')


class TestConflictResolution(unittest.TestCase):
    """ConflictResolution 枚举测试"""
    
    def test_enum_values(self):
        """测试枚举值"""
        self.assertEqual(ConflictResolution.OVERWRITE.value, "overwrite")
        self.assertEqual(ConflictResolution.SKIP.value, "skip")
        self.assertEqual(ConflictResolution.SHOW_DIFF.value, "diff")
        self.assertEqual(ConflictResolution.MERGE.value, "merge")


class TestOperationResult(unittest.TestCase):
    """OperationResult 数据类测试"""
    
    def test_default_values(self):
        """测试默认值"""
        result = OperationResult(True, "成功")
        self.assertTrue(result.success)
        self.assertEqual(result.message, "成功")
        self.assertFalse(result.skipped)
    
    def test_custom_values(self):
        """测试自定义值"""
        result = OperationResult(False, "失败", True)
        self.assertFalse(result.success)
        self.assertEqual(result.message, "失败")
        self.assertTrue(result.skipped)


class TestFileComparator(unittest.TestCase):
    """FileComparator 类测试"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.temp_path = Path(self.temp_dir)
    
    def tearDown(self):
        """清理测试环境"""
        import shutil
        shutil.rmtree(self.temp_dir)
    
    def test_files_are_same_both_exist_same_content(self):
        """测试相同内容的文件"""
        file1 = self.temp_path / "file1.txt"
        file2 = self.temp_path / "file2.txt"
        
        content = "相同内容"
        file1.write_text(content, encoding='utf-8')
        file2.write_text(content, encoding='utf-8')
        
        self.assertTrue(FileComparator.files_are_same(file1, file2))
    
    def test_files_are_same_different_content(self):
        """测试不同内容的文件"""
        file1 = self.temp_path / "file1.txt"
        file2 = self.temp_path / "file2.txt"
        
        file1.write_text("内容1", encoding='utf-8')
        file2.write_text("内容2", encoding='utf-8')
        
        self.assertFalse(FileComparator.files_are_same(file1, file2))
    
    def test_files_are_same_one_not_exist(self):
        """测试文件不存在的情况"""
        file1 = self.temp_path / "file1.txt"
        file2 = self.temp_path / "file2.txt"
        
        file1.write_text("内容", encoding='utf-8')
        # file2 不创建
        
        self.assertFalse(FileComparator.files_are_same(file1, file2))
    
    @patch('builtins.print')
    def test_show_file_diff_success(self, mock_print):
        """测试显示文件差异 - 成功"""
        file1 = self.temp_path / "file1.txt"
        file2 = self.temp_path / "file2.txt"
        
        file1.write_text("第一行\n第二行\n", encoding='utf-8')
        file2.write_text("第一行\n修改的第二行\n", encoding='utf-8')
        
        with patch.object(Color, 'print_colored') as mock_color_print:
            FileComparator.show_file_diff(file1, file2)
            mock_color_print.assert_called()
            mock_print.assert_called()
    
    @patch.object(Color, 'print_colored')
    def test_show_file_diff_exception(self, mock_color_print):
        """测试显示文件差异 - 异常处理"""
        file1 = self.temp_path / "non_existent1.txt"
        file2 = self.temp_path / "non_existent2.txt"
        
        FileComparator.show_file_diff(file1, file2)
        mock_color_print.assert_called_with("显示差异失败: [Errno 2] No such file or directory: '{}'".format(file1), Color.RED)


class TestProxyManager(unittest.TestCase):
    """ProxyManager 类测试"""
    
    def setUp(self):
        """设置测试环境"""
        self.temp_dir = tempfile.mkdtemp()
        self.claude_dir = Path(self.temp_dir)
        self.proxy_manager = ProxyManager(self.claude_dir)
    
    def tearDown(self):
        """清理测试环境"""
        import shutil
        shutil.rmtree(self.temp_dir)
    
    def test_default_proxy(self):
        """测试默认代理地址"""
        self.assertEqual(self.proxy_manager.default_proxy, "http://127.0.0.1:7890")
    
    def test_get_proxy_address_file_not_exist(self):
        """测试获取代理地址 - 文件不存在"""
        proxy = self.proxy_manager.get_proxy_address()
        self.assertEqual(proxy, self.proxy_manager.default_proxy)
    
    def test_get_proxy_address_file_exists(self):
        """测试获取代理地址 - 文件存在"""
        custom_proxy = "http://127.0.0.1:8080"
        self.proxy_manager.proxy_file.write_text(custom_proxy, encoding='utf-8')
        
        proxy = self.proxy_manager.get_proxy_address()
        self.assertEqual(proxy, custom_proxy)
    
    def test_get_proxy_address_file_read_error(self):
        """测试获取代理地址 - 读取错误"""
        # 创建文件但设置为不可读
        self.proxy_manager.proxy_file.write_text("test", encoding='utf-8')
        os.chmod(self.proxy_manager.proxy_file, 0o000)
        
        try:
            proxy = self.proxy_manager.get_proxy_address()
            self.assertEqual(proxy, self.proxy_manager.default_proxy)
        finally:
            # 恢复权限以便清理
            os.chmod(self.proxy_manager.proxy_file, 0o644)
    
    def test_save_proxy_address_success(self):
        """测试保存代理地址 - 成功"""
        custom_proxy = "http://127.0.0.1:8080"
        self.proxy_manager.save_proxy_address(custom_proxy)
        
        # 验证文件是否正确保存
        saved_proxy = self.proxy_manager.proxy_file.read_text(encoding='utf-8')
        self.assertEqual(saved_proxy, custom_proxy)
    
    @patch.object(Color, 'print_colored')
    def test_save_proxy_address_error(self, mock_color_print):
        """测试保存代理地址 - 错误处理"""
        # 创建一个无法写入的路径
        invalid_proxy_manager = ProxyManager(Path("/invalid_path_that_does_not_exist"))
        
        invalid_proxy_manager.save_proxy_address("http://test.com")
        mock_color_print.assert_called()
        args = mock_color_print.call_args[0]
        self.assertTrue(args[0].startswith("保存代理地址失败:"))
        self.assertEqual(args[1], Color.RED)
    
    @patch('builtins.input', side_effect=['', '测试输入'])
    @patch.object(Color, 'input_colored', side_effect=['', 'http://127.0.0.1:8080'])
    @patch.object(Color, 'print_colored')
    @patch('builtins.print')
    def test_prompt_for_proxy_default(self, mock_print, mock_color_print, mock_color_input, mock_input):
        """测试提示用户输入代理地址 - 使用默认"""
        result = self.proxy_manager.prompt_for_proxy()
        
        self.assertEqual(result, self.proxy_manager.default_proxy)
        # 验证文件是否保存
        saved_proxy = self.proxy_manager.proxy_file.read_text(encoding='utf-8')
        self.assertEqual(saved_proxy, self.proxy_manager.default_proxy)
    
    @patch.object(Color, 'input_colored', side_effect=['invalid_url', 'http://127.0.0.1:8080'])
    @patch.object(Color, 'print_colored')
    @patch('builtins.print')
    def test_prompt_for_proxy_custom_with_retry(self, mock_print, mock_color_print, mock_color_input):
        """测试提示用户输入代理地址 - 自定义地址，需要重试"""
        result = self.proxy_manager.prompt_for_proxy()
        
        self.assertEqual(result, 'http://127.0.0.1:8080')
        # 验证错误提示
        error_calls = [call for call in mock_print.call_args_list if "❌" in str(call)]
        self.assertTrue(len(error_calls) > 0)


if __name__ == '__main__':
    unittest.main()