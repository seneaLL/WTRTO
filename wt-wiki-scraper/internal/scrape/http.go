package scrape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const userAgent = "wtrto-limits-scraper/1.0 (+https://github.com/seneaLL/WTRTO)"

var (
	rateMu        sync.Mutex
	lastRequestAt time.Time
)

func politeWait(delay time.Duration) {
	rateMu.Lock()
	defer rateMu.Unlock()

	if wait := time.Until(lastRequestAt.Add(delay)); wait > 0 {
		time.Sleep(wait)
	}
	lastRequestAt = time.Now()
}

func Get(ctx context.Context, url string, delay time.Duration) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= 2; attempt++ {
		politeWait(delay)

		body, err := get(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
	}

	return "", lastErr
}

func get(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %s for %s", resp.Status, url)
	}

	return string(body), nil
}
