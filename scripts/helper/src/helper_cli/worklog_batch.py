"""Batch worklog processor for JIRA REST API.

Processes CSV files with time entries and posts them to JIRA.
Inspired by: https://www.youtube.com/watch?v=NP_e3-2FL4U
"""

import base64
import csv
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Callable, List, Optional

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry


@dataclass
class WorklogEntry:
    """Single worklog entry from CSV.

    CSV format: issue_key,time_spent,date,description,status
    Status flow: empty -> UPDATED -> DONE
    """

    issue_key: str
    time_spent: str
    date: str
    description: str = ""  # optional worklog comment
    status: str = ""  # empty/UPDATED/DONE
    row_number: int = 0

    def is_pending(self) -> bool:
        """Check if entry needs to be posted (empty status)."""
        return self.status == ""

    def is_updated(self) -> bool:
        """Check if entry was posted but needs verification."""
        return self.status.upper() == "UPDATED"

    def is_done(self) -> bool:
        """Check if entry is verified and synced."""
        return self.status.upper() == "DONE"

    def needs_processing(self) -> bool:
        """Check if entry needs any processing (pending or updated)."""
        return not self.is_done()


@dataclass
class WorklogResult:
    """Result of a single worklog submission."""

    entry: WorklogEntry
    success: bool
    error_message: Optional[str] = None
    jira_worklog_id: Optional[str] = None
    new_status: str = ""  # UPDATED or DONE


class WorklogBatchProcessor:
    """Process batch worklog entries from CSV file."""

    DEFAULT_CSV_PATH = Path.home() / ".config" / "helper-cli" / "worklogs.csv"

    def __init__(self, base_url: str, email: str, api_token: str):
        """Initialize processor with JIRA credentials.

        Args:
            base_url: JIRA base URL (e.g., https://company.atlassian.net)
            email: JIRA account email
            api_token: JIRA API token (Cloud) or Personal Access Token (Server/DC)
        """
        self.base_url = base_url.rstrip("/")
        self.email = email
        self.api_token = api_token

        # Detect if this is JIRA Cloud or Server/DC
        # Cloud URLs typically contain .atlassian.net
        self.is_cloud = ".atlassian.net" in base_url.lower()

        # Setup session with retry strategy
        self.session = requests.Session()
        retry = Retry(
            total=3,
            read=3,
            connect=3,
            backoff_factor=1,
            status_forcelist=(500, 502, 503, 504),
        )
        adapter = HTTPAdapter(max_retries=retry)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def _get_headers(self) -> dict:
        """Get headers for JIRA API requests."""
        if self.is_cloud:
            # JIRA Cloud: Basic auth with email:api_token
            auth = base64.b64encode(f"{self.email}:{self.api_token}".encode()).decode()
            return {
                "Authorization": f"Basic {auth}",
                "Accept": "application/json",
                "Content-Type": "application/json",
            }
        else:
            # JIRA Server/DC: Bearer token (Personal Access Token)
            return {
                "Authorization": f"Bearer {self.api_token}",
                "Accept": "application/json",
                "Content-Type": "application/json",
            }

    def parse_csv(self, file_path: Path, done_only: bool = False) -> List[WorklogEntry]:
        """Parse CSV file into WorklogEntry objects.

        CSV Format: issue_key,time_spent,date,description,status
        Date format: DD.MM.YYYY
        Status: empty = pending, UPDATED = needs verify, DONE = skip

        Args:
            file_path: Path to CSV file
            done_only: If False, only return entries that need processing (pending/updated)

        Returns:
            List of WorklogEntry objects
        """
        entries = []
        with open(file_path, newline="", encoding="utf-8") as csvfile:
            reader = csv.reader(csvfile)
            for row_num, row in enumerate(reader, start=1):
                if not row or len(row) < 3:
                    continue

                # Skip header row if present
                if row_num == 1 and row[0].lower() in ("issue_key", "issue", "ticket"):
                    continue

                issue_key = row[0].strip().upper()
                time_spent = row[1].strip()
                date = row[2].strip()
                description = row[3].strip() if len(row) > 3 else ""
                status = row[4].strip() if len(row) > 4 else ""

                entry = WorklogEntry(
                    issue_key=issue_key,
                    time_spent=time_spent,
                    date=date,
                    description=description,
                    status=status,
                    row_number=row_num,
                )

                if not done_only and entry.is_done():
                    continue

                entries.append(entry)

        return entries

    def parse_csv_all(self, file_path: Path) -> List[WorklogEntry]:
        """Parse all entries from CSV (including done ones).

        Used for rewriting the CSV file with updated statuses.
        """
        return self.parse_csv(file_path, done_only=True)

    def parse_date_to_iso(self, date_str: str, time_str: str = "09:00") -> str:
        """Convert DD.MM.YYYY to ISO 8601 format for JIRA.

        Args:
            date_str: Date in DD.MM.YYYY format
            time_str: Time in HH:MM format (default: 09:00)

        Returns:
            ISO 8601 formatted string for JIRA API
        """
        day, month, year = date_str.split(".")
        hour, minute = time_str.split(":")

        # Build datetime and format for JIRA
        dt = datetime(
            year=int(year),
            month=int(month),
            day=int(day),
            hour=int(hour),
            minute=int(minute),
        )
        return dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000")

    def post_worklog(
        self, entry: WorklogEntry, default_time: str = "09:00"
    ) -> WorklogResult:
        """Post single worklog to JIRA REST API.

        Args:
            entry: WorklogEntry to post
            default_time: Default start time (HH:MM)

        Returns:
            WorklogResult with success/failure info
        """
        # Use API v3 for Cloud, v2 for Server/DC
        api_version = "3" if self.is_cloud else "2"
        url = f"{self.base_url}/rest/api/{api_version}/issue/{entry.issue_key}/worklog"

        try:
            started = self.parse_date_to_iso(entry.date, default_time)
        except (ValueError, IndexError) as e:
            return WorklogResult(
                entry=entry,
                success=False,
                error_message=f"Invalid date format: {entry.date} ({e})",
            )

        # Build request body
        body = {
            "timeSpent": entry.time_spent,
            "started": started,
        }

        # Add description/comment if provided
        if entry.description:
            if self.is_cloud:
                # JIRA Cloud API v3 uses Atlassian Document Format
                body["comment"] = {
                    "type": "doc",
                    "version": 1,
                    "content": [
                        {
                            "type": "paragraph",
                            "content": [{"type": "text", "text": entry.description}],
                        }
                    ],
                }
            else:
                # JIRA Server/DC API v2 uses plain text
                body["comment"] = entry.description

        try:
            response = self.session.post(url, headers=self._get_headers(), json=body)

            if response.status_code == 201:
                data = response.json()
                return WorklogResult(
                    entry=entry,
                    success=True,
                    jira_worklog_id=data.get("id"),
                )
            elif response.status_code == 401:
                return WorklogResult(
                    entry=entry,
                    success=False,
                    error_message="Authentication failed - check credentials",
                )
            elif response.status_code == 403:
                return WorklogResult(
                    entry=entry,
                    success=False,
                    error_message="Permission denied - check access to issue",
                )
            elif response.status_code == 404:
                return WorklogResult(
                    entry=entry,
                    success=False,
                    error_message=f"Issue {entry.issue_key} not found",
                )
            else:
                error_detail = ""
                try:
                    error_data = response.json()
                    if "errorMessages" in error_data:
                        error_detail = "; ".join(error_data["errorMessages"])
                except Exception:
                    error_detail = response.text[:200]
                return WorklogResult(
                    entry=entry,
                    success=False,
                    error_message=f"HTTP {response.status_code}: {error_detail}",
                )

        except requests.RequestException as e:
            return WorklogResult(
                entry=entry,
                success=False,
                error_message=f"Network error: {e}",
            )

    def get_worklogs(self, issue_key: str, date: str) -> List[dict]:
        """Get all worklogs for an issue on a specific date.

        Args:
            issue_key: JIRA issue key (e.g., VIS-123)
            date: Date in DD/MM/YYYY format

        Returns:
            List of worklog dicts with id, timeSpent, started, comment
        """
        api_version = "3" if self.is_cloud else "2"
        url = f"{self.base_url}/rest/api/{api_version}/issue/{issue_key}/worklog"

        try:
            response = self.session.get(url, headers=self._get_headers())
            if response.status_code != 200:
                return []

            data = response.json()
            worklogs = data.get("worklogs", [])

            # Parse target date for comparison (DD.MM.YYYY -> YYYY-MM-DD)
            day, month, year = date.split(".")
            target_date = f"{year}-{month}-{day}"

            # Filter worklogs by date
            matching = []
            for wl in worklogs:
                started = wl.get("started", "")
                # started format: "2025-12-07T09:00:00.000+0000"
                if started.startswith(target_date):
                    # Extract comment text (handles both v2 plain text and v3 ADF)
                    comment = wl.get("comment", "")
                    if isinstance(comment, dict):
                        # JIRA Cloud ADF format - extract plain text
                        try:
                            content = comment.get("content", [])
                            if content and content[0].get("content"):
                                comment = content[0]["content"][0].get("text", "")
                            else:
                                comment = ""
                        except (IndexError, KeyError, TypeError):
                            comment = ""

                    matching.append(
                        {
                            "id": wl.get("id"),
                            "timeSpent": wl.get("timeSpent"),
                            "started": started,
                            "comment": comment or "",
                        }
                    )

            return matching

        except requests.RequestException:
            return []

    def delete_worklog(self, issue_key: str, worklog_id: str) -> bool:
        """Delete a worklog by ID.

        Args:
            issue_key: JIRA issue key
            worklog_id: Worklog ID to delete

        Returns:
            True if deleted successfully
        """
        api_version = "3" if self.is_cloud else "2"
        url = f"{self.base_url}/rest/api/{api_version}/issue/{issue_key}/worklog/{worklog_id}"

        try:
            response = self.session.delete(url, headers=self._get_headers())
            return response.status_code == 204
        except requests.RequestException:
            return False

    def sync_worklog(
        self, entry: WorklogEntry, default_time: str = "09:00"
    ) -> WorklogResult:
        """Sync worklog to JIRA with status verification.

        One worklog per issue per day. Status flow: empty -> UPDATED -> DONE

        For PENDING entries (empty status):
        - If JIRA has same time -> DONE
        - If JIRA has different time -> delete old, post new -> UPDATED
        - If JIRA has no worklog -> post new -> UPDATED

        For UPDATED entries:
        - Verify JIRA time matches CSV -> if yes, DONE
        - If time differs -> keep UPDATED (needs re-sync)

        Args:
            entry: WorklogEntry to sync
            default_time: Default start time (HH:MM)

        Returns:
            WorklogResult with success/failure info and new_status
        """
        # Fetch existing worklogs for this issue/date (should be at most one)
        existing = self.get_worklogs(entry.issue_key, entry.date)

        # Check if only ONE worklog exists with matching time
        if len(existing) == 1 and existing[0].get("timeSpent", "") == entry.time_spent:
            # Already synced correctly - mark as DONE
            return WorklogResult(
                entry=entry,
                success=True,
                jira_worklog_id=existing[0].get("id"),
                error_message="(verified)"
                if entry.is_updated()
                else "(already synced)",
                new_status="DONE",
            )

        # Need to sync: delete ALL existing worklogs for this date, then post new
        result = self.post_worklog(entry, default_time)
        if result.success and existing:
            # Delete all existing worklogs for this issue/date
            deleted_ids = []
            for wl in existing:
                wl_id = wl.get("id")
                if wl_id:
                    self.delete_worklog(entry.issue_key, wl_id)
                    deleted_ids.append(wl_id)
            if deleted_ids:
                result.error_message = f"(replaced {len(deleted_ids)} worklog(s))"
        result.new_status = "UPDATED" if result.success else ""
        return result

    def process_batch(
        self,
        entries: List[WorklogEntry],
        dry_run: bool = False,
        default_time: str = "09:00",
        progress_callback: Optional[Callable[[int, int, WorklogResult], None]] = None,
    ) -> List[WorklogResult]:
        """Process all entries, continuing on individual failures.

        Args:
            entries: List of WorklogEntry objects
            dry_run: If True, don't actually post (just validate)
            default_time: Default start time (HH:MM)
            progress_callback: Called with (current, total, result) for each entry

        Returns:
            List of WorklogResult objects
        """
        results = []
        total = len(entries)

        for idx, entry in enumerate(entries, start=1):
            if dry_run:
                # Validate only
                try:
                    self.parse_date_to_iso(entry.date, default_time)
                    result = WorklogResult(
                        entry=entry, success=True, new_status=entry.status or "UPDATED"
                    )
                except (ValueError, IndexError) as e:
                    result = WorklogResult(
                        entry=entry,
                        success=False,
                        error_message=f"Invalid date: {e}",
                    )
            else:
                result = self.sync_worklog(entry, default_time)

                # Stop on auth errors
                if not result.success and "Authentication" in (
                    result.error_message or ""
                ):
                    results.append(result)
                    if progress_callback:
                        progress_callback(idx, total, result)
                    break

            results.append(result)
            if progress_callback:
                progress_callback(idx, total, result)

        return results

    def update_csv_status(self, file_path: Path, results: List[WorklogResult]) -> None:
        """Update CSV file with new statuses (UPDATED/DONE).

        Args:
            file_path: Path to CSV file
            results: List of WorklogResult from process_batch
        """
        # Build mapping of row numbers to new status
        status_updates = {
            r.entry.row_number: r.new_status
            for r in results
            if r.success and r.new_status
        }
        if not status_updates:
            return

        # Read all lines from file
        lines = []
        with open(file_path, newline="", encoding="utf-8") as f:
            reader = csv.reader(f)
            for row_num, row in enumerate(reader, start=1):
                if row_num in status_updates:
                    # Ensure row has 5 columns, update status (col 4)
                    while len(row) < 4:
                        row.append("")
                    if len(row) < 5:
                        row.append(status_updates[row_num])
                    else:
                        row[4] = status_updates[row_num]
                lines.append(row)

        # Write back
        with open(file_path, "w", newline="", encoding="utf-8") as f:
            writer = csv.writer(f)
            writer.writerows(lines)

    @staticmethod
    def append_entry(
        file_path: Path,
        issue_key: str,
        time_spent: str,
        description: str = "",
        date: Optional[str] = None,
    ) -> None:
        """Append a new worklog entry to the CSV file.

        CSV format: issue_key,time_spent,date,description,status (5 columns)

        Args:
            file_path: Path to CSV file
            issue_key: JIRA issue key (e.g., VIS-123)
            time_spent: Time spent (e.g., 2h, 30m)
            description: Optional worklog description/comment
            date: Date in DD.MM.YYYY format (defaults to today)
        """
        if date is None:
            date = datetime.now().strftime("%d.%m.%Y")

        # Ensure directory exists
        file_path.parent.mkdir(parents=True, exist_ok=True)

        # Append new row (5 columns: issue, time, date, description, empty status)
        with open(file_path, "a", newline="", encoding="utf-8") as f:
            writer = csv.writer(f)
            writer.writerow([issue_key.upper(), time_spent, date, description, ""])
