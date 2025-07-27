#!/bin/bash

# Neovim Docker Test Environment
# Test latest Neovim with your configuration safely

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVIM_DIR="$(dirname "$SCRIPT_DIR")"
DOTFILES_DIR="$(dirname "$NVIM_DIR")"
IMAGE_NAME="nvim-test"
CONTAINER_NAME="nvim-test-container"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to build the Docker image
build_image() {
    info "Building Docker image with latest Neovim..."
    info "Using Neovim config from: $NVIM_DIR/.config/nvim"
    
    cd "$SCRIPT_DIR"
    
    docker build \
        --build-arg USER_ID=$(id -u) \
        --build-arg GROUP_ID=$(id -g) \
        --build-arg USERNAME=nvimuser \
        -t "$IMAGE_NAME" \
        -f Dockerfile \
        --build-context nvim="$NVIM_DIR" \
        .
    
    success "Docker image '$IMAGE_NAME' built successfully!"
}

# Function to run the container
run_container() {
    local workspace_dir="${1:-$(pwd)}"
    
    info "Starting Neovim test container..."
    info "Workspace: $workspace_dir"
    info "Neovim config: $NVIM_DIR/.config/nvim"
    
    # Remove existing container if it exists
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        warning "Removing existing container..."
        docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1
    fi
    
    # Run the container
    docker run -it --rm \
        --name "$CONTAINER_NAME" \
        -v "$workspace_dir:/workspace" \
        -v "$NVIM_DIR/.config/nvim:/home/nvimuser/.config/nvim" \
        "$IMAGE_NAME"
}

# Function to check health
health_check() {
    info "Running Neovim health check..."
    
    docker run --rm \
        -v "$NVIM_DIR/.config/nvim:/home/nvimuser/.config/nvim" \
        "$IMAGE_NAME" \
        nvim +checkhealth +qa
}

# Function to test specific functionality
test_functionality() {
    info "Testing Neovim functionality..."
    
    docker run --rm \
        -v "$NVIM_DIR/.config/nvim:/home/nvimuser/.config/nvim" \
        "$IMAGE_NAME" \
        nvim --headless -c "lua print('Neovim v0.11.3 test successful!')" -c "sleep 1" +qa
}

# Function to show help
show_help() {
    echo "Neovim Docker Test Environment"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  build                 Build the Docker image with latest Neovim"
    echo "  run [WORKSPACE_DIR]   Run interactive container (default: current directory)"
    echo "  health               Run Neovim health check"
    echo "  test                 Test basic Neovim functionality"
    echo "  clean                Remove Docker image"
    echo "  help                 Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                          # Build the image"
    echo "  $0 run                            # Run with current directory as workspace"
    echo "  $0 run ~/Dev/project             # Run with specific workspace"
    echo "  $0 health                        # Check Neovim health"
    echo "  $0 test                          # Test basic functionality"
    echo ""
    echo "Paths:"
    echo "  Neovim config: $NVIM_DIR/.config/nvim"
    echo "  Docker context: $SCRIPT_DIR"
}

# Function to clean up
clean_image() {
    warning "Removing Docker image '$IMAGE_NAME'..."
    
    # Stop and remove container if running
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1
    fi
    
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1
    fi
    
    # Remove image
    if docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        docker rmi "$IMAGE_NAME"
        success "Docker image removed successfully!"
    else
        warning "Image '$IMAGE_NAME' not found"
    fi
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    error "Docker is not installed or not in PATH"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    error "Docker daemon is not running"
    exit 1
fi

# Parse command line arguments
case "${1:-help}" in
    build)
        build_image
        ;;
    run)
        # Check if image exists
        if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
            warning "Image '$IMAGE_NAME' not found. Building..."
            build_image
        fi
        run_container "$2"
        ;;
    health)
        # Check if image exists
        if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
            error "Image '$IMAGE_NAME' not found. Run '$0 build' first."
            exit 1
        fi
        health_check
        ;;
    test)
        # Check if image exists
        if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
            error "Image '$IMAGE_NAME' not found. Run '$0 build' first."
            exit 1
        fi
        test_functionality
        ;;
    clean)
        clean_image
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac