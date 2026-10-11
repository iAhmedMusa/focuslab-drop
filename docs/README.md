# FocusLab Drop documentation

FocusLab Drop lets print-shop customers send photos from their phone to the counter computer over the shop Wi-Fi. No internet, no app, no WhatsApp downloads.

| Doc | Read it to |
|---|---|
| [overview.md](overview.md) | Understand what the app does and the decisions behind it |
| [setup.md](setup.md) | Set up the counter computer, router and QR codes |
| [operations.md](operations.md) | Run it day to day and fix common problems |
| [architecture.md](architecture.md) | See how the code, routes and folders fit together |
| [security.md](security.md) | Know what is protected, and the limits |
| [testing.md](testing.md) | Test it by hand and run the automated tests |

Quick start: build or pick a binary from `dist/`, run it on the counter computer, open `http://localhost:6789/lab-counter` (login `admin` / `admin`), and send a photo from a phone at the URL printed in the terminal.
