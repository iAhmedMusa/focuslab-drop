# FocusLab Drop on Windows

Windows runs the same server as the Mac, as one program with no installer. There is no tray app yet, so you start it by hand and keep its window open.

## Get the program

Download `focuslab-drop.exe` from the latest release on GitHub (Releases page), or build it on any computer with Go 1.24 or newer:

```
GOOS=windows GOARCH=amd64 go build -o dist/focuslab-drop.exe .
```

## First run

1. Put `focuslab-drop.exe` in a folder, for example `C:\FocusLabDrop`.
2. Double-click it. A black window opens and stays open. Do not close it while the shop is open.
3. If Windows SmartScreen says "Windows protected your PC", click **More info**, then **Run anyway**. The program is not signed.
4. When the firewall asks, tick **Private networks** and click **Allow access**. Without this, phones cannot reach the page.
5. The window prints the addresses:

```
saving photos to C:\Users\you\Prints
staff counter: http://localhost:6789/lab-counter
customer page: http://192.168.1.50:6789/drop
```

Open the staff counter in a browser on this computer. Default login is `admin` / `admin`. There is no screen to change it; the login only works on this computer, see [security.md](security.md).

## Every day

- Start: double-click `focuslab-drop.exe`.
- Stop: close the black window.
- Photos: `C:\Users\<you>\Prints\<date>\<customer folder>`. **Open folder** on the counter page opens it in Explorer.

## Start with Windows (optional)

1. Press `Win + R`, type `shell:startup`, press Enter.
2. Right-click in the folder, choose **New**, then **Shortcut**, and point it at `focuslab-drop.exe`.

## Options

Run from Command Prompt to change defaults, for example `focuslab-drop.exe -dir D:\Prints -port 6789`. The flags are in [setup.md](setup.md).

## Problems

| Problem | Fix |
|---|---|
| Phones cannot open the customer page | Firewall rule missing. In Windows Security, Firewall, allow `focuslab-drop.exe` on Private networks. Also check the Wi-Fi network is set to Private. |
| Page opens on the computer but not on phones | The phone is on a different network, or the QR uses the wrong IP address. Reserve the computer's IP on the router, see [setup.md](setup.md). |
| "Address already in use" | Another copy is running. Close it, or use `-port`. |
| Upload fails for a file named `NUL.jpg` or similar | Windows reserved name. See [security.md](security.md). |
