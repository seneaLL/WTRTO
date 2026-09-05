package telemetry

import (
	"encoding/json"
	"fmt"
)

type Indicators struct {
	Valid bool   `json:"valid"`
	Type  string `json:"type"`
	Army  string `json:"army"`

	Speed float64 `json:"speed"`
	TAS   float64 `json:"tas"`
	Mach  float64 `json:"mach"`
	Mach1 float64 `json:"mach1"`

	Vario            float64 `json:"vario"`
	AltitudeHour     float64 `json:"altitude_hour"`
	AltitudeMin      float64 `json:"altitude_min"`
	Altitude1Min     float64 `json:"altitude1_min"`
	Altitude10k      float64 `json:"altitude_10k"`
	RadioAltitude    float64 `json:"radio_altitude"`
	AviahorizonRoll  float64 `json:"aviahorizon_roll"`
	AviahorizonPitch float64 `json:"aviahorizon_pitch"`
	Bank             float64 `json:"bank"`
	Bank1            float64 `json:"bank1"`
	Turn             float64 `json:"turn"`
	Compass          float64 `json:"compass"`
	AoA              float64 `json:"aoa"`
	StickElevator    float64 `json:"stick_elevator"`
	StickAilerons    float64 `json:"stick_ailerons"`
	Gears            float64 `json:"gears"`
	GearLampUp       float64 `json:"gear_lamp_up"`
	GearLampOff      float64 `json:"gear_lamp_off"`
	GearLampDown     float64 `json:"gear_lamp_down"`
	Throttle         float64 `json:"throttle"`
	Throttle1        float64 `json:"throttle1"`
	Trimmer          float64 `json:"trimmer_indicator"`
	GMeter           float64 `json:"g_meter"`
	GMeterMin        float64 `json:"g_meter_min"`
	GMeterMax        float64 `json:"g_meter_max"`
	ClockHour        float64 `json:"clock_hour"`
	ClockMin         float64 `json:"clock_min"`
	ClockSec         float64 `json:"clock_sec"`

	WingSweepLever float64 `json:"wing_sweep_lever"`

	Stabilizer           float64 `json:"stabilizer"`
	Gear                 float64 `json:"gear"`
	GearNeutral          float64 `json:"gear_neutral"`
	HasSpeedWarning      float64 `json:"has_speed_warning"`
	RPM                  float64 `json:"rpm"`
	DrivingDirectionMode float64 `json:"driving_direction_mode"`
	CruiseControl        float64 `json:"cruise_control"`
	LWS                  float64 `json:"lws"`
	IRCM                 float64 `json:"ircm"`
	FirstStageAmmo       float64 `json:"first_stage_ammo"`
	CrewTotal            float64 `json:"crew_total"`
	CrewCurrent          float64 `json:"crew_current"`
	CrewDistance         float64 `json:"crew_distance"`
	GunnerState          float64 `json:"gunner_state"`
	DriverState          float64 `json:"driver_state"`
	EngineBroken         float64 `json:"engine_broken"`
	EngineDead           float64 `json:"engine_dead"`
	TrackBroken          float64 `json:"track_broken"`
	IsRepairingAuto      float64 `json:"is_repairing_auto"`
	RepairTime           float64 `json:"repair_time"`
	Burns                float64 `json:"burns"`

	Extra map[string]interface{} `json:"-"`
}

func (i Indicators) IsAircraft() bool { return i.Army == "air" }
func (i Indicators) IsTank() bool     { return i.Army == "tank" }

func (i *Indicators) UnmarshalJSON(data []byte) error {
	type alias Indicators
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*i = Indicators(a)
	i.Extra = raw

	return nil
}

type State map[string]interface{}

func (s State) Raw() map[string]interface{} { return s }

func (s State) float(key string) (float64, bool) {
	f, ok := s[key].(float64)

	return f, ok
}

func (s State) Valid() bool {
	b, _ := s["valid"].(bool)

	return b
}

func (s State) AltitudeM() (float64, bool)            { return s.float("H, m") }
func (s State) TrueAirspeedKmh() (float64, bool)      { return s.float("TAS, km/h") }
func (s State) IndicatedAirspeedKmh() (float64, bool) { return s.float("IAS, km/h") }
func (s State) Mach() (float64, bool)                 { return s.float("M") }
func (s State) AngleOfAttackDeg() (float64, bool)     { return s.float("AoA, deg") }
func (s State) AngleOfSideslipDeg() (float64, bool)   { return s.float("AoS, deg") }
func (s State) LateralG() (float64, bool)             { return s.float("Ny") }
func (s State) VerticalSpeedMs() (float64, bool)      { return s.float("Vy, m/s") }
func (s State) RollRateDegS() (float64, bool)         { return s.float("Wx, deg/s") }
func (s State) FuelMassKg() (float64, bool)           { return s.float("Mfuel, kg") }
func (s State) FuelMassInitialKg() (float64, bool)    { return s.float("Mfuel0, kg") }
func (s State) AileronPct() (float64, bool)           { return s.float("aileron, %") }
func (s State) ElevatorPct() (float64, bool)          { return s.float("elevator, %") }
func (s State) RudderPct() (float64, bool)            { return s.float("rudder, %") }
func (s State) FlapsPct() (float64, bool)             { return s.float("flaps, %") }
func (s State) GearPct() (float64, bool)              { return s.float("gear, %") }

func (s State) EngineThrottlePct(n int) (float64, bool) {
	return s.float(fmt.Sprintf("throttle %d, %%", n))
}
func (s State) EngineRPM(n int) (float64, bool) { return s.float(fmt.Sprintf("RPM %d", n)) }
func (s State) EnginePowerHp(n int) (float64, bool) {
	return s.float(fmt.Sprintf("power %d, hp", n))
}
func (s State) EngineManifoldPressureAtm(n int) (float64, bool) {
	return s.float(fmt.Sprintf("manifold pressure %d, atm", n))
}
func (s State) EngineOilTempC(n int) (float64, bool) {
	return s.float(fmt.Sprintf("oil temp %d, C", n))
}
func (s State) EngineWaterTempC(n int) (float64, bool) {
	return s.float(fmt.Sprintf("water temp %d, C", n))
}
func (s State) EnginePropPitchDeg(n int) (float64, bool) {
	return s.float(fmt.Sprintf("pitch %d, deg", n))
}
func (s State) EngineThrustKgs(n int) (float64, bool) {
	return s.float(fmt.Sprintf("thrust %d, kgs", n))
}
func (s State) EngineEfficiencyPct(n int) (float64, bool) {
	return s.float(fmt.Sprintf("efficiency %d, %%", n))
}
func (s State) EngineMagneto(n int) (float64, bool) {
	return s.float(fmt.Sprintf("magneto %d", n))
}
func (s State) EngineMixturePct(n int) (float64, bool) {
	return s.float(fmt.Sprintf("mixture %d, %%", n))
}
func (s State) EngineRadiatorPct(n int) (float64, bool) {
	return s.float(fmt.Sprintf("radiator %d, %%", n))
}
func (s State) EngineCompressorStage(n int) (float64, bool) {
	return s.float(fmt.Sprintf("compressor stage %d", n))
}

func (s State) EngineCount() int {
	n := 0
	for i := 1; i <= maxEngines; i++ {
		if _, ok := s.EngineThrottlePct(i); !ok {
			break
		}
		n = i
	}

	return n
}

const maxEngines = 8

type MapInfo struct {
	Valid         bool      `json:"valid"`
	GridSize      []float64 `json:"grid_size"`
	GridSteps     []float64 `json:"grid_steps"`
	GridZero      []float64 `json:"grid_zero"`
	HudType       int       `json:"hud_type"`
	MapGeneration int       `json:"map_generation"`
	MapMax        []float64 `json:"map_max"`
	MapMin        []float64 `json:"map_min"`
}

type MapObject struct {
	Type     string `json:"type"`
	Color    string `json:"color"`
	ColorRGB []int  `json:"color[]"`
	Blink    int    `json:"blink"`
	Icon     string `json:"icon"`
	IconBg   string `json:"icon_bg"`

	X  *float64 `json:"x,omitempty"`
	Y  *float64 `json:"y,omitempty"`
	DX *float64 `json:"dx,omitempty"`
	DY *float64 `json:"dy,omitempty"`

	SX *float64 `json:"sx,omitempty"`
	SY *float64 `json:"sy,omitempty"`
	EX *float64 `json:"ex,omitempty"`
	EY *float64 `json:"ey,omitempty"`
}

func (m MapObject) IsPlayer() bool { return m.Icon == "Player" }

type HudMsg struct {
	Events []interface{} `json:"events"`
	Damage []DamageEvent `json:"damage"`
}

type DamageEvent struct {
	ID     int    `json:"id"`
	Msg    string `json:"msg"`
	Sender string `json:"sender"`
	Enemy  bool   `json:"enemy"`
	Mode   string `json:"mode"`
}

type Mission struct {
	Status     string      `json:"status"`
	Objectives []Objective `json:"objectives"`
}

func (m Mission) Active() bool {
	switch m.Status {
	case "", "not_available", "0":
		return false
	}

	return true
}

type Objective struct {
	Primary bool   `json:"primary"`
	Status  string `json:"status"`
	Text    string `json:"text"`
}

type ChatMessage struct {
	ID     int    `json:"id"`
	Msg    string `json:"msg"`
	Sender string `json:"sender"`
	Enemy  bool   `json:"enemy"`
	Mode   string `json:"mode"`
}
