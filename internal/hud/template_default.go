package hud

var hudGreen = Color{R: 80, G: 255, B: 90, A: 255}

func textEl(id string, b Binding, label, unit string, x, y float64, precision int) Element {
	return Element{
		ID: id, Kind: KindText, Binding: b, Label: label, Unit: unit,
		X: x, Y: y, FontSize: 15, Precision: precision, Color: hudGreen,
	}
}

func speedZones() []Zone {
	return []Zone{
		{Threshold: 0, Color: Color{R: 80, G: 255, B: 90, A: 255}},
		{Threshold: 550, Color: Color{R: 255, G: 220, B: 60, A: 255}},
		{Threshold: 750, Color: Color{R: 255, G: 70, B: 60, A: 255}},
	}
}

func altitudeZones() []Zone {
	return []Zone{
		{Threshold: 0, Color: Color{R: 255, G: 70, B: 60, A: 255}},
		{Threshold: 500, Color: Color{R: 255, G: 220, B: 60, A: 255}},
		{Threshold: 1500, Color: Color{R: 80, G: 255, B: 90, A: 255}},
	}
}

func BuiltinTemplates() []Template {
	return []Template{
		DefaultAirTemplate(),
		MinimalAirTemplate(),
		ArcGaugesAirTemplate(),
	}
}

func IsBuiltin(name string) bool {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return true
		}
	}

	return false
}

func DefaultAirTemplate() Template {
	el := textEl

	return Template{
		Name: "Classic Jet HUD",
		Army: "air",
		Elements: []Element{
			el("thr", BindThrottlePct, "THR", "%", 0.28, 0.16, 0),
			el("ias", BindIAS, "IAS", "km/h", 0.28, 0.19, 0),
			el("spd", BindTAS, "SPD", "km/h", 0.28, 0.22, 0),
			el("mach", BindMach, "MACH", "", 0.28, 0.25, 2),
			el("alt", BindAltitude, "ALT", "m", 0.28, 0.28, 0),

			el("fuel_time", BindFuelTime, "FUEL", "min", 0.28, 0.34, 1),
			el("fuel_kg", BindFuelKg, "FUEL", "kg", 0.28, 0.37, 0),
			el("fuel_rate", BindFuelRate, "FUEL", "kg/min", 0.28, 0.40, 1),

			el("oil1", BindOilTemp1, "OIL1", "°C", 0.28, 0.46, 0),
			el("oil2", BindOilTemp2, "OIL2", "°C", 0.28, 0.49, 0),

			{
				ID: "compass", Kind: KindTapeH, Binding: BindCompass,
				X: 0.5, Y: 0.1, Length: 0.22, Range: 50, MinorStep: 5, MajorStep: 5, Wrap: 360,
				Color: hudGreen, Style: StyleStraight, Direction: DirCW,
			},

			{
				ID: "speed_tape", Kind: KindTapeV, Binding: BindIAS,
				X: 0.34, Y: 0.58, Length: 0.16, Range: 300, MinorStep: 25, MajorStep: 50,
				Color: hudGreen, Style: StyleStraight, Direction: DirUp, LabelSide: SideAuto,
				Zones: speedZones(),
			},
			{
				ID: "alt_tape", Kind: KindTapeV, Binding: BindAltitude,
				X: 0.62, Y: 0.58, Length: 0.16, Range: 600, MinorStep: 50, MajorStep: 100,
				Color: hudGreen, Style: StyleStraight, Direction: DirUp, LabelSide: SideAuto,
				Zones: altitudeZones(),
			},

			el("aoa", BindAoA, "AoA", "°", 0.44, 0.68, 1),
			el("gload", BindGLoad, "G", "", 0.44, 0.71, 1),
			el("aos", BindAoS, "AoS", "°", 0.56, 0.68, 1),
			el("vspeed", BindVSpeed, "VS", "m/s", 0.56, 0.71, 1),
		},
	}
}

func MinimalAirTemplate() Template {
	return Template{
		Name: "Minimal",
		Army: "air",
		Elements: []Element{
			{
				ID: "compass", Kind: KindTapeH, Binding: BindCompass,
				X: 0.5, Y: 0.08, Length: 0.2, Range: 50, MinorStep: 5, MajorStep: 5, Wrap: 360,
				Color: hudGreen, Style: StyleStraight, Direction: DirCW,
			},
			{
				ID: "speed_tape", Kind: KindTapeV, Binding: BindIAS,
				X: 0.3, Y: 0.55, Length: 0.2, Range: 300, MinorStep: 25, MajorStep: 50,
				Color: hudGreen, Style: StyleStraight, Direction: DirUp, LabelSide: SideAuto,
				Zones: speedZones(),
			},
			{
				ID: "alt_tape", Kind: KindTapeV, Binding: BindAltitude,
				X: 0.66, Y: 0.55, Length: 0.2, Range: 600, MinorStep: 50, MajorStep: 100,
				Color: hudGreen, Style: StyleStraight, Direction: DirUp, LabelSide: SideAuto,
				Zones: altitudeZones(),
			},
		},
	}
}

func ArcGaugesAirTemplate() Template {
	return Template{
		Name: "Arc Gauges",
		Army: "air",
		Elements: []Element{
			{
				ID: "compass", Kind: KindTapeH, Binding: BindCompass,
				X: 0.5, Y: 0.12, Length: 0.3, Range: 60, MinorStep: 5, MajorStep: 10, Wrap: 360,
				Color: hudGreen, Style: StyleStraight, Direction: DirCW,
			},
			{

				ID: "speed_tape", Kind: KindTapeV, Binding: BindIAS,
				X: 0.38, Y: 0.55, Length: 0.24, Range: 300, MinorStep: 25, MajorStep: 50,
				Color: hudGreen, Style: StyleArc, Direction: DirUp, LabelSide: SideAuto,
				Zones: speedZones(),
			},
			{
				ID: "alt_tape", Kind: KindTapeV, Binding: BindAltitude,
				X: 0.62, Y: 0.55, Length: 0.24, Range: 600, MinorStep: 50, MajorStep: 100,
				Color: hudGreen, Style: StyleArc, Direction: DirUp, LabelSide: SideAuto,
				Zones: altitudeZones(),
			},
			textEl("vspeed", BindVSpeed, "VS", "m/s", 0.5, 0.68, 1),
			textEl("gload", BindGLoad, "G", "", 0.5, 0.72, 1),
		},
	}
}
