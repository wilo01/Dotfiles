"""Domain and name availability checker for Helper CLI."""

import subprocess
import requests
import socket
import shutil
import csv
import json
from datetime import datetime
from pathlib import Path
from rich.table import Table
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.panel import Panel
from typing import Optional, List, Tuple, Dict


class DomainChecker:
    """Check domain and platform availability for project names."""
    
    def __init__(self):
        self.console = Console()
        self.tlds = ['.com', '.eu', '.io', '.dev', '.app', '.org', '.pl']
        # Check for available tools
        self.has_whois = shutil.which('whois') is not None
        self.has_dig = shutil.which('dig') is not None
        self.has_curl = shutil.which('curl') is not None
        # Try to import DNS resolver
        try:
            import dns.resolver
            self.dns_resolver = dns.resolver.Resolver()
            self.dns_resolver.timeout = 3
            self.dns_resolver.lifetime = 3
            self.has_dnspython = True
        except ImportError:
            self.has_dnspython = False
            self.dns_resolver = None
    
    def check_single_domain(self, domain: str, verbose: bool = False) -> Optional[bool]:
        """
        Check if a single domain is available using multiple verification methods.
        
        Args:
            domain: Full domain name (e.g., 'example.com')
            verbose: Show detailed output from each check
            
        Returns:
            True if available, False if taken, None if error
        """
        # Normalize domain to lowercase
        domain = domain.lower()
        
        # Show header
        self.console.print(f"\n[bold cyan]Checking domain: {domain}[/bold cyan]")
        self.console.print("━" * 50)
        
        results = []
        confidence_scores = []
        
        # 1. WHOIS Check
        self.console.print("\n[bold]WHOIS Check[/bold]")
        whois_result, whois_confidence, whois_details = self._check_whois_enhanced(domain, verbose)
        if whois_result is not None:
            if whois_result:
                self.console.print(f"✅ WHOIS: [green]AVAILABLE[/green] (Confidence: {whois_confidence}%)")
            else:
                self.console.print(f"❌ WHOIS: [red]RESERVED[/red] {whois_details}")
            results.append(whois_result)
            confidence_scores.append(whois_confidence)
        else:
            self.console.print(f"⚠️ WHOIS: [yellow]Unable to check[/yellow]")
        
        # 2. DNS Checks
        self.console.print("\n[bold]DNS Checks[/bold]")
        
        # DNS A Record
        dns_a_result, dns_a_ips = self._check_dns_a_record(domain, verbose)
        if dns_a_result is not None:
            if dns_a_result:
                self.console.print(f"✅ DNS A: [green]AVAILABLE[/green] (No A records)")
            else:
                ips_str = ", ".join(dns_a_ips) if dns_a_ips else "Records found"
                self.console.print(f"❌ DNS A: [red]RESERVED[/red] ({ips_str})")
            results.append(dns_a_result)
            confidence_scores.append(100 if dns_a_result is False else 80)
        else:
            self.console.print(f"⚠️ DNS A: [yellow]Unable to check[/yellow]")
        
        # DNS NS Record
        dns_ns_result, nameservers = self._check_dns_ns_record(domain, verbose)
        if dns_ns_result is not None:
            if dns_ns_result:
                self.console.print(f"✅ DNS NS: [green]AVAILABLE[/green] (No nameservers)")
            else:
                ns_str = ", ".join(nameservers[:2]) if nameservers else "Nameservers found"
                if len(nameservers) > 2:
                    ns_str += f", +{len(nameservers)-2} more"
                self.console.print(f"❌ DNS NS: [red]RESERVED[/red] ({ns_str})")
            results.append(dns_ns_result)
            confidence_scores.append(95 if dns_ns_result is False else 85)
        else:
            self.console.print(f"⚠️ DNS NS: [yellow]Unable to check[/yellow]")
        
        # 3. HTTP Check
        self.console.print("\n[bold]HTTP Check[/bold]")
        http_result, http_status = self._check_http_status(domain, verbose)
        if http_result is not None:
            if http_result:
                self.console.print(f"✅ HTTP: [green]AVAILABLE[/green] (No response)")
            else:
                self.console.print(f"❌ HTTP: [red]ACTIVE[/red] (Status: {http_status})")
            results.append(http_result)
            confidence_scores.append(90 if http_result is False else 70)
        else:
            self.console.print(f"⚠️ HTTP: [yellow]Unable to check[/yellow]")
        
        # Calculate final verdict
        self.console.print("\n" + "━" * 50)
        
        if not results:
            self.console.print("[yellow]⚠️ Unable to determine domain status[/yellow]")
            return None
        
        # Count votes
        available_votes = sum(1 for r in results if r is True)
        taken_votes = sum(1 for r in results if r is False)
        total_votes = len(results)
        
        # Calculate weighted confidence
        if confidence_scores:
            avg_confidence = sum(confidence_scores) / len(confidence_scores)
        else:
            avg_confidence = 0
        
        # Determine verdict
        if taken_votes > available_votes:
            is_available = False
            confidence = int((taken_votes / total_votes) * 100)
            verdict = f"[bold red]VERDICT: Domain '{domain}' is RESERVED[/bold red]"
        elif available_votes > taken_votes:
            is_available = True
            confidence = int((available_votes / total_votes) * 100)
            verdict = f"[bold green]VERDICT: Domain '{domain}' is LIKELY AVAILABLE[/bold green]"
        else:
            # Tie - use confidence scores
            is_available = False  # Conservative approach
            confidence = 50
            verdict = f"[bold yellow]VERDICT: Domain '{domain}' status is UNCERTAIN[/bold yellow]"
        
        self.console.print(verdict)
        self.console.print(f"Confidence: {confidence}% ({taken_votes if not is_available else available_votes}/{total_votes} checks indicate {'reserved' if not is_available else 'available'})")
        self.console.print("━" * 50)
        
        return is_available
    
    def _check_whois_enhanced(self, domain: str, verbose: bool = False) -> Tuple[Optional[bool], int, str]:
        """
        Enhanced WHOIS check with detailed parsing.
        
        Returns:
            Tuple of (is_available, confidence, details)
        """
        if not self.has_whois:
            if verbose:
                self.console.print("[dim]whois command not available. Install with: sudo dnf install whois[/dim]")
            return (None, 0, "")
        
        try:
            result = subprocess.run(
                ['whois', domain], 
                capture_output=True, 
                text=True,
                timeout=10
            )
            
            output = result.stdout
            output_lower = output.lower()
            
            # Enhanced availability indicators
            availability_indicators = [
                'no match for domain',
                'no match for "',
                'not found',
                'no data found',
                'no entries found',
                'status: free',
                'status: available',
                'domain does not exist',
                'no information available',
                'domain not found',
                'no matching record',
                'object does not exist',
                'is free',
                'available for registration'
            ]
            
            # Check for availability
            for indicator in availability_indicators:
                if indicator in output_lower:
                    if verbose:
                        self.console.print(f"[dim]Found availability indicator: '{indicator}'[/dim]")
                    return (True, 95, "")
            
            # Look for registrar info
            registrar = None
            for line in output.split('\n'):
                if 'registrar:' in line.lower():
                    registrar = line.split(':', 1)[1].strip()
                    break
            
            details = f"(Registrar: {registrar})" if registrar else "(Registered)"
            
            # Check for rate limiting
            if 'rate limit' in output_lower or 'quota exceeded' in output_lower:
                return (None, 0, "(Rate limited)")
            
            return (False, 90, details)
            
        except subprocess.TimeoutExpired:
            return (None, 0, "(Timeout)")
        except Exception as e:
            if verbose:
                self.console.print(f"[dim]WHOIS error: {e}[/dim]")
            return (None, 0, "")
    
    def _check_dns_a_record(self, domain: str, verbose: bool = False) -> Tuple[Optional[bool], List[str]]:
        """
        Check for DNS A records using dig or dnspython.
        
        Returns:
            Tuple of (is_available, list_of_ips)
        """
        # Try dig first
        if self.has_dig:
            try:
                result = subprocess.run(
                    ['dig', '+short', 'A', domain],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                
                ips = [line.strip() for line in result.stdout.split('\n') if line.strip() and self._is_valid_ip(line.strip())]
                
                if verbose and ips:
                    self.console.print(f"[dim]Found A records: {', '.join(ips)}[/dim]")
                
                return (len(ips) == 0, ips)
                
            except Exception as e:
                if verbose:
                    self.console.print(f"[dim]dig error: {e}[/dim]")
        
        # Fallback to dnspython
        if self.has_dnspython:
            try:
                import dns.resolver
                answers = self.dns_resolver.resolve(domain, 'A')
                ips = [str(rdata) for rdata in answers]
                return (False, ips)
            except dns.resolver.NXDOMAIN:
                return (True, [])
            except dns.resolver.NoAnswer:
                return (True, [])
            except Exception:
                pass
        
        # Fallback to basic socket
        try:
            ip = socket.gethostbyname(domain)
            return (False, [ip])
        except socket.gaierror:
            return (True, [])
        except Exception:
            return (None, [])
    
    def _check_dns_ns_record(self, domain: str, verbose: bool = False) -> Tuple[Optional[bool], List[str]]:
        """
        Check for DNS NS records using dig or dnspython.
        
        Returns:
            Tuple of (is_available, list_of_nameservers)
        """
        # Try dig first
        if self.has_dig:
            try:
                result = subprocess.run(
                    ['dig', '+short', 'NS', domain],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                
                nameservers = [line.strip().rstrip('.') for line in result.stdout.split('\n') 
                             if line.strip() and not line.startswith(';')]
                
                if verbose and nameservers:
                    self.console.print(f"[dim]Found nameservers: {', '.join(nameservers)}[/dim]")
                
                return (len(nameservers) == 0, nameservers)
                
            except Exception as e:
                if verbose:
                    self.console.print(f"[dim]dig error: {e}[/dim]")
        
        # Fallback to dnspython
        if self.has_dnspython:
            try:
                import dns.resolver
                answers = self.dns_resolver.resolve(domain, 'NS')
                nameservers = [str(rdata).rstrip('.') for rdata in answers]
                return (False, nameservers)
            except dns.resolver.NXDOMAIN:
                return (True, [])
            except dns.resolver.NoAnswer:
                return (True, [])
            except Exception:
                pass
        
        return (None, [])
    
    def _check_http_status(self, domain: str, verbose: bool = False) -> Tuple[Optional[bool], str]:
        """
        Check HTTP/HTTPS status using curl or requests.
        
        Returns:
            Tuple of (is_available, status_description)
        """
        # Try curl first
        if self.has_curl:
            for protocol in ['https', 'http']:
                try:
                    result = subprocess.run(
                        ['curl', '-I', '-s', '-m', '5', '-w', '%{http_code}', '-o', '/dev/null', f'{protocol}://{domain}'],
                        capture_output=True,
                        text=True,
                        timeout=6
                    )
                    
                    status_code = result.stdout.strip()
                    
                    if status_code and status_code.isdigit():
                        code = int(status_code)
                        if verbose:
                            self.console.print(f"[dim]HTTP status code: {code}[/dim]")
                        
                        if 200 <= code < 600:
                            status_desc = f"{code} {self._get_http_status_text(code)}"
                            return (False, status_desc)
                
                except Exception as e:
                    if verbose:
                        self.console.print(f"[dim]curl error: {e}[/dim]")
        
        # Fallback to requests
        for protocol in ['https', 'http']:
            try:
                response = requests.get(
                    f"{protocol}://{domain}",
                    timeout=5,
                    allow_redirects=True,
                    headers={'User-Agent': 'Mozilla/5.0'}
                )
                
                status_desc = f"{response.status_code} {self._get_http_status_text(response.status_code)}"
                return (False, status_desc)
                
            except requests.ConnectionError:
                continue
            except requests.Timeout:
                continue
            except Exception:
                continue
        
        return (True, "No response")
    
    def _is_valid_ip(self, ip: str) -> bool:
        """Check if string is a valid IP address."""
        try:
            parts = ip.split('.')
            return len(parts) == 4 and all(0 <= int(part) <= 255 for part in parts)
        except:
            return False
    
    def _get_http_status_text(self, code: int) -> str:
        """Get human-readable HTTP status text."""
        status_texts = {
            200: "OK",
            301: "Moved Permanently",
            302: "Found",
            304: "Not Modified",
            400: "Bad Request",
            401: "Unauthorized",
            403: "Forbidden",
            404: "Not Found",
            500: "Internal Server Error",
            502: "Bad Gateway",
            503: "Service Unavailable"
        }
        return status_texts.get(code, "")
    
    def _check_domain_with_dns_fallback(self, domain: str) -> Optional[bool]:
        """
        Check domain using multiple fallback methods when WHOIS fails.
        
        Fallback order:
        1. DNS A records (using dig)
        2. DNS NS records (using dig)
        3. DNS with Python resolver
        4. HTTP/HTTPS check
        5. Socket connection check
        
        Returns:
            True if likely available, False if likely taken, None only if all methods fail
        """
        # Method 1: Check DNS A records with dig
        if self.has_dig:
            try:
                result = subprocess.run(
                    ['dig', '+short', 'A', domain, '@8.8.8.8'],
                    capture_output=True,
                    text=True,
                    timeout=3
                )
                
                if result.returncode == 0:
                    output = result.stdout.strip()
                    # If we get IP addresses, domain is taken
                    if output and any(self._is_valid_ip(line.strip()) for line in output.split('\n') if line.strip()):
                        return False
                    # If we get NXDOMAIN or empty response, likely available
                    if not output or 'NXDOMAIN' in result.stderr:
                        # Double-check with NS records
                        ns_result = subprocess.run(
                            ['dig', '+short', 'NS', domain, '@8.8.8.8'],
                            capture_output=True,
                            text=True,
                            timeout=3
                        )
                        if ns_result.returncode == 0:
                            ns_output = ns_result.stdout.strip()
                            if ns_output and len(ns_output) > 5:
                                return False  # Has nameservers, domain is taken
                            else:
                                return True  # No nameservers, likely available
            except Exception:
                pass
        
        # Method 2: DNS with Python resolver
        if self.has_dnspython:
            try:
                import dns.resolver
                # Try to resolve A records
                try:
                    self.dns_resolver.resolve(domain, 'A')
                    return False  # Has A records, domain is taken
                except dns.resolver.NXDOMAIN:
                    return True  # Domain doesn't exist, available
                except dns.resolver.NoAnswer:
                    # No A records, check NS records
                    try:
                        self.dns_resolver.resolve(domain, 'NS')
                        return False  # Has nameservers, taken
                    except (dns.resolver.NXDOMAIN, dns.resolver.NoAnswer):
                        return True  # No NS records, likely available
            except Exception:
                pass
        
        # Method 3: HTTP/HTTPS check with curl
        if self.has_curl:
            try:
                for protocol in ['https', 'http']:
                    result = subprocess.run(
                        ['curl', '-I', '-s', '-m', '3', '--connect-timeout', '2',
                         '-w', '%{http_code}', '-o', '/dev/null', f'{protocol}://{domain}'],
                        capture_output=True,
                        text=True,
                        timeout=4
                    )
                    
                    if result.returncode == 0:
                        status_code = result.stdout.strip()
                        if status_code and status_code.isdigit():
                            code = int(status_code)
                            if code > 0:
                                return False  # Got HTTP response, domain exists
                # No HTTP response from any protocol
                return True  # Likely available
            except Exception:
                pass
        
        # Method 4: Basic socket check
        try:
            import socket
            # Try to resolve the domain
            socket.gethostbyname(domain)
            return False  # Resolution succeeded, domain exists
        except socket.gaierror:
            # Resolution failed, likely available
            return True
        except Exception:
            pass
        
        # Method 5: HTTP check with requests library
        try:
            for protocol in ['https', 'http']:
                response = requests.head(
                    f"{protocol}://{domain}",
                    timeout=3,
                    allow_redirects=False,
                    headers={'User-Agent': 'Mozilla/5.0'}
                )
                if response.status_code > 0:
                    return False  # Got response, domain exists
        except requests.exceptions.ConnectionError:
            # Connection failed, likely domain doesn't exist
            return True
        except requests.exceptions.Timeout:
            pass  # Timeout, try next method
        except Exception:
            pass
        
        # Method 6: Aggressive assumption for common patterns
        # For random-looking domains, they're likely available
        if len(domain.split('.')[0]) > 10 and any(char.isdigit() for char in domain):
            # Long domain with numbers, probably available
            return True
        
        # If all methods fail, return None
        return None
    
    def _check_single_domain_simple(self, domain: str) -> Optional[bool]:
        """
        Simple domain check for use in name-finder (original behavior).
        
        Args:
            domain: Full domain name (e.g., 'example.com')
            
        Returns:
            True if available, False if taken, None if error
        """
        # Normalize domain to lowercase
        domain = domain.lower()
        
        if not self.has_whois:
            return None
        
        try:
            # Run whois command
            result = subprocess.run(
                ['whois', domain], 
                capture_output=True, 
                text=True,
                timeout=10
            )
            
            output = result.stdout
            output_lower = output.lower()
            stderr_lower = result.stderr.lower()
            
            # Check for error conditions first
            error_indicators = [
                'request limit exceeded',
                'rate limit',
                'quota exceeded',
                'getaddrinfo',
                'name or service not known',
                'connection refused',
                'temporarily unavailable',
                'error',
                'timeout'
            ]
            
            for indicator in error_indicators:
                if indicator in output_lower or indicator in stderr_lower:
                    return None  # Return None for errors (will show as N/A)
            
            # Check for availability indicators
            availability_indicators = [
                'no match for',
                'not found',
                'no data found',
                'no entries found',
                'status: free',
                'domain does not exist',
                'no matching record',
                'object does not exist',
                'is free',
                'available for registration',
                'no information available'
            ]
            
            for indicator in availability_indicators:
                if indicator in output_lower:
                    return True
            
            # Check for taken indicators
            taken_indicators = [
                'registrant:',
                'registrar:',
                'created:',
                'creation date:',
                'domain name:',
                'registry domain id:',
                'updated date:',
                'expiry date:'
            ]
            
            for indicator in taken_indicators:
                if indicator in output_lower:
                    return False
            
            # If output is too short or empty, it's likely an error
            if len(output.strip()) < 50:
                return None
                
            # Default to None (unknown) instead of False
            return None
            
        except subprocess.TimeoutExpired:
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
    
    def check_gitlab(self, username: str) -> Optional[bool]:
        """
        Check if GitLab username is available.
        
        Args:
            username: GitLab username to check
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            response = requests.get(
                f"https://gitlab.com/{username}",
                timeout=5,
                allow_redirects=False,
                headers={'User-Agent': 'Helper-CLI-Checker/1.0'}
            )
            
            if response.status_code == 404:
                # Double-check with API
                api_response = requests.get(
                    f"https://gitlab.com/api/v4/users?username={username}",
                    timeout=5
                )
                if api_response.status_code == 200:
                    data = api_response.json()
                    return len(data) == 0  # Empty list means available
                return True
            elif response.status_code in [200, 301, 302]:
                return False
            else:
                return None
        except requests.RequestException:
            return None
    
    def check_instagram(self, username: str) -> Optional[bool]:
        """
        Check if Instagram handle is available.
        
        Args:
            username: Instagram handle to check
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            response = requests.get(
                f"https://www.instagram.com/{username}/",
                timeout=5,
                headers={'User-Agent': 'Mozilla/5.0'}
            )
            return response.status_code == 404
        except requests.RequestException:
            return None
    
    def check_twitter(self, username: str) -> Optional[bool]:
        """
        Check if Twitter/X handle is available.
        
        Args:
            username: Twitter handle to check (without @)
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            response = requests.get(
                f"https://x.com/{username}",
                timeout=5,
                headers={'User-Agent': 'Mozilla/5.0'},
                allow_redirects=True
            )
            return response.status_code == 404
        except requests.RequestException:
            return None
    
    def check_linkedin(self, username: str) -> Optional[bool]:
        """
        Check if LinkedIn company URL is available.
        
        Args:
            username: LinkedIn company handle to check
            
        Returns:
            True if available, False if taken, None if error
        """
        try:
            response = requests.get(
                f"https://www.linkedin.com/company/{username}",
                timeout=5,
                headers={'User-Agent': 'Mozilla/5.0'}
            )
            return response.status_code == 404
        except requests.RequestException:
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
    
    def check_names_table(self, names: List[str]) -> Tuple[Table, List[Dict]]:
        """
        Check multiple names and return a formatted table.
        
        Args:
            names: List of names to check
            
        Returns:
            Tuple of (Rich Table, List of result dictionaries)
        """
        from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TimeElapsedColumn
        
        # Create table with expanded columns
        table = Table(title="Domain, Platform & Social Media Availability", show_lines=True)
        table.add_column("Name", style="cyan", no_wrap=True)
        
        # Domain columns
        for tld in self.tlds:
            table.add_column(tld, justify="center", width=5)
        
        # Platform columns
        table.add_column("GitHub", justify="center", width=7)
        table.add_column("GitLab", justify="center", width=7)
        table.add_column("npm", justify="center", width=5)
        
        # Social media columns
        table.add_column("IG", justify="center", width=4)
        table.add_column("X", justify="center", width=4)
        table.add_column("LI", justify="center", width=4)
        
        table.add_column("Score", justify="center", width=10)
        
        all_results = []
        
        # Calculate total checks for progress bar
        total_checks_per_name = len(self.tlds) + 6  # domains + platforms + social
        total_checks = len(names) * total_checks_per_name
        
        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
            TimeElapsedColumn(),
            console=self.console
        ) as progress:
            overall_task = progress.add_task("[cyan]Checking availability...", total=total_checks)
            
            for name in names:
                name_task = progress.add_task(f"[yellow]Checking {name}...", total=total_checks_per_name)
                name = name.lower()
                results = {
                    'name': name,
                    'domains': {},
                    'platforms': {},
                    'available_count': 0,
                    'total_count': 0
                }
                
                row_data = [name]
                
                # Check each TLD
                for tld in self.tlds:
                    progress.update(name_task, description=f"[yellow]Checking {name}{tld}...")
                    domain = f"{name}{tld}"
                    # Try WHOIS first
                    status = self._check_single_domain_simple(domain)
                    
                    # If WHOIS fails, try DNS fallback
                    if status is None:
                        status = self._check_domain_with_dns_fallback(domain)
                    
                    results['domains'][domain] = status
                    results['total_count'] += 1
                    
                    if status is True:
                        row_data.append("[green]✅[/green]")
                        results['available_count'] += 1
                    elif status is False:
                        row_data.append("[red]❌[/red]")
                    elif status is None:
                        row_data.append("[yellow]N/A[/yellow]")
                    
                    progress.update(name_task, advance=1)
                    progress.update(overall_task, advance=1)
            
                # Check platforms - GitHub, GitLab, npm
                for platform_name, check_func in [
                    ('github', self.check_github),
                    ('gitlab', self.check_gitlab),
                    ('npm', self.check_npm)
                ]:
                    progress.update(name_task, description=f"[yellow]Checking {name} on {platform_name}...")
                    status = check_func(name)
                    results['platforms'][platform_name] = status
                    results['total_count'] += 1
                    if status is True:
                        row_data.append("[green]✅[/green]")
                        results['available_count'] += 1
                    elif status is False:
                        row_data.append("[red]❌[/red]")
                    else:
                        row_data.append("[yellow]N/A[/yellow]")
                    
                    progress.update(name_task, advance=1)
                    progress.update(overall_task, advance=1)
                
                # Check social media - Instagram, Twitter/X, LinkedIn
                results['social'] = {}
                for social_name, check_func in [
                    ('instagram', self.check_instagram),
                    ('twitter', self.check_twitter),
                    ('linkedin', self.check_linkedin)
                ]:
                    progress.update(name_task, description=f"[yellow]Checking {name} on {social_name}...")
                    status = check_func(name)
                    results['social'][social_name] = status
                    results['total_count'] += 1
                    if status is True:
                        row_data.append("[green]✅[/green]")
                        results['available_count'] += 1
                    elif status is False:
                        row_data.append("[red]❌[/red]")
                    else:
                        row_data.append("[yellow]N/A[/yellow]")
                    
                    progress.update(name_task, advance=1)
                    progress.update(overall_task, advance=1)
                
                # Calculate score (only count definitive results)
                all_values = list(results['domains'].values()) + list(results['platforms'].values()) + list(results.get('social', {}).values())
                definitive_count = sum(1 for v in all_values if v is not None)
                if definitive_count > 0:
                    score_percent = (results['available_count'] / definitive_count) * 100
                    if score_percent >= 80:
                        score_str = f"[green]{results['available_count']}/{definitive_count} ({score_percent:.0f}%)[/green]"
                    elif score_percent >= 50:
                        score_str = f"[yellow]{results['available_count']}/{definitive_count} ({score_percent:.0f}%)[/yellow]"
                    else:
                        score_str = f"[red]{results['available_count']}/{definitive_count} ({score_percent:.0f}%)[/red]"
                else:
                    score_str = "[dim]N/A[/dim]"
                    score_percent = 0
                
                row_data.append(score_str)
                table.add_row(*row_data)
                
                results['score_percent'] = score_percent
                results['definitive_count'] = definitive_count
                all_results.append(results)
                
                progress.update(name_task, description=f"[green]✓ Completed {name}")
                progress.remove_task(name_task)
        
        return table, all_results
    
    def export_results(self, results: List[Dict], format: str, output_file: str = None) -> bool:
        """
        Export results to various formats.
        
        Args:
            results: List of result dictionaries from check_names_table
            format: Export format ('csv', 'markdown', 'json')
            output_file: Output file path (optional, auto-generated if not provided)
            
        Returns:
            True if successful, False otherwise
        """
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        
        if format == 'csv':
            if not output_file:
                output_file = f"domain_check_{timestamp}.csv"
            return self._export_csv(results, output_file)
        elif format == 'markdown':
            if not output_file:
                output_file = f"domain_check_{timestamp}.md"
            return self._export_markdown(results, output_file)
        elif format == 'json':
            if not output_file:
                output_file = f"domain_check_{timestamp}.json"
            return self._export_json(results, output_file)
        else:
            self.console.print(f"[red]Unsupported format: {format}[/red]")
            return False
    
    def _export_csv(self, results: List[Dict], output_file: str) -> bool:
        """Export results to CSV format."""
        try:
            with open(output_file, 'w', newline='', encoding='utf-8') as csvfile:
                if not results:
                    return False
                
                # Prepare headers
                headers = ['Name']
                
                # Add domain headers
                for tld in self.tlds:
                    headers.append(tld)
                
                # Add platform headers
                headers.extend(['GitHub', 'GitLab', 'npm'])
                
                # Add social media headers
                headers.extend(['Instagram', 'Twitter/X', 'LinkedIn'])
                
                headers.append('Score %')
                
                writer = csv.DictWriter(csvfile, fieldnames=headers)
                writer.writeheader()
                
                for result in results:
                    row = {'Name': result['name']}
                    
                    # Add domain results
                    for tld in self.tlds:
                        domain = f"{result['name']}{tld}"
                        status = result['domains'].get(domain)
                        row[tld] = 'Available' if status is True else 'Taken' if status is False else 'N/A'
                    
                    # Add platform results
                    row['GitHub'] = 'Available' if result['platforms'].get('github') is True else 'Taken' if result['platforms'].get('github') is False else 'N/A'
                    row['GitLab'] = 'Available' if result['platforms'].get('gitlab') is True else 'Taken' if result['platforms'].get('gitlab') is False else 'N/A'
                    row['npm'] = 'Available' if result['platforms'].get('npm') is True else 'Taken' if result['platforms'].get('npm') is False else 'N/A'
                    
                    # Add social media results
                    row['Instagram'] = 'Available' if result.get('social', {}).get('instagram') is True else 'Taken' if result.get('social', {}).get('instagram') is False else 'N/A'
                    row['Twitter/X'] = 'Available' if result.get('social', {}).get('twitter') is True else 'Taken' if result.get('social', {}).get('twitter') is False else 'N/A'
                    row['LinkedIn'] = 'Available' if result.get('social', {}).get('linkedin') is True else 'Taken' if result.get('social', {}).get('linkedin') is False else 'N/A'
                    
                    row['Score %'] = f"{result.get('score_percent', 0):.0f}"
                    
                    writer.writerow(row)
            
            self.console.print(f"[green]✅ CSV exported to: {output_file}[/green]")
            return True
        except Exception as e:
            self.console.print(f"[red]Error exporting CSV: {e}[/red]")
            return False
    
    def _export_markdown(self, results: List[Dict], output_file: str) -> bool:
        """Export results to Markdown table format."""
        try:
            with open(output_file, 'w', encoding='utf-8') as mdfile:
                mdfile.write("# Domain, Platform & Social Media Availability Report\n\n")
                mdfile.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
                
                # Create table header
                headers = ['Name'] + self.tlds + ['GitHub', 'GitLab', 'npm', 'Instagram', 'Twitter/X', 'LinkedIn', 'Score']
                mdfile.write('| ' + ' | '.join(headers) + ' |\n')
                mdfile.write('|' + '---|' * len(headers) + '\n')
                
                # Add data rows
                for result in results:
                    row = [result['name']]
                    
                    # Add domain results
                    for tld in self.tlds:
                        domain = f"{result['name']}{tld}"
                        status = result['domains'].get(domain)
                        row.append('✅' if status is True else '❌' if status is False else 'N/A')
                    
                    # Add platform results
                    row.append('✅' if result['platforms'].get('github') is True else '❌' if result['platforms'].get('github') is False else 'N/A')
                    row.append('✅' if result['platforms'].get('gitlab') is True else '❌' if result['platforms'].get('gitlab') is False else 'N/A')
                    row.append('✅' if result['platforms'].get('npm') is True else '❌' if result['platforms'].get('npm') is False else 'N/A')
                    
                    # Add social media results
                    row.append('✅' if result.get('social', {}).get('instagram') is True else '❌' if result.get('social', {}).get('instagram') is False else 'N/A')
                    row.append('✅' if result.get('social', {}).get('twitter') is True else '❌' if result.get('social', {}).get('twitter') is False else 'N/A')
                    row.append('✅' if result.get('social', {}).get('linkedin') is True else '❌' if result.get('social', {}).get('linkedin') is False else 'N/A')
                    
                    # Add score
                    row.append(f"{result.get('score_percent', 0):.0f}%")
                    
                    mdfile.write('| ' + ' | '.join(row) + ' |\n')
                
                # Add summary
                mdfile.write('\n## Summary\n\n')
                if results:
                    best = max(results, key=lambda x: x.get('score_percent', 0))
                    mdfile.write(f"**Best option:** {best['name']} ({best.get('score_percent', 0):.0f}% availability)\n\n")
                
                mdfile.write('### Legend\n')
                mdfile.write('- ✅ Available\n')
                mdfile.write('- ❌ Taken\n')
                mdfile.write('- N/A Could not verify\n')
            
            self.console.print(f"[green]✅ Markdown exported to: {output_file}[/green]")
            return True
        except Exception as e:
            self.console.print(f"[red]Error exporting Markdown: {e}[/red]")
            return False
    
    def _export_json(self, results: List[Dict], output_file: str) -> bool:
        """Export results to JSON format."""
        try:
            export_data = {
                'timestamp': datetime.now().isoformat(),
                'results': results,
                'summary': {
                    'total_checked': len(results),
                    'best_option': max(results, key=lambda x: x.get('score_percent', 0)) if results else None
                }
            }
            
            with open(output_file, 'w', encoding='utf-8') as jsonfile:
                json.dump(export_data, jsonfile, indent=2)
            
            self.console.print(f"[green]✅ JSON exported to: {output_file}[/green]")
            return True
        except Exception as e:
            self.console.print(f"[red]Error exporting JSON: {e}[/red]")
            return False
    
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
        
        # Check domains using simple method for name-finder (to avoid too much output)
        for tld in self.tlds:
            domain = f"{name}{tld}"
            self.console.print(f"  • {domain:<20}", end="")
            status = self._check_single_domain_simple(domain)
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
                    status = self._check_single_domain_simple(domain)
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