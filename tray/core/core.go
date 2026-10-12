// Package core holds the tray app logic that does not call Windows, so it builds and tests anywhere.
package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"strconv"
)

type State int

const (
	Stopped State = iota
	Starting
	Running
	Failed
)

// On reports whether the server is running or about to run.
func (s State) On() bool { return s == Starting || s == Running }

func ValidPort(p int) bool { return p >= 1024 && p <= 65535 }

func CustomerURL(ip string, port int) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/drop"
}

func CounterURL(port int) string { return fmt.Sprintf("http://localhost:%d/lab-counter", port) }

// ServerArgs are the flags for focuslab-drop.exe; -exit-on-stdin-close makes it die with the tray app.
func ServerArgs(dir string, port int) []string {
	return []string{"-dir", dir, "-port", strconv.Itoa(port), "-exit-on-stdin-close"}
}

// PrivateIPv4s returns the private IPv4 addresses (the shop Wi-Fi or Ethernet), in the order given.
func PrivateIPv4s(addrs []net.Addr) []string {
	var out []string
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && n.IP.IsPrivate() {
			out = append(out, n.IP.String())
		}
	}
	return out
}

// Pick returns selected if it is still present, else the first address, else "".
func Pick(ips []string, selected string) string {
	for _, ip := range ips {
		if ip == selected {
			return ip
		}
	}
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

// Status is the one-line text at the top of the menu and the tray tooltip.
func Status(s State, problem string) string {
	switch s {
	case Starting:
		return "Starting..."
	case Running:
		return "Server on: customers can send photos"
	case Failed:
		return "Problem: " + problem
	}
	return "Server off"
}

// Icon is a 32x32 coloured dot as a Windows .ico (PNG inside). Failed adds a white "!".
func Icon(s State) []byte {
	c := map[State]color.RGBA{
		Running:  {0x22, 0xA8, 0x5A, 0xFF},
		Starting: {0xF2, 0xA1, 0x0B, 0xFF},
		Stopped:  {0xD6, 0x3B, 0x3B, 0xFF},
		Failed:   {0xD6, 0x3B, 0x3B, 0xFF},
	}[s]
	const n = 32
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			dx, dy := float64(x)-15.5, float64(y)-15.5
			if dx*dx+dy*dy <= 14*14 {
				img.SetRGBA(x, y, c)
			}
		}
	}
	if s == Failed {
		white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
		for y := 7; y <= 18; y++ {
			for x := 14; x <= 17; x++ {
				img.SetRGBA(x, y, white)
			}
		}
		for y := 22; y <= 25; y++ {
			for x := 14; x <= 17; x++ {
				img.SetRGBA(x, y, white)
			}
		}
	}
	var p bytes.Buffer
	png.Encode(&p, img)
	var ico bytes.Buffer
	binary.Write(&ico, binary.LittleEndian, []uint16{0, 1, 1})                 // ICONDIR: reserved, type icon, one image
	ico.Write([]byte{n, n, 0, 0})                                              // width, height, palette, reserved
	binary.Write(&ico, binary.LittleEndian, []uint16{1, 32})                   // planes, bits per pixel
	binary.Write(&ico, binary.LittleEndian, []uint32{uint32(p.Len()), 6 + 16}) // image size, offset
	ico.Write(p.Bytes())
	return ico.Bytes()
}
