//go:build windows

package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// run starts a helper without flashing a console window.
func run(stdin string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	cmd.Start() // best effort; the helper outlives us if needed
	go cmd.Wait()
}

func copyText(s string) { run(s, "clip") }
func openURL(u string)  { run("", "rundll32", "url.dll,FileProtocolHandler", u) }
func openLog()          { run("", "notepad", logPath()) }
func openFolder(dir string) {
	os.MkdirAll(dir, 0o755)
	run("", "explorer", dir)
}
