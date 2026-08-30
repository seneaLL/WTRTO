package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seneal/wtrto/internal/config"
	"github.com/seneal/wtrto/internal/hud"
	"github.com/seneal/wtrto/internal/i18n"
	"github.com/seneal/wtrto/internal/metrics"
	"github.com/seneal/wtrto/internal/native"
	"github.com/seneal/wtrto/internal/overlay"
	"github.com/seneal/wtrto/internal/telemetry"
	"github.com/seneal/wtrto/internal/update"
	"github.com/seneal/wtrto/internal/version"
)

const overlayChildArg = "--overlay-child"

var (
	colorBg          = native.Color{R: 13, G: 15, B: 19, A: 255}
	colorPanel       = native.Color{R: 22, G: 25, B: 32, A: 255}
	colorText        = native.Color{R: 236, G: 238, B: 242, A: 255}
	colorTextDim     = native.Color{R: 142, G: 150, B: 163, A: 255}
	colorAccent      = native.Color{R: 68, G: 224, B: 140, A: 255}
	colorAccentHover = native.Color{R: 92, G: 238, B: 158, A: 255}
	colorDanger      = native.Color{R: 228, G: 96, B: 96, A: 255}
	colorDangerHover = native.Color{R: 238, G: 116, B: 116, A: 255}
	colorPanelHover  = native.Color{R: 30, G: 34, B: 43, A: 255}
	colorBtnText     = native.Color{R: 12, G: 14, B: 17, A: 255}
)

var (
	updateMu        sync.Mutex
	updateAvailable bool
	updateLatestSHA string
)

func startUpdateCheck() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, err := update.Check(ctx, version.Version)
		if err != nil {
			return
		}
		updateMu.Lock()
		updateAvailable = res.Available
		updateLatestSHA = res.LatestSHA
		updateMu.Unlock()
	}()
}

func updateStatus() (bool, string) {
	updateMu.Lock()
	defer updateMu.Unlock()

	return updateAvailable, updateLatestSHA
}

const fpsUnlimited = -1

var fpsOptions = []int{30, 60, 120, 144, 240, fpsUnlimited}

func fpsLabelFor(v int) string {
	if v == fpsUnlimited {
		return i18n.T("fps.unlimited")
	}

	return fmt.Sprintf("%d FPS", v)
}

func nextFPS(v int) int {
	for i, o := range fpsOptions {
		if o == v {
			return fpsOptions[(i+1)%len(fpsOptions)]
		}
	}

	return fpsOptions[0]
}

func main() {
	if len(os.Args) > 2 && os.Args[1] == overlayChildArg {
		launcherPID, _ := strconv.Atoi(os.Args[2])
		runOverlay(launcherPID)

		return
	}
	runLauncher()
}

func applyLanguage(s config.State) {
	switch s.Language {
	case string(i18n.RU):
		i18n.Set(i18n.RU)
	case string(i18n.EN):
		i18n.Set(i18n.EN)
	}
}

func runOverlay(launcherPID int) {
	initial := config.Load()
	applyLanguage(initial)

	w, err := overlay.New("wtrto-overlay", initial.EffectiveFPSLimit())
	if err != nil {
		fmt.Fprintln(os.Stderr, "wtrto: overlay init failed:", err)
		os.Exit(1)
	}

	ownPIDs := []int32{int32(os.Getpid())}
	if launcherPID > 0 {
		ownPIDs = append(ownPIDs, int32(launcherPID))
	}
	sampler := metrics.NewSampler(ownPIDs)

	stop := make(chan struct{})
	go sampler.Run(stop)
	defer close(stop)

	debugOn := initial.DebugInfo
	hideFPS := initial.HideFPS
	editMode := initial.HUDEditMode
	w.SetClickThrough(!editMode)

	go config.Watch(1*time.Second, stop, func(s config.State) {
		debugOn = s.DebugInfo
		hideFPS = s.HideFPS
		applyLanguage(s)
		w.SetTargetFPS(s.EffectiveFPSLimit())
		if s.HUDEditMode != editMode {
			editMode = s.HUDEditMode
			w.SetClickThrough(!editMode)
		}
	})

	telClient := telemetry.NewClient(telemetry.DefaultBaseURL)
	tracker := hud.NewTracker()
	hud.EnsureBuiltins()
	activeTemplateName := initial.ActiveTemplate
	if activeTemplateName == "" {
		activeTemplateName = hud.DefaultAirTemplate().Name
	}
	tmpl, err := hud.Load(activeTemplateName)
	if err != nil {
		tmpl = hud.EnsureDefault(hud.DefaultAirTemplate())
	}
	editState := &hud.EditState{}

	var telMu sync.Mutex
	var latestInd *telemetry.Indicators
	var latestState telemetry.State
	var latestMission telemetry.Mission

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				ind, indErr := telClient.Indicators(ctx)
				st, stErr := telClient.State(ctx)
				cancel()
				telMu.Lock()
				if indErr == nil {
					latestInd = ind
				}
				if stErr == nil {
					latestState = st
				}
				telMu.Unlock()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				m, err := telClient.Mission(ctx)
				cancel()
				if err == nil && m != nil {
					telMu.Lock()
					latestMission = *m
					telMu.Unlock()
				}
			}
		}
	}()

	sw, sh := w.Size()
	hudActive := false
	debugFrame := overlay.NewHUD(sw, sh, sampler, func() bool { return debugOn }, func() bool { return !hideFPS }, func() bool { return hudActive })

	exitEditMode := func() {
		editMode = false
		w.SetClickThrough(true)
		s := config.Load()
		s.HUDEditMode = false
		config.Save(s)
	}

	w.Run(func(c *native.Canvas, in *native.Input) bool {
		telMu.Lock()
		ind, st := latestInd, latestState
		mission := latestMission
		telMu.Unlock()

		values := tracker.Update(ind, st)

		wtPID, wtRunning := sampler.WTPID()
		foreground := wtRunning && w.ActiveWindowPID() == wtPID
		gameplayVisible := mission.Active() && foreground

		hudActive = editMode || gameplayVisible
		if !editMode && !gameplayVisible {
			values.Valid = false
		}

		wasDragging := editState.Dragging != ""
		hud.Draw(c, sw, sh, &tmpl, values, editMode, editState, in)
		changed := wasDragging && editState.Dragging == ""

		if editMode {
			if hud.DrawPropertiesPanel(c, in, sw, sh, &tmpl, editState) {
				changed = true
			}

			if editState.PendingDialog != "" {
				w.Hide()
				hud.ResolvePendingDialog(&tmpl, editState)
				w.Show(true)
			}
		}
		if changed {
			hud.Save(tmpl)
		}

		if editMode {
			if in.KeyEscape {
				exitEditMode()
			}
			btnW, btnH := 160, 36
			doneRect := native.Rect{X: sw/2 - btnW/2, Y: 16, W: btnW, H: btnH}
			if native.Button(c, in, doneRect, i18n.T("hud.edit_done"),
				native.Color{R: 68, G: 224, B: 140, A: 255},
				native.Color{R: 92, G: 238, B: 158, A: 255},
				native.Color{R: 12, G: 14, B: 17, A: 255}, 14) {
				exitEditMode()
			}
		}

		return debugFrame(c, in)
	})
}

var (
	overlayCmd  *exec.Cmd
	overlayOn   bool
	debugInfo   bool
	fpsLimit    int
	hudEditMode bool
	hideFPS     bool

	hotkeyKeysym    uint
	hotkeyMods      uint
	hotkeyLabel     string
	capturingHotkey bool
)

const (
	defaultHotkeyMods  = native.ModCtrl | native.ModAlt
	defaultHotkeyLabel = "Ctrl+Alt+H"
)

var defaultHotkeyKeysym = func() uint {
	if runtime.GOOS == "windows" {
		return 0x48
	}

	return 0x68
}()

func modsLabel(mods uint) string {
	var parts []string
	if mods&native.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if mods&native.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mods&native.ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if mods&native.ModSuper != 0 {
		parts = append(parts, "Super")
	}

	return strings.Join(parts, "+")
}

func formatHotkeyLabel(mods uint, keyRune rune, keysym uint) string {
	key := fmt.Sprintf("Key0x%x", keysym)

	switch {
	case keysym >= 0x20 && keysym < 0x7f:
		key = strings.ToUpper(string(rune(keysym)))

	case keyRune >= 33 && keyRune < 127:
		key = strings.ToUpper(string(keyRune))
	}

	mp := modsLabel(mods)
	if mp == "" {
		return key
	}

	return mp + "+" + key
}

func startOverlay() {
	cmd := exec.Command(os.Args[0], overlayChildArg, strconv.Itoa(os.Getpid()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return
	}
	overlayCmd = cmd
	overlayOn = true

	go func() {
		cmd.Wait()
		overlayOn = false
		overlayCmd = nil
	}()
}

func stopOverlay() {
	if overlayCmd != nil && overlayCmd.Process != nil {
		overlayCmd.Process.Kill()
	}
	overlayCmd = nil
	overlayOn = false
}

func toggleOverlay() {
	if overlayOn {
		stopOverlay()

		return
	}
	startOverlay()
}

func saveState() {
	config.Save(config.State{
		DebugInfo:    debugInfo,
		Language:     string(i18n.Current()),
		FPSLimit:     fpsLimit,
		HUDEditMode:  hudEditMode,
		HideFPS:      hideFPS,
		HotkeyKeysym: hotkeyKeysym,
		HotkeyMods:   hotkeyMods,
		HotkeyLabel:  hotkeyLabel,
	})
}

func setLanguage(l i18n.Lang) {
	i18n.Set(l)
	saveState()
}

const (
	winW   = 520
	winH   = 590
	margin = 24
)

func launcherFrame(c *native.Canvas, in *native.Input) bool {
	curW, curH := c.Size()
	c.FillRect(native.Rect{X: 0, Y: 0, W: curW, H: curH}, colorBg)

	contentW := curW - 2*margin
	if contentW < 100 {
		contentW = 100
	}
	y := margin

	langCol := func(l i18n.Lang) native.Color {
		if i18n.Current() == l {
			return colorAccent
		}

		return colorTextDim
	}
	if native.Button(c, in, native.Rect{X: margin, Y: y, W: 48, H: 28}, "EN", colorPanel, colorPanelHover, langCol(i18n.EN), 12) {
		setLanguage(i18n.EN)
	}
	if native.Button(c, in, native.Rect{X: margin + 56, Y: y, W: 48, H: 28}, "RU", colorPanel, colorPanelHover, langCol(i18n.RU), 12) {
		setLanguage(i18n.RU)
	}
	y += 28 + 28

	c.TextBold(margin, y+22, colorText, 24, i18n.T("app.title"))
	y += 32
	c.Text(margin, y+16, colorTextDim, 14, i18n.T("app.subtitle"))
	y += 16 + 28

	statusText := i18n.T("status.off")
	statusCol := colorTextDim
	if overlayOn {
		statusText = i18n.T("status.on")
		statusCol = colorAccent
	}
	c.FillCircle(margin+6, y+8, 6, statusCol)
	c.Text(margin+22, y+13, colorTextDim, 14, statusText)
	y += 20 + 18

	btnLabel := i18n.T("button.enable")
	btnBg, btnBgHover := colorAccent, colorAccentHover
	if overlayOn {
		btnLabel = i18n.T("button.disable")
		btnBg, btnBgHover = colorDanger, colorDangerHover
	}
	if native.Button(c, in, native.Rect{X: margin, Y: y, W: contentW, H: 46}, btnLabel, btnBg, btnBgHover, colorBtnText, 15) {
		toggleOverlay()
	}
	y += 46 + 26

	if native.Checkbox(c, in, native.Rect{X: margin, Y: y, W: 22, H: 22}, debugInfo, colorTextDim, colorAccent, colorText, 14, i18n.T("debug.checkbox")) {
		debugInfo = !debugInfo
		saveState()
	}
	y += 22 + 16

	if native.Checkbox(c, in, native.Rect{X: margin, Y: y, W: 22, H: 22}, !hideFPS, colorTextDim, colorAccent, colorText, 14, i18n.T("overlay.show_fps")) {
		hideFPS = !hideFPS
		saveState()
	}
	y += 22 + 16

	if native.Checkbox(c, in, native.Rect{X: margin, Y: y, W: 22, H: 22}, hudEditMode, colorTextDim, colorAccent, colorText, 14, i18n.T("hud.edit_mode")) {
		hudEditMode = !hudEditMode
		saveState()
	}
	y += 22 + 18
	c.Text(margin+34, y, colorTextDim, 11, i18n.T("hud.edit_hint"))
	y += 11 + 22

	c.Text(margin, y+14, colorTextDim, 13, "Хоткей вкл/выкл оверлея")
	y += 14 + 10
	hkLabel := hotkeyLabel
	if capturingHotkey {
		hkLabel = "Нажмите комбинацию клавиш…"
	}
	if native.Button(c, in, native.Rect{X: margin, Y: y, W: 220, H: 34}, hkLabel, colorPanel, colorPanelHover, colorText, 13) {
		capturingHotkey = true
	}
	y += 34 + 28

	c.Text(margin, y+14, colorTextDim, 13, i18n.T("settings.fps_limit"))
	y += 14 + 10
	if native.Button(c, in, native.Rect{X: margin, Y: y, W: 150, H: 34}, fpsLabelFor(fpsLimit), colorPanel, colorPanelHover, colorText, 13) {
		fpsLimit = nextFPS(fpsLimit)
		saveState()
	}
	y += 34 + 28

	const repoURL = "github.com/seneaLL/WTRTO"
	rw, rh := c.TextSize(repoURL, 11)
	repoY := curH - margin - rh + 4
	c.Text(margin+contentW-rw, repoY, colorTextDim, 11, repoURL)

	buildText := version.String()
	bw, bh := c.TextSize(buildText, 11)
	buildY := repoY - bh - 4
	c.Text(margin+contentW-bw, buildY, colorTextDim, 11, buildText)

	if avail, latestSHA := updateStatus(); avail {
		updText := "Доступно обновление"
		if latestSHA != "" {
			updText += " (" + latestSHA + ")"
		}
		uw, _ := c.TextSize(updText, 11)
		c.Text(margin+contentW-bw-uw-12, buildY, colorAccent, 11, updText)
	}

	return true
}

func runLauncher() {
	s := config.Load()
	debugInfo = s.DebugInfo
	fpsLimit = s.EffectiveFPSLimit()
	hudEditMode = s.HUDEditMode
	hideFPS = s.HideFPS
	hotkeyKeysym = s.HotkeyKeysym
	hotkeyMods = s.HotkeyMods
	hotkeyLabel = s.HotkeyLabel
	if hotkeyKeysym == 0 {
		hotkeyKeysym = defaultHotkeyKeysym
		hotkeyMods = defaultHotkeyMods
		hotkeyLabel = defaultHotkeyLabel
	}
	applyLanguage(s)
	startUpdateCheck()

	w, err := native.NewWindow(native.WindowOptions{
		Title:     "WTRTO",
		X:         200,
		Y:         160,
		W:         winW,
		H:         winH,
		Decorated: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "wtrto: launcher init failed:", err)
		os.Exit(1)
	}
	w.SetFPS(30)
	w.GrabHotkey(hotkeyKeysym, hotkeyMods)
	defer w.UngrabHotkey()

	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go config.Watch(1*time.Second, stopWatch, func(s config.State) {

		hudEditMode = s.HUDEditMode
	})

	w.Run(func(c *native.Canvas, in *native.Input) bool {
		if in.HotkeyTriggered {
			toggleOverlay()
		}
		if capturingHotkey {
			if in.KeyEscape {
				capturingHotkey = false
			} else if in.KeyEvent && !native.IsModifierKeySym(in.KeySym) {
				hotkeyKeysym = in.KeySym
				hotkeyMods = in.KeyMods & (native.ModShift | native.ModCtrl | native.ModAlt | native.ModSuper)
				hotkeyLabel = formatHotkeyLabel(hotkeyMods, in.KeyRune, hotkeyKeysym)
				w.UngrabHotkey()
				w.GrabHotkey(hotkeyKeysym, hotkeyMods)
				capturingHotkey = false
				saveState()
			}
		}

		return launcherFrame(c, in)
	})

	stopOverlay()
}
