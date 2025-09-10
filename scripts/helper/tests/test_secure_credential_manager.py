"""Tests for secure credential manager."""

import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from src.helper_cli.config.secure_credential_manager import (
    SecureCredentialManager,
    SecureCredentialAdapter
)


class TestSecureCredentialManager(unittest.TestCase):
    """Test secure credential manager."""
    
    def setUp(self):
        """Set up test environment."""
        # Create temporary directory for testing
        self.temp_dir = tempfile.mkdtemp()
        self.config_dir = Path(self.temp_dir) / 'config'
        self.config_dir.mkdir(parents=True, exist_ok=True)
        
        # Create manager with test directory
        self.manager = SecureCredentialManager(config_dir=self.config_dir)
    
    def tearDown(self):
        """Clean up test environment."""
        # Remove temporary directory
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def test_initialization(self):
        """Test manager initialization."""
        self.assertTrue(self.config_dir.exists())
        self.assertTrue(self.manager.key_file.exists())
        self.assertIsNotNone(self.manager._cipher)
    
    def test_save_and_retrieve_credentials(self):
        """Test saving and retrieving credentials."""
        # Save credentials
        test_creds = {
            'username': 'testuser',
            'password': 'testpass123',
            'api_key': 'secret-key-123'
        }
        
        success = self.manager.save_credentials('test_service', test_creds)
        self.assertTrue(success)
        
        # Retrieve credentials
        retrieved = self.manager.get_credentials('test_service')
        self.assertIsNotNone(retrieved)
        self.assertEqual(retrieved, test_creds)
    
    def test_encryption(self):
        """Test that credentials are actually encrypted."""
        # Save credentials
        test_creds = {'secret': 'sensitive-data'}
        self.manager.save_credentials('test', test_creds)
        
        # Read raw file content
        with open(self.manager.credentials_file, 'rb') as f:
            raw_content = f.read()
        
        # Verify it's not plaintext
        self.assertNotIn(b'sensitive-data', raw_content)
        self.assertNotIn(b'secret', raw_content)
    
    def test_delete_credentials(self):
        """Test deleting credentials."""
        # Save multiple services
        self.manager.save_credentials('service1', {'key': 'value1'})
        self.manager.save_credentials('service2', {'key': 'value2'})
        
        # Delete one service
        success = self.manager.delete_credentials('service1')
        self.assertTrue(success)
        
        # Verify it's deleted
        self.assertIsNone(self.manager.get_credentials('service1'))
        
        # Verify other service still exists
        self.assertIsNotNone(self.manager.get_credentials('service2'))
    
    def test_list_services(self):
        """Test listing services."""
        # Initially empty
        self.assertEqual(self.manager.list_services(), [])
        
        # Add services
        self.manager.save_credentials('jira', {'token': '123'})
        self.manager.save_credentials('github', {'token': '456'})
        
        # List services
        services = self.manager.list_services()
        self.assertEqual(set(services), {'jira', 'github'})
    
    def test_export_import(self):
        """Test export and import with password."""
        # Save some credentials
        original_creds = {
            'jira': {'token': 'jira-123'},
            'sheets': {'key': 'sheets-456'}
        }
        
        for service, creds in original_creds.items():
            self.manager.save_credentials(service, creds)
        
        # Export with password
        password = 'export-password-123'
        export_string = self.manager.export_credentials(password)
        self.assertIsNotNone(export_string)
        
        # Create new manager
        new_dir = Path(self.temp_dir) / 'new_config'
        new_manager = SecureCredentialManager(config_dir=new_dir)
        
        # Import with correct password
        success = new_manager.import_credentials(export_string, password)
        self.assertTrue(success)
        
        # Verify imported credentials
        for service, creds in original_creds.items():
            imported = new_manager.get_credentials(service)
            self.assertEqual(imported, creds)
    
    def test_import_with_wrong_password(self):
        """Test import fails with wrong password."""
        # Export with one password
        self.manager.save_credentials('test', {'data': 'value'})
        export_string = self.manager.export_credentials('correct-password')
        
        # Try import with wrong password
        new_manager = SecureCredentialManager(
            config_dir=Path(self.temp_dir) / 'new'
        )
        success = new_manager.import_credentials(export_string, 'wrong-password')
        self.assertFalse(success)
    
    def test_file_permissions(self):
        """Test that files have secure permissions."""
        # Save credentials to create files
        self.manager.save_credentials('test', {'data': 'value'})
        
        # Check permissions (owner read/write only)
        key_stat = os.stat(self.manager.key_file)
        creds_stat = os.stat(self.manager.credentials_file)
        
        # Check that only owner has read/write permissions (0o600)
        self.assertEqual(key_stat.st_mode & 0o777, 0o600)
        self.assertEqual(creds_stat.st_mode & 0o777, 0o600)


class TestSecureCredentialAdapter(unittest.TestCase):
    """Test secure credential adapter for backward compatibility."""
    
    def setUp(self):
        """Set up test environment."""
        self.temp_dir = tempfile.mkdtemp()
        
        # Mock the secure manager
        with patch('src.helper_cli.config.secure_credential_manager.SecureCredentialManager'):
            self.adapter = SecureCredentialAdapter()
            self.adapter.secure_manager = MagicMock()
    
    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def test_get_jira_credentials(self):
        """Test getting JIRA credentials."""
        # Mock return value
        self.adapter.secure_manager.get_credentials.return_value = {
            'base_url': 'https://test.atlassian.net',
            'email': 'test@example.com',
            'api_token': 'secret-token'
        }
        
        # Get credentials
        creds = self.adapter.get_jira_credentials()
        
        self.assertIsNotNone(creds)
        self.assertEqual(creds['base_url'], 'https://test.atlassian.net')
        self.assertEqual(creds['email'], 'test@example.com')
        self.assertEqual(creds['api_token'], 'secret-token')
        
        # Verify correct service was queried
        self.adapter.secure_manager.get_credentials.assert_called_with('jira')
    
    def test_save_jira_credentials(self):
        """Test saving JIRA credentials."""
        # Save credentials
        self.adapter.secure_manager.save_credentials.return_value = True
        
        success = self.adapter.save_jira_credentials(
            'https://test.atlassian.net',
            'test@example.com',
            'secret-token'
        )
        
        self.assertTrue(success)
        
        # Verify correct data was saved
        self.adapter.secure_manager.save_credentials.assert_called_with(
            'jira',
            {
                'base_url': 'https://test.atlassian.net',
                'email': 'test@example.com',
                'api_token': 'secret-token'
            }
        )
    
    @patch.dict(os.environ, {
        'JIRA_URL': 'https://env.atlassian.net',
        'JIRA_EMAIL': 'env@example.com',
        'JIRA_API_TOKEN': 'env-token'
    })
    def test_fallback_to_environment(self):
        """Test fallback to environment variables."""
        # No credentials in secure storage
        self.adapter.secure_manager.get_credentials.return_value = None
        
        # Get credentials (should fall back to env)
        creds = self.adapter.get_jira_credentials()
        
        self.assertIsNotNone(creds)
        self.assertEqual(creds['base_url'], 'https://env.atlassian.net')
        self.assertEqual(creds['email'], 'env@example.com')
        self.assertEqual(creds['api_token'], 'env-token')
    
    def test_google_sheets_config(self):
        """Test Google Sheets configuration."""
        # Mock return value
        self.adapter.secure_manager.get_credentials.return_value = {
            'sheet_id': 'test-sheet-id',
            'credentials_file': '/path/to/creds.json'
        }
        
        # Get config
        config = self.adapter.get_google_sheets_config()
        
        self.assertIsNotNone(config)
        self.assertEqual(config['sheet_id'], 'test-sheet-id')
        
        # Save config
        self.adapter.secure_manager.save_credentials.return_value = True
        success = self.adapter.save_google_sheets_config(
            'new-sheet-id',
            '/new/path/creds.json'
        )
        
        self.assertTrue(success)
        self.adapter.secure_manager.save_credentials.assert_called_with(
            'google_sheets',
            {
                'sheet_id': 'new-sheet-id',
                'credentials_file': '/new/path/creds.json'
            }
        )


if __name__ == '__main__':
    unittest.main()