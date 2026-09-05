package scrape

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/seneaLL/WTRTO/wt-wiki-scraper/internal/limitsdata"
)

func unitURL(slug string) string {
	return "https://wiki.warthunder.com/unit/" + slug
}

type block struct {
	value string
	info  string
}

func findBlock(doc *goquery.Document, header string) (block, bool) {
	var found block
	var ok bool

	doc.Find(".game-unit_chars-block").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		headerText := strings.TrimSpace(s.Find(".game-unit_chars-header").First().Text())
		if headerText != header {
			return true
		}

		info := strings.TrimSpace(s.Find(".game-unit_chars-info").First().Text())
		value := strings.TrimSpace(s.Find(".game-unit_chars-line > .game-unit_chars-value").First().Text())
		if value == "" {
			value = strings.TrimSpace(s.Find(".game-unit_chars-subline .game-unit_chars-value").First().Text())
		}

		found = block{value: value, info: info}
		ok = true

		return false
	})

	return found, ok
}

type UnitResult struct {
	Slug    string
	Limits  *limitsdata.Aircraft
	Warning string
}

func ParseUnit(ctx context.Context, slug string, delay time.Duration) (UnitResult, error) {
	body, err := Get(ctx, unitURL(slug), delay)
	if err != nil {
		return UnitResult{Slug: slug}, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return UnitResult{Slug: slug}, err
	}

	hasLimits := false
	doc.Find(".form-text").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if strings.TrimSpace(s.Text()) == "Limits:" {
			hasLimits = true

			return false
		}

		return true
	})
	if !hasLimits {
		return UnitResult{Slug: slug, Warning: "no Limits section on page"}, nil
	}

	maxSpeed, hasMaxSpeed := findBlock(doc, "Max Speed Limit (IAS)")
	mach, hasMach := findBlock(doc, "Mach Number Limit")
	gLoad, hasG := findBlock(doc, "G limit")
	flap, hasFlap := findBlock(doc, "Flap Speed Limit (IAS)")
	gear, hasGear := findBlock(doc, "Gear Speed Limit (IAS)")

	if !hasMaxSpeed && !hasMach && !hasG && !hasFlap && !hasGear {
		return UnitResult{Slug: slug, Warning: "Limits section present but no known rows matched"}, nil
	}

	var warnings []string
	aircraft := limitsdata.Aircraft{}

	if hasMaxSpeed {
		if r, ok := ParseRangeValue(maxSpeed.value); ok {
			aircraft.MaxSpeedIASKmh = r
		} else {
			warnings = append(warnings, "could not parse max speed limit")
		}
	}
	if hasMach {
		if r, ok := ParseRangeValue(mach.value); ok {
			aircraft.MachLimit = r
		} else {
			warnings = append(warnings, "could not parse mach limit")
		}
	}
	if hasG {
		if pos, neg, ok := ParseGLimit(gLoad.value); ok {
			aircraft.GLimitPos, aircraft.GLimitNeg = pos, neg
		} else {
			warnings = append(warnings, "could not parse g limit")
		}
	}
	if hasFlap {
		fl, warn := ParseFlapLimit(flap.info, flap.value)
		aircraft.FlapSpeedKmh = fl
		if warn != "" {
			warnings = append(warnings, warn)
		}
	}
	if hasGear {
		if v, ok := ParseSingleNumber(gear.value); ok {
			aircraft.GearSpeedKmh = v
		}
	}

	return UnitResult{Slug: slug, Limits: &aircraft, Warning: strings.Join(warnings, "; ")}, nil
}
