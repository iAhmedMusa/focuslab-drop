# Mac app

FocusLab Drop for Mac wraps the server in a normal app: one round power button to turn it on and off, a status light, a folder picker, a printable QR poster, and an option to start by itself when the Mac starts. Windows has its own tray app, see [windows.md](windows.md).

## What you get

- **Menu-bar dot**, always visible at the top of the screen:
  - green: server on, customers can send photos
  - amber: starting
  - red: server off
  - red with `!`: a problem (the window says what)
- **Small window** (open it from the menu or by opening the app again):
  - the big light and the round power button (green when on, gray when off)
  - **Save photos to**: choose the folder, or show it in Finder
  - **Customer link**: the address customers use, with a Copy button. If the Mac has more than one network address, pick the Wi-Fi one.
  - **Print QR poster**: an A4 page with the upload QR code, ready to print or save as PDF
  - **Open counter page**: the staff page with the live list of uploads
  - **Launch at login** and **Start the server when this app opens**
  - **Advanced**: the port (default 6789)
- **Menu** with the same actions, plus Quit.

The app has no Dock icon. Closing the window does not stop the server. Quit from the menu to stop it, or press the power button to turn it off.

## Install

1. Open `FocusLabDrop.dmg` and drag **FocusLab Drop** onto **Applications**.
2. Open it from Applications.

### First open (the app is not signed by Apple)

The app is not signed with an Apple developer account, so macOS blocks it the first time. Do this once:

- **macOS 15 or newer:** double-click the app, close the warning, open **System Settings, Privacy & Security**, scroll to Security, and click **Open Anyway** next to FocusLab Drop. Confirm with your password.
- **macOS 13 or 14:** right-click the app, choose **Open**, then **Open** again.

Then allow the network prompts:

- "Do you want the application focuslab-drop to accept incoming network connections?" Click **Allow**.
- If macOS asks about the local network, click **Allow**. Phones cannot reach the server without it.

## Set it up

1. Open the window and check **Save photos to**. The default is `Prints` in your home folder. Click **Choose...** to use another folder. If the server is on, it restarts so uploads go to the new folder.
2. Check the **Customer link**. Reserve that address for this Mac in your router (see [setup.md](setup.md)) so the QR code never goes stale.
3. Click **Print QR poster...**, then **Print...** or **Save as PDF...**. Put the printed poster at the counter.
4. Turn on **Launch at login**. If macOS asks, approve FocusLab Drop under **System Settings, General, Login Items**.
5. Keep **Start the server when this app opens** on, so the server turns on by itself.

Do not change the port after printing the poster. A new port means a new QR code.

## Starting automatically

With **Launch at login** on, the app starts every time this Mac user logs in, starts the server, and stays quietly in the menu bar (no window).

Two things must be true for this to work after a power cut or restart:

- **The Mac must log in by itself.** Login items run after login, not at power-on. Open **System Settings, Users & Groups** and set **Automatically log in as** to the shop user. This option is not available when FileVault disk encryption is on. With FileVault on, someone must type the password after every restart.
- **The Mac must not fall asleep.** A sleeping Mac stops serving uploads. In **System Settings, Energy** (or Battery), prevent automatic sleeping while the display is off. The app does not keep the Mac awake for you.

## Where things are stored

| What | Where |
|---|---|
| Photos | The folder you chose |
| App settings (folder, port, options) | macOS preferences for `com.focuslab.drop` |
| Staff login (`admin` / `admin`) | `~/Library/Application Support/focuslab-drop/users.db` |
| Server log (replaced on each start) | `~/Library/Logs/FocusLab Drop/server.log`, or **Show server log** in the window |

## Problems

| Problem | Fix |
|---|---|
| Red light, "Port 6789 is already in use" | Another copy is running (for example one started from Terminal). Quit it, or change the port under Advanced. |
| Red light, "The server stopped by itself" | Open **Show server log**. Usually the folder cannot be written. Choose another folder. |
| Phones cannot open the customer link | Join the same Wi-Fi as the Mac, allow the network prompts above, and check the router has no client isolation. See [operations.md](operations.md). |
| "No network address found" | The Mac is not on Wi-Fi or Ethernet. Connect it. |
| Server was off after a restart | The Mac did not log in by itself, or Launch at login is off. See Starting automatically. |
| Launch at login says it is waiting | Approve FocusLab Drop under System Settings, General, Login Items. |

## Update or remove

- **Update:** quit the app, drag the new version onto Applications and replace the old one. Settings and photos are kept.
- **Remove:** turn off **Launch at login**, quit the app, and drag it from Applications to the Trash. Your photos stay in their folder.

## Build it yourself

You need Xcode and Go.

```
./macos/build.sh
```

This builds the server and the app for Apple Silicon and Intel, signs them with a local ad-hoc signature, and writes `dist/FocusLab Drop.app` and `dist/FocusLabDrop.dmg`.

Run the Swift tests with `cd macos && swift test`, and the server tests with `go test ./...`.

## How it fits together

- The app (Swift, in `macos/`) contains the same Go server as the command-line version, inside the app bundle.
- Pressing the power button when the server is off starts the server as a background process with the chosen folder and port, and waits until it really answers before showing green.
- The server is started with `-exit-on-stdin-close`. If the app crashes or is force-quit, the server notices and stops by itself, so it never keeps the port busy.
- Quitting the app, or pressing the power button when it is on, stops the server.

## Known limits

- Not signed by Apple, so the first-open steps above are needed on every Mac.
- No automatic updates.
- Built for Apple Silicon and Intel, but only run-tested on Apple Silicon so far.
- The Windows tray app is simpler: no folder picker, port setting or QR poster yet. See [windows.md](windows.md).
