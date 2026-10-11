# Test the Windows tray app on a real PC

The tray app has only been compiled on a Mac. This page is the full hand test to run on a Windows PC before the Windows app is merged or handed to the shop. It takes about 45 minutes. Use a pen or the result table at the end.

Pass means every step in a test behaves as the "Expect" line says. If one fails, stop that test, write down what you saw, and carry on with the next one.

## What you need

- A Windows 10 or 11 PC (64-bit). Ideally the real counter PC, or one set up the same way.
- A phone on the **same Wi-Fi** as the PC, with a few photos on it.
- `FocusLabDrop-windows.zip` (see below).
- A second browser tab and Task Manager (`Ctrl + Shift + Esc`).

## Get the zip onto the PC

1. On the Mac, in the project folder, run `./tray/build.sh`. It writes `dist/FocusLabDrop-windows.zip` (about 8 MB). If the release on GitHub already has the zip, download that instead.
2. Copy the zip to the Windows PC with a USB stick, a shared drive, or your own cloud folder.
3. On the PC, right-click the zip, choose **Properties**, tick **Unblock** at the bottom if you see it, and click **OK**. Without this, Windows adds an extra warning to every file inside.
4. Right-click the zip, **Extract All**, and extract to `C:\FocusLabDrop`.
5. Check the folder has both `FocusLabDrop.exe` and `focuslab-drop.exe`.

## Before you start: write down the PC

Open **Settings, System, About** and note:

- Windows version (10 or 11, and the build number)
- Account type: **Settings, Accounts, Your info** says "Local account" or shows an email (Microsoft account)
- Smart App Control (Windows 11 only): **Windows Security, App & browser control, Smart App Control settings**. Note Off, Evaluation or On. If it is **On**, the unsigned app will be blocked with no **Run anyway** button. Do not turn it off on a shop PC just for this test, because it cannot be turned back on. Test on another PC, or on a VM
- Wi-Fi profile: **Settings, Network & internet, Wi-Fi, your network** should say **Private network**

## Tests

Do them in order. Each one starts from "tray app running, server on, green dot" unless it says otherwise.

### 1. First run

1. Double-click `C:\FocusLabDrop\FocusLabDrop.exe`.
2. If you see "Windows protected your PC", click **More info**, then **Run anyway**.
3. If the firewall asks, tick **Private networks** and click **Allow access**.

Expect:
- no black console window appears at any point
- a dot appears in the tray, possibly under the `^` arrow next to the clock
- the dot goes amber for a moment, then green

Then drag the dot out of the `^` overflow onto the taskbar so it stays visible.

### 2. Status text and tooltip

1. Hover over the dot.
2. Click the dot to open the menu.

Expect:
- the tooltip starts with "FocusLab Drop:" and says "Server on: customers can send photos"
- the top menu line says the same, and is greyed out (it cannot be clicked)
- **Server on** has a tick
- the customer link line shows your PC's address, for example `http://192.168.1.50:6789/drop  (click to copy)`

### 3. Customer upload from the phone

1. Note the address in the menu line. On the phone, open it in the browser.
2. Pick 3 or more photos, enter a name, and press **Send**.
3. In Explorer, open `C:\Users\<you>\Prints`.

Expect:
- the phone page loads and each photo shows a progress bar and ends with "photos received"
- a folder named like `Name_Phone_HHMM` appears under today's date, holding the full-size photos

### 4. Power toggle

1. Open the menu and click **Server on** (untick it).
2. Reload the page on the phone.
3. Open the menu and click **Server on** again.

Expect:
- after step 1: the dot turns red, the top line says "Server off", the tick clears
- after step 2: the phone page does not load
- after step 3: amber, then green, the tick returns, and the phone page loads again
- in Task Manager (Details tab), `focuslab-drop.exe` is gone while off, and back while on

### 5. Menu items

Click each one in turn:

| Item | Expect |
|---|---|
| Customer link line | Nothing visible happens. Paste into the browser address bar with `Ctrl + V`: the customer link appears |
| Open counter page | The browser opens `http://localhost:6789/lab-counter`. Sign in with `admin` / `admin`. The page shows a green dot and "Live" |
| Open photos folder | Explorer opens the Prints folder (it is created if missing) |
| Show server log | Notepad opens `server.log`. It lists the "saving photos to", "staff counter" and "customer page" lines |

### 6. Staff page stays private

1. On the phone, open `http://<pc-address>:6789/lab-counter`.
2. On the PC, open `http://<pc-address>:6789/lab-counter` (the address, not `localhost`).

Expect: both show a "not found" page. Only `localhost` works.

### 7. Choose network address (skip if the PC has one address)

1. Make a second private address appear, for example by turning on a VPN, a Hyper-V or WSL virtual adapter, or by plugging in Ethernet next to Wi-Fi.
2. Wait 5 seconds and open the menu.

Expect:
- **Choose network address** appears, listing each address
- the current one has a tick, and the customer link line uses it
- clicking another address moves the tick and changes the link line

### 8. Port already in use

1. Open the menu, click **Quit**, then **Yes**.
2. Open Command Prompt, run `cd C:\FocusLabDrop`, then `focuslab-drop.exe`. Leave that black window open.
3. Start `FocusLabDrop.exe` again.

Expect:
- the dot is red with a white `!`
- the top line says "Problem: Port 6789 is already in use. Another copy may be running. Close it and try again."

Then close the black window, open the menu, and click **Server on**. Expect amber, then green.

### 9. Missing server program

1. **Quit** the tray app.
2. In `C:\FocusLabDrop`, rename `focuslab-drop.exe` to `focuslab-drop.exe.bak`.
3. Start `FocusLabDrop.exe`.

Expect: red `!` and "Problem: focuslab-drop.exe is missing. Keep it in the same folder as FocusLabDrop.exe."

Rename the file back, open the menu, and click **Server on**. Expect green.

### 10. The server stops by itself

1. With the dot green, open Task Manager, **Details** tab.
2. Right-click `focuslab-drop.exe` and choose **End task**.

Expect:
- within a second or two the dot turns red with `!`
- the top line starts "Problem: The server stopped by itself" and ends with "Check the log."
- clicking **Server on** brings it back to green

### 11. The tray app dies, the server must not be left behind

1. With the dot green, in Task Manager **Details** find `FocusLabDrop.exe` and choose **End task**.
2. Look at the list again, then run `tasklist | findstr /i focuslab` in Command Prompt.

Expect:
- the dot disappears (it may linger until you hover over it, that is a normal Windows quirk)
- `focuslab-drop.exe` is gone too. If it is still listed, this test failed: note it, then end it by hand

Start `FocusLabDrop.exe` again for the next tests.

### 12. Only one copy

1. With the tray app running, double-click `FocusLabDrop.exe` again, twice.

Expect: no second dot, no error, and Task Manager shows one `FocusLabDrop.exe`.

### 13. Quit

1. Open the menu and click **Quit**. A box asks "Quit FocusLab Drop?". Click **No**.
2. Click **Quit** again and click **Yes**.
3. Start the app again and untick **Server on**. Then click **Quit**.

Expect:
- step 1: nothing closes, the dot stays
- step 2: the dot goes and both programs are gone from Task Manager
- step 3: with the server already off, Quit closes at once with no question

### 14. Launch at login

1. Start the app. Open the menu and tick **Launch at login**.
2. In Command Prompt run:
   ```
   reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v FocusLabDrop
   ```
3. Sign out of Windows and sign back in. Wait 20 seconds.
4. Open **Settings, Apps, Startup**, find **FocusLabDrop**, and switch it **Off**. Wait 5 seconds and open the tray menu.
5. Tick **Launch at login** in the menu again.
6. Open the menu, untick **Launch at login**, and run the `reg query` command again.

Expect:
- step 2: shows `FocusLabDrop    REG_SZ    "C:\FocusLabDrop\FocusLabDrop.exe"`
- step 3: the dot appears by itself and turns green, with no black window
- step 4: the tick in the menu clears
- step 5: the tick returns, and Settings now shows FocusLabDrop **On**
- step 6: the command says it was unable to find the value

### 15. Moved folder

1. Tick **Launch at login**, then **Quit**.
2. Move the whole `C:\FocusLabDrop` folder to `C:\FocusLabDrop2` and start `FocusLabDrop.exe` from there.

Expect: **Launch at login** is not ticked (the saved path is the old one). Ticking it saves the new path. Untick it again before moving on.

### 16. Network lost and back

1. Turn Wi-Fi off on the PC (or unplug Ethernet). Wait 5 seconds and open the menu.
2. Turn it back on and wait 10 seconds.

Expect:
- during: the link line says "No network address found. Connect to the shop Wi-Fi." and is greyed out. The server keeps running
- after: the link returns, and the phone page loads again

### 17. Restart the PC (the real shop case)

Do this on the PC that will really be used, after the auto-login setup in [windows.md](windows.md) (Starting automatically). Skip it on a borrowed PC.

1. Tick **Launch at login**.
2. Restart the PC, and do not touch it.

Expect:
- Windows signs in by itself, the dot appears, and it goes green
- the customer page loads on the phone without anyone clicking anything

If Windows stops at the sign-in screen, auto-login is not set up. That is a PC setting, not an app fault.

### 18. Stays up for a while

Leave the app on, with the dot green, for at least an hour with the PC set to never sleep. Send a photo from the phone at the end.

Expect: the upload still works and the dot is still green.

### 19. Remove it

1. Untick **Launch at login**, then **Quit**.
2. Delete `C:\FocusLabDrop` (or `C:\FocusLabDrop2`).
3. Run `reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v FocusLabDrop`.

Expect: the folder deletes without "in use" errors, the registry value is not found, and the photos in `C:\Users\<you>\Prints` are still there.

## Result table

Copy this and fill it in:

| # | Test | Pass / Fail | What you saw if it failed |
|---|---|---|---|
| 1 | First run | | |
| 2 | Status text and tooltip | | |
| 3 | Customer upload from the phone | | |
| 4 | Power toggle | | |
| 5 | Menu items | | |
| 6 | Staff page stays private | | |
| 7 | Choose network address | | |
| 8 | Port already in use | | |
| 9 | Missing server program | | |
| 10 | The server stops by itself | | |
| 11 | The tray app dies | | |
| 12 | Only one copy | | |
| 13 | Quit | | |
| 14 | Launch at login | | |
| 15 | Moved folder | | |
| 16 | Network lost and back | | |
| 17 | Restart the PC | | |
| 18 | Stays up for a while | | |
| 19 | Remove it | | |

## What to send back

- the Windows version, account type, Smart App Control state, and whether the Wi-Fi is Private
- the filled result table
- for every failure: a screenshot of the tray dot and menu, and the text of `%LocalAppData%\FocusLab Drop\server.log` (paste `%LocalAppData%\FocusLab Drop\server.log` into the Explorer address bar to find it)
- anything that felt wrong even if it passed, for example a slow start or a menu that was hard to find

## Passing

The Windows app is ready to merge and hand to the shop when tests 1 to 6 and 8 to 14 pass. Tests 7, 15 to 19 can fail without blocking a merge, but a failure in 17 or 18 means the shop PC needs a setting changed (auto-login, sleep) before it is used, or the app needs a fix.
