# FocusLab Drop documentation

FocusLab Drop lets print-shop customers send photos from their phone to the counter computer over the shop Wi-Fi. No internet, no app, no WhatsApp downloads.

| Doc | Read it to |
|---|---|
| [overview.md](overview.md) | Understand what the app does and the decisions behind it |
| [macos-app.md](macos-app.md) | Install and use the Mac app (power button, folder, QR poster, auto start) |
| [windows.md](windows.md) | Install and use the Windows tray app (or the plain program) |
| [windows-tray-plan.md](windows-tray-plan.md) | See the plan, limits and risks behind the Windows tray app |
| [setup.md](setup.md) | Set up the counter computer, router and QR codes |
| [operations.md](operations.md) | Run it day to day and fix common problems |
| [architecture.md](architecture.md) | See how the code, routes and folders fit together |
| [security.md](security.md) | Know what is protected, and the limits |
| [testing.md](testing.md) | Test it by hand and run the automated tests |

Quick start on a Mac: open `FocusLabDrop.dmg`, drag the app to Applications, and follow [macos-app.md](macos-app.md).

Quick start from the command line: build or pick a binary from `dist/`, run it on the counter computer, open `http://localhost:6789/lab-counter` (login `admin` / `admin`), and send a photo from a phone at the URL printed in the terminal.
