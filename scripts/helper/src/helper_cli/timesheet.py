"""Timesheet logging functionality for Jira Cloud Timesheet Tracking plugin."""

import json
import requests
from datetime import datetime, timezone
from typing import Optional, Dict, Any
from pathlib import Path


class TimesheetLogger:
    """Log time entries to Jira Cloud Timesheet Tracking plugin."""

    def __init__(
        self,
        jira_url: str,
        session_token: Optional[str] = None,
        xsrf_token: Optional[str] = None,
        jsessionid: Optional[str] = None,
    ):
        """Initialize timesheet logger.

        Args:
            jira_url: Base Jira Cloud URL (e.g., https://acre-identity.atlassian.net)
            session_token: tenant.session.token cookie value
            xsrf_token: atlassian.xsrf.token cookie value
            jsessionid: JSESSIONID cookie value
        """
        self.jira_url = jira_url.rstrip("/")
        self.graphql_endpoint = f"{self.jira_url}/gateway/api/graphql"
        self.session = requests.Session()

        # Set up cookies
        if session_token:
            self.session.cookies.set("tenant.session.token", session_token)
        if xsrf_token:
            self.session.cookies.set("atlassian.xsrf.token", xsrf_token)
        if jsessionid:
            self.session.cookies.set("JSESSIONID", jsessionid)

        # Standard headers
        self.session.headers.update(
            {
                "Content-Type": "application/json",
                "Accept": "*/*",
                "apollographql-client-name": "GATEWAY",
                "cache-control": "no-cache",
                "pragma": "no-cache",
            }
        )

    def log_work(
        self,
        issue_key: str,
        time_spent: str,
        started: Optional[datetime] = None,
        comment: str = "",
        project_id: Optional[str] = None,
        issue_id: Optional[str] = None,
        context_token: Optional[str] = None,
        extension_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Log work time to a Jira issue.

        Args:
            issue_key: Jira issue key (e.g., 'TDT-2')
            time_spent: Time spent in Jira format (e.g., '30m', '2h', '1d 4h')
            started: When the work started (defaults to now)
            comment: Optional comment for the worklog
            project_id: Project ID (optional, will be fetched if not provided)
            issue_id: Issue ID (optional, will be fetched if not provided)
            context_token: Forge context token (optional)
            extension_id: Extension ID (optional)

        Returns:
            Response data from the GraphQL mutation
        """
        if started is None:
            started = datetime.now(timezone.utc)

        # Format timestamp for Jira
        started_str = started.strftime("%Y-%m-%dT%H:%M:%S.000%z")
        if started_str.endswith("+0000"):
            started_str = started_str.replace("+0000", "+0000")

        # Build the GraphQL mutation
        mutation = {
            "operationName": "forge_ui_invokeExtension",
            "variables": {
                "input": {
                    "contextIds": [],
                    "extensionId": extension_id
                    or "ari:cloud:ecosystem::extension/48cddd1c-9611-49cd-a49c-9ff6f9bee138/42ccec14-73af-418b-beb9-c76305bfe500/static/timesheet-global-page",
                    "payload": {
                        "call": {
                            "path": "/forge/worklog",
                            "method": "POST",
                            "body": {
                                "issueIdOrKey": issue_key,
                                "timeSpent": time_spent,
                                "started": started_str,
                                "newEstimate": "0m",
                                "attributes": {},
                            },
                            "invokeType": "ui-remote-fetch",
                        },
                        "context": {
                            "cloudId": "",
                            "localId": "ari:cloud:ecosystem::extension/48cddd1c-9611-49cd-a49c-9ff6f9bee138/42ccec14-73af-418b-beb9-c76305bfe500/static/timesheet-global-page",
                            "environmentId": "42ccec14-73af-418b-beb9-c76305bfe500",
                            "environmentType": "PRODUCTION",
                            "moduleKey": "timesheet-global-page",
                            "siteUrl": self.jira_url,
                            "appVersion": "4.9.0",
                            "extension": {
                                "type": "jira:globalPage",
                                "jira": {"isNewNavigation": True},
                            },
                        },
                    },
                    "entryPoint": "resolver",
                }
            },
            "query": """mutation forge_ui_invokeExtension($input: InvokeExtensionInput!) {
  invokeExtension(input: $input) {
    success
    response {
      body
      __typename
    }
    contextToken {
      jwt
      expiresAt
      __typename
    }
    errors {
      message
      extensions {
        errorType
        statusCode
        ... on InvokeExtensionPayloadErrorExtension {
          fields {
            authInfoUrl
            __typename
          }
          __typename
        }
        __typename
      }
      __typename
    }
    __typename
  }
}
""",
        }

        # Add optional fields
        if project_id:
            mutation["variables"]["input"]["payload"]["call"]["body"]["projectId"] = project_id
        if issue_id:
            mutation["variables"]["input"]["payload"]["call"]["body"]["issueId"] = issue_id
        if comment:
            mutation["variables"]["input"]["payload"]["call"]["body"]["comment"] = comment
        if context_token:
            mutation["variables"]["input"]["payload"]["contextToken"] = context_token

        # Make the request
        response = self.session.post(self.graphql_endpoint, json=mutation)
        response.raise_for_status()

        return response.json()

    @classmethod
    def from_config_file(cls, config_path: Optional[Path] = None) -> "TimesheetLogger":
        """Create logger from config file.

        Args:
            config_path: Path to config file (defaults to ~/.config/helper-cli/timesheet.json)

        Returns:
            Configured TimesheetLogger instance
        """
        if config_path is None:
            config_path = Path.home() / ".config" / "helper-cli" / "timesheet.json"

        if not config_path.exists():
            raise FileNotFoundError(
                f"Config file not found: {config_path}\n"
                "Create it with: helper timesheet setup"
            )

        with open(config_path) as f:
            config = json.load(f)

        return cls(
            jira_url=config["jira_url"],
            session_token=config.get("session_token"),
            xsrf_token=config.get("xsrf_token"),
            jsessionid=config.get("jsessionid"),
        )


def save_timesheet_config(
    jira_url: str,
    session_token: str,
    xsrf_token: str,
    jsessionid: str,
    config_path: Optional[Path] = None,
) -> Path:
    """Save timesheet configuration to file.

    Args:
        jira_url: Base Jira Cloud URL
        session_token: tenant.session.token cookie value
        xsrf_token: atlassian.xsrf.token cookie value
        jsessionid: JSESSIONID cookie value
        config_path: Path to save config (defaults to ~/.config/helper-cli/timesheet.json)

    Returns:
        Path to saved config file
    """
    if config_path is None:
        config_path = Path.home() / ".config" / "helper-cli" / "timesheet.json"

    config_path.parent.mkdir(parents=True, exist_ok=True)

    config = {
        "jira_url": jira_url,
        "session_token": session_token,
        "xsrf_token": xsrf_token,
        "jsessionid": jsessionid,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }

    with open(config_path, "w") as f:
        json.dump(config, f, indent=2)

    # Set restrictive permissions (user read/write only)
    config_path.chmod(0o600)

    return config_path
