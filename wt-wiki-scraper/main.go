package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seneaLL/WTRTO/wt-wiki-scraper/internal/limitsdata"
	"github.com/seneaLL/WTRTO/wt-wiki-scraper/internal/scrape"
)

func main() {
	listURL := flag.String("list-url", "https://wiki.warthunder.com/aviation?v=l", "aviation list page")
	out := flag.String("out", "../data/aircraft_limits.generated.json", "output file")
	concurrency := flag.Int("concurrency", 3, "concurrent requests")
	delay := flag.Duration("delay", 300*time.Millisecond, "minimum delay between requests")
	limit := flag.Int("limit", 0, "process only first N aircraft (0 = all)")
	only := flag.String("only", "", "comma-separated list of slugs to process instead of the full list")
	flag.Parse()

	ctx := context.Background()

	slugs, err := resolveSlugs(ctx, *only, *listURL, *delay)
	if err != nil {
		log.Fatalf("resolve aircraft list: %v", err)
	}

	targets := slugs
	if *limit > 0 && *limit < len(targets) {
		targets = targets[:*limit]
	}
	fmt.Printf("Found %d aircraft, processing %d\n", len(slugs), len(targets))

	results, fetchErrs := parseAll(ctx, targets, *concurrency, *delay)

	file := limitsdata.File{Version: time.Now().UTC().Format("2006-01-02"), Aircraft: make(map[string]limitsdata.Aircraft)}
	skipped := 0
	for i, res := range results {
		if fetchErrs[i] != nil || res.Limits == nil {
			skipped++

			continue
		}

		file.Aircraft[res.Slug] = *res.Limits
	}

	if err := writeFile(*out, file); err != nil {
		log.Fatalf("write output: %v", err)
	}

	fmt.Printf("Wrote %d aircraft to %s (%d skipped, no usable Limits section)\n", len(file.Aircraft), *out, skipped)
}

func resolveSlugs(ctx context.Context, only, listURL string, delay time.Duration) ([]string, error) {
	if only == "" {
		return scrape.FetchAviationSlugs(ctx, listURL, delay)
	}

	var slugs []string
	for _, s := range strings.Split(only, ",") {
		if s = strings.TrimSpace(s); s != "" {
			slugs = append(slugs, s)
		}
	}

	return slugs, nil
}

func parseAll(ctx context.Context, targets []string, concurrency int, delay time.Duration) ([]scrape.UnitResult, []error) {
	results := make([]scrape.UnitResult, len(targets))
	errs := make([]error, len(targets))
	total := len(targets)
	var done int64

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				slug := targets[i]
				log.Printf("parsing %s...", slug)

				res, err := scrape.ParseUnit(ctx, slug, delay)
				results[i], errs[i] = res, err

				n := atomic.AddInt64(&done, 1)
				switch {
				case err != nil:
					log.Printf("[%d/%d] %s: FAILED (%v)", n, total, slug, err)
				case res.Limits == nil:
					log.Printf("[%d/%d] %s: SKIPPED (%s)", n, total, slug, res.Warning)
				case res.Warning != "":
					log.Printf("[%d/%d] %s: OK, with warning (%s)", n, total, slug, res.Warning)
				default:
					log.Printf("[%d/%d] %s: OK", n, total, slug)
				}
			}
		}()
	}
	for i := range targets {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	return results, errs
}

func writeFile(path string, file limitsdata.File) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(absPath, append(data, '\n'), 0o644)
}
