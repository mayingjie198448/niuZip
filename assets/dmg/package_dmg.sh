#!/usr/bin/env bash
set -e

VOL_NAME="niuZip"
DMG_FINAL="build/niuZip-arm64.dmg"
DMG_TEMP="build/pack.temp.dmg"

echo "==> Preparing DMG staging..."
rm -f "${DMG_FINAL}" "${DMG_TEMP}"

# Create temporary writable DMG
hdiutil create -size 100m -volname "${VOL_NAME}" -fs HFS+ -ov "${DMG_TEMP}"

# Attach temporary DMG
DEVICE=$(hdiutil attach -readwrite -noverify -noautoopen "${DMG_TEMP}" | grep -E '^/dev/' | head -n 1 | awk '{print $1}')
echo "==> Attached DMG device: ${DEVICE}"

# Wait for volume to appear
sleep 1
VOL_PATH="/Volumes/${VOL_NAME}"

# Populate DMG
mkdir -p "${VOL_PATH}/.background"
cp assets/dmg/background.png "${VOL_PATH}/.background/background.png"
cp -R build/niuZip.app "${VOL_PATH}/niuZip.app"
ln -s /Applications "${VOL_PATH}/Applications"

# Set volume icon if appicon exists
if [ -f assets/icons/appicon.icns ]; then
    cp assets/icons/appicon.icns "${VOL_PATH}/.VolumeIcon.icns"
    SetFile -a C "${VOL_PATH}" 2>/dev/null || true
fi

# AppleScript to format Finder layout
echo "==> Customizing Finder window layout via AppleScript..."
osascript <<EOF || true
tell application "Finder"
    tell disk "${VOL_NAME}"
        open
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set the bounds of container window to {300, 200, 860, 560}
        set theViewOptions to the icon view options of container window
        set arrangement of theViewOptions to not arranged
        set icon size of theViewOptions to 96
        try
            set background picture of theViewOptions to file ".background:background.png"
        end try
        set position of item "niuZip.app" of container window to {140, 180}
        set position of item "Applications" of container window to {420, 180}
        update without registering applications
        delay 1
        close
    end tell
end tell
EOF

# Sync & Detach
sync
echo "==> Detaching temporary DMG..."
hdiutil detach "${DEVICE}" -force

# Convert to read-only compressed DMG
echo "==> Converting to final compressed DMG..."
hdiutil convert "${DMG_TEMP}" -format UDZO -imagekey zlib-level=9 -ov -o "${DMG_FINAL}"
rm -f "${DMG_TEMP}"

echo "==> DMG successfully created at ${DMG_FINAL}!"
ls -lh "${DMG_FINAL}"
