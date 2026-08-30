package hud

import (
	"time"

	"github.com/seneal/wtrto/internal/telemetry"
)

type Values struct {
	Valid bool

	ThrottlePct float64
	IAS         float64
	TAS         float64
	Mach        float64
	Altitude    float64

	FuelKg      float64
	FuelRateKgM float64
	FuelTimeMin float64

	OilTemp1 float64
	OilTemp2 float64

	Compass float64
	AoA     float64
	AoS     float64
	GLoad   float64
	VSpeed  float64
	IASRate float64

	Pitch float64
	Roll  float64
}

type Tracker struct {
	lastFuelKg   float64
	lastFuelTime time.Time
	fuelRate     float64
	haveFuel     bool

	lastIAS     float64
	lastIASTime time.Time
	iasRate     float64
	haveIAS     bool
}

func NewTracker() *Tracker {
	return &Tracker{}
}

func (t *Tracker) Update(ind *telemetry.Indicators, st telemetry.State) Values {
	var v Values
	if ind == nil || !ind.Valid || !ind.IsAircraft() {
		return v
	}
	v.Valid = true

	v.ThrottlePct = maxThrottle(st, ind)
	v.IAS, _ = st.IndicatedAirspeedKmh()
	v.TAS, _ = st.TrueAirspeedKmh()
	v.Mach, _ = st.Mach()
	v.Altitude, _ = st.AltitudeM()
	v.AoA, _ = st.AngleOfAttackDeg()
	v.AoS, _ = st.AngleOfSideslipDeg()
	v.GLoad, _ = st.LateralG()
	v.VSpeed, _ = st.VerticalSpeedMs()
	v.OilTemp1, _ = st.EngineOilTempC(1)
	v.OilTemp2, _ = st.EngineOilTempC(2)
	v.FuelKg, _ = st.FuelMassKg()
	v.Compass = ind.Compass
	v.Pitch = ind.AviahorizonPitch
	v.Roll = ind.AviahorizonRoll

	t.updateFuelRate(v.FuelKg)
	v.FuelRateKgM = t.fuelRate
	if t.fuelRate > 0.01 {
		v.FuelTimeMin = v.FuelKg / t.fuelRate
	}

	t.updateIASRate(v.IAS)
	v.IASRate = t.iasRate

	return v
}

func (t *Tracker) updateIASRate(ias float64) {
	now := time.Now()
	if !t.haveIAS {
		t.lastIAS = ias
		t.lastIASTime = now
		t.haveIAS = true

		return
	}

	dt := now.Sub(t.lastIASTime).Seconds()
	if dt < 0.1 {
		return
	}

	rate := (ias - t.lastIAS) / dt
	t.iasRate = t.iasRate*0.7 + rate*0.3
	t.lastIAS = ias
	t.lastIASTime = now
}

func (t *Tracker) updateFuelRate(fuelKg float64) {
	now := time.Now()
	if !t.haveFuel {
		t.lastFuelKg = fuelKg
		t.lastFuelTime = now
		t.haveFuel = true

		return
	}

	dt := now.Sub(t.lastFuelTime).Minutes()
	if dt < 0.1 {
		return
	}

	delta := t.lastFuelKg - fuelKg
	if delta >= 0 {
		rate := delta / dt
		t.fuelRate = t.fuelRate*0.7 + rate*0.3
	}
	t.lastFuelKg = fuelKg
	t.lastFuelTime = now
}

func maxThrottle(st telemetry.State, ind *telemetry.Indicators) float64 {
	max := 0.0
	for n := 1; n <= 2; n++ {
		if v, ok := st.EngineThrottlePct(n); ok && v > max {
			max = v
		}
	}
	if max == 0 {
		max = ind.Throttle * 100
	}

	return max
}
