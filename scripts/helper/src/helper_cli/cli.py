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
def jira(text, copy):
    """Generate JIRA branch name from commit message text."""
    branch_command = helper.generate_jira_branch_name(text)
    
    if copy:
        pyperclip.copy(branch_command)
        console.print(f"✅ Copied to clipboard: [bold green]{branch_command}[/bold green]")
    else:
        console.print(f"Branch command: [bold blue]{branch_command}[/bold blue]")


@main.command()
@click.argument('badge_id')
@click.option('--type', '-t', type=click.Choice(['rfid', 'qr']), default='rfid', 
              help='Badge type (rfid or qr)')
@click.option('--auto-type/--no-auto-type', default=True, 
              help='Automatically type the badge (requires pynput)')
def badge(badge_id, type, auto_type):
    """Generate and optionally auto-type badge strings."""
    if type == 'rfid':
        badge_string = f'@{badge_id}#'
    else:  # qr
        badge_string = f'${badge_id}#'
    
    console.print(f"[bold blue]{type.upper()} badge:[/bold blue] {badge_string}")
    
    if auto_type:
        helper.type_string_with_delay(badge_string)
        console.print("✅ Badge typed automatically")
    else:
        pyperclip.copy(badge_string)
        console.print("✅ Badge copied to clipboard")


@main.command()
@click.argument('text')
@click.option('--copy/--no-copy', default=True, help='Copy result to clipboard')
def stash(text, copy):
    """Generate git stash command with formatted message."""
    stash_command = helper.generate_stash_command(text)
    
    if copy:
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
def interactive():
    """Interactive mode for helper commands."""
    console.print(Panel.fit("🔧 Helper CLI - Interactive Mode", style="bold blue"))
    
    # Create options table
    table = Table(title="Available Commands")
    table.add_column("Command", justify="left", style="cyan")
    table.add_column("Description", justify="left")
    
    commands = [
        ("jira <text>", "Generate JIRA branch name"),
        ("badge <id>", "Generate badge string (RFID/QR)"),
        ("stash <text>", "Generate git stash command"),
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
                # Parse and execute command
                parts = user_input.split()
                if parts:
                    try:
                        main(parts, standalone_mode=False)
                    except SystemExit:
                        pass  # Click commands call sys.exit, ignore in interactive mode
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