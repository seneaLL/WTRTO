package limitsdata

import "encoding/json"

type FlapLimits struct {
	TakeoffLanding float64 `json:"takeoff_landing"`
	Combat         float64 `json:"combat"`
}

type Range struct {
	Min float64
	Max float64
}

func (r Range) MarshalJSON() ([]byte, error) {
	if r.Min == r.Max {
		return json.Marshal(r.Min)
	}

	return json.Marshal(struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}{r.Min, r.Max})
}

type Aircraft struct {
	MaxSpeedIASKmh Range      `json:"max_speed_ias_kmh"`
	MachLimit      Range      `json:"mach_limit"`
	GLimitPos      float64    `json:"g_limit_pos"`
	GLimitNeg      float64    `json:"g_limit_neg"`
	FlapSpeedKmh   FlapLimits `json:"flap_speed_kmh"`
	GearSpeedKmh   float64    `json:"gear_speed_kmh"`
}

type File struct {
	Version  string              `json:"version"`
	Aircraft map[string]Aircraft `json:"aircraft"`
}
