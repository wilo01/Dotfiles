#!/bin/bash

# Neovim Test Script
# Comprehensive testing of Neovim installation and configuration
# Author: DarO
# Usage: ./test-nvim.sh [--verbose] [--quick]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVIM_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Test flags
VERBOSE=false
QUICK=false

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

step() {
    echo -e "${PURPLE}▶${NC} $1"
}

verbose() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}  →${NC} $1"
    fi
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --verbose|-v)
                VERBOSE=true
                shift
                ;;
            --quick|-q)
                QUICK=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    echo "Neovim Test Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --verbose, -v    Show detailed output for each test"
    echo "  --quick, -q      Run only basic tests (faster)"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Test Categories:"
    echo "  1. Installation verification"
    echo "  2. Configuration loading"
    echo "  3. Plugin functionality"
    echo "  4. LSP integration"
    echo "  5. Performance checks"
}

# Test basic Neovim installation
test_installation() {
    step "Testing Neovim installation..."
    
    # Check if nvim command exists
    if ! command -v nvim &> /dev/null; then
        error "Neovim not found in PATH"
        return 1
    fi
    success "Neovim found in PATH"
    
    # Check version
    local version_output
    if ! version_output=$(nvim --version 2>&1); then
        error "Failed to get Neovim version"
        return 1
    fi
    
    local version=$(echo "$version_output" | head -1 | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+')
    success "Neovim version: v$version"
    verbose "Installation path: $(which nvim)"
    
    return 0
}

# Test configuration loading
test_configuration() {
    step "Testing configuration loading..."
    
    # Test basic startup
    if ! timeout 10s nvim --headless -c "quit" 2>/dev/null; then
        error "Neovim failed to start with configuration"
        return 1
    fi
    success "Configuration loads without errors"
    
    # Test Lua configuration
    if ! timeout 10s nvim --headless -c "lua print('Config test')" -c "quit" 2>/dev/null; then
        error "Lua configuration test failed"
        return 1
    fi
    success "Lua configuration working"
    
    return 0
}

# Test plugin manager and plugins
test_plugins() {
    step "Testing plugin functionality..."
    
    if [[ "$QUICK" == true ]]; then
        verbose "Skipping detailed plugin tests (quick mode)"
        return 0
    fi
    
    # Test Lazy.nvim plugin manager
    local lazy_test_output
    if ! lazy_test_output=$(timeout 30s nvim --headless -c "lua require('lazy').setup()" -c "quit" 2>&1); then
        warning "Plugin manager test had issues"
        verbose "Output: $lazy_test_output"
    else
        success "Plugin manager (Lazy.nvim) working"
    fi
    
    # Test treesitter
    if timeout 15s nvim --headless -c "lua vim.treesitter.get_parser('lua')" -c "quit" 2>/dev/null; then
        success "Treesitter functionality working"
    else
        warning "Treesitter test failed"
    fi
    
    return 0
}

# Test LSP functionality
test_lsp() {
    step "Testing LSP functionality..."
    
    if [[ "$QUICK" == true ]]; then
        verbose "Skipping LSP tests (quick mode)"
        return 0
    fi
    
    # Test LSP client availability
    if timeout 10s nvim --headless -c "lua print(vim.lsp)" -c "quit" 2>/dev/null; then
        success "LSP client available"
    else
        warning "LSP client test failed"
    fi
    
    return 0
}

# Test performance
test_performance() {
    step "Testing performance..."
    
    # Test startup time
    local startup_time
    startup_time=$(timeout 10s nvim --headless --startuptime /tmp/nvim_startup.log -c "quit" 2>/dev/null && \
                   tail -1 /tmp/nvim_startup.log | awk '{print $1}' 2>/dev/null || echo "unknown")
    
    if [[ "$startup_time" != "unknown" ]]; then
        # Convert to milliseconds if needed
        if [[ "$startup_time" =~ ^[0-9]+\.[0-9]+$ ]]; then
            local startup_ms=$(echo "$startup_time * 1000" | bc -l 2>/dev/null || echo "$startup_time")
            startup_ms=${startup_ms%.*}  # Remove decimal part
            success "Startup time: ${startup_ms}ms"
            
            if [[ "$startup_ms" -gt 1000 ]]; then
                warning "Startup time is quite slow (>${startup_ms}ms)"
            fi
        else
            success "Startup time: ${startup_time}ms"
        fi
    else
        warning "Could not measure startup time"
    fi
    
    # Cleanup
    rm -f /tmp/nvim_startup.log
    
    return 0
}

# Test file operations
test_file_operations() {
    step "Testing file operations..."
    
    local test_file="/tmp/nvim_test_file.txt"
    local test_content="Hello from Neovim test!"
    
    # Test file creation and editing
    if echo "$test_content" | timeout 10s nvim --headless -c "put" -c "write $test_file" -c "quit" 2>/dev/null; then
        if [[ -f "$test_file" ]] && grep -q "$test_content" "$test_file"; then
            success "File operations working"
        else
            warning "File write test failed"
        fi
    else
        warning "File operation test failed"
    fi
    
    # Cleanup
    rm -f "$test_file"
    
    return 0
}

# Test health check
test_health() {
    step "Running Neovim health check..."
    
    local health_output
    if health_output=$(timeout 30s nvim --headless -c "checkhealth" -c "quit" 2>&1); then
        verbose "Health check completed"
        
        # Check for critical errors
        if echo "$health_output" | grep -qi "error\|fail"; then
            warning "Health check found some issues"
            if [[ "$VERBOSE" == true ]]; then
                echo "$health_output" | grep -i "error\|fail" | head -5
            fi
        else
            success "Health check passed"
        fi
    else
        warning "Health check failed to complete"
    fi
    
    return 0
}

# Run comprehensive test suite
run_tests() {
    local failed_tests=0
    local total_tests=0
    
    echo "=== Neovim Test Suite ==="
    echo ""
    
    # Basic installation test
    ((total_tests++))
    if ! test_installation; then
        ((failed_tests++))
    fi
    echo ""
    
    # Configuration test
    ((total_tests++))
    if ! test_configuration; then
        ((failed_tests++))
    fi
    echo ""
    
    # Plugin tests
    ((total_tests++))
    if ! test_plugins; then
        ((failed_tests++))
    fi
    echo ""
    
    # File operations test
    ((total_tests++))
    if ! test_file_operations; then
        ((failed_tests++))
    fi
    echo ""
    
    # LSP test (unless quick mode)
    if [[ "$QUICK" == false ]]; then
        ((total_tests++))
        if ! test_lsp; then
            ((failed_tests++))
        fi
        echo ""
    fi
    
    # Performance test
    ((total_tests++))
    if ! test_performance; then
        ((failed_tests++))
    fi
    echo ""
    
    # Health check
    ((total_tests++))
    if ! test_health; then
        ((failed_tests++))
    fi
    echo ""
    
    # Summary
    echo "=== Test Results ==="
    local passed_tests=$((total_tests - failed_tests))
    
    if [[ $failed_tests -eq 0 ]]; then
        success "All tests passed! ($passed_tests/$total_tests)"
        return 0
    else
        warning "Some tests failed ($passed_tests/$total_tests passed)"
        return 1
    fi
}

# Main test function
main() {
    parse_args "$@"
    
    if [[ "$QUICK" == true ]]; then
        info "Running quick test suite..."
    elif [[ "$VERBOSE" == true ]]; then
        info "Running verbose test suite..."
    else
        info "Running standard test suite..."
    fi
    
    echo ""
    
    if run_tests; then
        echo ""
        success "Neovim installation is working correctly!"
        return 0
    else
        echo ""
        error "Some tests failed. Check the output above for details."
        return 1
    fi
}

# Run main function with all arguments
main "$@"