//go:build windows

package main

import (
	"net"
	"time"

	"fyne.io/systray"

	"focuslab-drop/tray/core"
)

const maxAddrs = 5

type app struct {
	srv      *server
	selected string // chosen network address; not saved between runs
	ips      []string
	lastIcon core.State
	iconSet  bool

	status, power, link, counter, folder, logItem, login, quit *systray.MenuItem
	pickMenu                                                   *systray.MenuItem
	pick                                                       [maxAddrs]*systray.MenuItem
}

func main() {
	if !acquireSingleInstance() {
		return
	}
	a := &app{srv: newServer()}
	systray.Run(a.ready, a.srv.stop)
}

func (a *app) ready() {
	systray.SetTitle("FocusLab Drop")
	a.status = systray.AddMenuItem("", "")
	a.status.Disable()
	systray.AddSeparator()
	a.power = systray.AddMenuItemCheckbox("Server on", "Turn the server on or off", false)
	a.link = systray.AddMenuItem("", "Click to copy the customer link")
	a.pickMenu = systray.AddMenuItem("Choose network address", "Pick the Wi-Fi address customers use")
	for i := range a.pick {
		a.pick[i] = a.pickMenu.AddSubMenuItemCheckbox("", "", false)
		a.pick[i].Hide()
	}
	systray.AddSeparator()
	a.counter = systray.AddMenuItem("Open counter page", "")
	a.folder = systray.AddMenuItem("Open photos folder", "")
	a.logItem = systray.AddMenuItem("Show server log", "")
	systray.AddSeparator()
	a.login = systray.AddMenuItemCheckbox("Launch at login", "Start FocusLab Drop when you sign in to Windows", false)
	a.quit = systray.AddMenuItem("Quit", "Stop the server and close the app")

	a.srv.start()
	a.refresh()
	go a.loop()
}

func (a *app) loop() {
	tick := time.NewTicker(5 * time.Second) // network or Startup settings can change outside the app
	defer tick.Stop()
	pickCh := make(chan int)
	for i, it := range a.pick {
		go func() {
			for range it.ClickedCh {
				pickCh <- i
			}
		}()
	}
	for {
		select {
		case <-a.srv.changed:
		case <-tick.C:
		case i := <-pickCh:
			a.selected = a.ips[i]
		case <-a.power.ClickedCh:
			if st, _ := a.srv.snapshot(); st.On() {
				a.srv.stop()
			} else {
				a.srv.start()
			}
		case <-a.link.ClickedCh:
			if ip := core.Pick(a.ips, a.selected); ip != "" {
				copyText(core.CustomerURL(ip, port))
			}
		case <-a.counter.ClickedCh:
			openURL(core.CounterURL(port))
		case <-a.folder.ClickedCh:
			openFolder(a.srv.dir)
		case <-a.logItem.ClickedCh:
			openLog()
		case <-a.login.ClickedCh:
			setAutoStart(!autoStartEnabled())
		case <-a.quit.ClickedCh:
			if st, _ := a.srv.snapshot(); !st.On() || confirmQuit() {
				systray.Quit()
				return
			}
		}
		a.refresh()
	}
}

func (a *app) refresh() {
	st, problem := a.srv.snapshot()
	text := core.Status(st, problem)
	a.status.SetTitle(text)
	systray.SetTooltip("FocusLab Drop: " + text)
	if !a.iconSet || a.lastIcon != st {
		systray.SetIcon(core.Icon(st))
		a.lastIcon, a.iconSet = st, true
	}
	setChecked(a.power, st.On())

	addrs, _ := net.InterfaceAddrs()
	a.ips = core.PrivateIPv4s(addrs)
	cur := core.Pick(a.ips, a.selected)
	if cur == "" {
		a.link.SetTitle("No network address found. Connect to the shop Wi-Fi.")
		a.link.Disable()
	} else {
		a.link.SetTitle(core.CustomerURL(cur, port) + "  (click to copy)")
		a.link.Enable()
	}
	if len(a.ips) > 1 {
		a.pickMenu.Show()
	} else {
		a.pickMenu.Hide()
	}
	for i, it := range a.pick {
		if i < len(a.ips) && i < maxAddrs {
			it.SetTitle(a.ips[i])
			it.Show()
			setChecked(it, a.ips[i] == cur)
		} else {
			it.Hide()
		}
	}
	setChecked(a.login, autoStartEnabled())
}

func setChecked(it *systray.MenuItem, on bool) {
	if on {
		it.Check()
	} else {
		it.Uncheck()
	}
}
