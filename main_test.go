package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T) *server {
	t.Helper()
	db, err := initDB(filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	s := newServer(filepath.Join(t.TempDir(), "Prints"), db)
	s.now = func() time.Time { return time.Date(2026, 10, 11, 14, 32, 0, 0, time.Local) }
	return s
}

func do(h http.Handler, method, target, remote string, mod func(*http.Request)) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	r.RemoteAddr = remote
	if mod != nil {
		mod(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func cookieOf(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func startSession(t *testing.T, h http.Handler, name, phone string) *http.Cookie {
	t.Helper()
	w := do(h, "POST", "/drop/start", "192.168.1.9:5000", func(r *http.Request) {
		r.Body = io.NopCloser(strings.NewReader(url.Values{"name": {name}, "phone": {phone}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	})
	c := cookieOf(w, "fl_drop")
	if w.Code != 204 || c == nil {
		t.Fatalf("start: %d", w.Code)
	}
	return c
}

func upload(h http.Handler, c *http.Cookie, filename string, data []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	fw.Write(data)
	mw.Close()
	r := httptest.NewRequest("POST", "/drop/upload", &buf)
	r.RemoteAddr = "192.168.1.9:5000"
	r.Header.Set("Content-Type", mw.FormDataContentType())
	if c != nil {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestSanitizers(t *testing.T) {
	cases := []struct{ got, want string }{
		{safeName("  Ayesha Khan "), "Ayesha_Khan"},
		{safeName("../../etc"), "etc"},
		{safeName(`..\..\x`), "x"},
		{safeName(""), ""},
		{safeName("আয়েশা খান"), "আয়েশা_খান"}, // Bangla vowel signs must survive
		{safeFile("ছবি১ (২).jpg"), "ছবি১ (২).jpg"},
		{safePhone("+92 300-123 4567abc"), "+923001234567"},
		{safePhone("০১৭১১-২২৩৩৪৪"), "01711223344"},
		{safeFile("../../etc/passwd.jpg"), "passwd.jpg"},
		{safeFile(`..\..\evil.png`), "evil.png"},
		{safeFile(".hidden.jpg"), "hidden.jpg"},
		{safeFile("IMG 1;rm -rf.jpg"), "IMG 1_rm -rf.jpg"},
		{safeFile(".."), ""},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q want %q", c.got, c.want)
		}
	}
	if within("/a/b", "/a/b/../c") || !within("/a/b", "/a/b/c/d") {
		t.Error("within is wrong")
	}
}

func TestUploadSavesByteIdentical(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	c := startSession(t, h, "Ayesha Khan", "0300 1234567")
	data := bytes.Repeat([]byte{0xff, 0xd8, 0x42}, 100000)
	if w := upload(h, c, "IMG_1.jpg", data); w.Code != 200 {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	// same name again must not overwrite
	if w := upload(h, c, "IMG_1.jpg", []byte("second")); w.Code != 200 {
		t.Fatalf("second upload: %d", w.Code)
	}
	dir := filepath.Join(s.dir, "2026-10-11", "Ayesha_Khan_03001234567_1432")
	got, err := os.ReadFile(filepath.Join(dir, "IMG_1.jpg"))
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("file missing or differs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "IMG_1 (2).jpg")); err != nil {
		t.Fatal("duplicate name was not suffixed")
	}
	if left, _ := filepath.Glob(filepath.Join(dir, "*.part")); len(left) != 0 {
		t.Fatalf("leftover .part files: %v", left)
	}
}

func TestGuestFolderWhenNameAndPhoneEmpty(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	c := startSession(t, h, "", "")
	if w := upload(h, c, "a.png", []byte("x")); w.Code != 200 {
		t.Fatalf("upload: %d", w.Code)
	}
	if _, err := os.Stat(filepath.Join(s.dir, "2026-10-11", "Guest_1432", "a.png")); err != nil {
		t.Fatal(err)
	}
}

func TestUploadRejects(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	if w := upload(h, nil, "a.jpg", []byte("x")); w.Code != 401 {
		t.Errorf("no cookie: %d", w.Code)
	}
	c := startSession(t, h, "", "")
	if w := upload(h, c, "run.exe", []byte("x")); w.Code != 415 {
		t.Errorf("exe: %d", w.Code)
	}
	if w := upload(h, c, "../../../escape.jpg", []byte("x")); w.Code != 200 {
		t.Errorf("traversal name should be flattened: %d", w.Code)
	}
	if _, err := os.Stat(filepath.Join(s.dir, "..", "escape.jpg")); err == nil {
		t.Error("file escaped the print directory")
	}
	// expire the session
	s.now = func() time.Time { return time.Date(2026, 10, 11, 15, 33, 0, 0, time.Local) }
	if w := upload(h, c, "late.jpg", []byte("x")); w.Code != 401 {
		t.Errorf("expired session: %d", w.Code)
	}
}

func TestCounterIsLoopbackOnly(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	for _, p := range []string{"/lab-counter", "/lab-counter/login", "/lab-counter/events", "/lab-counter/api/info"} {
		if w := do(h, "GET", p, "192.168.1.9:5000", func(r *http.Request) { r.Host = "localhost:6789" }); w.Code != 404 {
			t.Errorf("%s from LAN: %d, want 404", p, w.Code)
		}
	}
	// Staff on the counter machine reaching it via its LAN IP is also refused.
	if w := do(h, "GET", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) { r.Host = "192.168.1.50:6789" }); w.Code != 404 {
		t.Errorf("unexpected Host accepted: %d", w.Code)
	}
	if w := do(h, "GET", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) { r.Host = "localhost:6789" }); w.Code != 200 {
		t.Errorf("loopback login page: %d", w.Code)
	}
}

func TestCounterAuth(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	local := func(r *http.Request) { r.Host = "localhost:6789" }
	if w := do(h, "GET", "/lab-counter/login", "127.0.0.1:5000", local); w.Header().Get("Referrer-Policy") != "same-origin" {
		t.Error("Referrer-Policy must be same-origin so browsers send a real Origin on form posts")
	}
	// the Origin check: same-origin passes, a foreign or null origin is refused
	for origin, want := range map[string]int{"http://localhost:6789": 303, "http://evil.com": 403, "null": 403} {
		w := do(h, "POST", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) {
			local(r)
			r.Header.Set("Origin", origin)
			r.Body = io.NopCloser(strings.NewReader("username=a&password=b"))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		})
		if w.Code != want {
			t.Errorf("Origin %q: got %d want %d", origin, w.Code, want)
		}
	}
	if w := do(h, "GET", "/lab-counter", "127.0.0.1:5000", local); w.Code != 303 || w.Header().Get("Location") != "/lab-counter/login" {
		t.Errorf("unauthenticated: %d %s", w.Code, w.Header().Get("Location"))
	}
	login := func(user, pass string) *httptest.ResponseRecorder {
		return do(h, "POST", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) {
			local(r)
			r.Body = io.NopCloser(strings.NewReader(url.Values{"username": {user}, "password": {pass}}.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		})
	}
	bad := login("admin", "wrong")
	if bad.Header().Get("Location") != "/lab-counter/login?e=1" || cookieOf(bad, "fl_staff") != nil {
		t.Error("wrong password was accepted")
	}
	ok := login("admin", "admin")
	ck := cookieOf(ok, "fl_staff")
	if ck == nil || ok.Header().Get("Location") != "/lab-counter" {
		t.Fatal("admin/admin was rejected")
	}
	if w := do(h, "GET", "/lab-counter", "127.0.0.1:5000", func(r *http.Request) { local(r); r.AddCookie(ck) }); w.Code != 200 {
		t.Errorf("signed in: %d", w.Code)
	}
	// the open endpoint refuses paths outside the print directory
	w := do(h, "POST", "/lab-counter/open", "127.0.0.1:5000", func(r *http.Request) {
		local(r)
		r.AddCookie(ck)
		r.Body = io.NopCloser(strings.NewReader(`{"folder":"../../etc"}`))
	})
	if w.Code != 400 {
		t.Errorf("open outside dir: %d", w.Code)
	}
}

func TestNoFolderListingOrDownload(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	c := startSession(t, h, "A", "")
	upload(h, c, "secret.jpg", []byte("private"))
	for _, p := range []string{"/", "/Prints/", "/2026-10-11/A_1432/secret.jpg", "/drop/../secret.jpg", "/drop/assets/", "/drop/assets/../main.go"} {
		w := do(h, "GET", p, "192.168.1.9:5000", nil)
		if strings.Contains(w.Body.String(), "private") || w.Code == 200 {
			t.Errorf("%s leaked: %d", p, w.Code)
		}
	}
}

func TestPageAssetsAreServed(t *testing.T) {
	h := testServer(t).routes()
	for path, want := range map[string]string{
		"/drop/assets/hero.webp":           "image/webp",
		"/drop/assets/fonts/Kalpurush.ttf": "font/ttf",
		"/drop/assets/logo-192x192.png":    "image/png",
	} {
		w := do(h, "GET", path, "192.168.1.9:5000", nil)
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Errorf("%s: %d, %d bytes", path, w.Code, w.Body.Len())
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, want) {
			t.Errorf("%s: content type %q, want %q", path, ct, want)
		}
	}
	if w := do(h, "GET", "/drop/assets/shelf.jpg", "192.168.1.9:5000", nil); w.Code != 404 {
		t.Errorf("removed shelf.jpg should be gone: %d", w.Code)
	}
}

func TestReviewLink(t *testing.T) {
	for v, want := range map[string]bool{
		"": true,
		"https://g.page/r/CeQwDhZHiq-DEBM/review":    true,
		"https://maps.app.goo.gl/abc":                true,
		"http://g.page/r/x/review":                   false,
		"javascript:alert(1)":                        false,
		"https://evil.com/g.page":                    false,
		"https://g.page.evil.com/x":                  false,
		"https://user@g.page/x":                      false,
		"https://g.page/" + strings.Repeat("a", 500): false,
	} {
		if got := validReview(v); got != want {
			t.Errorf("validReview(%q) = %v, want %v", v, got, want)
		}
	}

	s := testServer(t)
	h := s.routes()
	local := func(r *http.Request) { r.Host = "localhost:6789" }
	login := do(h, "POST", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) {
		local(r)
		r.Body = io.NopCloser(strings.NewReader("username=admin&password=admin"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	})
	ck := cookieOf(login, "fl_staff")
	save := func(body string, c *http.Cookie) int {
		return do(h, "POST", "/lab-counter/review", "127.0.0.1:5000", func(r *http.Request) {
			local(r)
			if c != nil {
				r.AddCookie(c)
			}
			r.Body = io.NopCloser(strings.NewReader(body))
		}).Code
	}
	cfg := func() string {
		return strings.TrimSpace(do(h, "GET", "/drop/config", "192.168.1.9:5000", nil).Body.String())
	}
	if cfg() != `{"review":""}` {
		t.Errorf("default config: %s", cfg())
	}
	if c := save(`{"url":"https://g.page/r/abc/review"}`, nil); c != 401 {
		t.Errorf("save without login: %d", c)
	}
	if c := save(`{"url":"http://evil.com"}`, ck); c != 400 {
		t.Errorf("bad link: %d", c)
	}
	if c := save(`{"url":" https://g.page/r/abc/review "}`, ck); c != 204 {
		t.Fatalf("save: %d", c)
	}
	if cfg() != `{"review":"https://g.page/r/abc/review"}` {
		t.Errorf("config after save: %s", cfg())
	}
	if c := save(`{"url":""}`, ck); c != 204 || cfg() != `{"review":""}` {
		t.Errorf("clear: %d %s", c, cfg())
	}
}

func TestChangePassword(t *testing.T) {
	s := testServer(t)
	h := s.routes()
	local := func(r *http.Request) { r.Host = "localhost:6789" }
	login := func(pass string) *http.Cookie {
		return cookieOf(do(h, "POST", "/lab-counter/login", "127.0.0.1:5000", func(r *http.Request) {
			local(r)
			r.Body = io.NopCloser(strings.NewReader(url.Values{"username": {"admin"}, "password": {pass}}.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}), "fl_staff")
	}
	change := func(c *http.Cookie, cur, next string) int {
		return do(h, "POST", "/lab-counter/password", "127.0.0.1:5000", func(r *http.Request) {
			local(r)
			if c != nil {
				r.AddCookie(c)
			}
			r.Body = io.NopCloser(strings.NewReader(`{"current":"` + cur + `","new":"` + next + `"}`))
		}).Code
	}
	a, b := login("admin"), login("admin") // two signed-in sessions
	if a == nil || b == nil {
		t.Fatal("default login failed")
	}
	if c := change(nil, "admin", "longenough"); c != 401 {
		t.Errorf("without login: %d", c)
	}
	if c := change(a, "wrong", "longenough"); c != 403 {
		t.Errorf("wrong current: %d", c)
	}
	for _, bad := range []string{"short", "admin"} {
		if c := change(a, "admin", bad); c != 400 {
			t.Errorf("new %q: %d", bad, c)
		}
	}
	if c := change(a, "admin", "longenough"); c != 204 {
		t.Fatalf("change: %d", c)
	}
	if login("admin") != nil {
		t.Error("old password still works")
	}
	if login("longenough") == nil {
		t.Error("new password rejected")
	}
	get := func(c *http.Cookie) int {
		return do(h, "GET", "/lab-counter/api/info", "127.0.0.1:5000", func(r *http.Request) { local(r); r.AddCookie(c) }).Code
	}
	if get(a) != 200 {
		t.Error("the session that changed the password was signed out")
	}
	if get(b) != 401 {
		t.Error("another session survived the password change")
	}
	if w := do(h, "GET", "/lab-counter/api/info", "127.0.0.1:5000", func(r *http.Request) { local(r); r.AddCookie(a) }); !strings.Contains(w.Body.String(), `"defaultPass":false`) {
		t.Errorf("defaultPass: %s", w.Body.String())
	}
}
