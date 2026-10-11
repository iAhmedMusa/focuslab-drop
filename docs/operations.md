# Operations

## Daily use

1. Start the program on the counter computer and leave the window open.
2. Open `http://localhost:6789/lab-counter` and sign in. Sign-in lasts 12 hours, and again after every restart.
3. Customers scan the two QR codes and upload. New rows appear on the counter page as files arrive.
4. Click **Open folder** on a row to open that customer's folder in Finder or Explorer, then print from it.
5. Turn on **Sound** (button at the top) to hear a beep when photos arrive. Browsers allow sound only after you click, so click once after opening the page.
6. Stop with `Ctrl+C` in the terminal, or close the window.

The counter page lists today's folders. Older days stay on disk under `Prints/<date>/`.

## Staff login

- Default login: `admin` / `admin`.
- The counter page only opens on the counter computer itself, so the default password is not reachable from customer phones.
- There is no screen to change the password yet. To reset to the default, quit the program, delete `users.db` (location in [setup.md](setup.md)), and start it again. It recreates `admin` / `admin`.

## Limits

| Limit | Value |
|---|---|
| File types | jpg, jpeg, png, heic, heif, webp, tif, tiff, pdf |
| Size per file | 50 MB |
| Size per customer visit | 2 GB |
| Free disk space needed | at least 1 GB, otherwise uploads stop |
| Customer session | 1 hour |
| Staff session | 12 hours |

## Troubleshooting

| Problem | Likely cause and fix |
|---|---|
| Phone cannot open the customer page | Phone is on mobile data or another Wi-Fi. Join the guest Wi-Fi. If it still fails, the router may have "client isolation" on, or the computer firewall is blocking port 6789. |
| QR code worked yesterday, not today | The computer's IP changed. Set the router address reservation, then print a new QR. |
| Counter page shows 404 | You opened it with the computer's LAN address. Use `http://localhost:6789/lab-counter`. This is intentional. |
| Counter page keeps asking to sign in | The program was restarted, or 12 hours passed. Sign in again. |
| "Your 1 hour session ended" on the phone | Reload the page to start a new session. |
| "The lab computer is out of space" | Free at least 1 GB on the drive that holds `Prints`. |
| "Only photos and PDF files are accepted" | The file is another type, for example a video. Send photos only. |
| "Larger than 50 MB" | Send that file another way, or resize it. |
| Program says address already in use | Another copy is running, or something else uses the port. Stop it or start with `-port 6790` and print a new QR. |
| Photo shows as a grey box with JPG or HEIC on the phone | Preview only. The browser could not draw a thumbnail. The file still uploads. |
| iPhone photos arrive as .jpg instead of .heic | Safari converts HEIC to JPEG on upload. Fine for printing. |
| Upload stops when the phone screen locks | Ask customers to keep the page open until it says received. |
| Folder contains a `.part` file | An upload was interrupted. Staff can delete it. The customer can send the photo again. |
