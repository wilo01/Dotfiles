"""Tests for unified availability checker."""

import unittest
from unittest.mock import MagicMock, patch, Mock
import time

from src.helper_cli.services.unified_availability_checker import (
    UnifiedAvailabilityChecker,
    CheckStatus,
    CheckResult,
    check_availability,
    check_domain_availability
)


class TestCheckResult(unittest.TestCase):
    """Test CheckResult dataclass."""
    
    def test_initialization(self):
        """Test CheckResult initialization."""
        result = CheckResult(
            name="example.com",
            platform="domain",
            status=CheckStatus.AVAILABLE,
            url="https://example.com",
            price=9.99
        )
        
        self.assertEqual(result.name, "example.com")
        self.assertEqual(result.platform, "domain")
        self.assertEqual(result.status, CheckStatus.AVAILABLE)
        self.assertEqual(result.url, "https://example.com")
        self.assertEqual(result.price, 9.99)
        self.assertIsNotNone(result.checked_at)
    
    def test_is_available(self):
        """Test is_available property."""
        available = CheckResult("test", "domain", CheckStatus.AVAILABLE)
        self.assertTrue(available.is_available)
        
        taken = CheckResult("test", "domain", CheckStatus.TAKEN)
        self.assertFalse(taken.is_available)
        
        error = CheckResult("test", "domain", CheckStatus.ERROR)
        self.assertFalse(error.is_available)
    
    def test_to_dict(self):
        """Test conversion to dictionary."""
        result = CheckResult(
            name="test",
            platform="npm",
            status=CheckStatus.TAKEN,
            message="Package exists"
        )
        
        data = result.to_dict()
        self.assertEqual(data['name'], "test")
        self.assertEqual(data['platform'], "npm")
        self.assertEqual(data['status'], "taken")
        self.assertEqual(data['message'], "Package exists")


class TestUnifiedAvailabilityChecker(unittest.TestCase):
    """Test UnifiedAvailabilityChecker."""
    
    def setUp(self):
        """Set up test environment."""
        self.checker = UnifiedAvailabilityChecker(timeout=1, max_workers=2)
    
    def test_initialization(self):
        """Test checker initialization."""
        self.assertEqual(self.checker.timeout, 1)
        self.assertEqual(self.checker.max_workers, 2)
        self.assertIsNotNone(self.checker.session)
        self.assertEqual(len(self.checker._cache), 0)
    
    def test_validate_name(self):
        """Test name validation."""
        # Valid names
        self.assertTrue(self.checker._validate_name("example"))
        self.assertTrue(self.checker._validate_name("test-123"))
        self.assertTrue(self.checker._validate_name("my-app"))
        
        # Invalid names
        self.assertFalse(self.checker._validate_name(""))
        self.assertFalse(self.checker._validate_name("a"))  # Too short
        self.assertFalse(self.checker._validate_name("-test"))  # Starts with hyphen
        self.assertFalse(self.checker._validate_name("test-"))  # Ends with hyphen
        self.assertFalse(self.checker._validate_name("test_name"))  # Underscore
        self.assertFalse(self.checker._validate_name("test.name"))  # Period
        self.assertFalse(self.checker._validate_name("a" * 64))  # Too long
    
    @patch('socket.gethostbyname')
    def test_check_domain_dns_lookup(self, mock_gethostbyname):
        """Test domain checking via DNS lookup."""
        # Domain exists (has DNS record)
        mock_gethostbyname.return_value = "93.184.216.34"
        result = self.checker.check_domain("example.com")
        self.assertEqual(result.status, CheckStatus.TAKEN)
        self.assertEqual(result.url, "https://example.com")
        
        # Domain doesn't exist (no DNS record)
        mock_gethostbyname.side_effect = socket.gaierror("Name or service not known")
        result = self.checker.check_domain("nonexistent-domain-12345.com")
        self.assertEqual(result.status, CheckStatus.AVAILABLE)
        self.assertIsNone(result.url)
    
    @patch.object(UnifiedAvailabilityChecker, '_check_whois')
    @patch('socket.gethostbyname')
    def test_check_domain_with_whois(self, mock_gethostbyname, mock_whois):
        """Test domain checking with WHOIS fallback."""
        # No DNS record, but WHOIS says taken
        mock_gethostbyname.side_effect = socket.gaierror("Name or service not known")
        mock_whois.return_value = CheckStatus.TAKEN
        
        result = self.checker.check_domain("registered-but-no-dns.com")
        self.assertEqual(result.status, CheckStatus.TAKEN)
        
        # No DNS record, WHOIS says available
        mock_whois.return_value = CheckStatus.AVAILABLE
        result = self.checker.check_domain("truly-available.com")
        self.assertEqual(result.status, CheckStatus.AVAILABLE)
    
    def test_check_social_with_mock_response(self):
        """Test social media checking with mocked responses."""
        with patch.object(self.checker.session, 'head') as mock_head:
            # Username taken
            mock_response = Mock()
            mock_response.status_code = 200
            mock_head.return_value = mock_response
            
            result = self.checker.check_social("existinguser", "github")
            self.assertEqual(result.status, CheckStatus.TAKEN)
            self.assertEqual(result.platform, "github")
            self.assertEqual(result.url, "https://github.com/existinguser")
            
            # Username available
            mock_response.status_code = 404
            result = self.checker.check_social("availableuser", "twitter")
            self.assertEqual(result.status, CheckStatus.AVAILABLE)
            self.assertIsNone(result.url)
            
            # Rate limited
            mock_response.status_code = 429
            result = self.checker.check_social("testuser", "instagram")
            self.assertEqual(result.status, CheckStatus.RATE_LIMITED)
    
    def test_check_social_invalid_platform(self):
        """Test social media checking with invalid platform."""
        result = self.checker.check_social("testuser", "invalid-platform")
        self.assertEqual(result.status, CheckStatus.INVALID)
        self.assertIn("Unknown platform", result.message)
    
    def test_check_package_with_mock_response(self):
        """Test package registry checking with mocked responses."""
        with patch.object(self.checker.session, 'get') as mock_get:
            # Package exists
            mock_response = Mock()
            mock_response.status_code = 200
            mock_get.return_value = mock_response
            
            result = self.checker.check_package("express", "npm")
            self.assertEqual(result.status, CheckStatus.TAKEN)
            self.assertEqual(result.platform, "npm")
            
            # Package doesn't exist
            mock_response.status_code = 404
            result = self.checker.check_package("my-unique-package-name", "pypi")
            self.assertEqual(result.status, CheckStatus.AVAILABLE)
            
            # Rate limited
            mock_response.status_code = 429
            result = self.checker.check_package("test-package", "cargo")
            self.assertEqual(result.status, CheckStatus.RATE_LIMITED)
    
    def test_check_package_invalid_registry(self):
        """Test package checking with invalid registry."""
        result = self.checker.check_package("test-package", "invalid-registry")
        self.assertEqual(result.status, CheckStatus.INVALID)
        self.assertIn("Unknown registry", result.message)
    
    def test_caching(self):
        """Test result caching."""
        with patch.object(self.checker.session, 'head') as mock_head:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_head.return_value = mock_response
            
            # First call should hit the network
            result1 = self.checker.check_social("testuser", "github")
            self.assertEqual(mock_head.call_count, 1)
            
            # Second call should use cache
            result2 = self.checker.check_social("testuser", "github")
            self.assertEqual(mock_head.call_count, 1)  # Still 1, not 2
            
            # Results should be the same
            self.assertEqual(result1.status, result2.status)
            self.assertEqual(result1.name, result2.name)
    
    def test_suggest_alternatives(self):
        """Test alternative name suggestions."""
        suggestions = self.checker.suggest_alternatives("myapp")
        
        self.assertIn("myapp-app", suggestions)
        self.assertIn("myapp-io", suggestions)
        self.assertIn("myapp-dev", suggestions)
        self.assertIn("get-myapp", suggestions)
        self.assertIn("myapp1", suggestions)
        self.assertIn("myapp-1", suggestions)
        
        # Check max suggestions limit
        suggestions = self.checker.suggest_alternatives("test", max_suggestions=5)
        self.assertEqual(len(suggestions), 5)
    
    def test_export_results_json(self):
        """Test exporting results as JSON."""
        results = {
            'domains': [
                CheckResult("test.com", "domain", CheckStatus.AVAILABLE),
                CheckResult("test.org", "domain", CheckStatus.TAKEN)
            ],
            'social': [
                CheckResult("testuser", "github", CheckStatus.AVAILABLE)
            ]
        }
        
        json_output = self.checker.export_results(results, format='json')
        
        import json
        data = json.loads(json_output)
        self.assertIn('domains', data)
        self.assertIn('social', data)
        self.assertEqual(len(data['domains']), 2)
        self.assertEqual(data['domains'][0]['status'], 'available')
    
    def test_export_results_csv(self):
        """Test exporting results as CSV."""
        results = {
            'domains': [
                CheckResult("test.com", "domain", CheckStatus.AVAILABLE)
            ]
        }
        
        csv_output = self.checker.export_results(results, format='csv')
        
        lines = csv_output.strip().split('\n')
        self.assertIn('Category', lines[0])
        self.assertIn('Status', lines[0])
        self.assertIn('domains', lines[1])
        self.assertIn('available', lines[1])
    
    def test_export_results_markdown(self):
        """Test exporting results as Markdown."""
        results = {
            'domains': [
                CheckResult("test.com", "domain", CheckStatus.AVAILABLE),
                CheckResult("test.org", "domain", CheckStatus.TAKEN, url="https://test.org")
            ]
        }
        
        md_output = self.checker.export_results(results, format='markdown')
        
        self.assertIn('# Availability Check Results', md_output)
        self.assertIn('## Domains', md_output)
        self.assertIn('✅', md_output)  # Available emoji
        self.assertIn('❌', md_output)  # Taken emoji
        self.assertIn('[Link](https://test.org)', md_output)
    
    @patch('concurrent.futures.ThreadPoolExecutor')
    def test_check_all_integration(self, mock_executor):
        """Test check_all method integration."""
        # Mock executor to control execution
        mock_executor_instance = MagicMock()
        mock_executor.return_value.__enter__ = MagicMock(return_value=mock_executor_instance)
        mock_executor.return_value.__exit__ = MagicMock(return_value=None)
        
        # Mock futures
        mock_future1 = MagicMock()
        mock_future1.result.return_value = CheckResult("test.com", "domain", CheckStatus.AVAILABLE)
        
        mock_future2 = MagicMock()
        mock_future2.result.return_value = CheckResult("test", "github", CheckStatus.TAKEN)
        
        mock_executor_instance.submit.side_effect = [mock_future1, mock_future2]
        
        # Create mock for as_completed
        with patch('concurrent.futures.as_completed') as mock_as_completed:
            mock_as_completed.return_value = [mock_future1, mock_future2]
            
            results = self.checker.check_all(
                "test",
                check_domains=True,
                check_social=True,
                check_packages=False,
                tlds=['.com']
            )
            
            # Verify structure
            self.assertIn('domains', results)
            self.assertIn('social', results)
            self.assertIn('packages', results)


class TestConvenienceFunctions(unittest.TestCase):
    """Test convenience functions."""
    
    @patch('src.helper_cli.services.unified_availability_checker.UnifiedAvailabilityChecker')
    def test_check_availability(self, mock_checker_class):
        """Test check_availability convenience function."""
        mock_checker = MagicMock()
        mock_checker_class.return_value = mock_checker
        mock_checker.check_all.return_value = {'domains': [], 'social': []}
        
        result = check_availability("test", check_domains=True)
        
        mock_checker.check_all.assert_called_once_with(
            "test",
            check_domains=True
        )
    
    @patch('src.helper_cli.services.unified_availability_checker.UnifiedAvailabilityChecker')
    def test_check_domain_availability(self, mock_checker_class):
        """Test check_domain_availability convenience function."""
        mock_checker = MagicMock()
        mock_checker_class.return_value = mock_checker
        mock_result = CheckResult("test.com", "domain", CheckStatus.AVAILABLE)
        mock_checker.check_domain.return_value = mock_result
        
        result = check_domain_availability("test.com")
        
        mock_checker.check_domain.assert_called_once_with("test.com")
        self.assertEqual(result, mock_result)


if __name__ == '__main__':
    unittest.main()