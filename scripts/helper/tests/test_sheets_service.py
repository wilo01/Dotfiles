"""Tests for GoogleSheetsService."""

import unittest
from unittest.mock import Mock, patch, MagicMock, call
from datetime import datetime, timedelta
from pathlib import Path

import sys
sys.path.insert(0, str(Path(__file__).parent.parent / 'src'))

# Mock Google imports if not available
try:
    from googleapiclient.errors import HttpError
except ImportError:
    class HttpError(Exception):
        pass

from helper_cli.services.google_sheets_service import GoogleSheetsService


class TestGoogleSheetsService(unittest.TestCase):
    """Test cases for GoogleSheetsService."""

    def setUp(self):
        """Set up test fixtures."""
        self.service = GoogleSheetsService(
            sheet_id='test_sheet_id',
            credentials_path=None  # Don't initialize real service
        )

        # Mock the service
        self.mock_sheets_service = MagicMock()
        self.service.service = self.mock_sheets_service

    def test_find_today_cell(self):
        """Test finding today's cell in the sheet."""
        # Mock current date
        test_date = datetime(2025, 1, 15)

        with patch('helper_cli.services.google_sheets_service.datetime') as mock_datetime:
            mock_datetime.now.return_value = test_date
            mock_datetime.strftime = datetime.strftime

            # Mock sheet values
            mock_values = {
                'values': [
                    ['January 1'],
                    ['January 2'],
                    ['January 15'],  # Today
                    ['January 16']
                ]
            }

            self.mock_sheets_service.spreadsheets().values().get().execute.return_value = mock_values

            result = self.service.find_today_cell()

            # Should find row 3 (1-indexed)
            self.assertEqual(result, 'January 2025!B3')

    def test_find_today_cell_not_found(self):
        """Test when today's cell is not found."""
        test_date = datetime(2025, 1, 15)

        with patch('helper_cli.services.google_sheets_service.datetime') as mock_datetime:
            mock_datetime.now.return_value = test_date
            mock_datetime.strftime = datetime.strftime

            # Mock sheet values without today's date
            mock_values = {
                'values': [
                    ['January 1'],
                    ['January 2'],
                    ['January 16']
                ]
            }

            self.mock_sheets_service.spreadsheets().values().get().execute.return_value = mock_values

            result = self.service.find_today_cell()
            self.assertIsNone(result)

    def test_write_standup_notes(self):
        """Test writing standup notes."""
        notes = "Test standup notes"

        with patch.object(self.service, 'find_today_cell', return_value='January 2025!B15'):
            result = self.service.write_standup_notes(notes)

        self.assertTrue(result)

        # Verify update was called
        self.mock_sheets_service.spreadsheets().values().update.assert_called_once()
        call_args = self.mock_sheets_service.spreadsheets().values().update.call_args
        self.assertEqual(call_args[1]['spreadsheetId'], 'test_sheet_id')
        self.assertEqual(call_args[1]['range'], 'January 2025!B15')
        self.assertEqual(call_args[1]['body']['values'], [[notes]])

    def test_append_work_log(self):
        """Test appending work log to today's cell."""
        # Mock finding today's cell
        with patch.object(self.service, 'find_today_cell', return_value='January 2025!B15'):
            # Mock current cell content
            self.mock_sheets_service.spreadsheets().values().get().execute.return_value = {
                'values': [['Existing content']]
            }

            # Mock time
            test_time = datetime(2025, 1, 15, 14, 30)
            with patch('helper_cli.services.google_sheets_service.datetime') as mock_datetime:
                mock_datetime.now.return_value = test_time
                mock_datetime.strftime = datetime.strftime

                result = self.service.append_work_log('TEST-1', '2h', 'Working on feature')

        self.assertTrue(result)

        # Verify content was appended
        update_call = self.mock_sheets_service.spreadsheets().values().update.call_args
        updated_text = update_call[1]['body']['values'][0][0]
        self.assertIn('Existing content', updated_text)
        self.assertIn('TEST-1 (2h) - Working on feature', updated_text)
        self.assertIn('[14:30]', updated_text)

    def test_generate_standup_notes(self):
        """Test generating standup notes."""
        yesterday_logs = [
            {
                'ticket_key': 'TEST-1',
                'duration': '2h',
                'description': 'Fixed bug'
            },
            {
                'ticket_key': 'TEST-2',
                'duration': '1h',
                'description': 'Code review'
            }
        ]

        today_plan = [
            'TEST-3: Implement new feature',
            'TEST-4: Write tests'
        ]

        result = self.service.generate_standup_notes(yesterday_logs, today_plan)

        # Check structure
        self.assertIn('Standup', result)
        self.assertIn('Yesterday:', result)
        self.assertIn('TEST-1: Fixed bug (2h)', result)
        self.assertIn('TEST-2: Code review (1h)', result)
        self.assertIn('Today:', result)
        self.assertIn('TEST-3: Implement new feature', result)
        self.assertIn('TEST-4: Write tests', result)
        self.assertIn('Blockers:', result)

    def test_get_today_schedule(self):
        """Test getting today's schedule."""
        with patch.object(self.service, 'find_today_cell', return_value='January 2025!B15'):
            # Mock cell content
            self.mock_sheets_service.spreadsheets().values().get().execute.return_value = {
                'values': [['• TEST-1: Morning standup\n• TEST-2: Code review\n• TEST-3: Implementation']]
            }

            result = self.service.get_today_schedule()

        self.assertEqual(len(result), 3)
        self.assertIn('• TEST-1: Morning standup', result)
        self.assertIn('• TEST-2: Code review', result)
        self.assertIn('• TEST-3: Implementation', result)

    def test_create_month_sheet(self):
        """Test creating a new month sheet."""
        test_month = datetime(2025, 2, 1)

        # Mock existing sheets
        self.mock_sheets_service.spreadsheets().get().execute.return_value = {
            'sheets': [
                {'properties': {'title': 'January 2025'}}
            ]
        }

        result = self.service.create_month_sheet(test_month)

        self.assertTrue(result)

        # Verify batch update was called to create sheet
        batch_update = self.mock_sheets_service.spreadsheets().batchUpdate
        batch_update.assert_called_once()

        # Check sheet creation request
        call_args = batch_update.call_args
        request = call_args[1]['body']['requests'][0]
        self.assertEqual(request['addSheet']['properties']['title'], 'February 2025')

    def test_create_month_sheet_already_exists(self):
        """Test creating month sheet when it already exists."""
        test_month = datetime(2025, 1, 1)

        # Mock existing sheets including the target month
        self.mock_sheets_service.spreadsheets().get().execute.return_value = {
            'sheets': [
                {'properties': {'title': 'January 2025'}}
            ]
        }

        result = self.service.create_month_sheet(test_month)

        self.assertTrue(result)

        # Should not try to create the sheet
        self.mock_sheets_service.spreadsheets().batchUpdate.assert_not_called()

    def test_read_week_notes(self):
        """Test reading week notes."""
        # Mock sheet values
        self.mock_sheets_service.spreadsheets().values().get().execute.return_value = {
            'values': [
                ['January 10', 'Work on TEST-1'],
                ['January 11', 'Complete TEST-2'],
                ['January 12', 'Review PRs'],
                ['January 13', ''],  # Empty notes
                ['January 14', 'Sprint planning']
            ]
        }

        result = self.service.read_week_notes()

        self.assertEqual(len(result), 5)
        self.assertEqual(result[0]['date'], 'January 10')
        self.assertEqual(result[0]['notes'], 'Work on TEST-1')
        self.assertEqual(result[3]['notes'], '')

    @patch('helper_cli.services.google_sheets_service.GOOGLE_SHEETS_AVAILABLE', False)
    def test_service_unavailable(self):
        """Test behavior when Google Sheets libraries are not available."""
        service = GoogleSheetsService(
            sheet_id='test_sheet',
            credentials_path='/path/to/creds'
        )

        # Service should be None
        self.assertIsNone(service.service)

        # Methods should return defaults
        self.assertIsNone(service.find_today_cell())
        self.assertFalse(service.write_standup_notes('test'))
        self.assertEqual(service.read_week_notes(), [])
        self.assertFalse(service.append_work_log('TEST-1', '1h', 'test'))

    def test_error_handling(self):
        """Test error handling for API failures."""
        # Mock HttpError
        error = Exception('API Error')
        self.mock_sheets_service.spreadsheets().values().get().execute.side_effect = error

        # Should handle error gracefully
        result = self.service.find_today_cell()
        self.assertIsNone(result)

        result = self.service.read_week_notes()
        self.assertEqual(result, [])


if __name__ == '__main__':
    unittest.main()
