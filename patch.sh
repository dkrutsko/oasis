#!/bin/bash
# Vendors dependencies and applies the gogpu macOS key
# beep patch. Run this after go get or any go.mod change.
#
# Usage: ./tools/vendor_patch.sh

set -e

echo "Vendoring dependencies..."
go mod vendor

FILE="vendor/github.com/gogpu/gogpu/internal/platform/platform_darwin.go"

if [ ! -f "$FILE" ]; then
    echo "Warning: $FILE not found, skipping patch"
    exit 0
fi

# Skip if already patched
if grep -q "return false // suppress NSBeep" "$FILE"; then
    echo "Patch already applied."
    exit 0
fi

echo "Applying macOS key beep patch..."

# Replace the keyDown and keyUp blocks to add return false,
# preventing sendEvent: from forwarding key events to the
# NSView responder chain which triggers NSBeep.
sed -i '' 's/w\.dispatchCharFromEvent(event)$/w.dispatchCharFromEvent(event)\
		return false \/\/ suppress NSBeep/' "$FILE"

sed -i '' 's/w\.dispatchKeyEvent(key, w\.modifiers, false)$/w.dispatchKeyEvent(key, w.modifiers, false)\
		return false \/\/ suppress NSBeep/' "$FILE"

echo "Patch applied."
