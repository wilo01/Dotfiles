"""Basic smoke tests for timesheet module."""

import pytest
from pathlib import Path

import sys

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from helper_cli.timesheet import TimesheetLogger, save_timesheet_config


def test_timesheet_logger_init():
    """Test TimesheetLogger initialization."""
    logger = TimesheetLogger(jira_url="https://test.atlassian.net")
    assert logger.jira_url == "https://test.atlassian.net"
    assert logger.graphql_endpoint == "https://test.atlassian.net/gateway/api/graphql"


def test_timesheet_logger_init_with_trailing_slash():
    """Test that trailing slash is stripped from URL."""
    logger = TimesheetLogger(jira_url="https://test.atlassian.net/")
    assert logger.jira_url == "https://test.atlassian.net"


def test_timesheet_logger_sets_cookies():
    """Test that cookies are set correctly."""
    logger = TimesheetLogger(
        jira_url="https://test.atlassian.net",
        session_token="token1",
        xsrf_token="token2",
        jsessionid="token3",
    )
    cookies = logger.session.cookies.get_dict()
    assert cookies.get("tenant.session.token") == "token1"
    assert cookies.get("atlassian.xsrf.token") == "token2"
    assert cookies.get("JSESSIONID") == "token3"


def test_from_config_file_not_found():
    """Test error when config file doesn't exist."""
    with pytest.raises(FileNotFoundError) as exc_info:
        TimesheetLogger.from_config_file(Path("/nonexistent/path.json"))
    assert "Config file not found" in str(exc_info.value)


def test_save_config_creates_file(tmp_path):
    """Test that save_timesheet_config creates config file."""
    config_path = tmp_path / "timesheet.json"
    result = save_timesheet_config(
        jira_url="https://test.atlassian.net",
        session_token="token1",
        xsrf_token="token2",
        jsessionid="token3",
        config_path=config_path,
    )
    assert result == config_path
    assert config_path.exists()


def test_save_config_sets_permissions(tmp_path):
    """Test that config file has restrictive permissions."""
    config_path = tmp_path / "timesheet.json"
    save_timesheet_config(
        jira_url="https://test.atlassian.net",
        session_token="token1",
        xsrf_token="token2",
        jsessionid="token3",
        config_path=config_path,
    )
    # Check permissions (0o600 = user read/write only)
    mode = config_path.stat().st_mode & 0o777
    assert mode == 0o600
