"""Domain and name availability checker for Helper CLI."""

import subprocess
import requests
from datetime import datetime
from rich.table import Table
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn
from typing import Optional, List, Tuple, Dict


class DomainChecker:
    """Check domain and platform availability for project names."""
    
    def __init__(self):
        self.console = Console()
        self.tlds = ['.pl', '.com', '.app', '.io']
    
    def check_single_domain(self, domain: str) -> Optional[bool]:
        """
        Check if a single domain is available.
        
        Args:
            domain: Full domain name (e.g., 'example.com')
            
        Returns:
            True if available, False if taken, None if error
        """
        # Normalize domain to lowercase
        domain = domain.lower()
        
        try:
            # Run whois command
            result = subprocess.run(
                ['whois', domain], 
                capture_output=True, 
                text=True,
                timeout=10
            )
            
            # Check for availability indicators in whois output
            availability_indicators = [
                'No information available',
                'No match for',
                'NOT FOUND',
                'No Data Found',
                'No entries found',
                'Status: free',
                'domain does not exist'
            ]
            
            output_lower = result.stdout.lower()
            for indicator in availability_indicators:
                if indicator.lower() in output_lower:
                    return True
            
            # If we see common taken indicators
            if 'registrant' in output_lower or 'created' in output_lower:
                return False
                
            # Default to taken if we can't determine
            return False
            
        except subprocess.TimeoutExpired:
            return None
        except FileNotFoundError:
            # whois command not available
            self.console.print("[yellow]⚠ whois command not found. Install with: sudo dnf install whois[/yellow]")
            return None
        except Exception:
            return None
    
    def check_all_tlds(self, name: str) -> Table:
        """
        Check domain availability across all TLDs.
        
        Args:
            name: Base name to check (without TLD)
            
        Returns:
            Rich Table with results
        """
        # Normalize name to lowercase
        name = name.lower()
        
        table = Table(title=f"Domain Availability for: {name}")
        table.add_column("Domain", style="cyan", no_wrap=True)
        table.add_column("Status", justify="center")
        
        for tld in self.tlds:
            domain = f"{name}{tld}"
            status = self.check_single_domain(domain)
            
            if status is True:
                table.add_row(domain, "[green]✅ Available[/green]")
            elif status is False:
                table.add_row(domain, "[red]❌ Taken[/red]")
            else:
                table.add_row(domain, "[yellow]⚠️ Error[/yellow]")
        
        return table
    
    def check_github(self, username: str) -> Optional[bool]:
        """
        Check if GitHub username is available.
        
        Args:
            username: GitHub username to check
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            response = requests.get(
                f"https://github.com/{username}",
                timeout=5,
                headers={'User-Agent': 'Helper-CLI-Checker/1.0'}
            )
            return response.status_code == 404
        except requests.RequestException:
            return None
    
    def check_npm(self, package_name: str) -> Optional[bool]:
        """
        Check if npm package name is available.
        
        Args:
            package_name: npm package name to check
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            result = subprocess.run(
                ['npm', 'view', package_name],
                capture_output=True,
                text=True,
                timeout=10
            )
            # If package doesn't exist, npm returns error
            return 'E404' in result.stderr or 'code E404' in result.stderr
        except subprocess.TimeoutExpired:
            return None
        except FileNotFoundError:
            self.console.print("[yellow]⚠ npm not found. Install Node.js first.[/yellow]")
            return None
        except Exception:
            return None
    
    def generate_trademark_urls(self, name: str) -> Tuple[str, str]:
        """
        Generate trademark search URLs.
        
        Args:
            name: Name to search for
            
        Returns:
            Tuple of (UP RP URL, EUIPO URL)
        """
        # URL encode the name for search queries
        import urllib.parse
        encoded_name = urllib.parse.quote(name)
        
        uprp_url = f"https://ewyszukiwarka.pue.uprp.gov.pl/search/pwp-st?query={encoded_name}"
        euipo_url = f"https://euipo.europa.eu/eSearch/#basic/1+1+1+1/50+50+50+50/{encoded_name}"
        
        return uprp_url, euipo_url
    
    def check_name_with_progress(self, name: str) -> Dict:
        """
        Check all availability for a name with live progress output.
        
        Args:
            name: Name to check
            
        Returns:
            Dictionary with results and score
        """
        # Normalize to lowercase
        name = name.lower()
        results = {
            'name': name,
            'domains': {},
            'platforms': {},
            'available_count': 0,
            'total_count': 0
        }
        
        self.console.print(f"\n📊 Checking [bold cyan]{name}[/bold cyan]:")
        
        # Check domains
        for tld in self.tlds:
            domain = f"{name}{tld}"
            self.console.print(f"  • {domain:<20}", end="")
            status = self.check_single_domain(domain)
            results['total_count'] += 1
            
            if status is True:
                self.console.print("[green]✅ Available[/green]")
                results['domains'][domain] = True
                results['available_count'] += 1
            elif status is False:
                self.console.print("[red]❌ Taken[/red]")
                results['domains'][domain] = False
            else:
                self.console.print("[yellow]⚠️ Error[/yellow]")
                results['domains'][domain] = None
        
        # Check GitHub
        self.console.print(f"  • {'GitHub':<20}", end="")
        gh_status = self.check_github(name)
        results['total_count'] += 1
        
        if gh_status is True:
            self.console.print("[green]✅ Available[/green]")
            results['platforms']['github'] = True
            results['available_count'] += 1
        elif gh_status is False:
            self.console.print("[red]❌ Taken[/red]")
            results['platforms']['github'] = False
        else:
            self.console.print("[yellow]⚠️ Error[/yellow]")
            results['platforms']['github'] = None
        
        # Check npm
        self.console.print(f"  • {'npm':<20}", end="")
        npm_status = self.check_npm(name)
        results['total_count'] += 1
        
        if npm_status is True:
            self.console.print("[green]✅ Available[/green]")
            results['platforms']['npm'] = True
            results['available_count'] += 1
        elif npm_status is False:
            self.console.print("[red]❌ Taken[/red]")
            results['platforms']['npm'] = False
        else:
            self.console.print("[yellow]⚠️ Error[/yellow]")
            results['platforms']['npm'] = None
        
        # Calculate score
        if results['total_count'] > 0:
            score_percent = (results['available_count'] / results['total_count']) * 100
            score_text = f"{results['available_count']}/{results['total_count']} ({score_percent:.0f}%)"
            
            if score_percent >= 80:
                self.console.print(f"  [bold green]Score: {score_text} ⭐[/bold green]")
            elif score_percent >= 50:
                self.console.print(f"  [bold yellow]Score: {score_text}[/bold yellow]")
            else:
                self.console.print(f"  [bold red]Score: {score_text}[/bold red]")
        
        results['score_percent'] = score_percent if results['total_count'] > 0 else 0
        
        return results
    
    def generate_report(self, names: List[str], output_file: str = 'naming-report.md') -> bool:
        """
        Generate comprehensive verification report for multiple names.
        
        Args:
            names: List of names to check
            output_file: Output file path for the report
            
        Returns:
            True if successful, False otherwise
        """
        try:
            report = []
            report.append("# Name Verification Report - Name Finder\n")
            report.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M')}\n")
            report.append(f"Checked names: {', '.join(names)}\n")
            report.append("\n---\n")
            
            for name in names:
                # Normalize name to lowercase
                name_normalized = name.lower()
                report.append(f"\n## {name}\n")
                
                # Domain availability
                report.append("\n### Domain Availability\n")
                for tld in self.tlds:
                    domain = f"{name_normalized}{tld}"
                    status = self.check_single_domain(domain)
                    if status is True:
                        report.append(f"- {domain}: ✅ **Available**\n")
                    elif status is False:
                        report.append(f"- {domain}: ❌ Taken\n")
                    else:
                        report.append(f"- {domain}: ⚠️ Could not verify\n")
                
                # Platform availability
                report.append("\n### Platform Availability\n")
                
                # GitHub
                gh_status = self.check_github(name_normalized)
                if gh_status is True:
                    report.append(f"- **GitHub**: ✅ Available (github.com/{name_normalized})\n")
                elif gh_status is False:
                    report.append(f"- **GitHub**: ❌ Taken\n")
                else:
                    report.append(f"- **GitHub**: ⚠️ Could not verify\n")
                
                # npm
                npm_status = self.check_npm(name_normalized)
                if npm_status is True:
                    report.append(f"- **npm**: ✅ Available (npmjs.com/package/{name_normalized})\n")
                elif npm_status is False:
                    report.append(f"- **npm**: ❌ Taken\n")
                else:
                    report.append(f"- **npm**: ⚠️ Could not verify\n")
                
                # Trademark URLs
                report.append("\n### Trademark Verification\n")
                uprp_url, euipo_url = self.generate_trademark_urls(name)
                report.append(f"- [Check UP RP Database]({uprp_url})\n")
                report.append(f"- [Check EUIPO Database]({euipo_url})\n")
                
                report.append("\n---\n")
            
            # Summary section
            report.append("\n## Summary\n\n")
            report.append("### Next Steps\n")
            report.append("1. Verify trademark availability for preferred names\n")
            report.append("2. Register domains for selected name\n")
            report.append("3. Secure social media handles\n")
            report.append("4. Register npm package if applicable\n")
            report.append("5. Create GitHub organization/user\n")
            
            # Write report to file
            with open(output_file, 'w', encoding='utf-8') as f:
                f.writelines(report)
            
            return True
            
        except Exception as e:
            self.console.print(f"[red]Error generating report: {e}[/red]")
            return False