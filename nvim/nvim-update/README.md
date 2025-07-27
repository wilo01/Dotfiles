# Neovim Update Suite

A comprehensive set of scripts for safely updating Neovim to the latest version with automatic backup and rollback capabilities.

## 🚀 Features

- **Smart Installation**: Tries Flatpak first, falls back to GitHub releases
- **Automatic Backup**: Creates timestamped backups before updates
- **Easy Rollback**: Restore previous installation if issues arise
- **Comprehensive Testing**: Validates installation and configuration
- **Future-Proof**: Works for ongoing updates (v0.12, v0.13, etc.)

## 📁 Scripts

### `nvim-update.sh` - Main Update Script
The primary script that handles the entire update process.

```bash
# Update to latest version
./nvim-update.sh

# See what would be done without making changes
./nvim-update.sh --dry-run

# Force GitHub installation (skip Flatpak)
./nvim-update.sh --force-github

# Show help
./nvim-update.sh --help
```

**Features:**
- Detects current and latest versions
- Creates automatic backups
- Tries Flatpak installation first
- Falls back to GitHub releases
- Validates installation

### `test-nvim.sh` - Test & Validation Script
Comprehensive testing of Neovim installation and configuration.

```bash
# Run full test suite
./test-nvim.sh

# Quick tests only
./test-nvim.sh --quick

# Verbose output
./test-nvim.sh --verbose
```

**Tests:**
- Installation verification
- Configuration loading
- Plugin functionality
- LSP integration
- File operations
- Performance checks
- Health verification

### `rollback-nvim.sh` - Rollback Script
Safely restore previous Neovim installation from backup.

```bash
# List available backups
./rollback-nvim.sh

# Restore specific backup
./rollback-nvim.sh 20241127_143022
```

**Features:**
- Lists all available backups
- Validates backup integrity
- Removes current installation
- Restores data and configuration
- Reinstalls system Neovim

## 🔄 Update Methods

### Method 1: Flatpak (Recommended)
- **Source**: Flathub (io.neovim.nvim)
- **Pros**: Sandboxed, easy updates, official package
- **Cons**: Slight isolation (manageable with wrapper)

### Method 2: GitHub Releases
- **Source**: Official Neovim GitHub releases
- **Pros**: Latest version, direct from source
- **Cons**: Manual updates, no package management

## 📋 Quick Start

1. **Update Neovim**:
   ```bash
   cd ~/.Dotfiles/nvim/nvim-update
   ./nvim-update.sh
   ```

2. **Test Installation**:
   ```bash
   ./test-nvim.sh
   ```

3. **If Issues Arise**:
   ```bash
   ./rollback-nvim.sh
   ```

## 🛡️ Safety Features

### Automatic Backups
- Located in `~/.nvim-backups/`
- Timestamped format: `nvim-YYYYMMDD_HHMMSS`
- Includes:
  - `~/.local/share/nvim` (data)
  - `~/.cache/nvim` (cache)
  - Binary path information

### Validation
- Tests basic functionality before completion
- Verifies configuration loads correctly
- Checks plugin compatibility
- Validates version upgrade

### Rollback Protection
- Preserves system Neovim as fallback
- Creates additional backup before rollback
- Validates backup integrity
- Safe restoration process

## 📝 Installation Methods Detail

### Flatpak Installation
When using Flatpak, the script:
1. Adds Flathub repository if needed
2. Installs/updates `io.neovim.nvim`
3. Creates wrapper script at `~/.local/bin/nvim`
4. Sets up configuration symlinks for sandbox access
5. Ensures seamless integration with your existing config

### GitHub Installation
When using GitHub releases, the script:
1. Downloads latest `nvim-linux-x86_64.tar.gz`
2. Extracts to `~/.local/nvim-linux-x86_64/`
3. Creates symlink at `~/.local/bin/nvim`
4. Creates desktop entry
5. Manages PATH precedence

## 🔧 Configuration

### PATH Precedence
The scripts respect your PATH precedence:
1. `~/.local/bin/nvim` (user installation)
2. `/usr/local/bin/nvim` (system-wide manual)
3. `/usr/bin/nvim` (package manager)

### Wrapper Integration
For Flatpak installations, a wrapper script ensures:
- Proper argument passing
- Environment variable handling
- Configuration access
- Plugin compatibility

## 🚨 Troubleshooting

### Common Issues

**Flatpak Installation Fails**
- Check if Flatpak is installed: `flatpak --version`
- Verify Flathub repository: `flatpak remotes`
- Script automatically falls back to GitHub

**GitHub Download Fails**
- Check internet connection
- Verify GitHub API access
- Try again later (rate limiting)

**Configuration Issues**
- Run test script: `./test-nvim.sh --verbose`
- Check health: `nvim +checkhealth`
- Consider rollback if persistent

**Permission Issues**
- Ensure `~/.local/bin` is in PATH
- Check file permissions on scripts
- Verify backup directory access

### Recovery Steps

1. **Installation Test Failed**:
   ```bash
   ./test-nvim.sh --verbose
   # Check specific failing tests
   ```

2. **Configuration Problems**:
   ```bash
   # Test with minimal config
   nvim --clean
   ```

3. **Complete Rollback**:
   ```bash
   ./rollback-nvim.sh
   # Select appropriate backup
   ```

## 📊 Version History

The scripts maintain compatibility across Neovim versions:
- **Current Support**: v0.10.x → v0.11.x
- **Future Ready**: v0.12.x, v0.13.x, etc.
- **Tested Upgrades**: Fedora 41 system packages to latest

## 🔗 Integration

### Adding to PATH
If needed, add to your shell config:
```bash
# In ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

### Alias Setup
For convenience:
```bash
alias nvim-update='~/.Dotfiles/nvim/nvim-update/nvim-update.sh'
alias nvim-test='~/.Dotfiles/nvim/nvim-update/test-nvim.sh'
alias nvim-rollback='~/.Dotfiles/nvim/nvim-update/rollback-nvim.sh'
```

## 📈 Future Updates

These scripts are designed for ongoing use:
- Automatic latest version detection
- Backward compatible backup format
- Extensible test framework
- Method preference preservation

Run `./nvim-update.sh` anytime to get the latest Neovim version!