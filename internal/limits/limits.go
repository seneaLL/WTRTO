package limits

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const remoteURL = "https://raw.githubusercontent.com/seneaLL/WTRTO/master/data/aircraft_limits.json"

type FlapLimits struct {
	TakeoffLanding float64 `json:"takeoff_landing"`
	Combat         float64 `json:"combat"`
}

type Range struct {
	Min float64
	Max float64
}

func (r *Range) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		r.Min, r.Max = f, f

		return nil
	}

	var obj struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	r.Min, r.Max = obj.Min, obj.Max

	return nil
}

func (r Range) At(sweep float64) float64 {
	if sweep < 0 {
		sweep = 0
	}
	if sweep > 1 {
		sweep = 1
	}

	return r.Min + (r.Max-r.Min)*sweep
}

type Aircraft struct {
	MaxSpeedIASKmh Range      `json:"max_speed_ias_kmh"`
	MachLimit      Range      `json:"mach_limit"`
	GLimitPos      float64    `json:"g_limit_pos"`
	GLimitNeg      float64    `json:"g_limit_neg"`
	FlapSpeedKmh   FlapLimits `json:"flap_speed_kmh"`
	GearSpeedKmh   float64    `json:"gear_speed_kmh"`
}

type file struct {
	Version  string              `json:"version"`
	Aircraft map[string]Aircraft `json:"aircraft"`
}

var (
	mu      sync.RWMutex
	current file
)

func filePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(exe), "aircraft_limits.json"), nil
}

func Load() {
	p, err := filePath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return
	}

	mu.Lock()
	current = f
	mu.Unlock()
}

func Version() string {
	mu.RLock()
	defer mu.RUnlock()

	return current.Version
}

func Get(aircraftType string) (Aircraft, bool) {
	mu.RLock()
	defer mu.RUnlock()

	a, ok := current.Aircraft[aircraftType]

	return a, ok
}

func CheckAndUpdate(ctx context.Context) (updated bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "wtrto-limits-check")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("limits: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var remote file
	if err := json.Unmarshal(body, &remote); err != nil {
		return false, fmt.Errorf("limits: invalid remote file: %w", err)
	}

	mu.RLock()
	previousVersion := current.Version
	mu.RUnlock()

	p, err := filePath()
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(p, body, 0o644); err != nil {
		return false, err
	}

	Load()

	return previousVersion != remote.Version, nil
}
