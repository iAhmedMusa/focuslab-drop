//go:build windows

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"focuslab-drop/tray/core"
)

const (
	port           = 6789
	createNoWindow = 0x08000000
)

// server runs focuslab-drop.exe as a hidden child and reports whether it is really answering.
type server struct {
	mu      sync.Mutex
	state   core.State
	problem string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	done    chan struct{}
	runID   int
	changed chan struct{} // poked on every state change; the UI goroutine redraws
	dir     string
}

func newServer() *server {
	home, _ := os.UserHomeDir()
	return &server{changed: make(chan struct{}, 1), dir: filepath.Join(home, "Prints")}
}

func logPath() string {
	cache, _ := os.UserCacheDir() // %LocalAppData%
	return filepath.Join(cache, "FocusLab Drop", "server.log")
}

func serverBinary() string {
	if env := os.Getenv("FOCUSLAB_SERVER_BIN"); env != "" {
		return env
	}
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "focuslab-drop.exe")
}

func (s *server) snapshot() (core.State, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state, s.problem
}

func (s *server) setLocked(st core.State, problem string) {
	s.state, s.problem = st, problem
	select {
	case s.changed <- struct{}{}:
	default:
	}
}

func (s *server) start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.On() {
		return
	}
	bin := serverBinary()
	if _, err := os.Stat(bin); err != nil {
		s.setLocked(core.Failed, "focuslab-drop.exe is missing. Keep it in the same folder as FocusLabDrop.exe.")
		return
	}
	if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		s.setLocked(core.Failed, fmt.Sprintf("Port %d is already in use. Another copy may be running. Close it and try again.", port))
		return
	} else {
		ln.Close()
	}
	os.MkdirAll(filepath.Dir(logPath()), 0o755)
	logf, err := os.Create(logPath()) // replaced on each start
	if err != nil {
		s.setLocked(core.Failed, "Could not write the log: "+err.Error())
		return
	}
	cmd := exec.Command(bin, core.ServerArgs(s.dir, port)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	cmd.Stdout, cmd.Stderr = logf, logf
	in, _ := cmd.StdinPipe() // held open on purpose: when this app dies the pipe closes and the server exits
	if err := cmd.Start(); err != nil {
		logf.Close()
		s.setLocked(core.Failed, "Could not start the server: "+err.Error())
		return
	}
	s.runID++
	id := s.runID
	s.cmd, s.stdin = cmd, in
	done := make(chan struct{})
	s.done = done
	s.setLocked(core.Starting, "")
	go func() {
		err := cmd.Wait()
		logf.Close()
		close(done)
		s.exited(id, err)
	}()
	go s.waitAnswering(id)
}

func (s *server) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
	s.setLocked(core.Stopped, "")
}

func (s *server) stopLocked() {
	if s.cmd == nil {
		return
	}
	s.runID++ // the exit handler of this run is now stale and must not report a failure
	s.stdin.Close()
	select {
	case <-s.done:
	case <-time.After(2 * time.Second):
		s.cmd.Process.Kill()
		<-s.done
	}
	s.cmd, s.stdin = nil, nil
}

func (s *server) exited(id int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != s.runID {
		return
	}
	s.cmd, s.stdin = nil, nil
	detail := lastLogLine()
	if detail == "" && err != nil {
		detail = err.Error()
	}
	s.setLocked(core.Failed, "The server stopped by itself ("+detail+"). Check the log.")
}

func (s *server) waitAnswering(id int) {
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 40; i++ {
		s.mu.Lock()
		alive := id == s.runID && s.cmd != nil
		s.mu.Unlock()
		if !alive {
			return
		}
		if resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/drop", port)); err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				s.mu.Lock()
				if id == s.runID {
					s.setLocked(core.Running, "")
				}
				s.mu.Unlock()
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == s.runID {
		s.stopLocked()
		s.setLocked(core.Failed, "The server did not start in time. Try again.")
	}
}

func lastLogLine() string {
	b, err := os.ReadFile(logPath())
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
