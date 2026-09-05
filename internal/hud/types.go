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

	BindThrottle1 Binding = "throttle_1_pct"
	BindThrottle2 Binding = "throttle_2_pct"
	BindThrottle3 Binding = "throttle_3_pct"
	BindThrottle4 Binding = "throttle_4_pct"

	BindRPM1 Binding = "rpm_1"
	BindRPM2 Binding = "rpm_2"
	BindRPM3 Binding = "rpm_3"
	BindRPM4 Binding = "rpm_4"

	BindManifold1 Binding = "manifold_pressure_1_atm"
	BindManifold2 Binding = "manifold_pressure_2_atm"
	BindManifold3 Binding = "manifold_pressure_3_atm"
	BindManifold4 Binding = "manifold_pressure_4_atm"

	BindOilTemp3 Binding = "oil_temp_3"
	BindOilTemp4 Binding = "oil_temp_4"

	BindWaterTemp1 Binding = "water_temp_1"
	BindWaterTemp2 Binding = "water_temp_2"
	BindWaterTemp3 Binding = "water_temp_3"
	BindWaterTemp4 Binding = "water_temp_4"

	BindPower1 Binding = "power_1_hp"
	BindPower2 Binding = "power_2_hp"
	BindPower3 Binding = "power_3_hp"
	BindPower4 Binding = "power_4_hp"

	BindThrust1 Binding = "thrust_1_kgs"
	BindThrust2 Binding = "thrust_2_kgs"
	BindThrust3 Binding = "thrust_3_kgs"
	BindThrust4 Binding = "thrust_4_kgs"

	BindEfficiency1 Binding = "efficiency_1_pct"
	BindEfficiency2 Binding = "efficiency_2_pct"
	BindEfficiency3 Binding = "efficiency_3_pct"
	BindEfficiency4 Binding = "efficiency_4_pct"

	BindPropPitch1 Binding = "prop_pitch_1_deg"
	BindPropPitch2 Binding = "prop_pitch_2_deg"
	BindPropPitch3 Binding = "prop_pitch_3_deg"
	BindPropPitch4 Binding = "prop_pitch_4_deg"

	BindWingSweep Binding = "wing_sweep_pct"
)

var AllBindings = []Binding{
	BindThrottlePct, BindIAS, BindTAS, BindMach, BindAltitude,
	BindFuelKg, BindFuelTime, BindFuelRate, BindOilTemp1, BindOilTemp2,
	BindCompass, BindAoA, BindAoS, BindGLoad, BindVSpeed, BindIASRate,

	BindAileron, BindElevator, BindRudder, BindFlaps, BindGearPct,
	BindRollRate, BindFuelPct, BindTrimmer, BindRadioAlt, BindTurn,

	BindThrottle1, BindThrottle2, BindThrottle3, BindThrottle4,
	BindRPM1, BindRPM2, BindRPM3, BindRPM4,
	BindManifold1, BindManifold2, BindManifold3, BindManifold4,
	BindOilTemp3, BindOilTemp4,
	BindWaterTemp1, BindWaterTemp2, BindWaterTemp3, BindWaterTemp4,
	BindPower1, BindPower2, BindPower3, BindPower4,
	BindThrust1, BindThrust2, BindThrust3, BindThrust4,
	BindEfficiency1, BindEfficiency2, BindEfficiency3, BindEfficiency4,
	BindPropPitch1, BindPropPitch2, BindPropPitch3, BindPropPitch4,
	BindWingSweep,
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
