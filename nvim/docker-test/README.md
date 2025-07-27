# Neovim Docker Test Environment

A reusable Docker environment for testing the latest Neovim with your configuration without affecting your host installation.

## Features

- **Fedora 41** base image (matches your system)
- **Latest Neovim** (v0.11.3) from official GitHub releases
- **Your exact configuration** from `../.config/nvim/`
- **Isolated environment** - no impact on host Neovim installation
- **File editing support** - mount any directory as workspace
- **Health checks** - validate plugin compatibility

## Quick Start

```bash
# Build the Docker image
./nvim/docker-test/nvim-test.sh build

# Run interactive container with current directory as workspace
./nvim/docker-test/nvim-test.sh run

# Run with specific workspace directory
./nvim/docker-test/nvim-test.sh run /path/to/your/project

# Check Neovim health
./nvim/docker-test/nvim-test.sh health

# Test basic functionality
./nvim/docker-test/nvim-test.sh test

# Clean up (remove Docker image)
./nvim/docker-test/nvim-test.sh clean
```

## What's Included

### System Dependencies
- git, curl, wget
- Build tools (gcc, make, cmake, ninja)
- Node.js and npm (for plugins requiring Node)
- ripgrep and fd-find (for telescope and other search plugins)
- Python3 (for Python-based plugins)

### Neovim Configuration
- Your complete config from `nvim/.config/nvim/`
- All plugins via Lazy.nvim
- Preserved lazy-lock.json for consistent plugin versions
- Custom keymaps and settings

## Usage Examples

### Testing Plugin Compatibility
```bash
# Start container
./docker/docker-nvim.sh run

# Inside container, check plugin health
nvim +checkhealth +qa
```

### Editing Files
```bash
# Mount project directory
./docker/docker-nvim.sh run ~/Dev/my-project

# Inside container, edit files
nvim src/main.lua
```

### Testing New Configuration
```bash
# Modify your nvim config locally
# Run container to test changes
./docker/docker-nvim.sh run
```

## Container Details

- **User**: `nvimuser` (matches your host UID/GID)
- **Workspace**: `/workspace` (mounted from host)
- **Config**: `/home/nvimuser/.config/nvim` (read-only mount)
- **Neovim**: Latest version from GitHub releases

## Benefits

✅ **Safe Testing**: Test latest Neovim without breaking your setup  
✅ **Plugin Validation**: Check plugin compatibility before host upgrade  
✅ **Reproducible**: Consistent environment across different machines  
✅ **Isolated**: No interference with existing Neovim installation  
✅ **Easy Cleanup**: Remove container and image when done  

## Troubleshooting

### Permission Issues
The container user is created with your host UID/GID to avoid permission problems with mounted files.

### Plugin Installation
Lazy.nvim will automatically install plugins on first run. This happens inside the container and doesn't affect your host.

### Performance
For better performance with large projects, consider using Docker volumes instead of bind mounts for frequently accessed files.

## Customization

Edit `docker/Dockerfile` to:
- Add additional system packages
- Change base image
- Modify user setup
- Add environment variables

Edit `docker/docker-nvim.sh` to:
- Change default workspace directory
- Add custom mount points
- Modify container runtime options