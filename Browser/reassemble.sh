#!/bin/bash
# Reassemble split Chrome files

echo "Reassembling Chrome files..."

# Reassemble chrome-linux.zip if needed
if [ ! -f "chrome-linux.zip" ] && [ -f "chrome-linux.zip.part-aa" ]; then
    echo "Reassembling chrome-linux.zip..."
    cat chrome-linux.zip.part-* > chrome-linux.zip
    echo "✓ chrome-linux.zip reassembled"
fi

# Reassemble chrome binary if needed
if [ ! -f "chrome-linux/chrome" ] && [ -f "chrome-linux/chrome.part-aa" ]; then
    echo "Reassembling chrome binary..."
    cat chrome-linux/chrome.part-* > chrome-linux/chrome
    chmod +x chrome-linux/chrome
    echo "✓ chrome binary reassembled and made executable"
fi

echo "Done! Chrome files are ready to use."