# Setup

Do this once at the shop. On a Mac, the [Mac app](macos-app.md) does steps 1, 4 and 5 for you (it runs the server, picks the folder, and prints the QR poster). Steps 2 and 3 still apply.

## 1. Get the program

Use a binary from `dist/`, or build it (needs Go 1.24 or newer):

```
GOOS=darwin  GOARCH=arm64 go build -o dist/focuslab-drop-mac-arm64 .
GOOS=darwin  GOARCH=amd64 go build -o dist/focuslab-drop-mac-intel .
GOOS=windows GOARCH=amd64 go build -o dist/focuslab-drop.exe .
```

Pick the one for the counter computer: Apple Silicon Mac, Intel Mac, or Windows.

## 2. Give the counter computer a fixed address

The upload QR code contains the computer's IP address, so the address must never change.

1. Open the router admin page.
2. Find the DHCP or "address reservation" setting.
3. Reserve one address for the counter computer, for example `192.168.1.50`.
4. Reconnect the computer and check that it got that address.

## 3. Use a guest Wi-Fi network for customers

Create a guest network on the router so customers cannot see other devices. If the router can set a guest session or timeout, set it to 1 hour. This is the only real way to limit Wi-Fi time.

## 4. Run the program

macOS:

```
./focuslab-drop-mac-arm64
```

If macOS blocks it, right-click the file in Finder, choose Open, and confirm. Allow incoming connections when asked.

Windows: double-click `focuslab-drop.exe` and allow it through the firewall for private networks.

On start it prints:

```
saving photos to /Users/you/Prints
staff counter: http://localhost:6789/lab-counter
customer page: http://192.168.1.50:6789/drop
```

Options:

| Flag | Default | Meaning |
|---|---|---|
| `-dir` | `~/Prints` | Folder where photos are saved |
| `-port` | `6789` | Port to listen on |
| `-db` | in the user config folder | Staff login database |
| `-exit-on-stdin-close` | off | Exit when standard input closes. The Mac app uses this so a crashed app never leaves a server running. |

## 5. Print the QR codes

Make two QR codes with any QR generator and print them for the counter:

1. **Wi-Fi:** `WIFI:T:WPA;S:<guest network name>;P:<password>;;`
2. **Upload page:** `http://192.168.1.50:6789/drop` (use the address you reserved)

Test both with a phone before printing.

## 6. Optional: friendly name for staff

On the counter computer only, add this line to the hosts file so staff can open `http://focuslab:6789/lab-counter`:

- macOS: run `sudo nano /etc/hosts` and add `127.0.0.1 focuslab`
- Windows: open Notepad as administrator, open `C:\Windows\System32\drivers\etc\hosts`, and add `127.0.0.1 focuslab`

This changes names on that computer only. Phones ignore it, which is why the customer QR uses the IP address. To give phones a name too, set the Mac's local hostname to `focuslab` in System Settings, General, Sharing. Many phones can then open `http://focuslab.local:6789/drop`, but keep the IP address in the QR.

## Where files live

| What | macOS | Windows |
|---|---|---|
| Photos | `~/Prints` (or `-dir`) | `C:\Users\<you>\Prints` (or `-dir`) |
| Staff login database | `~/Library/Application Support/focuslab-drop/users.db` | `%AppData%\focuslab-drop\users.db` |
