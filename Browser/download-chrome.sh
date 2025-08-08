#!/bin/bash

# Download script for Chrome Linux standalone
# This file tracks the source of large binaries that are gitignored

CHROME_LINUX_URL="https://example.com/chrome-linux.zip"  # Replace with actual URL
CHROME_LINUX_MD5="YOUR_MD5_HERE"  # Replace with actual MD5

echo "Chrome Linux standalone files are not tracked in git due to size limits."
echo ""
echo "To restore chrome-linux.zip:"
echo "1. Download from your backup location or original source"
echo "2. Place in ~/.Dotfiles/Browser/"
echo "3. Extract with: unzip chrome-linux.zip"
echo ""
echo "File details:"
echo "- chrome-linux.zip: 147.28 MB"
echo "- chrome binary: 336.33 MB"
echo "- Date: February 2023"
echo ""
echo "The files are already present locally but gitignored."