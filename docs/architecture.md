# Architecture

## Stack

- Go, one binary, standard library plus `modernc.org/sqlite` (pure Go, no C compiler needed, so cross-compiling works).
- Web pages are plain HTML, CSS and JavaScript embedded in the binary with `embed`. There is no build step and no external fonts or CDNs, so it works without internet.
- Staff passwords are hashed with PBKDF2 (SHA-256, 200,000 rounds) from the Go standard library.

## Files

| File | Purpose |
|---|---|
| `main.go` | Server: routes, uploads, sessions, staff login, live events, startup |
| `disk_unix.go`, `disk_windows.go` | Free disk space check for macOS and Windows |
| `main_test.go` | Automated tests |
| `tray/core/` | Windows tray app logic with no Windows calls (URLs, server flags, address picking, status text, icon drawing) and its tests. Builds and tests on any OS |
| `tray/*_windows.go` | Windows tray app: menu and icons (`main_windows.go`), the hidden child server (`server_windows.go`), Launch at login in the registry (`autostart_windows.go`), single instance and quit prompt (`single_windows.go`), clipboard and open helpers (`open_windows.go`) |
| `tray/build.sh` | Builds `dist/FocusLabDrop-windows.zip` |
| `web/drop.html` | Customer upload page: Bangla by default with an English switch (remembered on the phone), sky-blue theme. Order: intro, name and phone, upload button, hero banner, footer |
| `web/login.html` | Staff sign in |
| `web/counter.html` | Staff live list: sticky header with a status light, newest 50 rows with a "Show more" button |
| `web/assets/` | Shared CSS (sky-blue theme, same as the customer page), logo, favicon, `hero.webp` banner, and the Bangla font in `fonts/` |

## Windows tray app

`FocusLabDrop.exe` does not contain the server. Like the Mac app, it starts `focuslab-drop.exe` from its own folder as a hidden child process with `-exit-on-stdin-close` and keeps the child's stdin pipe open. If the tray app dies, the pipe closes and the server exits, so no server is left behind. The server counts as running only after `http://127.0.0.1:6789/drop` answers 200. Output goes to `%LocalAppData%\FocusLab Drop\server.log`. The tray app uses `fyne.io/systray`, which needs no C compiler on Windows. It has no settings file; Launch at login is one `Run` registry value.

## Routes

| Route | Who | Guard |
|---|---|---|
| `GET /` | anyone | Redirects to `/drop` |
| `GET /drop` | customer | none |
| `GET /drop/assets/<file>` | anyone | embedded assets only, no listing |
| `POST /drop/start` | customer | none; creates a 1 hour session cookie |
| `POST /drop/upload` | customer | needs the session cookie |
| `GET, POST /lab-counter/login` | staff | counter computer only |
| `GET /lab-counter` | staff | counter computer only, signed in |
| `GET /lab-counter/events` | staff | counter computer only, signed in |
| `GET /lab-counter/api/info` | staff | counter computer only, signed in |
| `POST /lab-counter/open` | staff | counter computer only, signed in |
| anything else | anyone | 404 |

## Customer flow

1. The page loads. The customer may type a name and phone.
2. On **Send**, the page calls `POST /drop/start` once. The server stores the cleaned name and phone in memory and sets the `fl_drop` cookie (1 hour).
3. The page uploads files three at a time, one request per file, using `XMLHttpRequest` so it can show progress.
4. For each file the server:
   1. checks the session and the free disk space,
   2. cleans the file name and checks the extension,
   3. creates the customer folder on the first file,
   4. streams the body to `<name>.part`,
   5. renames it to the final name when complete,
   6. sends an event to the counter page.
5. A later batch from the same page reuses the session and the same folder.

## Staff flow

1. `/lab-counter/login` checks the username and password against SQLite and sets the `fl_staff` cookie (12 hours, in memory).
2. The counter page loads today's folders from `/lab-counter/api/info`, then listens to `/lab-counter/events` (server-sent events) for new files.
3. **Open folder** posts a folder path. The server checks it is inside the print folder, then runs `open` (macOS), `explorer` (Windows) or `xdg-open` on it.

## Data

| Data | Where | Lifetime |
|---|---|---|
| Photos | `-dir` on disk | Until staff delete them |
| Staff users | SQLite `users` table (`username`, `salt`, `hash`) | Permanent |
| Customer sessions | Memory | 1 hour, lost on restart |
| Staff sessions | Memory | 12 hours, lost on restart |

The filesystem is the record of uploads. There is no uploads table.

## Security headers

Every response sets `X-Content-Type-Options`, `X-Frame-Options: DENY`, `Referrer-Policy: same-origin` and a Content-Security-Policy that allows only this server's own scripts, styles and images. `Referrer-Policy` must stay `same-origin`: with `no-referrer`, browsers send `Origin: null` on form posts and the counter rejects the staff login.
