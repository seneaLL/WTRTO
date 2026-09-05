package scrape

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func FetchAviationSlugs(ctx context.Context, listURL string, delay time.Duration) ([]string, error) {
	body, err := Get(ctx, listURL, delay)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	doc.Find("a.wt-tree_item-link").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || !strings.HasPrefix(href, "/unit/") {
			return
		}
		if slug := strings.TrimSpace(strings.TrimPrefix(href, "/unit/")); slug != "" {
			seen[slug] = struct{}{}
		}
	})

	slugs := make([]string, 0, len(seen))
	for s := range seen {
		slugs = append(slugs, s)
	}
	sort.Strings(slugs)

	return slugs, nil
}
