# Windows tray app: plan

Status: plan only. Nothing in this document is built yet. [windows.md](windows.md) still describes the plain program.

## Goal

A small Windows app that lives in the system tray and does what the Mac menu-bar app does for daily use, without a console window.

The menu has exactly these items:

| Item | What it does |
|---|---|
| Status line | Shows "Server on", "Starting...", "Server off", or the problem text |
| Server on/off | Starts or stops the server (checkable item) |
| Customer link | Shows the link, for example `http://192.168.1.50:6789/drop`. Click copies it. If there are several network addresses, a submenu lets you pick the Wi-Fi one |
| Open counter page | Opens `http://localhost:6789/lab-counter` in the browser |
| Open photos folder | Opens the print folder in Explorer |
| Show server log | Opens `server.log` |
| Launch at login | Checkable item, adds or removes the startup entry |
| Quit | Stops the server and exits |

The tray icon is a coloured dot like the Mac app: green on, amber starting, red off, red with `!` for a problem.

## Out of scope for this version

- QR poster and QR code generation (the printed QR stays manual, see [setup.md](setup.md))
- Folder picker and port setting. The tray app uses the server defaults: `%USERPROFILE%\Prints` and port 6789. Anyone who needs other values can still run the plain program with `-dir` and `-port`
- A native window with the big round power button
- Code signing, an installer, auto-update
- Changing the staff password

## Design

Same idea as the Mac app: the tray program does not contain the server. It starts the existing `focuslab-drop.exe` as a hidden child process and watches it. `main.go` does not change.

```
FocusLabDrop.exe  (tray, no console window)
   |  starts hidden, with -exit-on-stdin-close, stdin pipe held open
   v
focuslab-drop.exe (the existing server)
```

- **Crash safety:** the server already exits when its stdin closes (`-exit-on-stdin-close`). If the tray program dies, the pipe closes and the server stops. No leftover server.
- **Start check:** after launch, poll `http://127.0.0.1:<port>/drop` every 250 ms for up to 10 s. A 200 means running. Same as `ServerController.swift`.
- **Port in use:** try to listen on the port first. If taken, show "Port 6789 is already in use. Another copy may be running." as the problem text.
- **Server exits by itself:** show the last log line in the status line and turn the icon red with `!`.
- **Log:** `%LocalAppData%\FocusLab Drop\server.log`, replaced on each start.
- **Launch at login:** one value under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` named `FocusLabDrop`, holding the quoted path to `FocusLabDrop.exe`. No admin rights and no approval step. The checkmark reflects the registry value, and also turns off if the user disabled it in Settings, Apps, Startup (`StartupApproved\Run`, first byte `03`). It is reread every 5 s.
- **Server starts when the app opens:** always. Turning it off is the power item. The app remembers nothing between runs except the login entry, so there is no settings file.
- **One copy only:** a named mutex. A second launch exits silently.
- **Copy link:** `clip.exe` with the link on stdin. **Open counter, folder, log:** `rundll32 url.dll,FileProtocolHandler` or `explorer.exe`.
- **Tray library:** `fyne.io/systray` (maintained fork of `getlantern/systray`). It needs no C compiler on Windows, so it cross-compiles from the Mac. Phase 0 confirms this on a real Windows PC before anything else is built.

## Files

New:

| File | Purpose |
|---|---|
| `tray/core/core.go` | Pure logic, no Windows calls: port check, customer/counter URL, private IPv4 list, state to status text. Builds and tests on any OS |
| `tray/core/core_test.go` | Unit tests for the above |
| `tray/main_windows.go` | Entry point, tray menu, icon switching, quit |
| `tray/server_windows.go` | Start, stop, watch the child server, write the log |
| `tray/autostart_windows.go` | Read and write the Run and StartupApproved registry values |
| `tray/single_windows.go` | Single-instance mutex |
| `tray/open_windows.go` | Copy to clipboard, open URL, folder and log |
| `tray/icons/*.ico` | Four tray icons (green, amber, red, red with `!`), embedded with `go:embed` |
| `tray/build.sh` | Cross-builds both exes and makes `dist/FocusLabDrop-windows.zip` |

Changed:

| File | Change |
|---|---|
| `go.mod`, `go.sum` | Add `fyne.io/systray` and `golang.org/x/sys` (already indirect) |
| `docs/windows.md` | Rewrite for the tray app once it exists |
| `docs/README.md` | Add this plan, later the user guide |
| `docs/architecture.md`, `docs/testing.md`, `docs/security.md`, `README.md` | Short Windows tray additions at the end |

`main.go` and the `web/` folder do not change.

## Build and ship

```
GOOS=windows GOARCH=amd64 go build -o dist/windows/focuslab-drop.exe .
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui -s -w" -o dist/windows/FocusLabDrop.exe ./tray
```

`-H=windowsgui` removes the console window. The zip holds both exes and must be extracted to one folder, for example `C:\FocusLabDrop`. The tray program looks for `focuslab-drop.exe` next to itself.

## Steps

0. **Spike on a real Windows PC.** A 30-line tray program with a colour-changing icon and a checkable item, cross-built from the Mac. Confirms the library, the icon format, and that left and right click both open the menu. Stop and re-plan here if it fails.
1. **Core and tests.** `tray/core` plus tests, run on the Mac with `go test ./tray/core`.
2. **Server controller.** Start, stop, watch, log, port check.
3. **Tray menu and icons.** The status line, power item, link submenu, open items, quit.
4. **Launch at login.** Registry value, Startup-disabled detection.
5. **Build script and zip.**
6. **Hand test on Windows** (checklist below).
7. **Docs.** Replace `windows.md` with the tray guide, update the other docs.
8. **Release** as a new version with the zip attached.

## Verification

Automated, on the Mac: `go vet ./... && go test ./...` and `GOOS=windows go vet ./tray/...`. This proves the logic and that the Windows code compiles. It does not prove it runs.

By hand, on a Windows 10 or 11 PC:

1. Unzip, double-click `FocusLabDrop.exe`. No console window. A tray icon appears (it may be under the `^` arrow).
2. Icon turns amber, then green. The customer link opens from a phone on the same Wi-Fi.
3. Power item off: icon red, phone page stops. On again: green.
4. Copy link, paste into a browser. Open counter, folder and log all open.
5. Start the plain `focuslab-drop.exe` first, then the tray app: red with `!` and the "port in use" text.
6. Kill `FocusLabDrop.exe` in Task Manager: `focuslab-drop.exe` disappears too.
7. Launch at login on, sign out and in: the app starts and the server is green. Turn it off in Settings, Apps, Startup: the checkmark clears.
8. Start it twice: only one tray icon.
9. Quit: both processes gone.

## Risks

| Risk | Effect | What we do |
|---|---|---|
| **Smart App Control** (Windows 11 clean installs) | Blocks unsigned programs with no "Run anyway" button | Check the PC before install. Options: sign the exes, or turn the setting off (cannot be turned back on without reinstalling Windows). Document it |
| **SmartScreen** | "Windows protected your PC" on first run | Same as today: **More info**, **Run anyway**. Document it |
| **Antivirus false positive** | Unsigned Go program flagged | Allow-list the folder. Signing reduces it |
| **Firewall prompt** | Phones cannot connect if it is not allowed | Tick **Private networks** on first run. Document it, same as today |
| **Auto-login with a Microsoft account** | After a power cut the PC waits at the sign-in screen, so Launch at login never runs | Use a local account for the shop user, or turn off "Require Windows Hello sign-in for Microsoft accounts" (Settings, Accounts, Sign-in options), then set auto-login with `netplwiz`. Does not work with a PIN. Document it |
| **Sleep** | A sleeping PC serves nothing | Set sleep to Never while plugged in. The app does not keep the PC awake |
| **Tray icon hidden** | Owner cannot see the status light | Drag the icon out of the `^` overflow once. Document it with a screenshot |
| **Library or icon problem** | Menu or icon does not behave on Windows | Phase 0 spike catches it before the rest is built |
| **Cannot test here** | Windows code is only compiled on the Mac | All hand tests run on a Windows PC or VM before release |

The app needs no Microsoft account and no internet. Nothing is sent anywhere.

## Open questions

- Is a Windows PC or VM available for the hand tests? Without one, step 6 cannot happen and the release stays a draft.
- Which Windows version does the shop PC run, and is it on a Microsoft or a local account? This decides how much of the auto-login section matters.
- Should a later version add the folder picker, port setting and QR poster? They are listed under out of scope on purpose.
