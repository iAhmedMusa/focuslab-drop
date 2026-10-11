#!/usr/bin/env bash
# Builds dist/FocusLabDrop-windows.zip: the tray app plus the server, both unsigned. Runs on any OS with Go and zip.
set -euo pipefail

cd "$(dirname "$0")/.."
OUT=dist/windows
rm -rf "$OUT" dist/FocusLabDrop-windows.zip
mkdir -p "$OUT"

export GOOS=windows GOARCH=amd64 CGO_ENABLED=0
go build -ldflags "-s -w" -o "$OUT/focuslab-drop.exe" .
go build -ldflags "-s -w -H=windowsgui" -o "$OUT/FocusLabDrop.exe" ./tray

(cd "$OUT" && zip -q ../FocusLabDrop-windows.zip FocusLabDrop.exe focuslab-drop.exe)
echo "Built dist/FocusLabDrop-windows.zip"
