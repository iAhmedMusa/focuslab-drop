# Testing

## Automated tests

```
go vet ./... && go test ./...
```

They cover:

- file, folder and phone name cleaning, including `../../etc` and `..\` tricks
- a path staying inside the print folder
- an upload saved byte for byte, with duplicate names suffixed and no `.part` left over
- the `Guest_<time>` folder when name and phone are empty
- no cookie, expired session, wrong file type and path-escape names being refused
- the staff page returning 404 from another address or a wrong `Host`
- staff sign in with `admin` / `admin`, wrong password, and the `Origin` and `Referrer-Policy` rules
- the open-folder endpoint refusing paths outside the print folder
- no folder listing or file download from the customer side

## Manual test on one computer and one phone

1. Start the program:
   ```
   ./dist/focuslab-drop-mac-arm64
   ```
   Note the customer URL it prints.
2. On the computer, open `http://localhost:6789/lab-counter` and sign in with `admin` / `admin`. Leave it open.
3. On the phone, join the same Wi-Fi and open the customer URL.
4. Choose 10 or more photos from the gallery. Enter a name on one run and leave it empty on another.
5. Press **Send** and check:
   - each photo shows its own progress bar,
   - the overall bar and the count ("3 of 10 sent") update,
   - the page ends with "photos received".
6. On the computer, check:
   - a new row appears on the counter page without reloading,
   - **Open folder** opens the right folder,
   - the files are there at full size, and the folder name matches (`Name_Phone_HHMM` or `Guest_HHMM`).
7. Guardrail checks:
   - on the phone, `http://<computer-ip>:6789/lab-counter` shows 404,
   - on the computer, the same address also shows 404, and `localhost` works.
8. Try a wrong password on the sign in page. It should show "Wrong username or password."
9. Optional: make a QR code of the customer URL, scan it with the phone camera, and repeat steps 4 to 6.

## Test on Windows

1. Copy `focuslab-drop.exe` to the Windows computer and run it. Allow the firewall prompt for private networks.
2. Repeat the manual test above. Check that photos land in `C:\Users\<you>\Prints\<date>\` and that **Open folder** opens Explorer.

## Edge cases to try

| Try | Expect |
|---|---|
| Send a video | Rejected with "Only photos and PDF files are accepted." |
| Send a file over 50 MB | Rejected with "Larger than 50 MB" before upload |
| Send two files with the same name | Second one saved as `name (2).jpg` |
| Turn off phone Wi-Fi mid upload | That photo shows "Connection lost. Tap Retry." |
| Press Retry after reconnecting | The photo uploads and the row turns to Sent |
| Add more photos after "received" (**Send more**) | Same folder, counts continue |
| Wait over 1 hour, then send | "Your 1 hour session ended. Reload this page to start again." |
| Phone in dark mode | Page uses the dark colours and stays readable |

## Checking a throttled upload

On a very fast local network the progress bar jumps. To see it move, test from a phone on Wi-Fi with a large file (20 MB or more), or send many photos at once.
