package hud

import (
	"time"

	"github.com/seneaLL/WTRTO/internal/limits"
	"github.com/seneaLL/WTRTO/internal/telemetry"
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
	FuelPct     float64

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

	Aileron   float64
	Elevator  float64
	Rudder    float64
	Flaps     float64
	GearPct   float64
	RollRate  float64
	Trimmer   float64
	RadioAlt  float64
	Turn      float64
	WingSweep float64

	SpeedWarning     bool
	SpeedLimitKnown  bool
	SpeedLimitRatio  float64
	SpeedLimitMaxKmh float64

	MachLimitKnown bool
	MachLimitRatio float64
	MachLimitMax   float64

	GLoadLimitKnown bool
	GLoadLimitRatio float64
	GLoadLimitPos   float64
	GLoadLimitNeg   float64

	EngineThrottle   [maxTrackedEngines]float64
	EngineRPM        [maxTrackedEngines]float64
	EngineManifold   [maxTrackedEngines]float64
	EngineOilTemp    [maxTrackedEngines]float64
	EngineWaterTemp  [maxTrackedEngines]float64
	EnginePower      [maxTrackedEngines]float64
	EngineThrust     [maxTrackedEngines]float64
	EngineEfficiency [maxTrackedEngines]float64
	EnginePropPitch  [maxTrackedEngines]float64
}

const maxTrackedEngines = 4

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
	if fuel0, ok := st.FuelMassInitialKg(); ok && fuel0 > 0 {
		v.FuelPct = v.FuelKg / fuel0 * 100
	}
	v.Compass = ind.Compass
	v.Pitch = ind.AviahorizonPitch
	v.Roll = ind.AviahorizonRoll

	v.Aileron, _ = st.AileronPct()
	v.Elevator, _ = st.ElevatorPct()
	v.Rudder, _ = st.RudderPct()
	v.Flaps, _ = st.FlapsPct()
	v.GearPct, _ = st.GearPct()
	v.RollRate, _ = st.RollRateDegS()
	v.Trimmer = ind.Trimmer * 100
	v.RadioAlt = ind.RadioAltitude
	v.Turn = ind.Turn
	v.WingSweep = ind.WingSweepIndicator * 100

	v.SpeedWarning = ind.HasSpeedWarning != 0
	if lim, ok := limits.Get(ind.Type); ok {
		sweep := ind.WingSweepIndicator

		if maxIASBase := lim.MaxSpeedIASKmh.At(sweep); maxIASBase > 0 {
			maxIAS := maxIASBase
			if v.GearPct > 0 && lim.GearSpeedKmh > 0 && lim.GearSpeedKmh < maxIAS {
				maxIAS = lim.GearSpeedKmh
			}
			if v.Flaps > 0 {
				if lim.FlapSpeedKmh.TakeoffLanding > 0 && lim.FlapSpeedKmh.TakeoffLanding < maxIAS {
					maxIAS = lim.FlapSpeedKmh.TakeoffLanding
				}
				if lim.FlapSpeedKmh.Combat > 0 && lim.FlapSpeedKmh.Combat < maxIAS {
					maxIAS = lim.FlapSpeedKmh.Combat
				}
			}
			v.SpeedLimitKnown = true
			v.SpeedLimitRatio = v.IAS / maxIAS
			v.SpeedLimitMaxKmh = maxIAS
		}

		if machLimit := lim.MachLimit.At(sweep); machLimit > 0 {
			v.MachLimitKnown = true
			v.MachLimitMax = machLimit
			v.MachLimitRatio = v.Mach / machLimit
		}

		if lim.GLimitPos > 0 && lim.GLimitNeg < 0 {
			v.GLoadLimitKnown = true
			v.GLoadLimitPos = lim.GLimitPos
			v.GLoadLimitNeg = lim.GLimitNeg
			if v.GLoad >= 0 {
				v.GLoadLimitRatio = v.GLoad / lim.GLimitPos
			} else {
				v.GLoadLimitRatio = v.GLoad / lim.GLimitNeg
			}
		}
	}

	for i := 0; i < maxTrackedEngines; i++ {
		n := i + 1
		v.EngineThrottle[i], _ = st.EngineThrottlePct(n)
		v.EngineRPM[i], _ = st.EngineRPM(n)
		v.EngineManifold[i], _ = st.EngineManifoldPressureAtm(n)
		v.EngineOilTemp[i], _ = st.EngineOilTempC(n)
		v.EngineWaterTemp[i], _ = st.EngineWaterTempC(n)
		v.EnginePower[i], _ = st.EnginePowerHp(n)
		v.EngineThrust[i], _ = st.EngineThrustKgs(n)
		v.EngineEfficiency[i], _ = st.EngineEfficiencyPct(n)
		v.EnginePropPitch[i], _ = st.EnginePropPitchDeg(n)
	}

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
	found := false
	for n := 1; n <= st.EngineCount(); n++ {
		if v, ok := st.EngineThrottlePct(n); ok {
			found = true
			if v > max {
				max = v
			}
		}
	}
	if !found {
		max = ind.Throttle * 100
	}

	return max
}
