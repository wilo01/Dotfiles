"""Tests for data models."""

import unittest
from datetime import datetime, timedelta
from pathlib import Path

import sys
sys.path.insert(0, str(Path(__file__).parent.parent / 'src'))

from helper_cli.models import (
    JiraTicket,
    TimeEntry,
    WorkNote,
    DailySchedule,
    SprintInfo,
    JiraConfig,
    GoogleSheetsConfig
)


class TestJiraTicket(unittest.TestCase):
    """Test cases for JiraTicket model."""

    def test_from_api_response(self):
        """Test creating JiraTicket from API response."""
        api_data = {
            'key': 'TEST-1',
            'fields': {
                'summary': 'Test issue',
                'status': {'name': 'In Progress'},
                'assignee': {'displayName': 'John Doe'},
                'reporter': {'displayName': 'Jane Doe'},
                'priority': {'name': 'High'},
                'issuetype': {'name': 'Bug'},
                'created': '2025-01-01T10:00:00+00:00',
                'updated': '2025-01-15T14:30:00+00:00',
                'description': 'Test description',
                'labels': ['backend', 'urgent']
            }
        }

        ticket = JiraTicket.from_api_response(api_data)

        self.assertEqual(ticket.key, 'TEST-1')
        self.assertEqual(ticket.summary, 'Test issue')
        self.assertEqual(ticket.status, 'In Progress')
        self.assertEqual(ticket.assignee, 'John Doe')
        self.assertEqual(ticket.reporter, 'Jane Doe')
        self.assertEqual(ticket.priority, 'High')
        self.assertEqual(ticket.issue_type, 'Bug')
        self.assertEqual(ticket.labels, ['backend', 'urgent'])

    def test_to_dict(self):
        """Test converting JiraTicket to dictionary."""
        ticket = JiraTicket(
            key='TEST-1',
            summary='Test issue',
            status='To Do',
            assignee='User 1',
            priority='Medium'
        )

        result = ticket.to_dict()

        self.assertEqual(result['key'], 'TEST-1')
        self.assertEqual(result['summary'], 'Test issue')
        self.assertEqual(result['status'], 'To Do')
        self.assertEqual(result['assignee'], 'User 1')
        self.assertEqual(result['priority'], 'Medium')


class TestTimeEntry(unittest.TestCase):
    """Test cases for TimeEntry model."""

    def test_duration_in_seconds(self):
        """Test duration conversion to seconds."""
        test_cases = [
            ('2h', 7200),
            ('30m', 1800),
            ('1h 30m', 5400),
            ('2h30m', 9000),
            ('3', 10800),  # Just number = hours
        ]

        for duration, expected_seconds in test_cases:
            entry = TimeEntry(
                ticket_key='TEST-1',
                duration=duration,
                description='Test',
                started_at=datetime.now()
            )
            self.assertEqual(entry.duration_in_seconds(), expected_seconds)

    def test_to_dict_from_dict(self):
        """Test dictionary conversion roundtrip."""
        original = TimeEntry(
            ticket_key='TEST-1',
            duration='2h 30m',
            description='Working on feature',
            started_at=datetime.now(),
            synced=True
        )

        dict_form = original.to_dict()
        restored = TimeEntry.from_dict(dict_form)

        self.assertEqual(restored.ticket_key, original.ticket_key)
        self.assertEqual(restored.duration, original.duration)
        self.assertEqual(restored.description, original.description)
        self.assertEqual(restored.synced, original.synced)


class TestWorkNote(unittest.TestCase):
    """Test cases for WorkNote model."""

    def test_to_standup_format(self):
        """Test formatting as standup notes."""
        note = WorkNote(
            date=datetime(2025, 1, 15),
            ticket_keys=['TEST-1', 'TEST-2'],
            summary='Daily standup',
            yesterday='Completed TEST-1\nReviewed PRs',
            today='Start TEST-2\nTeam meeting',
            blockers='Waiting for design approval'
        )

        result = note.to_standup_format()

        self.assertIn('January 15, 2025', result)
        self.assertIn('Yesterday:', result)
        self.assertIn('• Completed TEST-1', result)
        self.assertIn('• Reviewed PRs', result)
        self.assertIn('Today:', result)
        self.assertIn('• Start TEST-2', result)
        self.assertIn('• Team meeting', result)
        self.assertIn('Blockers:', result)
        self.assertIn('• Waiting for design approval', result)

    def test_no_blockers(self):
        """Test standup format with no blockers."""
        note = WorkNote(
            date=datetime.now(),
            ticket_keys=['TEST-1'],
            summary='Standup',
            yesterday='Work',
            today='More work'
        )

        result = note.to_standup_format()
        self.assertIn('• None', result)  # Should show "None" for blockers


class TestDailySchedule(unittest.TestCase):
    """Test cases for DailySchedule model."""

    def test_task_management(self):
        """Test task management functions."""
        schedule = DailySchedule(date=datetime.now())

        # Add planned tasks
        schedule.add_planned_task('Implement feature A')
        schedule.add_planned_task('Review PR')
        schedule.add_planned_task('Implement feature A')  # Duplicate

        self.assertEqual(len(schedule.planned_tasks), 2)  # No duplicates

        # Mark completed
        schedule.mark_completed('Implement feature A')
        self.assertIn('Implement feature A', schedule.completed_tasks)

        # Completion rate
        self.assertEqual(schedule.completion_rate(), 50.0)

    def test_time_logging(self):
        """Test time logging functionality."""
        schedule = DailySchedule(date=datetime.now())

        schedule.log_time('TEST-1', '2h')
        schedule.log_time('TEST-2', '1h 30m')
        schedule.log_time('TEST-1', '30m')  # Additional time

        self.assertEqual(schedule.time_logged['TEST-1'], '2h + 30m')
        self.assertEqual(schedule.time_logged['TEST-2'], '1h 30m')


class TestSprintInfo(unittest.TestCase):
    """Test cases for SprintInfo model."""

    def test_sprint_properties(self):
        """Test sprint computed properties."""
        start = datetime.now() - timedelta(days=7)
        end = datetime.now() + timedelta(days=7)

        sprint = SprintInfo(
            id=1,
            name='Sprint 1',
            state='active',
            start_date=start,
            end_date=end
        )

        self.assertTrue(sprint.is_active)
        self.assertEqual(sprint.days_remaining, 7)
        self.assertEqual(sprint.duration_days, 14)

    def test_closed_sprint(self):
        """Test closed sprint properties."""
        sprint = SprintInfo(
            id=2,
            name='Sprint 2',
            state='closed',
            start_date=datetime.now() - timedelta(days=14),
            end_date=datetime.now() - timedelta(days=1)
        )

        self.assertFalse(sprint.is_active)
        self.assertIsNone(sprint.days_remaining)  # Not active


class TestJiraConfig(unittest.TestCase):
    """Test cases for JiraConfig model."""

    def test_validation(self):
        """Test configuration validation."""
        # Valid config
        config = JiraConfig(
            base_url='https://test.atlassian.net',
            email='user@example.com',
            api_token='token123'
        )
        self.assertTrue(config.validate())

        # Invalid URL
        config = JiraConfig(
            base_url='not-a-url',
            email='user@example.com',
            api_token='token123'
        )
        self.assertFalse(config.validate())

        # Invalid email
        config = JiraConfig(
            base_url='https://test.atlassian.net',
            email='not-an-email',
            api_token='token123'
        )
        self.assertFalse(config.validate())

        # Missing token
        config = JiraConfig(
            base_url='https://test.atlassian.net',
            email='user@example.com',
            api_token=''
        )
        self.assertFalse(config.validate())


class TestGoogleSheetsConfig(unittest.TestCase):
    """Test cases for GoogleSheetsConfig model."""

    def test_validation(self):
        """Test configuration validation."""
        # Valid with JSON
        config = GoogleSheetsConfig(
            sheet_id='sheet123',
            credentials_json='{"type": "service_account"}'
        )
        self.assertTrue(config.validate())

        # Valid with file path
        config = GoogleSheetsConfig(
            sheet_id='sheet123',
            credentials_file='/path/to/creds.json'
        )
        self.assertTrue(config.validate())

        # Missing sheet ID
        config = GoogleSheetsConfig(
            sheet_id='',
            credentials_json='{"type": "service_account"}'
        )
        self.assertFalse(config.validate())

        # Missing credentials
        config = GoogleSheetsConfig(
            sheet_id='sheet123'
        )
        self.assertFalse(config.validate())

    def test_default_values(self):
        """Test default configuration values."""
        config = GoogleSheetsConfig(
            sheet_id='sheet123',
            credentials_json='{"test": true}'
        )

        self.assertEqual(config.sheet_name_pattern, '%B %Y')
        self.assertEqual(config.work_column, 'B')
        self.assertEqual(config.date_column, 'A')
        self.assertTrue(config.auto_create_monthly_sheet)


if __name__ == '__main__':
    unittest.main()
