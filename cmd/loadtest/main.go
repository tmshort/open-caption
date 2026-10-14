// Command loadtest opens many audience (SSE) connections against a server
// started with --dev and measures operator-to-client caption latency
// (docs/07-testing-and-acceptance.md, performance targets).
//
//	open-caption --dev --script big.md --http-addr :18080 &
//	loadtest -url http://localhost:18080 -clients 200 -messages 60
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		base     = flag.String("url", "http://localhost:8080", "server base URL")
		clients  = flag.Int("clients", 200, "concurrent audience connections")
		messages = flag.Int("messages", 60, "captions to send (script must have at least this many)")
		interval = flag.Duration("interval", 250*time.Millisecond, "time between NEXT commands")
	)
	flag.Parse()

	tr := &http.Transport{MaxIdleConnsPerHost: *clients + 8, MaxConnsPerHost: 0}
	hc := &http.Client{Transport: tr}

	var (
		wg        sync.WaitGroup
		ready     atomic.Int64
		failed    atomic.Int64
		rejected  atomic.Int64
		mu        sync.Mutex
		latencies []float64
		received  = make([]int, *clients)
	)
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := hc.Get(*base + "/events")
			if err != nil {
				failed.Add(1)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == 503 {
				rejected.Add(1)
				return
			}
			sc := bufio.NewScanner(resp.Body)
			sc.Buffer(make([]byte, 64*1024), 1<<20)
			var typ string
			var local []float64
			for sc.Scan() {
				line := sc.Text()
				switch {
				case strings.HasPrefix(line, "event: "):
					typ = line[7:]
				case strings.HasPrefix(line, "data: "):
					now := time.Now().UnixMilli()
					switch typ {
					case "snapshot":
						ready.Add(1)
					case "caption":
						received[i]++
						if ts, ok := field(line, `"ts":`); ok {
							local = append(local, float64(now-ts))
						}
						if received[i] >= *messages {
							mu.Lock()
							latencies = append(latencies, local...)
							mu.Unlock()
							return
						}
					}
				}
			}
			mu.Lock()
			latencies = append(latencies, local...)
			mu.Unlock()
		}(i)
	}

	deadline := time.Now().Add(20 * time.Second)
	for int(ready.Load()+failed.Load()+rejected.Load()) < *clients && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("connected %d/%d (failed %d, rejected 503: %d)\n", ready.Load(), *clients, failed.Load(), rejected.Load())

	for i := 0; i < *messages; i++ {
		resp, err := hc.Post(*base+"/dev/next", "", nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "next:", err)
			os.Exit(1)
		}
		resp.Body.Close()
		time.Sleep(*interval)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
	}

	complete, missing := 0, 0
	for i, n := range received {
		_ = i
		if n >= *messages {
			complete++
		} else {
			missing += *messages - n
		}
	}
	sort.Float64s(latencies)
	pct := func(p float64) float64 {
		if len(latencies) == 0 {
			return 0
		}
		return latencies[int(float64(len(latencies)-1)*p)]
	}
	fmt.Printf("clients that received all %d captions: %d/%d (missing deliveries: %d)\n", *messages, complete, ready.Load(), missing)
	fmt.Printf("latency ms  p50=%.0f  p95=%.0f  p99=%.0f  max=%.0f  (n=%d)\n", pct(.5), pct(.95), pct(.99), pct(1), len(latencies))
	if complete != int(ready.Load()) || pct(.95) > 250 {
		fmt.Println("RESULT: FAIL")
		os.Exit(2)
	}
	fmt.Println("RESULT: PASS")
}

func field(line, key string) (int64, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return 0, false
	}
	rest := line[i+len(key):]
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	n, err := strconv.ParseInt(rest[:j], 10, 64)
	return n, err == nil
}
