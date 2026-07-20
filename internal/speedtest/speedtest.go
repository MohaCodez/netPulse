package speedtest

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Result holds speed test measurements.
type Result struct {
	DownloadMbps float64
	UploadMbps   float64
	LatencyMs    float64
	JitterMs     float64
	Server       string
	Duration     time.Duration
	Timestamp    time.Time
}

// TestConfig configures a speed test run.
type TestConfig struct {
	DownloadURLs []string
	UploadURLs   []string // multiple fallbacks
	Duration     time.Duration
	Connections  int
}

// DefaultConfig returns default speed test configuration.
func DefaultConfig() *TestConfig {
	return &TestConfig{
		DownloadURLs: []string{
			"https://releases.ubuntu.com/22.04.5/ubuntu-22.04.5-desktop-amd64.iso",
			"https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.tar.xz",
		},
		UploadURLs: []string{
			"https://speed.cloudflare.com/__up",
			"https://httpbin.org/post",
		},
		Duration:    10 * time.Second,
		Connections: 6,
	}
}

// Runner manages scheduled and on-demand speed tests.
type Runner struct {
	config   *TestConfig
	interval time.Duration
	handler  func(*Result)

	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	last    *Result
}

// NewRunner creates a speed test runner.
func NewRunner(config *TestConfig, interval time.Duration, handler func(*Result)) *Runner {
	if config == nil {
		config = DefaultConfig()
	}
	return &Runner{
		config:   config,
		interval: interval,
		handler:  handler,
	}
}

// Start begins scheduled speed tests.
func (r *Runner) Start() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.mu.Unlock()

	go r.loop(ctx)
	log.Printf("[speedtest] runner started, interval=%s", r.interval)
}

// Stop halts scheduled speed tests.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.running {
		return
	}
	r.running = false
	r.cancel()
	log.Println("[speedtest] runner stopped")
}

// RunNow triggers an immediate speed test.
func (r *Runner) RunNow(ctx context.Context) (*Result, error) {
	return r.runTest(ctx)
}

// LastResult returns the most recent speed test result.
func (r *Runner) LastResult() *Result {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.last
}

func (r *Runner) loop(ctx context.Context) {
	select {
	case <-time.After(30 * time.Second):
	case <-ctx.Done():
		return
	}

	r.runAndStore(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runAndStore(ctx)
		}
	}
}

func (r *Runner) runAndStore(ctx context.Context) {
	result, err := r.runTest(ctx)
	if err != nil {
		log.Printf("[speedtest] test failed: %v", err)
		return
	}

	r.mu.Lock()
	r.last = result
	r.mu.Unlock()

	if r.handler != nil {
		r.handler(result)
	}
}

func (r *Runner) runTest(ctx context.Context) (*Result, error) {
	log.Println("[speedtest] starting test...")

	result := &Result{
		Timestamp: time.Now(),
	}

	// Measure latency
	latency, jitter, server, err := r.measureLatency(ctx)
	if err != nil {
		return nil, fmt.Errorf("latency measurement: %w", err)
	}
	result.LatencyMs = latency
	result.JitterMs = jitter
	result.Server = server

	// Download test
	downloadMbps, err := r.measureDownload(ctx)
	if err != nil {
		return nil, fmt.Errorf("download measurement: %w", err)
	}
	result.DownloadMbps = downloadMbps

	// Upload test (shorter duration to stay within total time budget)
	uploadMbps, err := r.measureUpload(ctx)
	if err != nil {
		log.Printf("[speedtest] upload failed (non-fatal): %v", err)
	} else {
		result.UploadMbps = uploadMbps
	}

	log.Printf("[speedtest] complete: ↓ %.1f Mbps | ↑ %.1f Mbps | latency: %.0fms | jitter: %.0fms",
		result.DownloadMbps, result.UploadMbps, result.LatencyMs, result.JitterMs)

	return result, nil
}

func (r *Runner) measureLatency(ctx context.Context) (latency, jitter float64, server string, err error) {
	if len(r.config.DownloadURLs) == 0 {
		return 0, 0, "", fmt.Errorf("no test URLs configured")
	}

	server = r.config.DownloadURLs[0]
	var latencies []float64
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < 5; i++ {
		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, "HEAD", server, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			// Try GET if HEAD fails
			req, _ = http.NewRequestWithContext(ctx, "GET", server, nil)
			start = time.Now()
			resp, err = client.Do(req)
			if err != nil {
				continue
			}
		}
		resp.Body.Close()
		elapsed := float64(time.Since(start)) / float64(time.Millisecond)
		latencies = append(latencies, elapsed)
	}

	if len(latencies) == 0 {
		return 0, 0, server, fmt.Errorf("all latency probes failed")
	}

	var sum float64
	for _, l := range latencies {
		sum += l
	}
	latency = sum / float64(len(latencies))

	var jitterSum float64
	for _, l := range latencies {
		diff := l - latency
		jitterSum += diff * diff
	}
	if len(latencies) > 1 {
		jitter = sqrt(jitterSum / float64(len(latencies)-1))
	}

	return latency, jitter, server, nil
}

func (r *Runner) measureDownload(ctx context.Context) (float64, error) {
	if len(r.config.DownloadURLs) == 0 {
		return 0, fmt.Errorf("no download URLs configured")
	}

	ctx, cancel := context.WithTimeout(ctx, r.config.Duration+5*time.Second)
	defer cancel()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	start := time.Now()
	deadline := start.Add(r.config.Duration)

	for i := 0; i < r.config.Connections; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			url := r.config.DownloadURLs[idx%len(r.config.DownloadURLs)]

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
				return
			}

			buf := make([]byte, 32*1024)
			for time.Now().Before(deadline) {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					mu.Lock()
					totalBytes += int64(n)
					mu.Unlock()
				}
				if err != nil {
					break
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	if totalBytes == 0 {
		return 0, fmt.Errorf("no data downloaded")
	}

	bytesPerSec := float64(totalBytes) / elapsed.Seconds()
	return (bytesPerSec * 8) / 1_000_000, nil
}

func (r *Runner) measureUpload(ctx context.Context) (float64, error) {
	if len(r.config.UploadURLs) == 0 {
		return 0, fmt.Errorf("no upload URL configured")
	}

	// Try each upload URL until one works
	for _, uploadURL := range r.config.UploadURLs {
		mbps, err := r.doUpload(ctx, uploadURL)
		if err == nil && mbps > 0 {
			return mbps, nil
		}
	}
	return 0, fmt.Errorf("all upload endpoints failed")
}

func (r *Runner) doUpload(ctx context.Context, uploadURL string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Duration+5*time.Second)
	defer cancel()

	// Use smaller chunks to avoid rate limiting (100KB each)
	dataSize := 100 * 1024
	data := make([]byte, dataSize)
	rand.Read(data)

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	start := time.Now()
	deadline := start.Add(r.config.Duration)

	// Use fewer connections for upload (servers are more restrictive)
	uploadConns := 3
	if r.config.Connections < uploadConns {
		uploadConns = r.config.Connections
	}

	for i := 0; i < uploadConns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: r.config.Duration + 5*time.Second}

			for time.Now().Before(deadline) {
				reader := bytes.NewReader(data)
				req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, reader)
				if err != nil {
					return
				}
				req.Header.Set("Content-Type", "application/octet-stream")

				resp, err := client.Do(req)
				if err != nil {
					return
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
					mu.Lock()
					totalBytes += int64(dataSize)
					mu.Unlock()
				} else {
					// Server rejected — stop trying
					return
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if totalBytes == 0 {
		return 0, fmt.Errorf("upload rejected by server")
	}

	bytesPerSec := float64(totalBytes) / elapsed.Seconds()
	return (bytesPerSec * 8) / 1_000_000, nil
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 20; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}
