"""Tests for JiraService."""

import unittest
from unittest.mock import Mock, patch, MagicMock
from datetime import datetime, timedelta
import json
import sqlite3
import tempfile
import os
from pathlib import Path

import sys
sys.path.insert(0, str(Path(__file__).parent.parent / 'src'))

from helper_cli.services.jira_service import JiraService


class TestJiraService(unittest.TestCase):
    """Test cases for JiraService."""

    def setUp(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.test_db = Path(self.temp_dir) / 'test_jira.db'

        # Mock environment
        self.patch_home = patch('helper_cli.services.jira_service.Path.home')
        self.mock_home = self.patch_home.start()
        self.mock_home.return_value = Path(self.temp_dir)

        # Create service instance
        self.service = JiraService(
            base_url='https://test.atlassian.net',
            email='test@example.com',
            api_token='test_token'
        )
        self.service.db_path = self.test_db

    def tearDown(self):
        """Clean up test fixtures."""
        self.patch_home.stop()
        # Clean up temp directory
        import shutil
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)

    def test_init_cache_db(self):
        """Test database initialization."""
        # Database should be created in setUp
        self.assertTrue(self.test_db.exists())

        # Check tables exist
        conn = sqlite3.connect(self.test_db)
        cursor = conn.cursor()
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = [row[0] for row in cursor.fetchall()]
        conn.close()

        self.assertIn('tickets', tables)
        self.assertIn('worklogs', tables)
        self.assertIn('sprints', tables)

    @patch('helper_cli.services.jira_service.requests.Session.get')
    def test_fetch_assigned_tickets_online(self, mock_get):
        """Test fetching tickets when online."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            'issues': [
                {
                    'key': 'TEST-1',
                    'fields': {
                        'summary': 'Test issue 1',
                        'status': {'name': 'In Progress'},
                        'assignee': {'displayName': 'Test User'},
                        'priority': {'name': 'High'},
                        'issuetype': {'name': 'Bug'}
                    }
                },
                {
                    'key': 'TEST-2',
                    'fields': {
                        'summary': 'Test issue 2',
                        'status': {'name': 'To Do'},
                        'assignee': {'displayName': 'Test User'},
                        'priority': {'name': 'Medium'},
                        'issuetype': {'name': 'Task'}
                    }
                }
            ]
        }
        mock_get.return_value = mock_response

        # Mock online check
        with patch.object(self.service, '_is_online', return_value=True):
            tickets = self.service.fetch_assigned_tickets()

        self.assertEqual(len(tickets), 2)
        self.assertEqual(tickets[0]['key'], 'TEST-1')
        self.assertEqual(tickets[1]['key'], 'TEST-2')

        # Check tickets were cached
        cached = self.service._get_cached_tickets()
        self.assertEqual(len(cached), 2)

    def test_fetch_assigned_tickets_with_jql(self):
        """Test fetching tickets with custom JQL."""
        with patch.object(self.service.session, 'get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {'issues': []}
            mock_get.return_value = mock_response

            with patch.object(self.service, '_is_online', return_value=True):
                self.service.fetch_assigned_tickets(
                    jql='project = TEST AND status = "In Progress"'
                )

            # Verify custom JQL was used
            mock_get.assert_called_once()
            call_args = mock_get.call_args
            self.assertEqual(
                call_args[1]['params']['jql'],
                'project = TEST AND status = "In Progress"'
            )

    def test_log_work(self):
        """Test logging work."""
        ticket_key = 'TEST-1'
        duration = '2h 30m'
        description = 'Working on feature'
        started_at = datetime.now()

        # Log work
        result = self.service.log_work(ticket_key, duration, description, started_at)
        self.assertTrue(result)

        # Check work was saved to cache
        worklogs = self.service.get_recent_worklogs(days=1)
        self.assertEqual(len(worklogs), 1)
        self.assertEqual(worklogs[0]['ticket_key'], ticket_key)
        self.assertEqual(worklogs[0]['duration'], duration)
        self.assertEqual(worklogs[0]['description'], description)
        self.assertFalse(worklogs[0]['synced'])

    @patch('helper_cli.services.jira_service.requests.Session.post')
    def test_sync_worklogs(self, mock_post):
        """Test syncing worklogs to Jira."""
        # Add unsynced worklog
        self.service.log_work('TEST-1', '2h', 'Test work')

        # Mock successful sync
        mock_response = Mock()
        mock_response.status_code = 201
        mock_post.return_value = mock_response

        with patch.object(self.service, '_is_online', return_value=True):
            result = self.service._sync_worklogs()

        self.assertTrue(result)

        # Check worklog was marked as synced
        worklogs = self.service.get_recent_worklogs(days=1)
        self.assertTrue(worklogs[0]['synced'])

    def test_parse_duration_to_seconds(self):
        """Test duration parsing."""
        test_cases = [
            ('2h', 7200),
            ('30m', 1800),
            ('1h 30m', 5400),
            ('2h30m', 9000),
            ('90m', 5400),
            ('1', 3600),  # Just number defaults to hours
            ('', 3600),   # Empty defaults to 1 hour
        ]

        for duration, expected_seconds in test_cases:
            result = self.service._parse_duration_to_seconds(duration)
            self.assertEqual(result, expected_seconds, f"Failed for {duration}")

    def test_get_recent_worklogs(self):
        """Test retrieving recent worklogs."""
        # Add multiple worklogs
        now = datetime.now()
        self.service.log_work('TEST-1', '1h', 'Old work', now - timedelta(days=10))
        self.service.log_work('TEST-2', '2h', 'Recent work', now - timedelta(days=2))
        self.service.log_work('TEST-3', '3h', 'Today work', now)

        # Get last 7 days
        recent = self.service.get_recent_worklogs(days=7)
        self.assertEqual(len(recent), 2)  # Should not include 10 days old

        # Get last 30 days
        all_logs = self.service.get_recent_worklogs(days=30)
        self.assertEqual(len(all_logs), 3)

    def test_search_tickets_with_pagination(self):
        """Test searching tickets with pagination."""
        with patch.object(self.service.session, 'get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {
                'issues': [
                    {'key': 'TEST-1', 'fields': {'summary': 'Issue 1'}},
                    {'key': 'TEST-2', 'fields': {'summary': 'Issue 2'}}
                ],
                'total': 10,
                'startAt': 0,
                'maxResults': 2
            }
            mock_get.return_value = mock_response

            with patch.object(self.service, '_is_online', return_value=True):
                result = self.service.search_tickets(
                    jql='project = TEST',
                    start_at=0,
                    max_results=2
                )

            self.assertEqual(len(result['issues']), 2)
            self.assertEqual(result['total'], 10)
            self.assertEqual(result['startAt'], 0)
            self.assertEqual(result['maxResults'], 2)

    def test_cache_tickets(self):
        """Test caching tickets in database."""
        tickets = [
            {
                'key': 'TEST-1',
                'fields': {
                    'summary': 'Test issue 1',
                    'status': {'name': 'In Progress'},
                    'assignee': {'displayName': 'User 1'}
                }
            },
            {
                'key': 'TEST-2',
                'fields': {
                    'summary': 'Test issue 2',
                    'status': {'name': 'Done'},
                    'assignee': {'displayName': 'User 2'}
                }
            }
        ]

        self.service._cache_tickets(tickets)

        # Retrieve cached tickets
        cached = self.service._get_cached_tickets()
        self.assertEqual(len(cached), 2)
        self.assertEqual(cached[0]['key'], 'TEST-1')
        self.assertEqual(cached[1]['key'], 'TEST-2')

    @patch('helper_cli.services.jira_service.requests.get')
    def test_is_online(self, mock_get):
        """Test online connectivity check."""
        # Test online
        mock_get.return_value = Mock()
        self.assertTrue(self.service._is_online())

        # Test offline
        mock_get.side_effect = Exception('Connection error')
        self.assertFalse(self.service._is_online())

    def test_offline_fallback(self):
        """Test fallback to cache when offline."""
        # Cache some tickets
        tickets = [
            {
                'key': 'CACHED-1',
                'fields': {'summary': 'Cached issue', 'status': {'name': 'To Do'}}
            }
        ]
        self.service._cache_tickets(tickets)

        # Mock offline
        with patch.object(self.service, '_is_online', return_value=False):
            result = self.service.fetch_assigned_tickets()

        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]['key'], 'CACHED-1')


if __name__ == '__main__':
    unittest.main()
