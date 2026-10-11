# Security

## Threat model

Customers are on the shop guest Wi-Fi and are not trusted. The counter computer holds the print folder and may hold other shop files. A customer must never be able to read, list, change or delete anything on it, or reach the staff page.

## Protections

### The staff page is closed to customers

- Every `/lab-counter` route accepts a request only if it comes from the computer itself (loopback address). From any phone the answer is 404, so the page does not even appear to exist.
- The `Host` header must be `localhost`, `127.0.0.1`, `::1` or `focuslab`. This stops a web page from tricking a browser on the counter computer into talking to the staff page under another name (DNS rebinding).
- Staff then sign in with a username and password. This is a second layer, not the main one.
- POST requests with an `Origin` header from another site (or `null`) are refused.
- The staff cookie is `HttpOnly` and `SameSite=Strict`.
- Opening the staff page through the computer's own LAN address also gives 404. Use `localhost`.

### Customers can only write

- There is no route that lists or serves files from the print folder. The server never uses a file server on it.
- Embedded assets are served from a fixed list, with directory listing blocked.

### File and folder names are cleaned

- File names keep only letters, digits, space and `. - _ ( )`. Folders and slashes are removed, and leading dots are stripped.
- Customer names keep letters, digits, `-` and `_`. Phones keep digits and `+`. Both are length limited.
- The final folder path is checked to be inside the print folder before anything is written.
- Only photo and PDF extensions are accepted.

### Writes are safe and bounded

- Files stream to disk, so large uploads do not fill memory.
- Each file is written as `.part` and renamed when complete, so staff never print a half-written file.
- An existing file is never overwritten. A duplicate name gets ` (2)`.
- Limits: 50 MB per file, 2 GB per customer visit, 5,000 open sessions, and uploads stop under 1 GB free disk.

### Passwords

Passwords are salted and hashed with PBKDF2-SHA-256 (200,000 rounds) and compared in constant time. Unknown users take the same time as wrong passwords. A failed sign-in waits half a second.

### The open-folder button

The server runs only `open`, `explorer` or `xdg-open` with one path that it has already checked is inside the print folder. No shell is involved.

## Known limits

- **Default password.** The login is `admin` / `admin` and there is no screen to change it. It only works from the counter computer. Treat anyone with access to that computer as trusted.
- **Soft time limit.** The 1 hour customer session can be restarted by reloading the page. A real 1 hour Wi-Fi limit must come from the router.
- **Anyone on the guest Wi-Fi can upload.** There is no per-customer login. Keep the Wi-Fi password at the counter, not public, and change it when needed.
- **Plain HTTP.** Traffic on the shop Wi-Fi is not encrypted. Photos could be seen by someone sniffing that Wi-Fi.
- **File content is not scanned.** Only the extension is checked, not that the file is really an image.
- **Sessions are in memory.** A restart signs everyone out.
- **No upload rate limit** beyond the size caps. A customer can fill the disk up to the 1 GB reserve.
- **Windows reserved names** such as `NUL.jpg` fail to save. That upload shows an error.

## Safe use checklist

- Put customers on a guest network, not the main shop network.
- Reserve the counter computer's address in the router.
- Keep the counter computer's firewall on and allow the program on private networks only.
- Do not expose port 6789 to the internet (no port forwarding).
- Check free disk space on the print drive now and then.
