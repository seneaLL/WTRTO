package hud

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
)

var AllBindings = []Binding{
	BindThrottlePct, BindIAS, BindTAS, BindMach, BindAltitude,
	BindFuelKg, BindFuelTime, BindFuelRate, BindOilTemp1, BindOilTemp2,
	BindCompass, BindAoA, BindAoS, BindGLoad, BindVSpeed, BindIASRate,
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
