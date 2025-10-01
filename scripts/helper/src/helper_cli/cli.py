"""Command line interface for Helper CLI."""

import click
import pyperclip
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from .core import HelperCore
from .domain_checker import DomainChecker
from .name_checker import NameAvailabilityChecker
from .services import JiraService, ContextDetector, GoogleSheetsService
from .config import CredentialManager
from .utils import safe_copy_to_clipboard, copy_or_print
from pathlib import Path
from datetime import datetime, timedelta

console = Console()
helper = HelperCore()
checker = DomainChecker()
name_checker = NameAvailabilityChecker()

# Initialize Jira services (lazy loading)
_jira_service = None
_context_detector = None
_sheets_service = None
_credential_manager = None


def get_jira_service():
    """Get or create Jira service instance."""
    global _jira_service, _credential_manager
    if _jira_service is None:
        if _credential_manager is None:
            _credential_manager = CredentialManager()
        creds = _credential_manager.get_jira_credentials()
        if creds:
            _jira_service = JiraService(
                base_url=creds["base_url"],
                email=creds["email"],
                api_token=creds["api_token"],
            )
        else:
            _jira_service = JiraService()  # Offline mode
    return _jira_service


def get_context_detector():
    """Get or create context detector instance."""
    global _context_detector
    if _context_detector is None:
        _context_detector = ContextDetector()
    return _context_detector


def get_sheets_service():
    """Get or create Google Sheets service instance."""
    global _sheets_service, _credential_manager

    # Check if GoogleSheetsService is available
    if GoogleSheetsService is None:
        return None

    if _sheets_service is None:
        if _credential_manager is None:
            _credential_manager = CredentialManager()
        config = _credential_manager.get_google_sheets_config()
        if config:
            _sheets_service = GoogleSheetsService(
                sheet_id=config.get("sheet_id"),
                credentials_path=config.get("credentials_file"),
            )
        else:
            _sheets_service = GoogleSheetsService()
    return _sheets_service


def get_credential_manager():
    """Get or create credential manager instance."""
    global _credential_manager
    if _credential_manager is None:
        _credential_manager = CredentialManager()
    return _credential_manager


@click.group()
@click.version_option()
def main():
    """Development helper CLI for JIRA, badges, URLs, and automation tasks."""
    pass


@main.command()
@click.argument("text")
@click.option("--copy/--no-copy", default=True, help="Copy result to clipboard")
@click.option(
    "--type", "-t", is_flag=True, help="Type command directly using wtype/ydotool"
)
def branch(text, copy, type):
    """Generate JIRA branch name from commit message text."""
    import os
    import shutil

    branch_command = helper.generate_jira_branch_name(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get("XDG_SESSION_TYPE") == "wayland":
            if not shutil.which("wtype") and not shutil.which("ydotool"):
                console.print(
                    "[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]"
                )
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(
                    branch_command,
                    True,
                    console,
                    f"✅ Copied to clipboard instead: [bold green]{branch_command}[/bold green]",
                )
                return

        helper.type_string_with_delay(branch_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{branch_command}[/bold green]")
    elif copy:
        copy_or_print(
            branch_command,
            True,
            console,
            f"✅ Copied to clipboard: [bold green]{branch_command}[/bold green]",
        )
    else:
        console.print(f"Branch command: [bold blue]{branch_command}[/bold blue]")


@main.command()
@click.argument("badge_string")
@click.option(
    "--copy", "-c", is_flag=True, help="Copy to clipboard instead of auto-typing"
)
def badge(badge_string, copy):
    """Type or copy badge string exactly as provided (e.g., '@1234#' or '$5678#')."""
    import os

    # Use the badge string exactly as provided by the user
    console.print(f"[bold blue]Badge string:[/bold blue] {badge_string}")

    # Check if we're on Wayland
    if not copy and os.environ.get("XDG_SESSION_TYPE") == "wayland":
        console.print(
            "[yellow]⚠ Auto-type not supported on Wayland. Copying to clipboard instead.[/yellow]"
        )
        copy_or_print(badge_string, True, console, "✅ Badge copied to clipboard")
    elif copy:
        copy_or_print(badge_string, True, console, "✅ Badge copied to clipboard")
    else:
        helper.type_string_with_delay(badge_string)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument("badge_id")
@click.option(
    "--copy", "-c", is_flag=True, help="Copy to clipboard instead of auto-typing"
)
def rfid(badge_id, copy):
    """Type or copy RFID badge ID exactly as provided."""
    import os
    import shutil

    # Use the badge_id exactly as provided - no prefix or suffix added
    console.print(f"[bold blue]RFID badge:[/bold blue] {badge_id}")

    if copy:
        copy_or_print(badge_id, True, console, "✅ Badge copied to clipboard")
    else:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get("XDG_SESSION_TYPE") == "wayland":
            if not shutil.which("wtype") and not shutil.which("ydotool"):
                console.print(
                    "[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]"
                )
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(
                    badge_id, True, console, "✅ Badge copied to clipboard instead"
                )
                return

        helper.type_string_with_delay(badge_id)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument("badge_id")
@click.option(
    "--copy", "-c", is_flag=True, help="Copy to clipboard instead of auto-typing"
)
def qr(badge_id, copy):
    """Type or copy QR badge ID exactly as provided."""
    import os
    import shutil

    # Use the badge_id exactly as provided - no prefix or suffix added
    console.print(f"[bold blue]QR badge:[/bold blue] {badge_id}")

    if copy:
        copy_or_print(badge_id, True, console, "✅ Badge copied to clipboard")
    else:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get("XDG_SESSION_TYPE") == "wayland":
            if not shutil.which("wtype") and not shutil.which("ydotool"):
                console.print(
                    "[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]"
                )
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(
                    badge_id, True, console, "✅ Badge copied to clipboard instead"
                )
                return

        helper.type_string_with_delay(badge_id)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument("text")
@click.option("--copy/--no-copy", default=True, help="Copy result to clipboard")
@click.option(
    "--type", "-t", is_flag=True, help="Type command directly using wtype/ydotool"
)
def stash(text, copy, type):
    """Generate git stash command with formatted message."""
    import os
    import shutil

    stash_command = helper.generate_stash_command(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get("XDG_SESSION_TYPE") == "wayland":
            if not shutil.which("wtype") and not shutil.which("ydotool"):
                console.print(
                    "[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]"
                )
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(
                    stash_command,
                    True,
                    console,
                    f"✅ Copied to clipboard instead: [bold green]{stash_command}[/bold green]",
                )
                console.print(
                    f"✅ Copied to clipboard instead: [bold green]{stash_command}[/bold green]"
                )
                return

        helper.type_string_with_delay(stash_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{stash_command}[/bold green]")
    elif copy:
        copy_or_print(
            stash_command,
            True,
            console,
            f"✅ Copied to clipboard: [bold green]{stash_command}[/bold green]",
        )
        console.print(
            f"✅ Copied to clipboard: [bold green]{stash_command}[/bold green]"
        )
    else:
        console.print(f"Stash command: [bold blue]{stash_command}[/bold blue]")


@main.command()
@click.argument("url")
@click.option("--copy/--no-copy", default=True, help="Copy QR command to clipboard")
def ngrok(url, copy):
    """Generate QR code command for NGROK URL."""
    full_url, qr_command = helper.generate_ngrok_qr(url)

    console.print(f"[bold blue]Full URL:[/bold blue] {full_url}")
    console.print(f"[bold blue]QR Command:[/bold blue] {qr_command}")

    if copy:
        copy_or_print(qr_command, True, console, "✅ QR command copied to clipboard")
        console.print("✅ QR command copied to clipboard")


@main.command()
@click.argument("url")
@click.option("--copy/--no-copy", default=True, help="Copy QR command to clipboard")
def vsc(url, copy):
    """Generate QR code command for VSC URL."""
    full_url, qr_command = helper.generate_vsc_qr(url)

    console.print(f"[bold blue]Full URL:[/bold blue] {full_url}")
    console.print(f"[bold blue]QR Command:[/bold blue] {qr_command}")

    if copy:
        copy_or_print(qr_command, True, console, "✅ QR command copied to clipboard")
        console.print("✅ QR command copied to clipboard")


@main.command()
@click.argument("rt_name")
def rt(rt_name):
    """Set up RT temporary directory with file copying."""
    helper.setup_rt_temp_directory(rt_name)
    console.print(f"✅ RT directory setup complete: [bold green]{rt_name}[/bold green]")


@main.command()
@click.argument("apex_file")
@click.argument("port")
@click.argument("auth")
@click.option("--copy/--no-copy", default=True, help="Copy command to clipboard")
def apex(apex_file, port, auth, copy):
    """Generate APEX upload curl command."""
    upload_command = helper.generate_apex_upload_command(apex_file, port, auth)

    console.print(f"[bold blue]APEX Upload Command:[/bold blue]")
    console.print(Panel(upload_command, expand=False))

    if copy:
        copy_or_print(
            upload_command, True, console, "✅ Upload command copied to clipboard"
        )
        console.print("✅ Upload command copied to clipboard")


@main.command("custom_default_data")
@click.option("--verify", "-v", is_flag=True, help="Verify paths without copying")
def custom_default_data(verify):
    """Sync custom_default_data.sql to TDS Suite Liquibase directory."""
    from pathlib import Path

    SOURCE_PATH = (
        Path.home() / "Dev" / "Private" / "DB-Scripts" / "custom_default_data.sql"
    )
    TARGET_DIR = (
        Path.home()
        / "Dev"
        / "branch-opener"
        / "branches"
        / "tds-suite"
        / "source"
        / "server"
        / "database"
        / "sql"
        / "safe"
        / "after-install"
    )

    if verify:
        console.print("[bold blue]🔍 Verification Mode[/bold blue]\n")
        console.print(f"[dim]Source:[/dim] {SOURCE_PATH}")
        console.print(f"[dim]Target:[/dim] {TARGET_DIR / 'custom_default_data.sql'}\n")

        if SOURCE_PATH.exists():
            file_size = SOURCE_PATH.stat().st_size
            console.print(f"✅ Source file exists ({file_size:,} bytes)")
        else:
            console.print(f"❌ Source file not found")

        if TARGET_DIR.exists():
            console.print(f"✅ Target directory exists")
        else:
            console.print(f"❌ Target directory not found")
        return

    success, message, file_size = helper.custom_default_data()

    if success:
        console.print(f"✅ [bold green]{message}[/bold green]")
        if file_size:
            console.print(f"   [dim]File size: {file_size:,} bytes[/dim]")
        console.print("   [dim]File will be picked up by Liquibase on next run[/dim]")
    else:
        console.print(f"❌ [bold red]{message}[/bold red]")


@main.command()
@click.argument("text")
@click.option("--copy/--no-copy", default=False, help="Copy result to clipboard")
@click.option(
    "--type", "-t", is_flag=True, help="Type result directly using wtype/ydotool"
)
def dash(text, copy, type):
    """Convert text to dash-separated format for filenames."""
    import os
    import shutil

    dash_result = helper.convert_to_dash_format(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get("XDG_SESSION_TYPE") == "wayland":
            if not shutil.which("wtype") and not shutil.which("ydotool"):
                console.print(
                    "[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'.[/yellow]"
                )
                copy_or_print(
                    dash_result,
                    True,
                    console,
                    f"✅ Copied to clipboard instead: [bold green]{dash_result}[/bold green]",
                )
                console.print(
                    f"✅ Copied to clipboard instead: [bold green]{dash_result}[/bold green]"
                )
                return

        helper.type_string_with_delay(dash_result, delay=0)
        console.print(f"✅ Typed: [bold green]{dash_result}[/bold green]")
    elif copy:
        copy_or_print(
            dash_result,
            True,
            console,
            f"✅ Copied to clipboard: [bold green]{dash_result}[/bold green]",
        )
        console.print(f"✅ Copied to clipboard: [bold green]{dash_result}[/bold green]")
    else:
        console.print(f"{dash_result}")


@main.command()
@click.argument("text")
@click.option("--copy/--no-copy", default=True, help="Copy result to clipboard")
def pr(text, copy):
    """Generate PR title with uppercased ticket numbers."""
    pr_title = helper.generate_pr_title(text)

    if copy:
        copy_or_print(
            pr_title,
            True,
            console,
            f"✅ Copied to clipboard: [bold green]{pr_title}[/bold green]",
        )
        console.print(f"✅ Copied to clipboard: [bold green]{pr_title}[/bold green]")
    else:
        console.print(f"PR Title: [bold blue]{pr_title}[/bold blue]")


@main.command()
@click.argument("text")
@click.option("--copy/--no-copy", default=True, help="Copy result to clipboard")
def filename(text, copy):
    """Generate clean filename from text (lowercase, with .md extension)."""
    filename_result = helper.generate_filename(text)

    if copy:
        copy_or_print(
            filename_result,
            True,
            console,
            f"✅ Copied to clipboard: [bold green]{filename_result}[/bold green]",
        )
        console.print(
            f"✅ Copied to clipboard: [bold green]{filename_result}[/bold green]"
        )
    else:
        console.print(f"Filename: [bold blue]{filename_result}[/bold blue]")


@main.command()
@click.argument("ticket_text", required=False)
@click.argument("agent_name", required=False)
@click.option(
    "--output", "-o", help="Agent output (if not provided, will prompt for input)"
)
@click.option("--auto", "-a", is_flag=True, help="Auto-detect ticket from context")
@click.option(
    "--append/--no-append",
    default=True,
    help="Append to existing log file instead of creating new one (default: append)",
)
def log(ticket_text, agent_name, output, auto, append):
    """Create work log for VIS tickets with agent output."""

    if auto or not ticket_text:
        detected_ticket, ticket_number = helper.detect_vis_ticket_from_context()
        if detected_ticket:
            if not ticket_text:
                ticket_text = detected_ticket
                console.print(
                    f"🔍 Auto-detected ticket: [bold green]{ticket_text}[/bold green]"
                )
            else:
                console.print(
                    f"🔍 Context shows: [dim]{detected_ticket}[/dim] (using provided: [bold blue]{ticket_text}[/bold blue])"
                )
        elif not ticket_text:
            console.print(
                "❌ No VIS ticket detected in context. Please provide ticket text."
            )
            return

    if not agent_name:
        agent_name = console.input("🤖 Enter agent name: ").strip()
        if not agent_name:
            console.print("❌ Agent name is required.")
            return

    if not output:
        console.print(f"📝 Creating work log for: [bold blue]{ticket_text}[/bold blue]")
        console.print(f"🤖 Agent: [bold cyan]{agent_name}[/bold cyan]")
        console.print("")
        console.print(
            "Please paste the complete agent output below (press Ctrl+D when done):"
        )
        console.print("─" * 60)

        import sys

        agent_output = sys.stdin.read().strip()

        console.print("")
        console.print("─" * 60)
        console.print("🔄 Processing...")
    else:
        agent_output = output

    success, result = helper.create_work_log(
        ticket_text, agent_name, agent_output, append=append
    )

    if success:
        console.print(f"✅ Work log created: [bold green]{result}[/bold green]")
    else:
        console.print(f"❌ Error creating work log: [bold red]{result}[/bold red]")


@main.command("check-domain")
@click.argument("domain")
@click.option("--copy", "-c", is_flag=True, help="Copy result to clipboard")
@click.option(
    "--verbose", "-v", is_flag=True, help="Show detailed output from each check"
)
def check_domain(domain, copy, verbose):
    """Check availability of a specific domain (e.g., cargolink.pl)"""
    available = checker.check_single_domain(domain, verbose=verbose)

    if copy:
        if available is True:
            safe_copy_to_clipboard(f"{domain} - available", console)
            console.print("\n📋 Result copied to clipboard")
        elif available is False:
            safe_copy_to_clipboard(f"{domain} - taken", console)
            console.print("\n📋 Result copied to clipboard")


@main.command("check-domains")
@click.argument("name")
def check_domains(name):
    """Check name across all popular TLDs (.pl, .com, .app, .io)"""
    table = checker.check_all_tlds(name)
    console.print(table)


@main.command("check-github")
@click.argument("username")
def check_github(username):
    """Check GitHub username availability"""
    status = checker.check_github(username)

    if status is True:
        console.print(f"✅ GitHub: [bold green]{username}[/bold green] is available")
        console.print(f"   URL would be: github.com/{username}")
    elif status is False:
        console.print(f"❌ GitHub: [bold red]{username}[/bold red] is taken")
    else:
        console.print(f"⚠️ Could not check GitHub availability")


@main.command("check-npm")
@click.argument("package_name")
def check_npm(package_name):
    """Check npm package name availability"""
    status = checker.check_npm(package_name)

    if status is True:
        console.print(f"✅ npm: [bold green]{package_name}[/bold green] is available")
        console.print(f"   Package would be: npmjs.com/package/{package_name}")
    elif status is False:
        console.print(f"❌ npm: [bold red]{package_name}[/bold red] is taken")
    else:
        console.print(f"⚠️ Could not check npm availability")


@main.command("check-trademark")
@click.argument("name")
@click.option("--open", "-o", is_flag=True, help="Open URLs in browser")
def check_trademark(name, open):
    """Search trademark databases for a name"""
    uprp_url, euipo_url = checker.generate_trademark_urls(name)

    console.print(f"🔍 Check '[bold cyan]{name}[/bold cyan]' in trademark databases:\n")
    console.print(f"  🇵🇱 UP RP: {uprp_url}")
    console.print(f"  🇪🇺 EUIPO: {euipo_url}")

    if open:
        import webbrowser

        console.print("\n🌐 Opening in browser...")
        webbrowser.open(uprp_url)
        webbrowser.open(euipo_url)


@main.group(name="jira")
def jira_group():
    """Jira workflow automation commands for time tracking and standup notes."""
    pass


@jira_group.command("start")
@click.option("--ticket", "-t", help="Ticket to work on (auto-detects if not provided)")
def jira_start(ticket):
    """Start working on a ticket (auto-detects from git branch)."""
    from rich.panel import Panel
    from rich.table import Table

    context_detector = get_context_detector()

    if not ticket:
        console.print("🔍 Auto-detecting ticket from git branch...")
        ticket = context_detector.detect_ticket_from_git()

        if not ticket:
            console.print("[yellow]⚠ Could not detect ticket from git branch.[/yellow]")
            ticket = (
                console.input("Enter ticket number (e.g., VIS-4703): ").strip().upper()
            )

    # Get full context
    context = context_detector.get_full_context()

    # Display detected context
    table = Table(title=f"Starting work on {ticket}", show_header=False, box=None)
    table.add_column("", style="dim")
    table.add_column("")

    if context["project"].get("repository"):
        table.add_row("📁 Repository:", context["project"]["repository"])
    if context["project"].get("branch"):
        table.add_row("🌿 Branch:", context["project"]["branch"])
    if context["suggestion"].get("start_time"):
        table.add_row(
            "⏰ Session start:", context["suggestion"]["start_time"].strftime("%H:%M")
        )

    console.print(table)
    console.print(
        Panel.fit(f"🚀 Starting work on [bold cyan]{ticket}[/bold cyan]", style="green")
    )


@jira_group.command("log")
@click.argument("duration", required=False)
@click.argument("description", required=False)
@click.option(
    "--ticket", "-t", help="Ticket to log time for (auto-detects if not provided)"
)
def jira_log(duration, description, ticket):
    """Log work time to Jira and Google Sheets."""
    from rich.progress import Progress, SpinnerColumn, TextColumn

    context_detector = get_context_detector()
    jira_service = get_jira_service()
    sheets_service = get_sheets_service()

    # Auto-detect or get suggestions
    suggestion = context_detector.suggest_time_entry()

    if not ticket:
        ticket = suggestion["ticket"]
        if ticket:
            console.print(f"🎯 Auto-detected ticket: [bold cyan]{ticket}[/bold cyan]")
        else:
            ticket = console.input("Enter ticket number: ").strip().upper()

    if not duration:
        suggested_duration = suggestion["duration"]
        duration = (
            console.input(f"⏰ Duration [{suggested_duration}]: ").strip()
            or suggested_duration
        )

    if not description:
        suggested_desc = suggestion["description"]
        if suggested_desc:
            console.print(f"📝 Suggested: [dim]{suggested_desc}[/dim]")
        description = (
            console.input("📝 What did you work on? ").strip() or suggested_desc
        )

    # Log with progress indicator
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        transient=True,
    ) as progress:
        task = progress.add_task("Logging work...", total=3)

        # Save to local cache first
        progress.update(task, description="💾 Saving to local cache...")
        success = jira_service.log_work(ticket, duration, description)
        progress.update(task, advance=1)

        # Sync to Jira if online
        progress.update(task, description="🔄 Syncing to Jira...")
        if jira_service._is_online():
            jira_service._sync_worklogs()
            progress.update(task, advance=1)
        else:
            console.print(
                "[yellow]📵 Offline - will sync when connection restored[/yellow]"
            )
            progress.update(task, advance=1)

        # Update Google Sheets
        progress.update(task, description="📊 Updating Google Sheets...")
        if sheets_service and sheets_service.service:
            sheets_service.append_work_log(ticket, duration, description)
        progress.update(task, advance=1)

    console.print(
        f"✅ Logged [bold green]{duration}[/bold green] to [bold cyan]{ticket}[/bold cyan]"
    )
    console.print(f"   {description}")


@jira_group.command("standup")
@click.option("--copy", "-c", is_flag=True, help="Copy to clipboard", default=True)
def jira_standup(copy):
    """Generate standup notes from yesterday's work and today's plan."""
    jira_service = get_jira_service()
    sheets_service = get_sheets_service()

    # Get yesterday's work from cache
    yesterday = datetime.now() - timedelta(days=1)
    recent_logs = jira_service.get_recent_worklogs(days=1)

    # Filter for yesterday's logs
    yesterday_logs = [
        log
        for log in recent_logs
        if log.get("started_at")
        and yesterday.date() == datetime.fromisoformat(log["started_at"]).date()
    ]

    # Get today's planned work (from tickets in progress)
    today_plan = []
    try:
        tickets = jira_service.fetch_assigned_tickets()
        for ticket in tickets[:3]:  # Top 3 tickets
            status = ticket.get("fields", {}).get("status", {}).get("name", "")
            if status in ["In Progress", "To Do"]:
                summary = ticket.get("fields", {}).get("summary", "")
                today_plan.append(f"{ticket['key']}: {summary}")
    except:
        today_plan = ["Continue current tasks"]

    # Generate standup notes
    if sheets_service:
        standup_text = sheets_service.generate_standup_notes(yesterday_logs, today_plan)
    else:
        # Fallback standup generation without sheets service
        standup_text = f"""📅 Standup - {datetime.now().strftime("%B %d, %Y")}

Yesterday:
"""
        if yesterday_logs:
            for log in yesterday_logs:
                standup_text += (
                    f"• {log['ticket_key']}: {log['description']} ({log['duration']})\n"
                )
        else:
            standup_text += "• No logged work\n"

        standup_text += "\nToday:\n"
        if today_plan:
            for item in today_plan:
                standup_text += f"• {item}\n"
        else:
            standup_text += "• Continue current tasks\n"

        standup_text += "\nBlockers:\n• None"

    console.print(Panel(standup_text, title="Standup Notes", style="blue"))

    # Try to sync to Google Sheets
    if sheets_service and sheets_service.service:
        if sheets_service.write_standup_notes(standup_text):
            console.print("📊 Synced to Google Sheets")

    if copy:
        safe_copy_to_clipboard(standup_text, console)
        console.print("✅ Copied to clipboard!")


@jira_group.command("status")
def jira_status():
    """Show today's work status and time logged."""
    from rich.table import Table

    jira_service = get_jira_service()

    # Get today's work logs
    today_logs = []
    recent_logs = jira_service.get_recent_worklogs(days=1)
    today = datetime.now().date()

    total_seconds = 0

    table = Table(title=f"Work Status - {datetime.now().strftime('%B %d, %Y')}")
    table.add_column("Ticket", style="cyan")
    table.add_column("Status", justify="center")
    table.add_column("Time", justify="right")
    table.add_column("Description")

    for log in recent_logs:
        if log.get("started_at"):
            log_date = datetime.fromisoformat(log["started_at"]).date()
            if log_date == today:
                status = "✅ Logged" if log.get("synced") else "💾 Cached"
                table.add_row(
                    log["ticket_key"],
                    status,
                    log["duration"],
                    log["description"][:40] + "..."
                    if len(log["description"]) > 40
                    else log["description"],
                )
                # Calculate total time
                duration_sec = jira_service._parse_duration_to_seconds(log["duration"])
                total_seconds += duration_sec

    # Add total row
    if total_seconds > 0:
        hours = total_seconds / 3600
        total_str = f"{hours:.1f}h" if hours >= 1 else f"{int(total_seconds / 60)}m"
        table.add_row("", "", f"[bold]{total_str}[/bold]", "[bold]Total[/bold]")
    else:
        table.add_row("", "", "[dim]0h[/dim]", "[dim]No work logged today[/dim]")

    console.print(table)

    # Show sync status
    if not jira_service._is_online():
        console.print(
            "[yellow]📵 Offline mode - logs will sync when connection restored[/yellow]"
        )


@jira_group.command("config")
@click.option(
    "--interactive", "-i", is_flag=True, help="Interactive configuration setup"
)
def jira_config(interactive):
    """Configure Jira and Google Sheets credentials."""
    credential_manager = get_credential_manager()

    if not interactive:
        # Show current configuration status
        console.print(Panel.fit("⚙️ Jira Configuration Status", style="bold blue"))

        jira_creds = credential_manager.get_jira_credentials()
        sheets_config = credential_manager.get_google_sheets_config()

        from rich.table import Table

        table = Table(show_header=False, box=None)
        table.add_column("", style="bold")
        table.add_column("")

        if jira_creds:
            table.add_row("Jira URL:", jira_creds["base_url"])
            table.add_row("Jira Email:", jira_creds["email"])
            table.add_row("Jira Token:", "✅ Configured")
        else:
            table.add_row("Jira:", "[red]Not configured[/red]")

        if sheets_config:
            table.add_row("Google Sheets ID:", sheets_config.get("sheet_id", "Not set"))
        else:
            table.add_row("Google Sheets:", "[red]Not configured[/red]")

        console.print(table)
        console.print("\nRun with --interactive for guided setup")
        return

    # Interactive configuration
    console.print(Panel.fit("⚙️ Jira Configuration Setup", style="bold blue"))

    # Jira configuration
    console.print("\n[bold]Jira Configuration[/bold]")
    console.print(
        "Get your API token from: https://id.atlassian.com/manage-profile/security/api-tokens"
    )

    base_url = console.input("Jira URL (e.g., https://company.atlassian.net): ").strip()
    email = console.input("Jira email: ").strip()
    api_token = console.input("Jira API token: ").strip()

    if base_url and email and api_token:
        if credential_manager.save_jira_credentials(base_url, email, api_token):
            console.print("[green]✅ Jira credentials saved securely[/green]")
        else:
            console.print("[red]❌ Failed to save Jira credentials[/red]")

    # Google Sheets configuration
    console.print("\n[bold]Google Sheets Configuration (optional)[/bold]")
    sheet_id = console.input("Google Sheets ID (or press Enter to skip): ").strip()

    if sheet_id:
        console.print("For Google Sheets access, you need a service account JSON file.")
        creds_path = console.input(
            "Path to service account JSON (or Enter to skip): "
        ).strip()

        creds_json = None
        if creds_path and Path(creds_path).exists():
            creds_json = Path(creds_path).read_text()

        if credential_manager.save_google_sheets_config(sheet_id, creds_json):
            console.print("[green]✅ Google Sheets configuration saved[/green]")
        else:
            console.print("[yellow]⚠ Google Sheets partially configured[/yellow]")

    console.print("\n[green]✅ Configuration complete![/green]")
    console.print("You can now use 'helper jira log' and other commands.")


@main.command("name-finder")
@click.argument("names", nargs=-1, required=True)
@click.option(
    "--output",
    "-o",
    default="naming-report",
    help="Base filename for exports (without extension)",
)
@click.option(
    "--format",
    "-f",
    multiple=True,
    default=["table"],
    type=click.Choice(["table", "csv", "markdown", "json", "all"]),
    help="Export formats (can specify multiple)",
)
@click.option(
    "--no-social", is_flag=True, help="Skip social media checks for faster results"
)
@click.option("--no-cache", is_flag=True, help="Disable caching for fresh results")
@click.option(
    "--workers", "-w", default=10, type=int, help="Number of parallel workers"
)
@click.option("--verbose", "-v", is_flag=True, help="Show detailed debug information")
def name_finder(names, output, format, no_social, no_cache, workers, verbose):
    """Find and verify the best name for your project with comprehensive checking"""
    console.print(
        f"🔍 [bold blue]Name Finder[/bold blue] - Analyzing {len(names)} name{'s' if len(names) > 1 else ''}..."
    )
    console.print(
        f"[dim]Checking domains, platforms{', and social media' if not no_social else ''}...[/dim]\n"
    )

    # Use the new refactored checker
    checker = NameAvailabilityChecker(
        cache_enabled=not no_cache, parallel_workers=workers, verbose=verbose
    )

    try:
        # Determine what to check
        social_platforms = (
            None if no_social else ["instagram", "twitter", "linkedin", "reddit"]
        )

        # Check names
        results = checker.check_names(
            list(names),
            domains=[".com", ".org", ".net", ".io", ".dev", ".eu"],
            platforms=["npm", "github", "gitlab", "pypi", "dockerhub"],
            social=social_platforms,
            show_progress=True,
        )

        # Display table if requested
        if "table" in format or not format:
            console.print()
            checker.display_table(results)

        # Calculate best option
        if results:
            best_score = 0
            best_name = None

            for name, services in results.items():
                available = sum(1 for r in services.values() if r.is_available)
                total = len(services)
                score = (available / total * 100) if total > 0 else 0

                if score > best_score:
                    best_score = score
                    best_name = name

            if best_name:
                console.print("\n" + "=" * 50)
                if best_score >= 80:
                    console.print(
                        f"🏆 [bold green]Best option: {best_name} ({best_score:.0f}% availability)[/bold green]"
                    )
                elif best_score >= 50:
                    console.print(
                        f"🏆 [bold yellow]Best option: {best_name} ({best_score:.0f}% availability)[/bold yellow]"
                    )
                else:
                    console.print(
                        f"⚠️ [bold red]Best option: {best_name} ({best_score:.0f}% availability)[/bold red]"
                    )

        # Export in requested formats
        export_formats = list(format)
        if "all" in export_formats:
            export_formats = ["csv", "markdown", "json"]

        if any(f in export_formats for f in ["csv", "markdown", "json"]):
            console.print(f"\n[bold]Exporting results...[/bold]")

            for fmt in export_formats:
                if fmt == "csv":
                    filepath = Path(f"{output}.csv")
                    checker.export_csv(results, filepath)
                    console.print(
                        f"📄 CSV saved to: [bold green]{filepath}[/bold green]"
                    )
                elif fmt == "markdown":
                    filepath = Path(f"{output}.md")
                    checker.export_markdown(results, filepath)
                    console.print(
                        f"📄 Markdown saved to: [bold green]{filepath}[/bold green]"
                    )
                elif fmt == "json":
                    filepath = Path(f"{output}.json")
                    checker.export_json(results, filepath)
                    console.print(
                        f"📄 JSON saved to: [bold green]{filepath}[/bold green]"
                    )

    finally:
        checker.close()


@main.command()
def interactive():
    """Interactive mode for helper commands."""
    console.print(Panel.fit("🔧 Helper CLI - Interactive Mode", style="bold blue"))

    table = Table(title="Available Commands")
    table.add_column("Command", justify="left", style="cyan")
    table.add_column("Description", justify="left")

    commands = [
        ("branch <text>", "Generate JIRA branch name"),
        ("rfid <id>", "Generate RFID badge (@id#)"),
        ("qr <id>", "Generate QR badge ($id#)"),
        ("badge <string>", "Type or copy badge string as-is"),
        ("stash <text>", "Generate git stash command"),
        ("dash <text>", "Convert text to dash-separated format"),
        (
            "log [ticket] [agent]",
            "Create work log for VIS tickets (auto-detect with -a)",
        ),
        ("ngrok <url>", "Generate QR code for NGROK URL"),
        ("vsc <url>", "Generate QR code for VSC URL"),
        ("rt <name>", "Setup RT directory"),
        ("apex <file> <port> <auth>", "Generate APEX upload command"),
        ("custom_default_data", "Sync custom_default_data.sql to TDS Suite"),
        ("check-domain <domain>", "Check domain availability"),
        ("check-domains <name>", "Check all TLDs for a name"),
        ("check-github <username>", "Check GitHub username"),
        ("check-npm <package>", "Check npm package name"),
        ("check-trademark <name>", "Search trademark databases"),
        ("name-finder <names...>", "Comprehensive name verification"),
    ]

    for cmd, desc in commands:
        table.add_row(cmd, desc)

    console.print(table)
    console.print("\n[dim]Type 'exit' to quit interactive mode[/dim]")

    while True:
        try:
            user_input = console.input("\n[bold blue]helper>[/bold blue] ").strip()
            if user_input.lower() in ["exit", "quit", "q"]:
                console.print("👋 Goodbye!")
                break
            elif user_input:
                parts = user_input.split()
                if parts:
                    try:
                        main(parts, standalone_mode=False)
                    except SystemExit:
                        pass
                    except Exception as e:
                        console.print(f"[bold red]Error:[/bold red] {e}")
        except KeyboardInterrupt:
            console.print("\n👋 Goodbye!")
            break
        except EOFError:
            console.print("\n👋 Goodbye!")
            break


if __name__ == "__main__":
    main()
