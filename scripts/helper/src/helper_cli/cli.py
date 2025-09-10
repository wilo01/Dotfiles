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
                base_url=creds['base_url'],
                email=creds['email'],
                api_token=creds['api_token']
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
                sheet_id=config.get('sheet_id'),
                credentials_path=config.get('credentials_file')
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
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
@click.option('--type', '-t', is_flag=True, help='Type command directly using wtype/ydotool')
def branch(text, copy, type):
    """Generate JIRA branch name from commit message text."""
    import os
    import shutil
    branch_command = helper.generate_jira_branch_name(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(branch_command, True, console, f"✅ Copied to clipboard instead: [bold green]{branch_command}[/bold green]")
                return

        helper.type_string_with_delay(branch_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{branch_command}[/bold green]")
    elif copy:
        copy_or_print(branch_command, True, console, f"✅ Copied to clipboard: [bold green]{branch_command}[/bold green]")
    else:
        console.print(f"Branch command: [bold blue]{branch_command}[/bold blue]")


@main.command()
@click.argument('badge_string')
@click.option('--copy', '-c', is_flag=True, help='Copy to clipboard instead of auto-typing')
def badge(badge_string, copy):
    """Type or copy badge string exactly as provided (e.g., '@1234#' or '$5678#')."""
    import os
    # Use the badge string exactly as provided by the user
    console.print(f"[bold blue]Badge string:[/bold blue] {badge_string}")

    # Check if we're on Wayland
    if not copy and os.environ.get('XDG_SESSION_TYPE') == 'wayland':
        console.print("[yellow]⚠ Auto-type not supported on Wayland. Copying to clipboard instead.[/yellow]")
        copy_or_print(badge_string, True, console, "✅ Badge copied to clipboard")
    elif copy:
        copy_or_print(badge_string, True, console, "✅ Badge copied to clipboard")
    else:
        helper.type_string_with_delay(badge_string)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument('badge_id')
@click.option('--copy', '-c', is_flag=True, help='Copy to clipboard instead of auto-typing')
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
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(badge_id, True, console, "✅ Badge copied to clipboard instead")
                return

        helper.type_string_with_delay(badge_id)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument('badge_id')
@click.option('--copy', '-c', is_flag=True, help='Copy to clipboard instead of auto-typing')
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
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(badge_id, True, console, "✅ Badge copied to clipboard instead")
                return

        helper.type_string_with_delay(badge_id)
        console.print("✅ Badge typed automatically")


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
@click.option('--type', '-t', is_flag=True, help='Type command directly using wtype/ydotool')
def stash(text, copy, type):
    """Generate git stash command with formatted message."""
    import os
    import shutil
    stash_command = helper.generate_stash_command(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                copy_or_print(stash_command, True, console, f"✅ Copied to clipboard instead: [bold green]{stash_command}[/bold green]")
                console.print(f"✅ Copied to clipboard instead: [bold green]{stash_command}[/bold green]")
                return

        helper.type_string_with_delay(stash_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{stash_command}[/bold green]")
    elif copy:
        copy_or_print(stash_command, True, console, f"✅ Copied to clipboard: [bold green]{stash_command}[/bold green]")
        console.print(f"✅ Copied to clipboard: [bold green]{stash_command}[/bold green]")
    else:
        console.print(f"Stash command: [bold blue]{stash_command}[/bold blue]")


@main.command()
@click.argument('url')
@click.option('--copy/--no-copy', default=True, help='Copy QR command to clipboard')
def ngrok(url, copy):
    """Generate QR code command for NGROK URL."""
    full_url, qr_command = helper.generate_ngrok_qr(url)

    console.print(f"[bold blue]Full URL:[/bold blue] {full_url}")
    console.print(f"[bold blue]QR Command:[/bold blue] {qr_command}")

    if copy:
        copy_or_print(qr_command, True, console, "✅ QR command copied to clipboard")
        console.print("✅ QR command copied to clipboard")


@main.command()
@click.argument('url')
@click.option('--copy/--no-copy', default=True, help='Copy QR command to clipboard')
def vsc(url, copy):
    """Generate QR code command for VSC URL."""
    full_url, qr_command = helper.generate_vsc_qr(url)

    console.print(f"[bold blue]Full URL:[/bold blue] {full_url}")
    console.print(f"[bold blue]QR Command:[/bold blue] {qr_command}")

    if copy:
        copy_or_print(qr_command, True, console, "✅ QR command copied to clipboard")
        console.print("✅ QR command copied to clipboard")


@main.command()
@click.argument('rt_name')
def rt(rt_name):
    """Set up RT temporary directory with file copying."""
    helper.setup_rt_temp_directory(rt_name)
    console.print(f"✅ RT directory setup complete: [bold green]{rt_name}[/bold green]")


@main.command()
@click.argument('apex_file')
@click.argument('port')
@click.argument('auth')
@click.option('--copy/--no-copy', default=True, help='Copy command to clipboard')
def apex(apex_file, port, auth, copy):
    """Generate APEX upload curl command."""
    upload_command = helper.generate_apex_upload_command(apex_file, port, auth)

    console.print(f"[bold blue]APEX Upload Command:[/bold blue]")
    console.print(Panel(upload_command, expand=False))

    if copy:
        copy_or_print(upload_command, True, console, "✅ Upload command copied to clipboard")
        console.print("✅ Upload command copied to clipboard")


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=False, help='Copy result to clipboard')
@click.option('--type', '-t', is_flag=True, help='Type result directly using wtype/ydotool')
def dash(text, copy, type):
    """Convert text to dash-separated format for filenames."""
    import os
    import shutil
    dash_result = helper.convert_to_dash_format(text)

    if type:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'.[/yellow]")
                copy_or_print(dash_result, True, console, f"✅ Copied to clipboard instead: [bold green]{dash_result}[/bold green]")
                console.print(f"✅ Copied to clipboard instead: [bold green]{dash_result}[/bold green]")
                return

        helper.type_string_with_delay(dash_result, delay=0)
        console.print(f"✅ Typed: [bold green]{dash_result}[/bold green]")
    elif copy:
        copy_or_print(dash_result, True, console, f"✅ Copied to clipboard: [bold green]{dash_result}[/bold green]")
        console.print(f"✅ Copied to clipboard: [bold green]{dash_result}[/bold green]")
    else:
        console.print(f"{dash_result}")


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
def pr(text, copy):
    """Generate PR title with uppercased ticket numbers."""
    pr_title = helper.generate_pr_title(text)

    if copy:
        copy_or_print(pr_title, True, console, f"✅ Copied to clipboard: [bold green]{pr_title}[/bold green]")
        console.print(f"✅ Copied to clipboard: [bold green]{pr_title}[/bold green]")
    else:
        console.print(f"PR Title: [bold blue]{pr_title}[/bold blue]")


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
def filename(text, copy):
    """Generate clean filename from text (lowercase, with .md extension)."""
    filename_result = helper.generate_filename(text)

    if copy:
        copy_or_print(filename_result, True, console, f"✅ Copied to clipboard: [bold green]{filename_result}[/bold green]")
        console.print(f"✅ Copied to clipboard: [bold green]{filename_result}[/bold green]")
    else:
        console.print(f"Filename: [bold blue]{filename_result}[/bold blue]")


@main.command()
@click.argument('ticket_text', required=False)
@click.argument('agent_name', required=False)
@click.option('--output', '-o', help='Agent output (if not provided, will prompt for input)')
@click.option('--auto', '-a', is_flag=True, help='Auto-detect ticket from context')
@click.option('--append/--no-append', default=True, help='Append to existing log file instead of creating new one (default: append)')
def log(ticket_text, agent_name, output, auto, append):
    """Create work log for VIS tickets with agent output."""

    if auto or not ticket_text:
        detected_ticket, ticket_number = helper.detect_vis_ticket_from_context()
        if detected_ticket:
            if not ticket_text:
                ticket_text = detected_ticket
                console.print(f"🔍 Auto-detected ticket: [bold green]{ticket_text}[/bold green]")
            else:
                console.print(f"🔍 Context shows: [dim]{detected_ticket}[/dim] (using provided: [bold blue]{ticket_text}[/bold blue])")
        elif not ticket_text:
            console.print("❌ No VIS ticket detected in context. Please provide ticket text.")
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
        console.print("Please paste the complete agent output below (press Ctrl+D when done):")
        console.print("─" * 60)

        import sys
        agent_output = sys.stdin.read().strip()

        console.print("")
        console.print("─" * 60)
        console.print("🔄 Processing...")
    else:
        agent_output = output

    success, result = helper.create_work_log(ticket_text, agent_name, agent_output, append=append)

    if success:
        console.print(f"✅ Work log created: [bold green]{result}[/bold green]")
    else:
        console.print(f"❌ Error creating work log: [bold red]{result}[/bold red]")


@main.command('check-domain')
@click.argument('domain')
@click.option('--copy', '-c', is_flag=True, help='Copy result to clipboard')
@click.option('--verbose', '-v', is_flag=True, help='Show detailed output from each check')
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


@main.command('check-domains')
@click.argument('name')
def check_domains(name):
    """Check name across all popular TLDs (.pl, .com, .app, .io)"""
    table = checker.check_all_tlds(name)
    console.print(table)


@main.command('check-github')
@click.argument('username')
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


@main.command('check-npm')
@click.argument('package_name')
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


@main.command('check-trademark')
@click.argument('name')
@click.option('--open', '-o', is_flag=True, help='Open URLs in browser')
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


@main.group(name='jira')
def jira_group():
    """Jira workflow automation commands for time tracking and standup notes."""
    pass


@jira_group.command('start')
@click.option('--ticket', '-t', help='Ticket to work on (auto-detects if not provided)')
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
            ticket = console.input("Enter ticket number (e.g., VIS-4703): ").strip().upper()

    # Get full context
    context = context_detector.get_full_context()

    # Display detected context
    table = Table(title=f"Starting work on {ticket}", show_header=False, box=None)
    table.add_column("", style="dim")
    table.add_column("")

    if context['project'].get('repository'):
        table.add_row("📁 Repository:", context['project']['repository'])
    if context['project'].get('branch'):
        table.add_row("🌿 Branch:", context['project']['branch'])
    if context['suggestion'].get('start_time'):
        table.add_row("⏰ Session start:", context['suggestion']['start_time'].strftime("%H:%M"))

    console.print(table)
    console.print(Panel.fit(f"🚀 Starting work on [bold cyan]{ticket}[/bold cyan]", style="green"))


@jira_group.command('log')
@click.argument('duration', required=False)
@click.argument('description', required=False)
@click.option('--ticket', '-t', help='Ticket to log time for (auto-detects if not provided)')
def jira_log(duration, description, ticket):
    """Log work time to Jira and Google Sheets."""
    from rich.progress import Progress, SpinnerColumn, TextColumn

    context_detector = get_context_detector()
    jira_service = get_jira_service()
    sheets_service = get_sheets_service()

    # Auto-detect or get suggestions
    suggestion = context_detector.suggest_time_entry()

    if not ticket:
        ticket = suggestion['ticket']
        if ticket:
            console.print(f"🎯 Auto-detected ticket: [bold cyan]{ticket}[/bold cyan]")
        else:
            ticket = console.input("Enter ticket number: ").strip().upper()

    if not duration:
        suggested_duration = suggestion['duration']
        duration = console.input(f"⏰ Duration [{suggested_duration}]: ").strip() or suggested_duration

    if not description:
        suggested_desc = suggestion['description']
        if suggested_desc:
            console.print(f"📝 Suggested: [dim]{suggested_desc}[/dim]")
        description = console.input("📝 What did you work on? ").strip() or suggested_desc

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
            console.print("[yellow]📵 Offline - will sync when connection restored[/yellow]")
            progress.update(task, advance=1)

        # Update Google Sheets
        progress.update(task, description="📊 Updating Google Sheets...")
        if sheets_service and sheets_service.service:
            sheets_service.append_work_log(ticket, duration, description)
        progress.update(task, advance=1)

    console.print(f"✅ Logged [bold green]{duration}[/bold green] to [bold cyan]{ticket}[/bold cyan]")
    console.print(f"   {description}")


@jira_group.command('standup')
@click.option('--tui/--no-tui', default=True, help='Use TUI spreadsheet interface')
def jira_standup(tui):
    """Generate interactive standup notes with your current tickets."""
    jira_service = get_jira_service()

    if tui:
        # Use the new TUI interface
        try:
            from helper_cli.tui.standup_app import run_standup_tui

            # Get tickets
            console.print("📋 Loading tickets...", style="dim")
            tickets = jira_service.get_relevant_tickets_for_standup()

            if not tickets:
                console.print("❌ No tickets found. Try 'helper jira fetch' first.", style="red")
                return

            # Launch TUI
            run_standup_tui(jira_service=jira_service, tickets=tickets)
            return
        except ImportError:
            console.print("⚠️ TUI not available. Install with: pip install textual", style="yellow")
            console.print("Falling back to interactive mode...\n")

    # Get relevant tickets (in progress, recently updated)
    console.print("📋 Fetching your tickets...", style="dim")
    tickets = jira_service.get_relevant_tickets_for_standup()

    if not tickets:
        console.print("❌ No tickets found. Try 'helper jira fetch' first.", style="red")
        return

    # Organize tickets (remove duplicates, prefer parent tickets)
    seen_keys = set()
    unique_tickets = []
    for ticket in tickets:
        key = ticket['key']
        if key not in seen_keys:
            seen_keys.add(key)
            unique_tickets.append(ticket)

    # Interactive prompt for descriptions
    standup_notes = []
    console.print("\n🎯 What did you work on for each ticket? (Press Enter to skip)\n")

    for ticket in unique_tickets[:8]:  # Limit to 8 tickets for standup
        key = ticket['key']
        summary = ticket.get('fields', {}).get('summary', 'No summary')

        # Truncate long summaries
        if len(summary) > 60:
            summary = summary[:57] + "..."

        # Get previous description as suggestion
        previous_desc = jira_service.get_recent_standup_description(key)
        prompt = f"[cyan]{key}[/cyan]: {summary}"

        if previous_desc:
            prompt += f"\n  [dim](Previous: {previous_desc[:40]}...)[/dim]"

        console.print(prompt)
        description = input("What did you work on? > ").strip()

        if description:
            standup_notes.append({
                'ticket_key': key,
                'ticket_summary': summary,
                'description': description
            })

        print()  # Empty line for readability

    if not standup_notes:
        console.print("❌ No work descriptions provided.", style="red")
        return

    # Save notes to database
    jira_service.save_standup_notes(standup_notes)

    # Generate both formats
    today_str = datetime.now().strftime('%B %d, %Y')

    # Terminal format
    terminal_output = f"📅 Standup Notes - {today_str}\n\n"
    terminal_output += "✅ Work Summary:\n"
    for note in standup_notes:
        terminal_output += f"• {note['ticket_key']}: {note['description']}\n"

    # Sheets format (tab-separated)
    sheets_output = ""
    for note in standup_notes:
        sheets_output += f"{note['ticket_key']}\t{note['ticket_summary']}\t{note['description']}\n"

    # Display both formats
    console.print("\n" + "="*60)
    console.print(Panel(terminal_output, title="Standup Summary", style="green"))

    console.print("\n📋 For Google Sheets (copied to clipboard):", style="blue")
    console.print(Panel(sheets_output.strip(), style="dim"))

    # Copy sheets format to clipboard
    safe_copy_to_clipboard(sheets_output, console)
    console.print("\n✅ Sheets format copied to clipboard!", style="green bold")


@jira_group.command('status')
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
        if log.get('started_at'):
            log_date = datetime.fromisoformat(log['started_at']).date()
            if log_date == today:
                status = "✅ Logged" if log.get('synced') else "💾 Cached"
                table.add_row(
                    log['ticket_key'],
                    status,
                    log['duration'],
                    log['description'][:40] + "..." if len(log['description']) > 40 else log['description']
                )
                # Calculate total time
                duration_sec = jira_service._parse_duration_to_seconds(log['duration'])
                total_seconds += duration_sec

    # Add total row
    if total_seconds > 0:
        hours = total_seconds / 3600
        total_str = f"{hours:.1f}h" if hours >= 1 else f"{int(total_seconds/60)}m"
        table.add_row("", "", f"[bold]{total_str}[/bold]", "[bold]Total[/bold]")
    else:
        table.add_row("", "", "[dim]0h[/dim]", "[dim]No work logged today[/dim]")

    console.print(table)

    # Show sync status
    if not jira_service._is_online():
        console.print("[yellow]📵 Offline mode - logs will sync when connection restored[/yellow]")


def display_smart_grouped_table(families, standalone, console, current_user=None):
    """Display tickets in a smart grouped table format."""
    from rich.table import Table
    from rich.text import Text

    # Helper functions for formatting
    def format_status(status):
        """Format status with color and icon."""
        if status == 'Done':
            return Text("✓ Done", style="green")
        elif status == 'In Progress':
            return Text("⚡ Progress", style="yellow")
        elif status in ['To Do', 'Backlog']:
            return Text(f"○ {status}", style="dim")
        else:
            return Text(status)

    def format_priority(priority):
        """Format priority with color."""
        if priority == 'Critical':
            return Text(priority, style="bold red")
        elif priority == 'High':
            return Text(priority, style="red")
        elif priority == 'Medium':
            return Text(priority, style="yellow")
        elif priority == 'Low':
            return Text(priority, style="green")
        else:
            return Text(priority, style="dim")

    def get_type_icon(issue_type):
        """Get icon for issue type."""
        icons = {
            'Epic': '🎯', 'Story': '📖', 'Bug': '🐛',
            'Task': '🔧', 'Sub-task': '📎', 'QA Test': '🧪'
        }
        return icons.get(issue_type, '')

    # Count totals
    total_count = len(standalone)
    for family in families:
        total_count += 1  # Parent
        total_count += len(family.get('children', []))

    # Create table
    table = Table(
        title=f"Jira Tickets ({total_count} tickets, {len(families)} groups)",
        show_lines=True,
        show_edge=True,
        box=None
    )

    table.add_column("Key", style="cyan", no_wrap=True, width=16)
    table.add_column("Summary", overflow="fold", min_width=40)
    table.add_column("Status", justify="center", width=12)
    table.add_column("Priority", justify="center", width=10)
    table.add_column("Type", justify="center", width=10)

    # Display families
    for i, family in enumerate(families):
        parent = family.get('parent')
        children = family.get('children', [])

        if parent:
            fields = parent['fields']

            # Calculate progress
            done_count = sum(1 for c in children if c['fields']['status']['name'] == 'Done')
            progress = f" [{done_count}/{len(children)}]" if children else ""

            # Parent row
            issue_type = fields.get('issuetype', {}).get('name', '')
            icon = get_type_icon(issue_type)
            assigned_marker = "●" if parent.get('_assigned_to_me') else " "

            table.add_row(
                Text(f"{icon} {parent['key']} {assigned_marker}", style="bold cyan"),
                Text(f"{fields.get('summary', '')[:50]}{progress}", style="bold"),
                format_status(fields.get('status', {}).get('name', '')),
                format_priority(fields.get('priority', {}).get('name', '')),
                Text(issue_type, style="bold")
            )

        # Children rows
        for j, child in enumerate(children):
            is_last = (j == len(children) - 1)
            prefix = "└─" if is_last else "├─"

            fields = child['fields']
            issue_type = fields.get('issuetype', {}).get('name', '')
            assigned_marker = "●" if child.get('_assigned_to_me') else " "

            table.add_row(
                Text(f"  {prefix} {child['key']} {assigned_marker}", style="cyan"),
                Text(f"  {fields.get('summary', '')[:48]}", style=""),
                format_status(fields.get('status', {}).get('name', '')),
                format_priority(fields.get('priority', {}).get('name', '')),
                Text(issue_type, style="dim")
            )

        # Add separator between groups (but not after last group)
        if i < len(families) - 1 or standalone:
            table.add_section()

    # Display standalone tickets
    if standalone:
        if families:
            # Add header for standalone section
            table.add_row(
                Text("─ Standalone ─", style="dim italic"),
                Text("", style="dim"),
                Text("", style="dim"),
                Text("", style="dim"),
                Text("", style="dim")
            )

        for ticket in standalone:
            fields = ticket['fields']
            issue_type = fields.get('issuetype', {}).get('name', '')
            icon = get_type_icon(issue_type)
            assigned_marker = "●" if ticket.get('_assigned_to_me') else " "

            table.add_row(
                Text(f"{icon} {ticket['key']} {assigned_marker}", style="cyan"),
                Text(fields.get('summary', '')[:50]),
                format_status(fields.get('status', {}).get('name', '')),
                format_priority(fields.get('priority', {}).get('name', '')),
                Text(issue_type)
            )

    console.print(table)

    # Legend
    console.print("\n[dim]● = Assigned to you | Icons: 🐛 Bug 📖 Story 🔧 Task 🧪 QA Test[/dim]")


def display_grouped_tickets(tickets, console):
    """Display tickets in a hierarchical grouped format."""
    from rich.table import Table
    from rich.rule import Rule
    from rich.panel import Panel

    def format_status(status_name):
        """Format status with color."""
        if status_name == 'Done':
            return f"[green]✓ {status_name}[/green]"
        elif status_name == 'In Progress':
            return f"[yellow]⚡ {status_name}[/yellow]"
        elif status_name in ['To Do', 'Backlog']:
            return f"[dim]○ {status_name}[/dim]"
        else:
            return status_name

    def format_priority(priority):
        """Format priority with color."""
        if priority == 'Critical':
            return f"[bold red]🔴 {priority}[/bold red]"
        elif priority == 'High':
            return f"[red]🟠 {priority}[/red]"
        elif priority == 'Medium':
            return f"[yellow]🟡 {priority}[/yellow]"
        elif priority == 'Low':
            return f"[green]🟢 {priority}[/green]"
        else:
            return f"⚪ {priority}"

    def get_issue_icon(issue_type):
        """Get icon for issue type."""
        icons = {
            'Epic': '🎯',
            'Story': '📖',
            'Bug': '🐛',
            'Task': '🔧',
            'Sub-task': '📎',
            'QA Test': '🧪'
        }
        return icons.get(issue_type, '📋')

    # Separate into groups
    parent_groups = []
    standalone_tasks = []
    orphaned_subtasks = []

    for ticket in tickets:
        if ticket.get('_is_parent'):
            parent_groups.append(ticket)
        elif ticket.get('_orphaned_subtask'):
            orphaned_subtasks.append(ticket)
        elif not ticket.get('_has_parent'):
            standalone_tasks.append(ticket)

    # Display parent groups with children
    for parent in parent_groups:
        issue_type = parent['fields'].get('issuetype', {}).get('name', '-')
        icon = get_issue_icon(issue_type)
        summary = parent['fields'].get('summary', 'No summary')
        status = parent['fields'].get('status', {}).get('name', 'Unknown')
        priority = parent['fields'].get('priority', {}).get('name', '-')

        # Create header for parent
        console.print(Rule(style="bright_blue"))
        header = f"{icon} [bold cyan]{parent['key']}[/bold cyan]: {summary[:60]}{'...' if len(summary) > 60 else ''}"
        console.print(header)
        console.print(f"   {format_status(status)}  {format_priority(priority)}  [dim]{issue_type}[/dim]")

        # Display children
        children = parent.get('_children', [])
        if children:
            console.print()
            for i, child in enumerate(children):
                is_last = i == len(children) - 1
                prefix = "└─" if is_last else "├─"

                child_key = child['key']
                child_summary = child['fields'].get('summary', 'No summary')
                child_status = child['fields'].get('status', {}).get('name', 'Unknown')
                child_type = child['fields'].get('issuetype', {}).get('name', '-')

                # Format child display
                console.print(f"   {prefix} [cyan]{child_key}[/cyan]: {child_summary[:50]}{'...' if len(child_summary) > 50 else ''}")
                console.print(f"   {'  ' if is_last else '│ '}    {format_status(child_status)}  [dim]{child_type}[/dim]")
        console.print()

    # Display orphaned subtasks if any
    if orphaned_subtasks:
        console.print(Rule("Subtasks (Parent not in view)", style="yellow"))
        for ticket in orphaned_subtasks:
            parent_info = ticket['fields'].get('parent', {})
            parent_key = parent_info.get('key', 'Unknown')
            parent_summary = parent_info.get('fields', {}).get('summary', '') if parent_info else ''

            console.print(f"  📎 [cyan]{ticket['key']}[/cyan]: {ticket['fields'].get('summary', 'No summary')[:50]}")
            console.print(f"      Parent: [dim]{parent_key}: {parent_summary[:30]}{'...' if len(parent_summary) > 30 else ''}[/dim]")
            console.print(f"      {format_status(ticket['fields'].get('status', {}).get('name', 'Unknown'))}")
            console.print()

    # Display standalone tasks
    if standalone_tasks:
        console.print(Rule("Standalone Tasks", style="green"))
        table = Table(show_header=True, box=None)
        table.add_column("Key", style="cyan")
        table.add_column("Summary", overflow="fold")
        table.add_column("Status")
        table.add_column("Type")

        for ticket in standalone_tasks:
            issue_type = ticket['fields'].get('issuetype', {}).get('name', '-')
            icon = get_issue_icon(issue_type)

            table.add_row(
                f"{icon} {ticket['key']}",
                ticket['fields'].get('summary', 'No summary')[:60],
                format_status(ticket['fields'].get('status', {}).get('name', 'Unknown')),
                issue_type
            )

        console.print(table)

    # Summary
    total = len(tickets)
    console.print()
    console.print(f"[dim]Total: {total} tickets[/dim]")


def organize_tickets_for_grouped_display(tickets, current_user_id=None):
    """Organize tickets into family groups for smart display."""
    from collections import defaultdict

    # Build mappings
    ticket_map = {t['key']: t for t in tickets}
    families = {}  # parent_key -> {'parent': ticket, 'children': [tickets]}
    standalone = []

    # First pass: identify all parent-child relationships
    for ticket in tickets:
        fields = ticket['fields']
        is_subtask = fields.get('issuetype', {}).get('subtask', False)
        parent_info = fields.get('parent')

        if is_subtask and parent_info:
            parent_key = parent_info.get('key')
            if parent_key:
                # Initialize family if needed
                if parent_key not in families:
                    families[parent_key] = {
                        'parent': None,
                        'children': []
                    }
                families[parent_key]['children'].append(ticket)

                # Mark ticket as assigned to user
                if current_user_id:
                    assignee = fields.get('assignee', {})
                    ticket['_assigned_to_me'] = assignee and assignee.get('accountId') == current_user_id

    # Second pass: assign parent tickets and identify standalone
    processed_keys = set()

    for ticket in tickets:
        ticket_key = ticket['key']
        is_subtask = ticket['fields'].get('issuetype', {}).get('subtask', False)

        # Skip if already processed as a child
        if is_subtask:
            processed_keys.add(ticket_key)
            continue

        # Check if this ticket is a parent
        if ticket_key in families:
            families[ticket_key]['parent'] = ticket
            processed_keys.add(ticket_key)

            # Mark if assigned to user
            if current_user_id:
                assignee = ticket['fields'].get('assignee', {})
                ticket['_assigned_to_me'] = assignee and assignee.get('accountId') == current_user_id
        elif ticket_key not in processed_keys:
            # It's a standalone ticket
            standalone.append(ticket)
            processed_keys.add(ticket_key)

            # Mark if assigned to user
            if current_user_id:
                assignee = ticket['fields'].get('assignee', {})
                ticket['_assigned_to_me'] = assignee and assignee.get('accountId') == current_user_id

    # Sort families by priority and status
    def get_sort_key(family):
        parent = family.get('parent')
        if parent:
            priority_order = {'Critical': 0, 'High': 1, 'Medium': 2, 'Low': 3}
            status_order = {'In Progress': 0, 'To Do': 1, 'Backlog': 2, 'Done': 3}

            priority = parent['fields'].get('priority', {}).get('name', 'Low')
            status = parent['fields'].get('status', {}).get('name', 'Backlog')

            return (priority_order.get(priority, 4), status_order.get(status, 4))
        return (5, 5)  # Put families without parents at the end

    sorted_families = sorted(families.values(), key=get_sort_key)

    # Sort standalone tickets
    standalone.sort(key=lambda t: (
        {'In Progress': 0, 'To Do': 1, 'Backlog': 2, 'Done': 3}.get(
            t['fields'].get('status', {}).get('name', 'Backlog'), 4
        ),
        t['key']
    ))

    return sorted_families, standalone


@jira_group.command('fetch')
@click.option('--jql', help='Custom JQL query to filter tickets')
@click.option('--sprint', default='current', help='Sprint filter: current, all, or sprint ID')
@click.option('--limit', default=50, help='Maximum number of tickets to fetch')
@click.option('--format', default='smart', type=click.Choice(['smart', 'table', 'json', 'simple', 'grouped']), help='Output format')
@click.option('--flat', is_flag=True, help='Disable smart grouping (flat table)')
@click.option('--no-parents', is_flag=True, help='Skip fetching parent tickets')
def jira_fetch(jql, sprint, limit, format, flat, no_parents):
    """Fetch Jira tickets with smart grouping and parent resolution."""
    from rich.table import Table
    from rich.progress import Progress, SpinnerColumn, TextColumn
    import json as json_module

    jira_service = get_jira_service()

    # Get current user for marking assigned tickets
    try:
        creds = jira_service.email
        current_user_id = None  # Would need to fetch from API, simplified for now
    except:
        current_user_id = None

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        transient=True,
    ) as progress:
        task = progress.add_task("Fetching tickets...", total=None)

        try:
            # Smart fetch with parent resolution (unless disabled)
            tickets = jira_service.fetch_assigned_tickets(
                jql=jql,
                sprint_filter=sprint,
                max_results=limit,
                fetch_parents=not no_parents
            )
        except Exception as e:
            console.print(f"[red]Error fetching tickets: {e}[/red]")
            return

    if not tickets:
        console.print("[yellow]No tickets found matching your criteria[/yellow]")
        return

    # Handle different output formats
    if format == 'json':
        console.print_json(json_module.dumps(tickets, indent=2, default=str))
    elif format == 'simple':
        for ticket in tickets:
            key = ticket['key']
            summary = ticket['fields'].get('summary', 'No summary')
            status = ticket['fields'].get('status', {}).get('name', 'Unknown')
            console.print(f"{key}: {summary} [{status}]")
    elif format == 'grouped':
        # Legacy grouped view
        display_grouped_tickets(tickets, console)
    elif format == 'smart' and not flat:
        # Smart grouped table (default)
        families, standalone = organize_tickets_for_grouped_display(tickets, current_user_id)
        display_smart_grouped_table(families, standalone, console, current_user_id)
    else:  # table format or flat
        table = Table(title=f"Jira Tickets ({len(tickets)} found)")
        table.add_column("Key", style="cyan", no_wrap=True)
        table.add_column("Parent", style="blue", no_wrap=True)
        table.add_column("Summary", overflow="fold")
        table.add_column("Status", justify="center")
        table.add_column("Priority", justify="center")
        table.add_column("Type", justify="center")

        for ticket in tickets:
            priority = ticket['fields'].get('priority', {}).get('name', '-')
            issue_type = ticket['fields'].get('issuetype', {}).get('name', '-')

            # Check if this is a subtask
            is_subtask = issue_type.lower() in ['sub-task', 'subtask']

            # Get parent ticket if this is a subtask
            parent_display = ''
            if 'parent' in ticket['fields'] and ticket['fields']['parent']:
                parent_key = ticket['fields']['parent'].get('key', '')
                parent_summary = ticket['fields']['parent'].get('fields', {}).get('summary', '')
                if parent_summary:
                    # Truncate parent summary if too long
                    if len(parent_summary) > 30:
                        parent_summary = parent_summary[:27] + '...'
                    parent_display = f"{parent_key}: {parent_summary}"
                else:
                    parent_display = parent_key

            # Color code priority
            if priority == 'Critical':
                priority = f"[bold red]{priority}[/bold red]"
            elif priority == 'High':
                priority = f"[red]{priority}[/red]"
            elif priority == 'Medium':
                priority = f"[yellow]{priority}[/yellow]"
            elif priority == 'Low':
                priority = f"[green]{priority}[/green]"

            # Color code status
            status_name = ticket['fields'].get('status', {}).get('name', 'Unknown')
            if status_name == 'Done':
                status = f"[green]{status_name}[/green]"
            elif status_name == 'In Progress':
                status = f"[yellow]{status_name}[/yellow]"
            elif status_name == 'To Do':
                status = f"[dim]{status_name}[/dim]"
            else:
                status = status_name

            # Format ticket key based on type
            if is_subtask:
                ticket_key = f"  └─ {ticket['key']}"  # Indent subtasks
            else:
                ticket_key = ticket['key']

            # Color code issue type
            if is_subtask:
                issue_type = f"[dim]{issue_type}[/dim]"
            elif issue_type == 'Story':
                issue_type = f"[cyan]{issue_type}[/cyan]"
            elif issue_type == 'Bug':
                issue_type = f"[red]{issue_type}[/red]"
            elif issue_type == 'Task':
                issue_type = f"[yellow]{issue_type}[/yellow]"

            table.add_row(
                ticket_key,
                parent_display if parent_display else '-',
                ticket['fields'].get('summary', 'No summary'),
                status,
                priority,
                issue_type
            )

        console.print(table)

        # Show cache status if offline
        if not jira_service._is_online():
            console.print("[yellow]📵 Showing cached results (offline mode)[/yellow]")


@jira_group.command('recent')
@click.option('--days', default=7, help='Number of days to look back')
@click.option('--group-by', default='day', type=click.Choice(['day', 'ticket', 'none']), help='How to group results')
def jira_recent(days, group_by):
    """Show recent work activities from the last N days."""
    from rich.table import Table
    from collections import defaultdict

    jira_service = get_jira_service()
    worklogs = jira_service.get_recent_worklogs(days=days)

    if not worklogs:
        console.print(f"[yellow]No work logged in the last {days} days[/yellow]")
        return

    if group_by == 'day':
        # Group by day
        logs_by_day = defaultdict(list)
        for log in worklogs:
            date = datetime.fromisoformat(log['started_at']).date()
            logs_by_day[date].append(log)

        for date in sorted(logs_by_day.keys(), reverse=True):
            console.print(f"\n[bold blue]📅 {date.strftime('%A, %B %d, %Y')}[/bold blue]")

            table = Table(show_header=False, box=None)
            table.add_column("", style="dim", width=10)
            table.add_column("")

            total_seconds = 0
            for log in logs_by_day[date]:
                duration_sec = jira_service._parse_duration_to_seconds(log['duration'])
                total_seconds += duration_sec

                sync_icon = "✅" if log['synced'] else "💾"
                table.add_row(
                    f"{sync_icon} {log['ticket_key']}",
                    f"{log['duration']}: {log['description']}"
                )

            console.print(table)
            hours = total_seconds / 3600
            total_str = f"{hours:.1f}h" if hours >= 1 else f"{int(total_seconds/60)}m"
            console.print(f"[dim]Total: {total_str}[/dim]")

    elif group_by == 'ticket':
        # Group by ticket
        logs_by_ticket = defaultdict(list)
        for log in worklogs:
            logs_by_ticket[log['ticket_key']].append(log)

        table = Table(title=f"Work by Ticket (Last {days} days)")
        table.add_column("Ticket", style="cyan")
        table.add_column("Total Time", justify="right")
        table.add_column("Entries", justify="center")
        table.add_column("Last Worked")

        for ticket_key in sorted(logs_by_ticket.keys()):
            logs = logs_by_ticket[ticket_key]
            total_seconds = sum(jira_service._parse_duration_to_seconds(log['duration']) for log in logs)
            hours = total_seconds / 3600
            total_str = f"{hours:.1f}h" if hours >= 1 else f"{int(total_seconds/60)}m"

            last_date = max(datetime.fromisoformat(log['started_at']) for log in logs)

            table.add_row(
                ticket_key,
                total_str,
                str(len(logs)),
                last_date.strftime('%b %d')
            )

        console.print(table)

    else:  # no grouping
        table = Table(title=f"Recent Work (Last {days} days)")
        table.add_column("Date", style="dim")
        table.add_column("Ticket", style="cyan")
        table.add_column("Duration", justify="right")
        table.add_column("Description")
        table.add_column("Status", justify="center")

        for log in worklogs:
            date = datetime.fromisoformat(log['started_at']).strftime('%b %d')
            status = "✅ Synced" if log['synced'] else "💾 Cached"

            table.add_row(
                date,
                log['ticket_key'],
                log['duration'],
                log['description'][:40] + "..." if len(log['description']) > 40 else log['description'],
                status
            )

        console.print(table)


@jira_group.command('today')
@click.option('--copy', '-c', is_flag=True, help='Copy summary to clipboard')
def jira_today(copy):
    """Show today's schedule and work logged."""
    from rich.table import Table

    jira_service = get_jira_service()
    sheets_service = get_sheets_service()

    today = datetime.now()

    # Get today's logged work
    worklogs = jira_service.get_recent_worklogs(days=1)
    today_logs = [
        log for log in worklogs
        if datetime.fromisoformat(log['started_at']).date() == today.date()
    ]

    # Create display
    console.print(Panel.fit(f"📅 {today.strftime('%A, %B %d, %Y')}", style="bold blue"))

    # Today's work logged
    if today_logs:
        table = Table(title="Work Logged Today")
        table.add_column("Ticket", style="cyan")
        table.add_column("Duration", justify="right")
        table.add_column("Description")
        table.add_column("Status", justify="center")

        total_seconds = 0
        for log in today_logs:
            duration_sec = jira_service._parse_duration_to_seconds(log['duration'])
            total_seconds += duration_sec
            status = "✅" if log['synced'] else "💾"

            table.add_row(
                log['ticket_key'],
                log['duration'],
                log['description'][:40] + "..." if len(log['description']) > 40 else log['description'],
                status
            )

        console.print(table)

        hours = total_seconds / 3600
        total_str = f"{hours:.1f}h" if hours >= 1 else f"{int(total_seconds/60)}m"
        console.print(f"\n[bold]Total logged: {total_str}[/bold]")
    else:
        console.print("[yellow]No work logged yet today[/yellow]")

    # Today's schedule from Google Sheets (if available)
    if sheets_service and sheets_service.service:
        try:
            schedule = sheets_service.get_today_schedule()
            if schedule:
                console.print("\n[bold]Today's Schedule:[/bold]")
                for item in schedule:
                    console.print(f"  • {item}")
        except:
            pass

    # Current sprint tickets in progress
    try:
        tickets = jira_service.fetch_assigned_tickets(sprint_filter='current', max_results=5)
        in_progress = [
            t for t in tickets
            if t['fields'].get('status', {}).get('name') == 'In Progress'
        ]

        if in_progress:
            console.print("\n[bold]In Progress:[/bold]")
            for ticket in in_progress[:3]:
                console.print(f"  • {ticket['key']}: {ticket['fields'].get('summary', '')}")
    except:
        pass

    if copy:
        summary = f"Today ({today.strftime('%b %d')}): "
        if today_logs:
            summary += f"{total_str} logged across {len(today_logs)} tickets"
        else:
            summary += "No work logged yet"

        safe_copy_to_clipboard(summary, console)
        console.print("\n✅ Summary copied to clipboard")


@jira_group.command('sync')
@click.option('--force', '-f', is_flag=True, help='Force sync even if already synced')
def jira_sync(force):
    """Sync offline worklogs with Jira and Google Sheets."""
    from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn

    jira_service = get_jira_service()
    sheets_service = get_sheets_service()

    # Check if online
    if not jira_service._is_online():
        console.print("[red]❌ Cannot sync - no internet connection[/red]")
        return

    # Get unsynced worklogs
    worklogs = jira_service.get_recent_worklogs(days=30)
    unsynced = [log for log in worklogs if not log['synced']] if not force else worklogs

    if not unsynced:
        console.print("[green]✅ All worklogs are already synced[/green]")
        return

    console.print(f"[bold]Found {len(unsynced)} worklog(s) to sync[/bold]")

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
    ) as progress:
        task = progress.add_task("Syncing to Jira...", total=len(unsynced))

        success_count = 0
        failed = []

        for log in unsynced:
            try:
                # Sync to Jira (this will be handled by the service)
                if jira_service._sync_worklogs():
                    success_count += 1
                else:
                    failed.append(log)
            except Exception as e:
                failed.append(log)
                console.print(f"[red]Failed to sync {log['ticket_key']}: {e}[/red]")

            progress.update(task, advance=1)

    # Sync to Google Sheets if available
    if sheets_service and sheets_service.service:
        console.print("📊 Syncing to Google Sheets...")
        try:
            for log in unsynced:
                sheets_service.append_work_log(
                    log['ticket_key'],
                    log['duration'],
                    log['description']
                )
            console.print("[green]✅ Google Sheets updated[/green]")
        except Exception as e:
            console.print(f"[yellow]⚠ Google Sheets sync failed: {e}[/yellow]")

    # Summary
    console.print(f"\n[bold]Sync Complete:[/bold]")
    console.print(f"  ✅ Successfully synced: {success_count}/{len(unsynced)}")
    if failed:
        console.print(f"  ❌ Failed: {len(failed)}")


@jira_group.command('sprint')
@click.option('--list', 'list_all', is_flag=True, help='List all sprints')
@click.option('--board-id', help='Board ID (will try to auto-detect if not provided)')
def jira_sprint(list_all, board_id):
    """Show current sprint information or list all sprints."""
    from rich.table import Table

    jira_service = get_jira_service()

    if list_all:
        console.print("[yellow]Sprint listing not fully implemented yet[/yellow]")
        # This would require fetching from /rest/agile/1.0/board/{boardId}/sprint
        return

    # Get current sprint
    sprint = jira_service.get_current_sprint()

    if not sprint:
        console.print("[yellow]No active sprint found[/yellow]")
        return

    console.print(Panel.fit(f"🏃 Current Sprint: {sprint.get('name', 'Unknown')}", style="bold green"))

    # Show sprint details if available
    if sprint.get('start_date') and sprint.get('end_date'):
        start = datetime.fromisoformat(sprint['start_date'])
        end = datetime.fromisoformat(sprint['end_date'])
        days_left = (end - datetime.now()).days

        table = Table(show_header=False, box=None)
        table.add_column("", style="dim")
        table.add_column("")

        table.add_row("Start Date:", start.strftime('%B %d, %Y'))
        table.add_row("End Date:", end.strftime('%B %d, %Y'))
        table.add_row("Days Remaining:", f"{days_left} days" if days_left > 0 else "Sprint ended")

        console.print(table)

    # Get sprint tickets
    try:
        tickets = jira_service.fetch_assigned_tickets(sprint_filter='current', max_results=50)

        if tickets:
            # Group by status
            by_status = {}
            for ticket in tickets:
                status = ticket['fields'].get('status', {}).get('name', 'Unknown')
                if status not in by_status:
                    by_status[status] = []
                by_status[status].append(ticket)

            console.print(f"\n[bold]Sprint Progress:[/bold]")
            for status in ['To Do', 'In Progress', 'Done']:
                if status in by_status:
                    count = len(by_status[status])
                    console.print(f"  {status}: {count} ticket(s)")
    except:
        pass

@jira_group.command('sheets-auth')
@click.option('--force', is_flag=True, help='Force re-authentication even if token exists')
@click.option('--revoke', is_flag=True, help='Revoke existing token')
@click.option('--status', is_flag=True, help='Show authentication status')
def jira_sheets_auth(force, revoke, status):
    """Ultra-simple Google Sheets authentication - just a few clicks!"""
    from helper_cli.services.simple_sheets_oauth import SimpleGoogleSheets
    from rich.table import Table
    from rich.panel import Panel

    # Get sheet ID from config
    credential_manager = get_credential_manager()
    sheets_config = credential_manager.get_google_sheets_config()
    sheet_id = sheets_config.get('sheet_id') if sheets_config else None

    oauth = SimpleGoogleSheets(sheet_id=sheet_id)

    # Show status
    if status:
        auth_status = oauth.check_auth_status()

        table = Table(show_header=False, box=None)
        table.add_column("", style="bold")
        table.add_column("")

        table.add_row("Authenticated:", "✅ Yes" if auth_status['authenticated'] else "❌ No")
        table.add_row("Token exists:", "✅ Yes" if auth_status['has_token'] else "❌ No")
        
        if auth_status['has_credentials']:
            if auth_status['using_test_credentials']:
                table.add_row("Credentials:", "✅ Using Google's test app (auto-configured)")
            elif auth_status['has_custom_credentials']:
                table.add_row("Credentials:", "✅ Using custom OAuth app")
            else:
                table.add_row("Credentials:", "✅ Found")
        else:
            table.add_row("Credentials:", "❌ Not found (will auto-create)")
        
        table.add_row("Token path:", auth_status['token_path'])
        
        if auth_status.get('error'):
            table.add_row("Error:", f"❌ {auth_status['error']}")

        console.print(Panel(table, title="🔐 OAuth2 Authentication Status", border_style="blue"))
        return

    # Revoke token
    if revoke:
        if oauth.revoke_auth():
            console.print("✅ Token revoked. Run 'helper jira sheets-auth' to re-authenticate.", style="green")
        else:
            console.print("ℹ️ No token found to revoke.", style="yellow")
        return

    # Authenticate
    try:
        console.print(Panel.fit("🚀 Zero-Config Google Sheets Authentication!", style="bold blue"))

        # Check credential status
        custom_creds = oauth.config_dir / 'oauth_credentials.json'
        
        if not custom_creds.exists():
            console.print("\n✨ First time? No problem!", style="green bold")
            console.print("We'll automatically set up Google's test OAuth app for you.", style="green")
            console.print("Perfect for personal use - no Google Cloud Console needed!\n", style="dim")
        else:
            console.print("\n📋 Using your custom OAuth credentials", style="yellow")
            console.print(f"  From: {custom_creds}", style="dim")
        
        console.print("What happens next:", style="yellow")
        console.print("1. Your browser will open", style="dim")
        console.print("2. Sign in with your Google account", style="dim")  
        console.print("3. Click 'Allow' to grant access", style="dim")
        console.print("4. Done forever! Token saved for future use", style="dim")
        
        console.print("\n🌐 Opening browser for authentication...\n", style="cyan")

        # Authenticate
        client = oauth.authenticate(force_reauth=force)

        # Test connection
        if oauth.test_connection():
            console.print("\n✅ Authentication successful! Token saved.", style="green bold")
            
            # Test with sheet if available  
            if sheet_id:
                console.print(f"\n🔍 Testing connection to your sheet: {sheet_id}", style="dim")
                try:
                    spreadsheet = oauth.open_sheet(sheet_id)
                    console.print(f"✅ Successfully connected to: {spreadsheet.title}", style="green")
                    
                    # Show available worksheets
                    console.print("\n📋 Available worksheets:", style="cyan")
                    for i, ws in enumerate(spreadsheet.worksheets()):
                        prefix = "  →" if hasattr(oauth, 'current_worksheet') and oauth.current_worksheet and ws.title == oauth.current_worksheet.title else "   "
                        console.print(f"{prefix} {ws.title}", style="dim")
                    
                    if hasattr(oauth, 'current_worksheet') and oauth.current_worksheet:
                        console.print(f"\n✅ Auto-selected worksheet: {oauth.current_worksheet.title}", style="green")
                    
                    console.print("\n🎉 You're all set! Run 'helper jira sheets' to start.", style="cyan bold")
                except Exception as e:
                    console.print(f"⚠️ Couldn't access sheet: {e}", style="yellow")
                    console.print("Make sure the sheet exists and you have access.", style="dim")
            else:
                console.print("\n💡 No sheet ID configured yet.", style="yellow")
                console.print("Add your Google Sheet ID to the config and run 'helper jira sheets-tui'", style="dim")
        else:
            console.print("⚠️ Authentication completed but connection test failed.", style="yellow")

    except ImportError:
        console.print("❌ gspread not installed. Run: pip install gspread", style="red")
    except Exception as e:
        console.print(f"❌ Authentication failed: {e}", style="red")
        console.print("Try 'helper jira sheets-auth --force' to re-authenticate.", style="yellow")


@jira_group.command('sheets')
@click.option('--sheet-id', envvar='GOOGLE_SHEET_ID', help='Google Sheet ID')
@click.option('--use-oauth/--use-service-account', default=True,
              help='Use OAuth2 (simpler) or Service Account')
@click.option('--credentials', envvar='GOOGLE_CREDENTIALS_PATH',
              default='~/.config/helper-cli/google-credentials.json',
              help='Path to Google service account credentials (if not using OAuth)')
def jira_sheets(sheet_id, use_oauth, credentials):
    """Launch Google Sheets Calendar for daily schedule tracking."""
    from pathlib import Path

    # Try to get sheet_id from stored config if not provided
    if not sheet_id:
        credential_manager = get_credential_manager()
        sheets_config = credential_manager.get_google_sheets_config()
        if sheets_config and sheets_config.get('sheet_id'):
            sheet_id = sheets_config['sheet_id']
            console.print(f"📊 Using Sheet ID from config: {sheet_id}", style="dim")

    # Check requirements
    if not sheet_id:
        console.print("❌ No Google Sheet ID provided. Set GOOGLE_SHEET_ID or use --sheet-id", style="red")
        console.print("Run 'helper jira config --interactive' to set up", style="yellow")
        return

    if use_oauth:
        # Use OAuth2 authentication (simpler!)
        from helper_cli.services.simple_sheets_oauth import SimpleGoogleSheets

        console.print("🔐 Using OAuth2 authentication (recommended)", style="green")
        oauth = SimpleGoogleSheets(sheet_id=sheet_id)

        # Check auth status
        auth_status = oauth.check_auth_status()
        if not auth_status['authenticated'] and not auth_status['has_token']:
            console.print("\n⚠️ Not authenticated yet. Let's set up OAuth2:", style="yellow")
            console.print("Run: helper jira sheets-auth", style="cyan")
            console.print("\nThis is a one-time setup that uses your Google account!", style="dim")
            return

        try:
            # Test authentication
            if not oauth.test_connection():
                console.print("⚠️ Could not access sheet. Trying to re-authenticate...", style="yellow")
                oauth.authenticate(force_reauth=True)
        except Exception as e:
            console.print(f"❌ Authentication error: {e}", style="red")
            console.print("Run: helper jira sheets-auth --force", style="yellow")
            return

        # Launch TUI with OAuth
        try:
            console.print("🚀 Launching Google Sheets Calendar TUI (OAuth2)...", style="green")
            console.print(f"📊 Sheet ID: {sheet_id}", style="dim")
            
            # Try to open the sheet first to test access
            try:
                spreadsheet = oauth.open_sheet(sheet_id)
                console.print(f"✅ Successfully connected to Google Sheets: {spreadsheet.title}", style="green")
                
                # List available worksheets
                if hasattr(oauth, 'current_worksheet') and oauth.current_worksheet:
                    console.print(f"📋 Using worksheet: {oauth.current_worksheet.title}", style="cyan")
                
                # Use better sheets TUI with OAuth
                import sys
                try:
                    from helper_cli.tui.better_sheets_tui import BetterSheetsApp
                    app = BetterSheetsApp(demo_mode=False, oauth_service=oauth)
                    app.run()
                except ImportError as e:
                    console.print(f"❌ Failed to load TUI: {e}", style="red")
                    console.print("Try: pip install textual", style="yellow")
                except KeyboardInterrupt:
                    pass  # User pressed Ctrl+C
                except Exception as e:
                    console.print(f"❌ TUI error: {e}", style="red")
                    import traceback
                    if "--debug" in sys.argv:
                        traceback.print_exc()
                finally:
                    # Clean up terminal
                    sys.stdout.write("\033[?1000l\033[?1003l\033[?1015l\033[?1006l")
                    sys.stdout.flush()
                
            except Exception as e:
                error_msg = str(e)
                if "quota project" in error_msg.lower():
                    console.print("\n⚠️  Using test credentials with limitations.", style="yellow")
                    console.print("📝 For full access, create your own OAuth app:", style="cyan")
                    console.print("   https://console.cloud.google.com/apis/credentials/oauthclient", style="dim")
                    console.print("\n🎯 Launching simplified TUI...", style="green")
                    
                    # Use better sheets TUI in demo mode
                    import sys
                    try:
                        from helper_cli.tui.better_sheets_tui import BetterSheetsApp
                        app = BetterSheetsApp(demo_mode=True, oauth_service=None)
                        app.run()
                    except Exception:
                        pass  # Silently handle any TUI exit exceptions
                    finally:
                        # Clean up terminal
                        sys.stdout.write("\033[?1000l\033[?1003l\033[?1015l\033[?1006l")
                        sys.stdout.flush()
                    return
                elif "permission" in error_msg.lower() or isinstance(e, PermissionError):
                    console.print(f"\n⚠️  Cannot access sheet: {sheet_id}", style="yellow")
                    console.print("\nPossible causes:", style="cyan")
                    console.print("1. Sheet doesn't exist or ID is incorrect", style="dim")
                    console.print("2. Sheet isn't shared with your Google account", style="dim")
                    console.print("3. Test credentials have limitations", style="dim")
                    console.print("\n💡 Tips:", style="green")
                    console.print("• Check the sheet exists at:", style="dim")
                    console.print(f"  https://docs.google.com/spreadsheets/d/{sheet_id}", style="cyan")
                    console.print("• Make sure you're logged in with the same Google account", style="dim")
                    console.print("• Try creating a new sheet and sharing it with yourself", style="dim")
                    console.print("\n🎯 Launching demo TUI (without sheet access)...", style="green")
                    
                    # Use better sheets TUI in demo mode
                    import sys
                    try:
                        from helper_cli.tui.better_sheets_tui import BetterSheetsApp
                        app = BetterSheetsApp(demo_mode=True, oauth_service=None)
                        app.run()
                    except Exception:
                        pass  # Silently handle any TUI exit exceptions
                    finally:
                        # Clean up terminal
                        sys.stdout.write("\033[?1000l\033[?1003l\033[?1015l\033[?1006l")
                        sys.stdout.flush()
                    return
                else:
                    console.print(f"❌ Error opening sheet: {e}", style="red")
                    console.print("Please check your sheet ID and permissions.", style="yellow")
                    return

        except ImportError as ie:
            # Fall back to service account version if OAuth app not available
            console.print(f"⚠️ OAuth TUI import error: {ie}", style="yellow")
            console.print("Falling back to service account version", style="yellow")
            use_oauth = False

    if not use_oauth:
        # Use Service Account authentication (original)
        credentials_path = Path(credentials).expanduser()
        if not credentials_path.exists():
            console.print(f"❌ Google credentials not found at {credentials_path}", style="red")
            console.print("\n💡 Tip: OAuth2 is much simpler! Run:", style="yellow")
            console.print("   helper jira sheets-auth", style="cyan")
            console.print("\nOr set up Service Account:", style="dim")
            console.print("   1. Go to https://console.cloud.google.com", style="dim")
            console.print("   2. Create a service account and download JSON key", style="dim")
            console.print(f"   3. Save it to {credentials_path}", style="dim")
            console.print("   4. Share your Google Sheet with the service account email", style="dim")
            return

        try:
            from helper_cli.tui.sheets_calendar_app import GoogleSheetsCalendarApp

            console.print("🚀 Launching Google Sheets Calendar TUI (Service Account)...", style="green")
            console.print(f"📊 Sheet ID: {sheet_id}", style="dim")

            # Launch the TUI app
            app = GoogleSheetsCalendarApp(sheet_id, str(credentials_path))
            app.run()

        except ImportError as e:
            console.print(f"❌ Missing dependencies: {e}", style="red")
            console.print("Install with: pip install textual google-api-python-client google-auth", style="yellow")
        except Exception as e:
            console.print(f"❌ Error launching TUI: {e}", style="red")


@jira_group.command('config')
@click.option('--interactive', '-i', is_flag=True, help='Interactive configuration setup')
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
            table.add_row("Jira URL:", jira_creds['base_url'])
            table.add_row("Jira Email:", jira_creds['email'])
            table.add_row("Jira Token:", "✅ Configured")
        else:
            table.add_row("Jira:", "[red]Not configured[/red]")

        if sheets_config:
            table.add_row("Google Sheets ID:", sheets_config.get('sheet_id', 'Not set'))
        else:
            table.add_row("Google Sheets:", "[red]Not configured[/red]")

        console.print(table)
        console.print("\nRun with --interactive for guided setup")
        return

    # Interactive configuration
    console.print(Panel.fit("⚙️ Jira Configuration Setup", style="bold blue"))

    # Jira configuration
    console.print("\n[bold]Jira Configuration[/bold]")
    console.print("Get your API token from: https://id.atlassian.com/manage-profile/security/api-tokens")

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
        creds_path = console.input("Path to service account JSON (or Enter to skip): ").strip()

        creds_json = None
        if creds_path and Path(creds_path).exists():
            creds_json = Path(creds_path).read_text()

        if credential_manager.save_google_sheets_config(sheet_id, creds_json):
            console.print("[green]✅ Google Sheets configuration saved[/green]")
        else:
            console.print("[yellow]⚠ Google Sheets partially configured[/yellow]")

    console.print("\n[green]✅ Configuration complete![/green]")
    console.print("You can now use 'helper jira log' and other commands.")


@main.command('name-finder')
@click.argument('names', nargs=-1, required=True)
@click.option('--output', '-o', default='naming-report', help='Base filename for exports (without extension)')
@click.option('--format', '-f', multiple=True, default=['table'],
              type=click.Choice(['table', 'csv', 'markdown', 'json', 'all']),
              help='Export formats (can specify multiple)')
@click.option('--no-social', is_flag=True, help='Skip social media checks for faster results')
@click.option('--no-cache', is_flag=True, help='Disable caching for fresh results')
@click.option('--workers', '-w', default=10, type=int, help='Number of parallel workers')
@click.option('--verbose', '-v', is_flag=True, help='Show detailed debug information')
def name_finder(names, output, format, no_social, no_cache, workers, verbose):
    """Find and verify the best name for your project with comprehensive checking"""
    console.print(f"🔍 [bold blue]Name Finder[/bold blue] - Analyzing {len(names)} name{'s' if len(names) > 1 else ''}...")
    console.print(f"[dim]Checking domains, platforms{', and social media' if not no_social else ''}...[/dim]\n")

    # Use the new refactored checker
    checker = NameAvailabilityChecker(cache_enabled=not no_cache, parallel_workers=workers, verbose=verbose)

    try:
        # Determine what to check
        social_platforms = None if no_social else ['instagram', 'twitter', 'linkedin', 'reddit']

        # Check names
        results = checker.check_names(
            list(names),
            domains=['.com', '.org', '.net', '.io', '.dev', '.eu'],
            platforms=['npm', 'github', 'gitlab', 'pypi', 'dockerhub'],
            social=social_platforms,
            show_progress=True
        )

        # Display table if requested
        if 'table' in format or not format:
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
                console.print("\n" + "="*50)
                if best_score >= 80:
                    console.print(f"🏆 [bold green]Best option: {best_name} ({best_score:.0f}% availability)[/bold green]")
                elif best_score >= 50:
                    console.print(f"🏆 [bold yellow]Best option: {best_name} ({best_score:.0f}% availability)[/bold yellow]")
                else:
                    console.print(f"⚠️ [bold red]Best option: {best_name} ({best_score:.0f}% availability)[/bold red]")

        # Export in requested formats
        export_formats = list(format)
        if 'all' in export_formats:
            export_formats = ['csv', 'markdown', 'json']

        if any(f in export_formats for f in ['csv', 'markdown', 'json']):
            console.print(f"\n[bold]Exporting results...[/bold]")

            for fmt in export_formats:
                if fmt == 'csv':
                    filepath = Path(f"{output}.csv")
                    checker.export_csv(results, filepath)
                    console.print(f"📄 CSV saved to: [bold green]{filepath}[/bold green]")
                elif fmt == 'markdown':
                    filepath = Path(f"{output}.md")
                    checker.export_markdown(results, filepath)
                    console.print(f"📄 Markdown saved to: [bold green]{filepath}[/bold green]")
                elif fmt == 'json':
                    filepath = Path(f"{output}.json")
                    checker.export_json(results, filepath)
                    console.print(f"📄 JSON saved to: [bold green]{filepath}[/bold green]")

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
        ("log [ticket] [agent]", "Create work log for VIS tickets (auto-detect with -a)"),
        ("ngrok <url>", "Generate QR code for NGROK URL"),
        ("vsc <url>", "Generate QR code for VSC URL"),
        ("rt <name>", "Setup RT directory"),
        ("apex <file> <port> <auth>", "Generate APEX upload command"),
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
            if user_input.lower() in ['exit', 'quit', 'q']:
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


if __name__ == '__main__':
    main()
