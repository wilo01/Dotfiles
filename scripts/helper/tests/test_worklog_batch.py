"""Basic smoke tests for worklog_batch module."""

from pathlib import Path

import sys

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from helper_cli.worklog_batch import WorklogEntry, WorklogBatchProcessor


class TestWorklogEntry:
    """Tests for WorklogEntry dataclass."""

    def test_pending_entry(self):
        """Test pending entry status."""
        entry = WorklogEntry(issue_key="VIS-1", time_spent="2h", date="07.12.2025")
        assert entry.is_pending() is True
        assert entry.is_updated() is False
        assert entry.is_done() is False
        assert entry.needs_processing() is True

    def test_updated_entry(self):
        """Test updated entry status."""
        entry = WorklogEntry(
            issue_key="VIS-1", time_spent="2h", date="07.12.2025", status="UPDATED"
        )
        assert entry.is_pending() is False
        assert entry.is_updated() is True
        assert entry.is_done() is False
        assert entry.needs_processing() is True

    def test_done_entry(self):
        """Test done entry status."""
        entry = WorklogEntry(
            issue_key="VIS-1", time_spent="2h", date="07.12.2025", status="DONE"
        )
        assert entry.is_pending() is False
        assert entry.is_updated() is False
        assert entry.is_done() is True
        assert entry.needs_processing() is False

    def test_status_case_insensitive(self):
        """Test that status checking is case-insensitive."""
        entry_lower = WorklogEntry(
            issue_key="VIS-1", time_spent="2h", date="07.12.2025", status="done"
        )
        entry_upper = WorklogEntry(
            issue_key="VIS-1", time_spent="2h", date="07.12.2025", status="DONE"
        )
        assert entry_lower.is_done() is True
        assert entry_upper.is_done() is True


class TestWorklogBatchProcessor:
    """Tests for WorklogBatchProcessor."""

    def test_is_cloud_detection_atlassian_net(self):
        """Test JIRA Cloud detection for .atlassian.net domains."""
        processor = WorklogBatchProcessor(
            base_url="https://company.atlassian.net",
            email="test@test.com",
            api_token="token",
        )
        assert processor.is_cloud is True

    def test_is_cloud_detection_localhost(self):
        """Test JIRA Server detection for localhost."""
        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        assert processor.is_cloud is False

    def test_is_cloud_detection_custom_domain(self):
        """Test JIRA Server detection for custom domains."""
        processor = WorklogBatchProcessor(
            base_url="https://jira.mycompany.com",
            email="test@test.com",
            api_token="token",
        )
        assert processor.is_cloud is False

    def test_parse_date_to_iso(self):
        """Test date parsing to ISO format."""
        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        result = processor.parse_date_to_iso("07.12.2025", "09:00")
        assert result == "2025-12-07T09:00:00.000+0000"

    def test_parse_date_to_iso_with_custom_time(self):
        """Test date parsing with custom time."""
        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        result = processor.parse_date_to_iso("25.11.2025", "14:30")
        assert result == "2025-11-25T14:30:00.000+0000"

    def test_parse_csv_creates_entries(self, tmp_path):
        """Test CSV parsing creates WorklogEntry objects."""
        csv_file = tmp_path / "worklogs.csv"
        csv_file.write_text("VIS-123,2h,07.12.2025,Test description,\n")

        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        entries = processor.parse_csv(csv_file)

        assert len(entries) == 1
        assert entries[0].issue_key == "VIS-123"
        assert entries[0].time_spent == "2h"
        assert entries[0].date == "07.12.2025"
        assert entries[0].description == "Test description"
        assert entries[0].status == ""

    def test_parse_csv_skips_done_entries(self, tmp_path):
        """Test that parse_csv skips DONE entries by default."""
        csv_file = tmp_path / "worklogs.csv"
        csv_file.write_text(
            "VIS-123,2h,07.12.2025,Pending,\n"
            "VIS-124,1h,07.12.2025,Done,DONE\n"
            "VIS-125,3h,07.12.2025,Updated,UPDATED\n"
        )

        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        entries = processor.parse_csv(csv_file)

        assert len(entries) == 2
        assert entries[0].issue_key == "VIS-123"
        assert entries[1].issue_key == "VIS-125"

    def test_parse_csv_skips_header(self, tmp_path):
        """Test that CSV header row is skipped."""
        csv_file = tmp_path / "worklogs.csv"
        csv_file.write_text(
            "issue_key,time_spent,date,description,status\n"
            "VIS-123,2h,07.12.2025,Test,\n"
        )

        processor = WorklogBatchProcessor(
            base_url="http://localhost:8080",
            email="test@test.com",
            api_token="token",
        )
        entries = processor.parse_csv(csv_file)

        assert len(entries) == 1
        assert entries[0].issue_key == "VIS-123"
