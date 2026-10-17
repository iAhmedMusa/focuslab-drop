// FocusLab Drop: customers on the shop Wi-Fi upload photos into a folder on this computer.
package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	_ "modernc.org/sqlite"
)

//go:embed web
var webFS embed.FS

const (
	maxFileBytes    = 50 << 20
	maxSessionBytes = 2 << 30
	minFreeBytes    = 1 << 30
	maxSessions     = 5000
	customerTTL     = time.Hour
	staffTTL        = 12 * time.Hour
	pbkdf2Iter      = 200_000
)

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".heic": true, ".heif": true,
	".webp": true, ".tif": true, ".tiff": true, ".pdf": true,
}

// Admin pages answer only to these Host headers, which blocks DNS rebinding from a web page.
var adminHosts = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true, "focuslab": true}

type custSession struct {
	mu      sync.Mutex
	label   string // sanitized name and phone, may be empty
	folder  string // absolute path, created on the first upload
	bytes   int64
	expires time.Time
}

type event struct {
	Folder string    `json:"folder"` // relative to the print directory, slash separated
	File   string    `json:"file"`
	Size   int64     `json:"size"`
	At     time.Time `json:"at"`
}

type folderInfo struct {
	Folder string    `json:"folder"`
	Files  int       `json:"files"`
	Bytes  int64     `json:"bytes"`
	At     time.Time `json:"at"`
}

type server struct {
	dir string
	db  *sql.DB
	now func() time.Time

	mu    sync.Mutex
	cust  map[string]*custSession
	staff map[string]time.Time
	subs  map[chan []byte]struct{}
}

func newServer(dir string, db *sql.DB) *server {
	return &server{
		dir: dir, db: db, now: time.Now,
		cust:  map[string]*custSession{},
		staff: map[string]time.Time{},
		subs:  map[chan []byte]struct{}{},
	}
}

func token() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ---- sanitizing ----

func truncRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// safeName turns a customer name into a folder fragment: letters, digits, "-" and "_".
func safeName(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '-': // marks: Bangla vowel signs are not letters
			b.WriteRune(r)
		case unicode.IsSpace(r) || r == '_':
			b.WriteByte('_')
		}
	}
	return truncRunes(strings.Trim(b.String(), "_"), 30)
}

func safePhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '০' && r <= '৯' { // Bangla digits, typed on a Bangla keyboard
			r = '0' + (r - '০')
		}
		if (r >= '0' && r <= '9') || r == '+' {
			b.WriteRune(r)
		}
	}
	return truncRunes(b.String(), 15)
}

// safeFile keeps only the base name and a conservative character set.
func safeFile(name string) string {
	name = path.Base(strings.ReplaceAll(name, `\`, "/"))
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || strings.ContainsRune(" .-_()", r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	s := strings.TrimLeft(b.String(), ". ")
	ext := filepath.Ext(s)
	return truncRunes(strings.TrimSuffix(s, ext), 120) + ext
}

func within(base, p string) bool {
	rel, err := filepath.Rel(base, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// ---- storage ----

func (s *server) ensureFolder(c *custSession) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.folder != "" {
		return c.folder, nil
	}
	t := s.now()
	label := c.label
	if label == "" {
		label = "Guest"
	}
	base := filepath.Join(s.dir, t.Format("2006-01-02"), label+"_"+t.Format("1504"))
	if !within(s.dir, base) {
		return "", errors.New("bad folder")
	}
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		return "", err
	}
	for i := 1; i < 100; i++ {
		p := base
		if i > 1 {
			p = fmt.Sprintf("%s-%d", base, i)
		}
		err := os.Mkdir(p, 0o755)
		if err == nil {
			c.folder = p
			return p, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	return "", errors.New("no free folder name")
}

// createUnique claims "<name>.part" exclusively; the caller renames it when the upload is complete.
func createUnique(dir, name string) (*os.File, string, error) {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 1; i < 1000; i++ {
		n := name
		if i > 1 {
			n = fmt.Sprintf("%s (%d)%s", stem, i, ext)
		}
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			continue
		}
		f, err := os.OpenFile(filepath.Join(dir, n+".part"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return f, n, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, "", err
		}
	}
	return nil, "", errors.New("too many files with the same name")
}

func (s *server) today() []folderInfo {
	root := filepath.Join(s.dir, s.now().Format("2006-01-02"))
	out := []folderInfo{}
	ents, _ := os.ReadDir(root)
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		fi := folderInfo{Folder: s.now().Format("2006-01-02") + "/" + e.Name()}
		files, _ := os.ReadDir(filepath.Join(root, e.Name()))
		for _, f := range files {
			info, err := f.Info()
			if err != nil || f.IsDir() || strings.HasSuffix(f.Name(), ".part") {
				continue
			}
			fi.Files++
			fi.Bytes += info.Size()
			if info.ModTime().After(fi.At) {
				fi.At = info.ModTime()
			}
		}
		if fi.Files > 0 {
			out = append(out, fi)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// ---- customer side ----

func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *server) custFrom(r *http.Request) *custSession {
	ck, err := r.Cookie("fl_drop")
	if err != nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.cust[ck.Value]
	if c == nil || !c.expires.After(s.now()) {
		return nil
	}
	return c
}

func (s *server) start(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var parts []string
	for _, p := range []string{safeName(r.PostFormValue("name")), safePhone(r.PostFormValue("phone"))} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	tok := token()
	s.mu.Lock()
	if len(s.cust) >= maxSessions {
		s.mu.Unlock()
		http.Error(w, "busy", http.StatusTooManyRequests)
		return
	}
	s.cust[tok] = &custSession{label: strings.Join(parts, "_"), expires: s.now().Add(customerTTL)}
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "fl_drop", Value: tok, Path: "/drop", MaxAge: int(customerTTL.Seconds()), HttpOnly: true, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) upload(w http.ResponseWriter, r *http.Request) {
	c := s.custFrom(r)
	if c == nil {
		jsonErr(w, http.StatusUnauthorized, "Your 1 hour session ended. Reload this page to start again.")
		return
	}
	c.mu.Lock()
	used := c.bytes
	c.mu.Unlock()
	if used >= maxSessionBytes {
		jsonErr(w, http.StatusRequestEntityTooLarge, "Upload limit reached for this visit.")
		return
	}
	if free, err := freeBytes(s.dir); err == nil && free < minFreeBytes {
		jsonErr(w, http.StatusInsufficientStorage, "The lab computer is out of space. Please tell our staff.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFileBytes+(1<<20))
	mr, err := r.MultipartReader()
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "Bad upload.")
		return
	}
	for {
		part, err := mr.NextPart()
		if err != nil {
			jsonErr(w, http.StatusBadRequest, "No file received.")
			return
		}
		if part.FormName() == "file" {
			s.saveFile(w, c, part)
			return
		}
	}
}

func (s *server) saveFile(w http.ResponseWriter, c *custSession, part *multipart.Part) {
	name := safeFile(part.FileName())
	if !allowedExt[strings.ToLower(filepath.Ext(name))] {
		jsonErr(w, http.StatusUnsupportedMediaType, "Only photos and PDF files are accepted.")
		return
	}
	dir, err := s.ensureFolder(c)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "Could not save. Please tell our staff.")
		return
	}
	f, final, err := createUnique(dir, name)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "Could not save. Please tell our staff.")
		return
	}
	tmp := f.Name()
	n, err := io.Copy(f, part)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > maxFileBytes {
		err = &http.MaxBytesError{Limit: maxFileBytes}
	}
	if err == nil {
		err = os.Rename(tmp, filepath.Join(dir, final))
	}
	if err != nil {
		os.Remove(tmp)
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			jsonErr(w, http.StatusRequestEntityTooLarge, "This file is larger than 50 MB.")
		} else {
			jsonErr(w, http.StatusBadRequest, "Upload interrupted. Please retry.")
		}
		return
	}
	c.mu.Lock()
	c.bytes += n
	c.mu.Unlock()
	rel, _ := filepath.Rel(s.dir, dir)
	s.publish(event{Folder: filepath.ToSlash(rel), File: final, Size: n, At: s.now()})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "name": final, "size": n})
}

// ---- counter side ----

func hashPass(pass string, salt []byte) []byte {
	h, _ := pbkdf2.Key(sha256.New, pass, salt, pbkdf2Iter, 32)
	return h
}

func (s *server) checkLogin(user, pass string) bool {
	var salt, hash []byte
	err := s.db.QueryRow(`SELECT salt, hash FROM users WHERE username = ?`, user).Scan(&salt, &hash)
	if err != nil {
		salt = make([]byte, 16) // same work for unknown users
	}
	return subtle.ConstantTimeCompare(hashPass(pass, salt), hash) == 1 && err == nil
}

func (s *server) isStaff(r *http.Request) bool {
	ck, err := r.Cookie("fl_staff")
	if err != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.staff[ck.Value]
	return ok && exp.After(s.now())
}

// counter allows loopback requests with an admin Host only; unknown routes look like a 404.
func (s *server) counter(auth bool, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		host := r.Host
		if hh, _, err := net.SplitHostPort(host); err == nil {
			host = hh
		}
		host = strings.Trim(strings.ToLower(host), "[]")
		if p := net.ParseIP(ip); p == nil || !p.IsLoopback() || !adminHosts[host] {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			if o := r.Header.Get("Origin"); o != "" {
				if u, err := url.Parse(o); err != nil || u.Host != r.Host {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
		}
		if auth && !s.isStaff(r) {
			if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/lab-counter/api") && r.URL.Path != "/lab-counter/events" {
				http.Redirect(w, r, "/lab-counter/login", http.StatusSeeOther)
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}
		h(w, r)
	}
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if r.ParseForm() != nil || !s.checkLogin(r.PostFormValue("username"), r.PostFormValue("password")) {
		time.Sleep(500 * time.Millisecond)
		http.Redirect(w, r, "/lab-counter/login?e=1", http.StatusSeeOther)
		return
	}
	tok := token()
	s.mu.Lock()
	s.staff[tok] = s.now().Add(staffTTL)
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "fl_staff", Value: tok, Path: "/lab-counter", MaxAge: int(staffTTL.Seconds()), HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, "/lab-counter", http.StatusSeeOther)
}

var reviewHosts = map[string]bool{"g.page": true, "maps.app.goo.gl": true, "google.com": true, "www.google.com": true, "search.google.com": true}

// validReview accepts an empty string (clears the link) or an https link on a Google host.
func validReview(v string) bool {
	if v == "" {
		return true
	}
	u, err := url.Parse(v)
	return len(v) <= 500 && err == nil && u.Scheme == "https" && reviewHosts[strings.ToLower(u.Hostname())] && u.User == nil
}

func (s *server) review() string {
	var v string
	s.db.QueryRow(`SELECT value FROM settings WHERE key = 'review'`).Scan(&v)
	return v
}

func (s *server) config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]string{"review": s.review()})
}

func (s *server) setReview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		URL string `json:"url"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	in.URL = strings.TrimSpace(in.URL)
	if !validReview(in.URL) {
		http.Error(w, "use an https link from Google (g.page, maps.app.goo.gl or google.com)", http.StatusBadRequest)
		return
	}
	if _, err := s.db.Exec(`INSERT INTO settings (key, value) VALUES ('review', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, in.URL); err != nil {
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) info(w http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		port = "6789"
	}
	urls := []string{}
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && n.IP.IsPrivate() {
			urls = append(urls, "http://"+net.JoinHostPort(n.IP.String(), port)+"/drop")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"dir": s.dir, "urls": urls, "sessions": s.today(), "review": s.review()})
}

func (s *server) publish(e event) {
	b, _ := json.Marshal(e)
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- b:
		default:
		}
	}
}

func (s *server) events(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan []byte, 32)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()
	fmt.Fprint(w, ": ok\n\n")
	fl.Flush()
	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case m := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", m)
		case <-tick.C:
			fmt.Fprint(w, ": ping\n\n")
		}
		fl.Flush()
	}
}

// open reveals a folder inside the print directory in Finder or Explorer.
func (s *server) open(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Folder string `json:"folder"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p := filepath.Join(s.dir, filepath.FromSlash(in.Folder))
	if !within(s.dir, p) {
		http.Error(w, "bad folder", http.StatusBadRequest)
		return
	}
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	cmd := map[string]string{"windows": "explorer", "darwin": "open"}[runtime.GOOS]
	if cmd == "" {
		cmd = "xdg-open"
	}
	c := exec.Command(cmd, p)
	if c.Start() != nil {
		http.Error(w, "could not open", http.StatusInternalServerError)
		return
	}
	go c.Wait()
	w.WriteHeader(http.StatusNoContent)
}

// ---- wiring ----

func (s *server) page(name string) http.HandlerFunc {
	b, err := webFS.ReadFile("web/" + name)
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(b)
	}
}

func (s *server) routes() http.Handler {
	assets, _ := fs.Sub(webFS, "web/assets")
	files := http.StripPrefix("/drop/assets/", http.FileServerFS(assets))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/drop", http.StatusFound) })
	mux.HandleFunc("GET /drop", s.page("drop.html"))
	mux.HandleFunc("GET /drop/assets/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") { // no directory listing
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		files.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /drop/config", s.config)
	mux.HandleFunc("POST /drop/start", s.start)
	mux.HandleFunc("POST /drop/upload", s.upload)

	mux.HandleFunc("GET /lab-counter/login", s.counter(false, s.page("login.html")))
	mux.HandleFunc("POST /lab-counter/login", s.counter(false, s.login))
	mux.HandleFunc("GET /lab-counter", s.counter(true, s.page("counter.html")))
	mux.HandleFunc("GET /lab-counter/events", s.counter(true, s.events))
	mux.HandleFunc("GET /lab-counter/api/info", s.counter(true, s.info))
	mux.HandleFunc("POST /lab-counter/open", s.counter(true, s.open))
	mux.HandleFunc("POST /lab-counter/review", s.counter(true, s.setReview))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		// not "no-referrer": browsers then send "Origin: null" on form posts and the counter's Origin check rejects them
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' blob: data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		mux.ServeHTTP(w, r)
	})
}

func (s *server) janitor() {
	for range time.Tick(5 * time.Minute) {
		now := s.now()
		s.mu.Lock()
		for k, c := range s.cust {
			if !c.expires.After(now) {
				delete(s.cust, k)
			}
		}
		for k, exp := range s.staff {
			if !exp.After(now) {
				delete(s.staff, k)
			}
		}
		s.mu.Unlock()
	}
}

func initDB(file string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, salt BLOB NOT NULL, hash BLOB NOT NULL)`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		return nil, err
	}
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	if n == 0 {
		salt := make([]byte, 16)
		rand.Read(salt)
		if _, err := db.Exec(`INSERT INTO users VALUES ('admin', ?, ?)`, salt, hashPass("admin", salt)); err != nil {
			return nil, err
		}
		log.Println("created default staff login admin / admin")
	}
	return db, nil
}

func main() {
	home, _ := os.UserHomeDir()
	cfg, _ := os.UserConfigDir()
	dir := flag.String("dir", filepath.Join(home, "Prints"), "folder where customer photos are saved")
	port := flag.Int("port", 6789, "port to listen on")
	dbFile := flag.String("db", filepath.Join(cfg, "focuslab-drop", "users.db"), "staff login database")
	exitOnStdinClose := flag.Bool("exit-on-stdin-close", false, "exit when stdin closes (the Mac app sets this so a crashed app never leaves a server behind)")
	flag.Parse()
	if *exitOnStdinClose {
		go func() { io.Copy(io.Discard, os.Stdin); os.Exit(0) }()
	}

	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		log.Fatal(err)
	}
	db, err := initDB(*dbFile)
	if err != nil {
		log.Fatal(err)
	}
	s := newServer(abs, db)
	go s.janitor()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("saving photos to %s", abs)
	log.Printf("staff counter: http://localhost:%d/lab-counter", *port)
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && n.IP.IsPrivate() {
			log.Printf("customer page: http://%s:%d/drop", n.IP, *port)
		}
	}
	srv := &http.Server{Handler: s.routes(), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 2 * time.Minute}
	log.Fatal(srv.Serve(ln))
}
