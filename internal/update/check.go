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
	repoOwner     = "seneaLL"
	repoName      = "WTRTO"
	defaultBranch = "master"
)

type Result struct {
	Available bool
	LatestSHA string
}

type compareFile struct {
	Filename string `json:"filename"`
}

type compareCommit struct {
	SHA string `json:"sha"`
}

type compareResult struct {
	Status  string          `json:"status"`
	AheadBy int             `json:"ahead_by"`
	Commits []compareCommit `json:"commits"`
	Files   []compareFile   `json:"files"`
}

func Check(ctx context.Context, localSHA string) (Result, error) {
	localSHA = strings.TrimSpace(localSHA)
	if localSHA == "" || localSHA == "dev" {
		return Result{}, fmt.Errorf("update: build has no embedded commit SHA")
	}

	compareURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/compare/%s...%s",
		repoOwner, repoName, localSHA, defaultBranch,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, compareURL, nil)
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

	var cmp compareResult
	if err := json.NewDecoder(resp.Body).Decode(&cmp); err != nil {
		return Result{}, fmt.Errorf("update: unexpected GitHub API response")
	}

	if cmp.AheadBy == 0 || len(cmp.Commits) == 0 {
		return Result{Available: false}, nil
	}

	appCodeChanged := false
	for _, f := range cmp.Files {
		if !strings.HasPrefix(f.Filename, "data/") {
			appCodeChanged = true

			break
		}
	}
	if !appCodeChanged {
		return Result{Available: false}, nil
	}

	latest := cmp.Commits[len(cmp.Commits)-1].SHA
	short := latest
	if len(short) > 7 {
		short = short[:7]
	}

	return Result{Available: true, LatestSHA: short}, nil
}
