# Overview

## The problem

Customers sent photos to the shop on WhatsApp. Staff had to open WhatsApp Desktop and download each photo, WhatsApp compressed the photos, and it used the shop's limited internet.

## What FocusLab Drop does

1. The customer joins the shop guest Wi-Fi and scans a printed QR code.
2. Their phone browser opens the upload page. They optionally type a name and phone number, then pick photos from the gallery or files.
3. Each photo shows a progress bar while it uploads.
4. The original files are saved straight into a folder on the counter computer. Staff print from that folder.

Everything stays on the shop network. Nothing is sent over the internet.

## Who uses what

| Person | Page | Where |
|---|---|---|
| Customer | `/drop` | Any phone on the shop Wi-Fi |
| Staff | `/lab-counter` | Only on the counter computer |

## Decisions

- **Local server, one program.** A single Go binary with the web pages inside it. It runs on macOS and Windows with nothing else to install.
- **Fixed QR codes.** One QR for the guest Wi-Fi and one for the upload page. Both are printed once and managed by hand. The app does not generate or change them.
- **QR points at the IP address.** Phones do not read the counter computer's hosts file, so a name like `focuslab` would not work from a phone. See [setup.md](setup.md).
- **Name and phone are optional.** With neither, the folder is named `Guest_<time>`.
- **Bangla first.** The customer page opens in Bangla (font: Kalpurush) with a one-tap switch to English. Bangla names, Bangla digits in phone numbers and Bangla file names are kept as typed.
- **Staff login in SQLite.** A small database holds the staff user. Default login is `admin` / `admin`.
- **Customers can only upload.** No listing, no download, no access to other customers' photos.
- **1 hour limit is soft.** The upload session expires after 1 hour. A real 1 hour Wi-Fi limit has to come from the router's guest network settings.

## Not included

- Generating or changing QR codes, or managing the Wi-Fi password.
- A Wi-Fi time limit or captive portal.
- Changing the staff password from the screen (see [operations.md](operations.md)).
- Automatic printing, order tracking, or customer notifications.
- Starting automatically when the computer boots.

## Where the folders go

```
Prints/
  2026-10-11/
    Ayesha_Khan_03001234567_1432/
      IMG_0001.jpg
      IMG_0002.jpg
    Guest_1507/
      photo.heic
```

Folder name is `<name>_<phone>_<HHMM>` using whichever parts the customer gave. Two files with the same name in one folder become `name.jpg` and `name (2).jpg`, so nothing is overwritten.
