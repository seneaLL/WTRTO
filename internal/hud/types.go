package hud

import "fmt"

type Binding string

const (
	BindThrottlePct Binding = "throttle_pct"
	BindIAS         Binding = "ias_kmh"
	BindTAS         Binding = "tas_kmh"
	BindMach        Binding = "mach"
	BindAltitude    Binding = "altitude_m"
	BindFuelKg      Binding = "fuel_kg"
	BindFuelTime    Binding = "fuel_time"
	BindFuelRate    Binding = "fuel_rate"
	BindOilTemp1    Binding = "oil_temp_1"
	BindOilTemp2    Binding = "oil_temp_2"
	BindCompass     Binding = "compass_deg"
	BindAoA         Binding = "aoa_deg"
	BindAoS         Binding = "aos_deg"
	BindGLoad       Binding = "g_load"
	BindVSpeed      Binding = "vspeed_ms"
	BindIASRate     Binding = "accel_kmh_s"

	BindAileron  Binding = "aileron_pct"
	BindElevator Binding = "elevator_pct"
	BindRudder   Binding = "rudder_pct"
	BindFlaps    Binding = "flaps_pct"
	BindGearPct  Binding = "gear_pct"
	BindRollRate Binding = "roll_rate_deg_s"
	BindFuelPct  Binding = "fuel_pct"
	BindTrimmer  Binding = "trimmer_pct"
	BindRadioAlt Binding = "radio_altitude_m"
	BindTurn     Binding = "turn"

	BindWingSweep Binding = "wing_sweep_pct"
)

const MaxEngines = 8

type engineMetric struct {
	name     string
	format   string
	labelKey string
	get      func(v Values, i int) float64
}

var engineMetrics = []engineMetric{
	{"throttle", "throttle_%d_pct", "editor.metric.throttle", func(v Values, i int) float64 { return v.EngineThrottle[i] }},
	{"rpm", "rpm_%d", "editor.metric.rpm", func(v Values, i int) float64 { return v.EngineRPM[i] }},
	{"manifold", "manifold_pressure_%d_atm", "editor.metric.manifold", func(v Values, i int) float64 { return v.EngineManifold[i] }},
	{"oil_temp", "oil_temp_%d", "editor.metric.oil_temp", func(v Values, i int) float64 { return v.EngineOilTemp[i] }},
	{"water_temp", "water_temp_%d", "editor.metric.water_temp", func(v Values, i int) float64 { return v.EngineWaterTemp[i] }},
	{"power", "power_%d_hp", "editor.metric.power", func(v Values, i int) float64 { return v.EnginePower[i] }},
	{"thrust", "thrust_%d_kgs", "editor.metric.thrust", func(v Values, i int) float64 { return v.EngineThrust[i] }},
	{"efficiency", "efficiency_%d_pct", "editor.metric.efficiency", func(v Values, i int) float64 { return v.EngineEfficiency[i] }},
	{"prop_pitch", "prop_pitch_%d_deg", "editor.metric.prop_pitch", func(v Values, i int) float64 { return v.EnginePropPitch[i] }},
}

func engineBinding(m engineMetric, n int) Binding {
	return Binding(fmt.Sprintf(m.format, n))
}

var AllBindings = buildAllBindings()

func buildAllBindings() []Binding {
	list := []Binding{
		BindThrottlePct, BindIAS, BindTAS, BindMach, BindAltitude,
		BindFuelKg, BindFuelTime, BindFuelRate, BindOilTemp1, BindOilTemp2,
		BindCompass, BindAoA, BindAoS, BindGLoad, BindVSpeed, BindIASRate,

		BindAileron, BindElevator, BindRudder, BindFlaps, BindGearPct,
		BindRollRate, BindFuelPct, BindTrimmer, BindRadioAlt, BindTurn,
	}

	for _, m := range engineMetrics {
		start := 1
		if m.name == "oil_temp" {
			start = 3
		}
		for n := start; n <= 4; n++ {
			list = append(list, engineBinding(m, n))
		}
	}

	list = append(list, BindWingSweep)

	for n := 5; n <= MaxEngines; n++ {
		for _, m := range engineMetrics {
			list = append(list, engineBinding(m, n))
		}
	}

	return list
}

type engineBindingRef struct {
	metric engineMetric
	n      int
}

var engineBindingIndex = buildEngineBindingIndex()

func buildEngineBindingIndex() map[Binding]engineBindingRef {
	idx := make(map[Binding]engineBindingRef, MaxEngines*len(engineMetrics))
	for n := 1; n <= MaxEngines; n++ {
		for _, m := range engineMetrics {
			idx[engineBinding(m, n)] = engineBindingRef{metric: m, n: n}
		}
	}

	return idx
}

type Style string

const (
	StyleStraight Style = "straight"
	StyleArc      Style = "arc"
)

type Direction string

const (
	DirUp   Direction = "up"
	DirDown Direction = "down"
	DirCW   Direction = "cw"
	DirCCW  Direction = "ccw"
)

type LabelSide string

const (
	SideAuto  LabelSide = "auto"
	SideLeft  LabelSide = "left"
	SideRight LabelSide = "right"
)

type Zone struct {
	Threshold float64 `json:"threshold"`
	Color     Color   `json:"color"`
}

type ElementKind string

const (
	KindText    ElementKind = "text"
	KindHorizon ElementKind = "horizon"
	KindTapeV   ElementKind = "tape_v"
	KindTapeH   ElementKind = "tape_h"
)

type Color struct {
	R, G, B, A uint8
}

type Element struct {
	ID        string      `json:"id"`
	Kind      ElementKind `json:"kind"`
	Binding   Binding     `json:"binding,omitempty"`
	Label     string      `json:"label,omitempty"`
	Unit      string      `json:"unit,omitempty"`
	X         float64     `json:"x"`
	Y         float64     `json:"y"`
	FontSize  int         `json:"font_size,omitempty"`
	Precision int         `json:"precision"`
	Color     Color       `json:"color"`
	Size      float64     `json:"size,omitempty"`
	Bold      bool        `json:"bold,omitempty"`
	AutoColor bool        `json:"auto_color,omitempty"`

	Length    float64 `json:"length,omitempty"`
	Range     float64 `json:"range,omitempty"`
	MinorStep float64 `json:"minor_step,omitempty"`
	MajorStep float64 `json:"major_step,omitempty"`
	Wrap      float64 `json:"wrap,omitempty"`

	Style     Style     `json:"style,omitempty"`
	Direction Direction `json:"direction,omitempty"`
	LabelSide LabelSide `json:"label_side,omitempty"`
	Zones     []Zone    `json:"zones,omitempty"`

	Thickness int `json:"thickness,omitempty"`

	GlowEnabled   bool    `json:"glow_enabled,omitempty"`
	GlowUseOwn    bool    `json:"glow_use_own_color,omitempty"`
	GlowColor     Color   `json:"glow_color,omitempty"`
	GlowIntensity float64 `json:"glow_intensity,omitempty"`
}

type Template struct {
	Name     string    `json:"name"`
	Army     string    `json:"army"`
	Elements []Element `json:"elements"`
}
