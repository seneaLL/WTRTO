package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	repoOwner = "seneaLL"
	repoName  = "WTRTO"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/commits?per_page=1"
)

type Result struct {
	Available bool
	LatestSHA string
}

func Check(ctx context.Context, localSHA string) (Result, error) {
	localSHA = strings.TrimSpace(localSHA)
	if localSHA == "" || localSHA == "dev" {
		return Result{}, fmt.Errorf("update: build has no embedded commit SHA")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "wtrto-update-check")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("update: GitHub API returned %s", resp.Status)
	}

	var commits []struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
		return Result{}, fmt.Errorf("update: unexpected GitHub API response")
	}

	latest := commits[0].SHA
	short := latest
	if len(short) > 7 {
		short = short[:7]
	}

	same := strings.HasPrefix(latest, localSHA) || strings.HasPrefix(localSHA, latest)

	return Result{Available: !same, LatestSHA: short}, nil
}
