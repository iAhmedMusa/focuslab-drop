# FocusLab Drop on Windows

Windows has a small tray app. It sits by the clock, shows a coloured dot, and runs the server in the background with no black window. The plain program is still available, see the end of this page.

> **Status:** the tray app is built and compiles for Windows, but it has not been tried on a real Windows PC yet. Follow the hand test in [testing.md](testing.md) before relying on it in the shop.

## What you get

- **Tray dot** next to the clock:
  - green: server on, customers can send photos
  - amber: starting
  - red: server off
  - red with `!`: a problem (hover over the dot or open the menu to read it)
- **Menu** (click the dot):
  - the status line
  - **Server on**: tick to turn the server on, untick to turn it off
  - the **customer link**, for example `http://192.168.1.50:6789/drop`. Click it to copy. If the PC has more than one network address, use **Choose network address** and pick the Wi-Fi one
  - **Open counter page**: the staff page with the live list of uploads (login `admin` / `admin`, see [security.md](security.md))
  - **Open photos folder**: opens `C:\Users\<you>\Prints` in Explorer
  - **Show server log**: opens the log in Notepad
  - **Launch at login**: starts the app when you sign in to Windows
  - **Quit**: stops the server and closes the app. It asks first if the server is on

Closing nothing is needed: there is no window to close. The server stops only when you untick **Server on**, choose **Quit**, or the PC shuts down.

Not in the Windows app yet: the QR poster, a folder picker and a port setting. It always saves to `C:\Users\<you>\Prints` and uses port 6789. To change either, use the plain program below.

## Install

1. Download `FocusLabDrop-windows.zip` from the latest release on GitHub, or build it (see below).
2. Unzip it into one folder, for example `C:\FocusLabDrop`. Keep `FocusLabDrop.exe` and `focuslab-drop.exe` **in the same folder**.
3. Double-click `FocusLabDrop.exe`.
4. If Windows SmartScreen says "Windows protected your PC", click **More info**, then **Run anyway**. The app is not signed.
5. When the firewall asks, tick **Private networks** and click **Allow access**. Without this, phones cannot reach the page.
6. Find the dot. Windows may hide new tray icons under the `^` arrow. Drag the dot out onto the taskbar so you can always see the status.
7. Right-click or click the dot, check the customer link, and make a QR code from it as in [setup.md](setup.md).

### If Windows blocks it completely (Smart App Control)

Some Windows 11 PCs, mostly ones set up from scratch, have **Smart App Control** on. It blocks unsigned programs and shows no **Run anyway** button. Check **Windows Security, App & browser control, Smart App Control settings**. If it is on, either sign the app (needs a paid code-signing certificate) or turn Smart App Control off. It cannot be turned back on without reinstalling Windows, so decide before you change it.

## Starting automatically

Tick **Launch at login**. From then on the app starts every time this Windows user signs in, turns the server on, and sits quietly in the tray. It writes one entry under `HKEY_CURRENT_USER\...\Run`. No administrator rights are needed.

If you switch it off in **Settings, Apps, Startup**, the tick in the menu clears too. Tick it again to switch it back on.

Three things must be true for this to work after a power cut or restart:

- **The PC must sign in by itself.** Startup apps run after sign-in, not at power-on.
  - Use a local Windows account for the shop user, or turn off **Require Windows Hello sign-in for Microsoft accounts** in **Settings, Accounts, Sign-in options**.
  - Press `Win + R`, type `netplwiz`, untick **Users must enter a user name and password to use this computer**, click **OK**, and type the password once.
  - This does not work with a PIN, face or fingerprint sign-in. With a Microsoft account and Windows Hello, the PC waits at the sign-in screen and the server stays off until someone signs in.
- **The PC must not sleep.** A sleeping PC stops serving uploads. In **Settings, System, Power**, set sleep to **Never** while plugged in.
- **The firewall must allow the program on Private networks**, and the Wi-Fi must be set to **Private**.

## Where things are stored

| What | Where |
|---|---|
| Photos | `C:\Users\<you>\Prints\<date>\<customer folder>` |
| Staff login (`admin` / `admin`) | `%AppData%\focuslab-drop\users.db` |
| Server log (replaced on each start) | `%LocalAppData%\FocusLab Drop\server.log`, or **Show server log** |
| Launch at login entry | `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`, value `FocusLabDrop` |

The app has no settings file. The only thing it remembers is the Launch at login entry.

## Problems

| Problem | Fix |
|---|---|
| Red dot with `!`, "Port 6789 is already in use" | Another copy is running, for example the plain program. Close it, then untick and tick **Server on** |
| Red dot with `!`, "focuslab-drop.exe is missing" | `focuslab-drop.exe` must sit next to `FocusLabDrop.exe`. Unzip both into one folder |
| Red dot with `!`, "The server stopped by itself" | Open **Show server log**. Usually the Prints folder cannot be written |
| Cannot see the dot | Click the `^` arrow by the clock and drag the dot out |
| Phones cannot open the customer page | Firewall rule missing: in Windows Security, Firewall, allow `focuslab-drop.exe` on Private networks. Also check the Wi-Fi network is set to Private |
| Page opens on the PC but not on phones | The phone is on a different network, or the QR uses the wrong IP address. Pick the right one under **Choose network address**, and reserve the PC's IP on the router, see [setup.md](setup.md) |
| "No network address found" in the menu | The PC is not on Wi-Fi or Ethernet. Connect it |
| Server was off after a restart | The PC did not sign in by itself, or **Launch at login** is off. See Starting automatically |
| Windows says it is blocked and offers no **Run anyway** | Smart App Control, see above |
| Upload fails for a file named `NUL.jpg` or similar | Windows reserved name. See [security.md](security.md) |

## Update or remove

- **Update:** choose **Quit**, replace both `.exe` files with the new ones, and start `FocusLabDrop.exe` again. Photos and the staff login are not touched.
- **Remove:** untick **Launch at login**, choose **Quit**, and delete the folder. Photos in `Prints` stay until you delete them.

## Build it yourself

On any computer with Go 1.24 or newer and `zip`:

```
./tray/build.sh
```

This writes `dist/FocusLabDrop-windows.zip` with `FocusLabDrop.exe` (the tray app, no console window) and `focuslab-drop.exe` (the server).

## The plain program

The server also runs on its own, without the tray app. Use it for another folder or port, or to debug.

1. Double-click `focuslab-drop.exe`. A black window opens and stays open. **Closing it stops the server.**
2. The window prints the addresses:

```
saving photos to C:\Users\you\Prints
staff counter: http://localhost:6789/lab-counter
customer page: http://192.168.1.50:6789/drop
```

3. To change defaults, run it from Command Prompt, for example `focuslab-drop.exe -dir D:\Prints -port 6789`. The flags are in [setup.md](setup.md).

Do not run the plain program and the tray app at the same time. They use the same port.
