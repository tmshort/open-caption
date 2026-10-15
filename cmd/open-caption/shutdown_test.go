package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestInterruptShutsDownCleanly starts the real program, puts a show on air, interrupts it the way the
// operating system does (Ctrl+C / Ctrl+Break / SIGINT) and checks that it saves and exits with status 0.
// It exercises the signal handling on every platform, including Windows.
func TestInterruptShutsDownCleanly(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "open-caption")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	data := filepath.Join(dir, "data")
	os.MkdirAll(data, 0o755)
	os.WriteFile(filepath.Join(data, "config.yaml"), []byte("admin_password: adminpw1\noperator_password: operatorpw1\n"), 0o600)
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := l.Addr().String()
	l.Close()

	var out bytes.Buffer
	cmd := exec.Command(bin, "--data", data, "--http-addr", addr, "--no-browser")
	cmd.Stdout, cmd.Stderr = &out, &out
	prepareInterruptible(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	t.Cleanup(func() { cmd.Process.Kill() })

	base := "http://" + addr
	jar, _ := cookiejar.New(nil)
	hc := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for {
		if r, err := hc.Get(base + "/healthz"); err == nil {
			r.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("did not start:\n%s", out.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
	post := func(path string, body any) map[string]any {
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", base+path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		r, err := hc.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var m map[string]any
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &m)
		if r.StatusCode >= 300 {
			t.Fatalf("%s: %d %s", path, r.StatusCode, raw)
		}
		return m
	}
	post("/api/login", map[string]string{"role": "admin", "password": "adminpw1"})
	show := post("/api/admin/shows", map[string]any{"sample": true})
	post("/api/admin/shows/"+fmt.Sprint(show["id"])+"/activate", map[string]any{"force": true})
	post("/api/operator/claim", map[string]bool{"takeover": true})
	for i := 0; i < 3; i++ {
		post("/api/operator/next", map[string]any{})
	}

	if err := interrupt(cmd); err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("cannot interrupt the program: %v", err) // CI is where this must work
		}
		t.Skipf("this environment cannot send an interrupt to a child process: %v", err)
	}
	select {
	case err := <-exited:
		if err != nil {
			t.Fatalf("exit: %v\n%s", err, out.String())
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("did not exit after the interrupt:\n%s", out.String())
	}
	log := out.String()
	for _, want := range []string{"shutting down", "stopped"} {
		if !strings.Contains(log, want) {
			t.Errorf("the log lacks %q:\n%s", want, log)
		}
	}
	live, err := os.ReadFile(filepath.Join(data, "live.json"))
	if err != nil || !strings.Contains(string(live), "cursor") {
		t.Errorf("the show position was not saved: %v %s", err, live)
	}
}
