package scrape

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/seneaLL/WTRTO/wt-wiki-scraper/internal/limitsdata"
)

var numberRe = regexp.MustCompile(`-?\d{1,3}(?:,\d{3})+(?:\.\d+)?|-?\d+(?:\.\d+)?`)

func ParseNumbers(s string) []float64 {
	matches := numberRe.FindAllString(s, -1)
	nums := make([]float64, 0, len(matches))
	for _, m := range matches {
		clean := strings.ReplaceAll(m, ",", "")
		if f, err := strconv.ParseFloat(clean, 64); err == nil {
			nums = append(nums, f)
		}
	}

	return nums
}

func ParseSingleNumber(s string) (float64, bool) {
	nums := ParseNumbers(s)
	if len(nums) == 0 {
		return 0, false
	}

	return nums[0], true
}

func ParseRangeValue(s string) (limitsdata.Range, bool) {
	nums := ParseNumbers(s)
	if len(nums) == 0 {
		return limitsdata.Range{}, false
	}

	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		min = math.Min(min, n)
		max = math.Max(max, n)
	}

	return limitsdata.Range{Min: min, Max: max}, true
}

func ParseGLimit(s string) (pos, neg float64, ok bool) {
	nums := ParseNumbers(s)
	if len(nums) < 2 {
		return 0, 0, false
	}

	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		min = math.Min(min, n)
		max = math.Max(max, n)
	}

	return max, min, true
}

func ParseFlapLimit(labelText, valueText string) (limitsdata.FlapLimits, string) {
	values := ParseNumbers(valueText)
	if len(values) == 0 {
		return limitsdata.FlapLimits{}, "no flap speed values found"
	}

	if labelText != "" {
		var labels []string
		for _, l := range strings.Split(labelText, "/") {
			if l = strings.ToUpper(strings.TrimSpace(l)); l != "" {
				labels = append(labels, l)
			}
		}

		if len(labels) == len(values) {
			byLabel := make(map[string]float64, len(labels))
			for i, l := range labels {
				byLabel[l] = values[i]
			}
			landing, hasLanding := byLabel["L"]
			takeoff, hasTakeoff := byLabel["T"]
			combat, hasCombat := byLabel["C"]

			if hasLanding || hasTakeoff || hasCombat {
				takeoffLanding := values[0]
				switch {
				case hasLanding && hasTakeoff:
					takeoffLanding = math.Min(landing, takeoff)
				case hasLanding:
					takeoffLanding = landing
				case hasTakeoff:
					takeoffLanding = takeoff
				}

				c := values[len(values)-1]
				if hasCombat {
					c = combat
				}

				return limitsdata.FlapLimits{TakeoffLanding: takeoffLanding, Combat: c}, ""
			}
		}

		return limitsdata.FlapLimits{TakeoffLanding: values[0], Combat: values[len(values)-1]},
			fmt.Sprintf("unrecognized flap labels %q for values %q, falling back to first/last", labelText, valueText)
	}

	if len(values) == 1 {
		return limitsdata.FlapLimits{TakeoffLanding: values[0], Combat: values[0]}, "single flap speed value, applied to both buckets"
	}

	return limitsdata.FlapLimits{TakeoffLanding: values[0], Combat: values[len(values)-1]}, ""
}
