//go:build windows

package main

import "golang.org/x/sys/windows"

// acquireSingleInstance returns false if another tray app already runs for this user.
func acquireSingleInstance() bool {
	name, _ := windows.UTF16PtrFromString(`Local\FocusLabDropTray`)
	_, err := windows.CreateMutex(nil, false, name) // handle stays open until the process exits
	return err != windows.ERROR_ALREADY_EXISTS
}

func confirmQuit() bool {
	text, _ := windows.UTF16PtrFromString("The server will stop, and customers cannot send photos until the app is opened again.")
	title, _ := windows.UTF16PtrFromString("Quit FocusLab Drop?")
	r, _ := windows.MessageBox(0, text, title, windows.MB_YESNO|windows.MB_ICONQUESTION|windows.MB_SETFOREGROUND)
	return r == 6 // IDYES
}
