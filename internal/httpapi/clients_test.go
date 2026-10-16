package httpapi

import (
	"bufio"
	"net/http"
	"testing"
	"time"
)

func TestDescribeDevice(t *testing.T) {
	cases := []struct {
		ua, kind, os, browser string
		projector             bool
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1", "phone", "iPhone", "Safari", false},
		{"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0 Mobile/15E148 Safari/604.1", "tablet", "iPad", "Chrome", false},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36", "phone", "Android", "Chrome", false},
		{"Mozilla/5.0 (Linux; Android 13; SM-X700) AppleWebKit/537.36 Chrome/124.0 Safari/537.36", "tablet", "Android", "Chrome", false},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0", "computer", "Windows", "Edge", false},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Gecko/20100101 Firefox/125.0", "computer", "Mac", "Firefox", false},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36", "computer", "Linux", "Chrome", false},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/17.0 Safari/605.1.15", "projector", "Mac", "Safari", true},
		{"curl/8.0", "other", "", "", false},
		{"", "other", "", "", false},
	}
	for _, c := range cases {
		k, o, b := describeDevice(c.ua, c.projector)
		if k != c.kind || o != c.os || b != c.browser {
			t.Errorf("%q: %s %s %s, want %s %s %s", c.ua, k, o, b, c.kind, c.os, c.browser)
		}
	}
}

// The Setup and operator pages can list who is connected: device, address and language.
func TestConnectedDevicesAreListed(t *testing.T) {
	s, ts, c := standbyServer(t)
	_, r := c.do("POST", "/api/admin/shows", map[string]any{"sample": true})
	c.do("POST", "/api/admin/shows/"+r["id"].(string)+"/activate", map[string]any{"start": true})
	open := func(path, lang, ua string) *http.Response {
		req, _ := http.NewRequest("GET", ts.URL+path, nil)
		req.Header.Set("User-Agent", ua)
		if lang != "" {
			req.Header.Set("Accept-Language", lang)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		bufio.NewReader(resp.Body).ReadString('\n') // the stream has started
		return resp
	}
	a := open("/events", "en", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) Version/17.4 Mobile/15E148 Safari/604.1")
	defer a.Body.Close()
	b := open("/events?display=projector", "es", "Mozilla/5.0 (Windows NT 10.0) Chrome/124.0 Safari/537.36")
	defer b.Body.Close()
	time.Sleep(100 * time.Millisecond)
	rows := s.rt.Clients()
	if len(rows) != 2 {
		t.Fatalf("%+v", rows)
	}
	if rows[0].Lang != "en" || rows[0].Kind != "phone" || rows[0].OS != "iPhone" || rows[0].IP != "127.0.0.1" || rows[0].LangName != "English" {
		t.Errorf("first: %+v", rows[0])
	}
	if rows[1].Lang != "es" || rows[1].Kind != "projector" || rows[1].Browser != "Chrome" {
		t.Errorf("second: %+v", rows[1])
	}
	code, body := c.do("GET", "/api/operator/clients", nil)
	if code != 200 || body["total"].(float64) != 2 {
		t.Errorf("the API: %d %v", code, body)
	}
}
