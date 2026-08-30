package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIndicatorsAndState(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/indicators", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"valid": true,
			"type": "P-51D-5",
			"speed": 420,
			"mach": 0.34,
			"altitude_hour": 3,
			"altitude_min": 1500,
			"gears": 1,
			"weapon2": 0.5
		}`))
	})
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"valid": true,
			"H, m": 1500.5,
			"TAS, km/h": 620,
			"IAS, km/h": 510,
			"Mfuel, kg": 220.3,
			"RPM 1": 2800
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL)
	ctx := context.Background()

	ind, err := c.Indicators(ctx)
	if err != nil {
		t.Fatalf("Indicators: %v", err)
	}
	if !ind.Valid || ind.Type != "P-51D-5" {
		t.Fatalf("unexpected indicators: %+v", ind)
	}
	if ind.Speed != 420 {
		t.Errorf("Speed = %v, want 420", ind.Speed)
	}
	if v, ok := ind.Extra["weapon2"]; !ok || v != 0.5 {
		t.Errorf("Extra[weapon2] = %v (ok=%v), want 0.5", v, ok)
	}

	st, err := c.State(ctx)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if !st.Valid() {
		t.Fatal("State.Valid() = false, want true")
	}
	if alt, ok := st.AltitudeM(); !ok || alt != 1500.5 {
		t.Errorf("AltitudeM() = %v (ok=%v), want 1500.5", alt, ok)
	}
	if tas, ok := st.TrueAirspeedKmh(); !ok || tas != 620 {
		t.Errorf("TrueAirspeedKmh() = %v (ok=%v), want 620", tas, ok)
	}
	if fuel, ok := st.FuelMassKg(); !ok || fuel != 220.3 {
		t.Errorf("FuelMassKg() = %v (ok=%v), want 220.3", fuel, ok)
	}
	if rpm, ok := st.float("RPM 1"); !ok || rpm != 2800 {
		t.Errorf("RPM 1 = %v (ok=%v), want 2800", rpm, ok)
	}
}

func TestClientGameNotRunning(t *testing.T) {

	c := NewClient("http://127.0.0.1:1")
	_, err := c.Indicators(context.Background())
	if err == nil {
		t.Fatal("expected error when nothing is listening")
	}
}

func TestClientEmptyBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hudmsg", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/gamechat", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL)
	ctx := context.Background()

	if _, err := c.HudMsg(ctx); err != nil {
		t.Errorf("HudMsg with empty body: %v", err)
	}
	if _, err := c.GameChat(ctx); err != nil {
		t.Errorf("GameChat with empty body: %v", err)
	}
}

func TestClientTankIndicators(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/indicators", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"valid": true,
			"army": "tank",
			"type": "tankModels/ussr_t_90m_arena_m",
			"stabilizer": 0.0,
			"gear": 0.0,
			"speed": 42.5,
			"rpm": 2100,
			"crew_total": 3.0,
			"crew_current": 3.0,
			"gunner_state": 0.0,
			"driver_state": 0.0,
			"engine_broken": 0.0,
			"track_broken": 0.0
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL)
	ind, err := c.Indicators(context.Background())
	if err != nil {
		t.Fatalf("Indicators: %v", err)
	}
	if !ind.IsTank() || ind.IsAircraft() {
		t.Fatalf("expected tank, got army=%q", ind.Army)
	}
	if ind.CrewTotal != 3 || ind.CrewCurrent != 3 {
		t.Errorf("crew = %v/%v, want 3/3", ind.CrewCurrent, ind.CrewTotal)
	}
	if ind.Speed != 42.5 {
		t.Errorf("Speed = %v, want 42.5", ind.Speed)
	}
}
