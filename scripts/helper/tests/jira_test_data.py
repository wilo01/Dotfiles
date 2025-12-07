#!/usr/bin/env python3
"""Generate 100 JIRA tickets + 90 worklogs for testing batch sync."""

import random
import requests  # [ ] TODO: Library stubs not installed for "requests"
from datetime import datetime
from pathlib import Path

JIRA_URL = "http://localhost:8080"
PROJECT_KEY = "VIS"  # Changed from TDT
SESSION_ID = "A03D08E7142CF4B23DE1DCBB55C503BD"

HEADERS = {
    "Cookie": f"JSESSIONID={SESSION_ID}",
    "Content-Type": "application/json",
}

TASK_TOPICS = [
    "Refactor authentication", "Update dependencies", "Implement caching",
    "Review PR changes", "Document API endpoints", "Optimize queries",
    "Add logging", "Configure CI/CD", "Setup monitoring", "Create tests",
    "Fix unit tests", "Update config", "Migrate data", "Add validation",
    "Improve error handling", "Refactor service layer", "Add metrics",
]

BUG_TOPICS = [
    "Fix login timeout", "Resolve null pointer", "Handle edge case",
    "Fix memory leak", "Correct validation", "Patch security issue",
    "Fix race condition", "Handle encoding", "Fix date parsing",
]

STORY_TOPICS = [
    "User dashboard", "Export feature", "Search functionality",
    "Notification system", "Report generator", "User preferences",
    "Audit logging", "Batch processing", "API integration",
]


def generate_random_time() -> str:
    """Generate random time between 30m and 4h."""
    minutes = random.randint(30, 240)
    if minutes >= 60:
        h, m = divmod(minutes, 60)
        return f"{h}h {m}m" if m else f"{h}h"
    return f"{minutes}m"


def create_single_ticket(issue_type: str, number: int, topic: str) -> str:
    """Create a single ticket and return its key."""
    payload = {
        "fields": {
            "project": {"key": PROJECT_KEY},
            "summary": f"{issue_type} #{number} - {topic}",
            "issuetype": {"name": issue_type},
            "description": "Auto-generated for worklog sync testing"
        }
    }

    resp = requests.post(
        f"{JIRA_URL}/rest/api/2/issue",
        json=payload,
        headers=HEADERS
    )

    if resp.ok:
        key = resp.json()["key"]
        print(f"  Created: {key} - {issue_type} #{number}")
        return key
    else:
        print(f"  FAILED: {issue_type} #{number} - {resp.status_code}: {resp.text[:100]}")
        return None  # [ ] TODO: Type "None" is not assignable to return type "str"   "None" is not assignab...; Incompatible return value type (got "None", expected "str")


def create_tickets() -> list:
    """Create 100 tickets: 60 Task, 25 Bug, 15 Story."""
    keys = []

    print("Creating 60 Tasks...")
    for i in range(60):
        key = create_single_ticket("Task", i + 1, random.choice(TASK_TOPICS))
        if key:
            keys.append(key)

    print("Creating 25 Bugs...")
    for i in range(25):
        key = create_single_ticket("Bug", i + 1, random.choice(BUG_TOPICS))
        if key:
            keys.append(key)

    print("Creating 15 Stories...")
    for i in range(15):
        key = create_single_ticket("Story", i + 1, random.choice(STORY_TOPICS))
        if key:
            keys.append(key)

    return keys


def log_worklogs(ticket_keys: list, percentage: float = 0.9):
    """Log worklogs on 90% of tickets with random times."""
    count = int(len(ticket_keys) * percentage)
    selected = random.sample(ticket_keys, count)

    print(f"\nLogging worklogs on {count} tickets...")

    success = 0
    for key in selected:
        time_spent = generate_random_time()
        payload = {
            "timeSpent": time_spent,
            "started": datetime.now().strftime("%Y-%m-%dT09:00:00.000+0000"),
            "comment": "Auto-generated worklog for testing"
        }
        resp = requests.post(
            f"{JIRA_URL}/rest/api/2/issue/{key}/worklog",
            json=payload,
            headers=HEADERS
        )
        status = "OK" if resp.ok else f"FAIL({resp.status_code})"
        print(f"  {key}: {time_spent} - {status}")
        if resp.ok:
            success += 1

    print(f"\nWorklogs created: {success}/{count}")
    return selected


def generate_csv(ticket_keys: list, csv_path: str):
    """Generate CSV for batch sync testing."""
    count = int(len(ticket_keys) * 0.9)
    selected = random.sample(ticket_keys, count)
    today = datetime.now().strftime("%d/%m/%Y")

    path = Path(csv_path).expanduser()
    path.parent.mkdir(parents=True, exist_ok=True)

    with open(path, "w") as f:
        f.write("issue_key,time_spent,date,status\n")
        for key in selected:
            time_spent = generate_random_time()
            f.write(f"{key},{time_spent},{today},\n")

    print(f"\nCSV written: {path} ({count} entries)")


if __name__ == "__main__":
    print("=" * 50)
    print("JIRA Test Data Generator")
    print("=" * 50)
    print(f"Target: {JIRA_URL} / Project: {PROJECT_KEY}")
    print()

    print("Phase 1: Creating 100 tickets...")
    keys = create_tickets()
    print(f"\nCreated: {len(keys)} tickets")

    if keys:
        print("\nPhase 2: Logging worklogs on 90% of tickets...")
        log_worklogs(keys)

        print("\nPhase 3: Generating test CSV...")
        generate_csv(keys, "~/.config/helper-cli/worklogs-test.csv")

    print("\n" + "=" * 50)
    print("Done!")
    print("=" * 50)
