# FocusLab Drop

Customers on the shop Wi-Fi scan a QR code, pick photos in their phone browser, and the originals land in a folder on the counter computer (Mac or Windows). No internet needed, nothing to install on the phone.

- Customer page: `http://<counter-ip>:6789/drop`
- Staff counter (live list of uploads, Open folder button): `http://localhost:6789/lab-counter`, default login `admin` / `admin`, reachable only from the counter computer itself.

Full documentation is in [docs/](docs/README.md).

Files are saved as `Prints/<date>/<name>_<phone>_<HHMM>/<file>`. Name and phone are optional; with neither, the folder is `Guest_<HHMM>`.

## Run

Pick the binary for the counter computer (build below, or `go run .`):

```
./focuslab-drop-mac-arm64                      # Apple Silicon
./focuslab-drop-mac-intel                      # Intel Mac
focuslab-drop.exe                              # Windows
```

Options: `-dir ~/Prints` (where photos go), `-port 6789`, `-db <file>` (staff login database). On start it prints the customer URL for each network it is on.

- Windows: allow the firewall prompt for private networks the first time.
- macOS: if blocked, right-click the binary, Open. Allow incoming connections when asked.

## One-time shop setup

1. **Fixed address for the counter computer.** In the router admin, reserve a DHCP address for it (for example `192.168.1.50`). The QR code contains this address, so it must not change.
2. **Make the QR code once.** Encode `http://192.168.1.50:6789/drop` with any QR generator and print it. A second QR for the guest Wi-Fi (`WIFI:T:WPA;S:<name>;P:<password>;;`) goes next to it. Both are fixed; this app does not generate or change them.
3. **Optional friendly name on the counter computer.** Add this line to the hosts file so staff can use `http://focuslab:6789/lab-counter`:
   - macOS: `sudo nano /etc/hosts`, add `127.0.0.1 focuslab`
   - Windows (run Notepad as administrator): `C:\Windows\System32\drivers\etc\hosts`, add `127.0.0.1 focuslab`

   This only affects the counter computer. Phones do not read its hosts file, which is why the QR code uses the IP address. To get a name that works on phones too, set the Mac's local hostname to `focuslab` (System Settings, General, Sharing). Many phones then open `http://focuslab.local:6789/drop`.
4. Use the router's **guest network** for customers so they cannot see other devices.

## What protects the counter computer

- `/lab-counter` answers only requests from the computer itself (loopback) with a `localhost`, `127.0.0.1` or `focuslab` host name. From any phone it returns 404. Opening it through the computer's own LAN IP also returns 404, by design.
- The staff login is a second layer. There is no screen to change the default password yet (see [docs/operations.md](docs/operations.md)).
- Customers can only write. There is no file listing or download, and file and folder names are sanitized so nothing lands outside `-dir`.
- Only photos and PDF (`jpg jpeg png heic heif webp tif tiff pdf`) up to 50 MB each and 2 GB per visit. Uploads stop when the disk has under 1 GB free.
- A customer session lasts 1 hour. This is a soft limit: the 1 hour Wi-Fi limit itself has to come from the router (guest network timeout), and nothing here can disconnect a phone.

## Build and test

```
go vet ./... && go test ./...
GOOS=darwin  GOARCH=arm64 go build -o dist/focuslab-drop-mac-arm64 .
GOOS=darwin  GOARCH=amd64 go build -o dist/focuslab-drop-mac-intel .
GOOS=windows GOARCH=amd64 go build -o dist/focuslab-drop.exe .
```

Needs Go 1.24 or newer. Brand images in `web/assets/` come from the Focus Digital Color Lab site.
