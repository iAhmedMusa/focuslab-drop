//go:build windows

package main

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey      = `Software\Microsoft\Windows\CurrentVersion\Run`
	approvedKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
	runValue    = "FocusLabDrop"
)

func quotedExe() string {
	exe, _ := os.Executable()
	return `"` + exe + `"`
}

// autoStartEnabled is true only if our entry points at this exe and Windows has not switched it off
// in Settings, Apps, Startup (that sets the low bit of the first byte under StartupApproved).
func autoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetStringValue(runValue)
	if err != nil || !strings.EqualFold(v, quotedExe()) {
		return false
	}
	if a, err := registry.OpenKey(registry.CURRENT_USER, approvedKey, registry.QUERY_VALUE); err == nil {
		defer a.Close()
		if b, _, err := a.GetBinaryValue(runValue); err == nil && len(b) > 0 && b[0]&1 == 1 {
			return false
		}
	}
	return true
}

func setAutoStart(on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		if err := k.DeleteValue(runValue); err != nil && err != registry.ErrNotExist {
			return err
		}
		return nil
	}
	if err := k.SetStringValue(runValue, quotedExe()); err != nil {
		return err
	}
	if a, err := registry.OpenKey(registry.CURRENT_USER, approvedKey, registry.SET_VALUE); err == nil {
		a.DeleteValue(runValue) // forget an earlier "disabled in Startup settings"
		a.Close()
	}
	return nil
}
