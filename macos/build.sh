#!/usr/bin/env bash
# Builds "FocusLab Drop.app" and dist/FocusLabDrop.dmg (Apple Silicon + Intel, unsigned apart from an ad-hoc signature).
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT=$PWD
DIST=$ROOT/dist
WORK=$ROOT/build
APP="$DIST/FocusLab Drop.app"
DMG=$DIST/FocusLabDrop.dmg
GO=${GO:-$(command -v go || echo /opt/homebrew/bin/go)}

rm -rf "$WORK" "$APP" "$DMG"
mkdir -p "$WORK" "$DIST"

echo "==> Go server (arm64 + amd64)"
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 "$GO" build -o "$WORK/server-arm64" .
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 "$GO" build -o "$WORK/server-amd64" .
lipo -create -output "$WORK/focuslab-drop" "$WORK/server-arm64" "$WORK/server-amd64"

echo "==> Swift app (arm64 + x86_64)"
(cd macos && swift build -c release --arch arm64 --arch x86_64)
APP_BIN=$ROOT/macos/.build/apple/Products/Release/FocusLabDrop

echo "==> Icon"
ICONSET=$WORK/AppIcon.iconset
mkdir -p "$ICONSET"
for s in 16 32 128 256 512; do
  sips -z $s $s macos/Resources/logo.png --out "$ICONSET/icon_${s}x${s}.png" >/dev/null
  sips -z $((s * 2)) $((s * 2)) macos/Resources/logo.png --out "$ICONSET/icon_${s}x${s}@2x.png" >/dev/null
done
iconutil -c icns "$ICONSET" -o "$WORK/AppIcon.icns"

echo "==> Bundle"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$APP_BIN" "$APP/Contents/MacOS/FocusLabDrop"
cp "$WORK/focuslab-drop" "$APP/Contents/Resources/focuslab-drop"
cp macos/Resources/Info.plist "$APP/Contents/Info.plist"
cp macos/Resources/logo.png "$APP/Contents/Resources/logo.png"
cp "$WORK/AppIcon.icns" "$APP/Contents/Resources/AppIcon.icns"

echo "==> Ad-hoc signature (no Apple developer account, so macOS will ask the owner to approve it once)"
codesign --force --sign - "$APP/Contents/Resources/focuslab-drop"
codesign --force --sign - "$APP"
codesign --verify --strict "$APP"

echo "==> DMG"
STAGE=$WORK/dmg
mkdir -p "$STAGE"
cp -R "$APP" "$STAGE/"
ln -s /Applications "$STAGE/Applications"
hdiutil create -volname "FocusLab Drop" -srcfolder "$STAGE" -ov -format UDZO "$DMG" >/dev/null

echo
echo "Built: $APP"
echo "Built: $DMG ($(du -h "$DMG" | cut -f1))"
for f in "$APP/Contents/MacOS/FocusLabDrop" "$APP/Contents/Resources/focuslab-drop"; do echo "$(basename "$f"): $(lipo -archs "$f")"; done
