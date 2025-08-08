"""Command line interface for Helper CLI."""

import click
import pyperclip
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from .core import HelperCore

console = Console()
helper = HelperCore()


@click.group()
@click.version_option()
def main():
    """Development helper CLI for JIRA, badges, URLs, and automation tasks."""
    pass


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
@click.option('--type', '-t', is_flag=True, help='Type command directly using wtype/ydotool')
def jira(text, copy, type):
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
                pyperclip.copy(branch_command)
                console.print(f"✅ Copied to clipboard instead: [bold green]{branch_command}[/bold green]")
                return

        helper.type_string_with_delay(branch_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{branch_command}[/bold green]")
    elif copy:
        pyperclip.copy(branch_command)
        console.print(f"✅ Copied to clipboard: [bold green]{branch_command}[/bold green]")
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
        pyperclip.copy(badge_string)
        console.print("✅ Badge copied to clipboard")
    elif copy:
        pyperclip.copy(badge_string)
        console.print("✅ Badge copied to clipboard")
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
        pyperclip.copy(badge_id)
        console.print("✅ Badge copied to clipboard")
    else:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                pyperclip.copy(badge_id)
                console.print("✅ Badge copied to clipboard instead")
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
        pyperclip.copy(badge_id)
        console.print("✅ Badge copied to clipboard")
    else:
        # Check if we're on Wayland and no typing tools available
        if os.environ.get('XDG_SESSION_TYPE') == 'wayland':
            if not shutil.which('wtype') and not shutil.which('ydotool'):
                console.print("[yellow]⚠ Auto-type on Wayland requires 'wtype' or 'ydotool'. Install with:[/yellow]")
                console.print("[dim]  sudo dnf install wtype  # or[/dim]")
                console.print("[dim]  sudo dnf install ydotool[/dim]")
                pyperclip.copy(badge_id)
                console.print("✅ Badge copied to clipboard instead")
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
                pyperclip.copy(stash_command)
                console.print(f"✅ Copied to clipboard instead: [bold green]{stash_command}[/bold green]")
                return

        helper.type_string_with_delay(stash_command, delay=0)  # No delay for commands
        console.print(f"✅ Typed command: [bold green]{stash_command}[/bold green]")
    elif copy:
        pyperclip.copy(stash_command)
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
        pyperclip.copy(qr_command)
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
        pyperclip.copy(qr_command)
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
        pyperclip.copy(upload_command)
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
                pyperclip.copy(dash_result)
                console.print(f"✅ Copied to clipboard instead: [bold green]{dash_result}[/bold green]")
                return

        helper.type_string_with_delay(dash_result, delay=0)
        console.print(f"✅ Typed: [bold green]{dash_result}[/bold green]")
    elif copy:
        pyperclip.copy(dash_result)
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
        pyperclip.copy(pr_title)
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
        pyperclip.copy(filename_result)
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


@main.command()
def interactive():
    """Interactive mode for helper commands."""
    console.print(Panel.fit("🔧 Helper CLI - Interactive Mode", style="bold blue"))

    table = Table(title="Available Commands")
    table.add_column("Command", justify="left", style="cyan")
    table.add_column("Description", justify="left")

    commands = [
        ("jira <text>", "Generate JIRA branch name"),
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
