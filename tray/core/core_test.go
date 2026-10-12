package core

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"net"
	"reflect"
	"testing"
)

func TestValidPort(t *testing.T) {
	for p, want := range map[int]bool{0: false, 80: false, 1023: false, 1024: true, 6789: true, 65535: true, 65536: false} {
		if ValidPort(p) != want {
			t.Errorf("ValidPort(%d) = %v, want %v", p, !want, want)
		}
	}
}

func TestURLsAndArgs(t *testing.T) {
	if got := CustomerURL("192.168.1.50", 6789); got != "http://192.168.1.50:6789/drop" {
		t.Error(got)
	}
	if got := CounterURL(6789); got != "http://localhost:6789/lab-counter" {
		t.Error(got)
	}
	want := []string{"-dir", `C:\Prints`, "-port", "6789", "-exit-on-stdin-close"}
	if got := ServerArgs(`C:\Prints`, 6789); !reflect.DeepEqual(got, want) {
		t.Error(got)
	}
}

func TestPrivateIPv4sAndPick(t *testing.T) {
	cidr := func(s string) net.Addr { ip, n, _ := net.ParseCIDR(s); n.IP = ip; return n }
	got := PrivateIPv4s([]net.Addr{cidr("192.168.1.50/24"), cidr("8.8.8.8/24"), cidr("fe80::1/64"), cidr("10.0.0.4/8"), &net.IPAddr{}})
	if !reflect.DeepEqual(got, []string{"192.168.1.50", "10.0.0.4"}) {
		t.Fatal(got)
	}
	if Pick(got, "10.0.0.4") != "10.0.0.4" || Pick(got, "1.1.1.1") != "192.168.1.50" || Pick(nil, "x") != "" {
		t.Error("Pick chose the wrong address")
	}
}

func TestStatus(t *testing.T) {
	if Status(Stopped, "") != "Server off" || Status(Failed, "port in use") != "Problem: port in use" {
		t.Error("wrong status text")
	}
	if !Starting.On() || !Running.On() || Stopped.On() || Failed.On() {
		t.Error("On() wrong")
	}
}

func TestIconIsValidICO(t *testing.T) {
	for _, s := range []State{Stopped, Starting, Running, Failed} {
		b := Icon(s)
		if binary.LittleEndian.Uint16(b[2:]) != 1 || binary.LittleEndian.Uint16(b[4:]) != 1 {
			t.Fatalf("state %d: bad ICO header", s)
		}
		size, off := binary.LittleEndian.Uint32(b[14:]), binary.LittleEndian.Uint32(b[18:])
		if int(off+size) != len(b) {
			t.Fatalf("state %d: size/offset do not match file length", s)
		}
		if _, err := png.Decode(bytes.NewReader(b[off:])); err != nil {
			t.Fatalf("state %d: %v", s, err)
		}
	}
	if bytes.Equal(Icon(Stopped), Icon(Failed)) {
		t.Error("failed icon should differ from stopped")
	}
}
