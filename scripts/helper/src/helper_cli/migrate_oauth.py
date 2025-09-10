#!/usr/bin/env python3
"""Migration script for consolidating OAuth implementations."""

import os
import json
import shutil
import logging
from pathlib import Path
from typing import Dict, List, Tuple

logging.basicConfig(level=logging.INFO, format='%(message)s')
logger = logging.getLogger(__name__)


class OAuthMigrator:
    """Migrate from multiple OAuth implementations to unified auth."""
    
    def __init__(self):
        """Initialize migrator."""
        self.config_dir = Path.home() / '.config' / 'helper-cli'
        self.backup_dir = self.config_dir / 'oauth_backup'
        
        # Old token locations
        self.old_tokens = [
            self.config_dir / 'gspread_token.json',
            self.config_dir / 'oauth_token.json',
            self.config_dir / 'google_token.json',
            Path.home() / '.config' / 'gspread' / 'authorized_user.json',
        ]
        
        # Old credential locations
        self.old_credentials = [
            self.config_dir / 'oauth_credentials.json',
            self.config_dir / 'google_credentials.json',
            self.config_dir / 'google_service_account.json',
            Path.home() / '.config' / 'gspread' / 'credentials.json',
        ]
        
        # New unified locations
        self.new_token = self.config_dir / 'google_token.json'
        self.new_credentials = self.config_dir / 'google_credentials.json'
        
        # Files to archive (old implementations)
        self.files_to_archive = [
            'src/helper_cli/services/google_oauth_service.py',
            'src/helper_cli/services/oauth_setup_helper.py',
            'src/helper_cli/services/simple_sheets_oauth.py',
            'src/helper_cli/services/sheets_api.py',
        ]
        
        # Documentation to archive
        self.docs_to_archive = [
            'docs/FINAL_OAUTH_SOLUTION.md',
            'docs/GOOGLE_SHEETS_CREDENTIALS_SETUP.md',
            'docs/GOOGLE_SHEETS_TUI_INTEGRATION.md',
            'docs/IMPLEMENTATION_GUIDE.md',
            'docs/IMPLEMENTATION_SUMMARY.md',
            'docs/OAUTH2_SETUP_GUIDE.md',
            'docs/QUICKEST_OAUTH_SETUP.md',
            'docs/SHEETS_TUI_SETUP.md',
            'docs/ULTRA_SIMPLE_OAUTH.md',
        ]
    
    def run(self) -> bool:
        """Run the migration process."""
        logger.info("🔄 Starting OAuth Migration Process")
        logger.info("=" * 50)
        
        # Step 1: Create backup directory
        if not self.create_backup_dir():
            return False
        
        # Step 2: Find and backup existing OAuth files
        tokens_found, creds_found = self.find_existing_auth()
        
        # Step 3: Migrate to unified locations
        if tokens_found or creds_found:
            if not self.migrate_auth_files(tokens_found, creds_found):
                return False
        
        # Step 4: Archive old implementations
        if not self.archive_old_implementations():
            return False
        
        # Step 5: Archive old documentation
        if not self.archive_old_docs():
            return False
        
        # Step 6: Update imports
        if not self.update_imports():
            return False
        
        logger.info("\n✅ Migration completed successfully!")
        logger.info("\nNext steps:")
        logger.info("1. Test authentication: helper jira sheets-auth --status")
        logger.info("2. If needed, re-authenticate: helper jira sheets-auth")
        logger.info("3. Remove backup after testing: rm -rf " + str(self.backup_dir))
        
        return True
    
    def create_backup_dir(self) -> bool:
        """Create backup directory."""
        try:
            self.backup_dir.mkdir(parents=True, exist_ok=True)
            logger.info(f"✓ Created backup directory: {self.backup_dir}")
            return True
        except Exception as e:
            logger.error(f"✗ Failed to create backup directory: {e}")
            return False
    
    def find_existing_auth(self) -> Tuple[List[Path], List[Path]]:
        """Find existing OAuth files."""
        tokens_found = []
        creds_found = []
        
        logger.info("\n📍 Searching for existing OAuth files...")
        
        # Check for tokens
        for token_path in self.old_tokens:
            if token_path.exists():
                tokens_found.append(token_path)
                logger.info(f"  Found token: {token_path}")
        
        # Check for credentials
        for cred_path in self.old_credentials:
            if cred_path.exists():
                creds_found.append(cred_path)
                logger.info(f"  Found credentials: {cred_path}")
        
        if not tokens_found and not creds_found:
            logger.info("  No existing OAuth files found")
        
        return tokens_found, creds_found
    
    def migrate_auth_files(self, tokens: List[Path], credentials: List[Path]) -> bool:
        """Migrate authentication files to unified locations."""
        logger.info("\n📦 Migrating authentication files...")
        
        try:
            # Migrate token (prefer gspread token if available)
            if tokens:
                # Prefer gspread token as it's most likely to be valid
                gspread_token = Path.home() / '.config' / 'gspread' / 'authorized_user.json'
                if gspread_token in tokens:
                    source_token = gspread_token
                else:
                    source_token = tokens[0]  # Use first found
                
                # Backup existing if present
                if self.new_token.exists():
                    shutil.copy2(self.new_token, self.backup_dir / 'google_token.json.backup')
                
                # Copy to new location
                shutil.copy2(source_token, self.new_token)
                os.chmod(self.new_token, 0o600)
                logger.info(f"  ✓ Migrated token to: {self.new_token}")
            
            # Migrate credentials (prefer custom over default)
            if credentials:
                source_creds = credentials[0]  # Use first found
                
                # Backup existing if present
                if self.new_credentials.exists():
                    shutil.copy2(self.new_credentials, self.backup_dir / 'google_credentials.json.backup')
                
                # Copy to new location
                shutil.copy2(source_creds, self.new_credentials)
                os.chmod(self.new_credentials, 0o600)
                logger.info(f"  ✓ Migrated credentials to: {self.new_credentials}")
            
            # Backup all old files
            for file_path in tokens + credentials:
                backup_path = self.backup_dir / file_path.name
                shutil.copy2(file_path, backup_path)
                logger.info(f"  ✓ Backed up: {file_path.name}")
            
            return True
            
        except Exception as e:
            logger.error(f"✗ Migration failed: {e}")
            return False
    
    def archive_old_implementations(self) -> bool:
        """Archive old OAuth implementation files."""
        logger.info("\n📁 Archiving old implementations...")
        
        archive_dir = Path('src/helper_cli/services/_archived')
        
        try:
            for file_path in self.files_to_archive:
                source = Path(file_path)
                if source.exists():
                    archive_dir.mkdir(parents=True, exist_ok=True)
                    dest = archive_dir / source.name
                    shutil.move(str(source), str(dest))
                    logger.info(f"  ✓ Archived: {source.name}")
            
            return True
            
        except Exception as e:
            logger.error(f"✗ Failed to archive implementations: {e}")
            return False
    
    def archive_old_docs(self) -> bool:
        """Archive old documentation files."""
        logger.info("\n📚 Archiving old documentation...")
        
        archive_dir = Path('docs/_archived_oauth')
        
        try:
            for doc_path in self.docs_to_archive:
                source = Path(doc_path)
                if source.exists():
                    archive_dir.mkdir(parents=True, exist_ok=True)
                    dest = archive_dir / source.name
                    shutil.move(str(source), str(dest))
                    logger.info(f"  ✓ Archived: {source.name}")
            
            return True
            
        except Exception as e:
            logger.error(f"✗ Failed to archive documentation: {e}")
            return False
    
    def update_imports(self) -> bool:
        """Update imports in Python files to use unified auth."""
        logger.info("\n🔧 Updating imports...")
        
        # Map old imports to new
        import_map = {
            'from .google_oauth_service import GoogleOAuthService': 
                'from .unified_google_auth import UnifiedGoogleAuth as GoogleOAuthService',
            'from .simple_sheets_oauth import SimpleGoogleSheets':
                'from .unified_google_auth import UnifiedGoogleAuth',
            'from helper_cli.services.google_oauth_service':
                'from helper_cli.services.unified_google_auth',
            'from helper_cli.services.simple_sheets_oauth':
                'from helper_cli.services.unified_google_auth',
        }
        
        # Files to check and update
        files_to_update = [
            'src/helper_cli/cli.py',
            'src/helper_cli/services/__init__.py',
            'src/helper_cli/services/google_sheets_service.py',
        ]
        
        try:
            for file_path in files_to_update:
                path = Path(file_path)
                if path.exists():
                    content = path.read_text()
                    original = content
                    
                    # Replace imports
                    for old_import, new_import in import_map.items():
                        if old_import in content:
                            content = content.replace(old_import, new_import)
                            logger.info(f"  ✓ Updated import in: {path.name}")
                    
                    # Write back if changed
                    if content != original:
                        path.write_text(content)
            
            return True
            
        except Exception as e:
            logger.error(f"✗ Failed to update imports: {e}")
            return False


def main():
    """Run the migration."""
    migrator = OAuthMigrator()
    
    print("\n" + "=" * 60)
    print("  Google Sheets OAuth Migration Tool")
    print("=" * 60)
    print("\nThis tool will:")
    print("1. Backup existing OAuth files")
    print("2. Migrate to unified authentication")
    print("3. Archive old implementations")
    print("4. Update import statements")
    print("\nYour existing authentication will be preserved.")
    
    response = input("\nProceed with migration? (y/n): ")
    
    if response.lower() == 'y':
        success = migrator.run()
        if not success:
            print("\n⚠️  Migration incomplete. Check errors above.")
            print("Your original files are backed up in:", migrator.backup_dir)
    else:
        print("\nMigration cancelled.")


if __name__ == "__main__":
    main()