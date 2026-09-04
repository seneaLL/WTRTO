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

const releasesURL = "https://github.com/seneaLL/WTRTO/releases"

func openURL(url string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start", "", url)
	} else {
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	runtime.LockOSThread()
}

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

var fpsOptions = []int{30, 60, 75, 90, 120, 144, 165, 180, 240, 360, fpsUnlimited}

func fpsLabelFor(v int) string {
	if v == fpsUnlimited {
		return i18n.T("fps.vsync")
	}

	return fmt.Sprintf("%d FPS", v)
}

func fpsOptionLabels() []string {
	labels := make([]string, len(fpsOptions))
	for i, v := range fpsOptions {
		labels[i] = fpsLabelFor(v)
	}

	return labels
}

func fpsIndex(v int) int {
	for i, o := range fpsOptions {
		if o == v {
			return i
		}
	}

	return 0
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
	var latestMapInfo telemetry.MapInfo

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
				mi, miErr := telClient.MapInfo(ctx)
				cancel()
				telMu.Lock()
				if err == nil && m != nil {
					latestMission = *m
				}
				if miErr == nil && mi != nil {
					latestMapInfo = *mi
				}
				telMu.Unlock()
			}
		}
	}()

	sw, sh := w.Size()
	hudActive := false
	renderActive := false
	debugFrame := overlay.NewHUD(sw, sh, sampler, func() bool { return debugOn }, func() bool { return !hideFPS }, func() bool { return renderActive })

	exitEditMode := func() {
		editMode = false
		w.SetClickThrough(true)
		s := config.Load()
		s.HUDEditMode = false
		config.Save(s)
	}

	const hudFadeSeconds = 0.25
	var hudAlpha float64
	var lastFrameAt time.Time
	var lastValidValues hud.Values

	w.Run(func(c *native.Canvas, in *native.Input) bool {
		now := time.Now()
		dt := 0.0
		if !lastFrameAt.IsZero() {
			dt = now.Sub(lastFrameAt).Seconds()
		}
		lastFrameAt = now

		telMu.Lock()
		ind, st := latestInd, latestState
		mission := latestMission
		mapInfo := latestMapInfo
		telMu.Unlock()

		values := tracker.Update(ind, st)
		if values.Valid {
			lastValidValues = values
		}

		wtPID, wtRunning := sampler.WTPID()
		foreground := wtRunning && w.ActiveWindowPID() == wtPID

		gameplayVisible := mission.Active() && mapInfo.Valid && foreground

		hudActive = editMode || gameplayVisible

		target := 0.0
		if hudActive {
			target = 1.0
		}
		switch {
		case dt <= 0:
			hudAlpha = target
		case target > hudAlpha:
			hudAlpha = min(target, hudAlpha+dt/hudFadeSeconds)
		case target < hudAlpha:
			hudAlpha = max(target, hudAlpha-dt/hudFadeSeconds)
		}
		w.SetAlpha(hudAlpha)

		renderActive = hudActive || hudAlpha > 0

		renderValues := values
		if !editMode && !gameplayVisible {
			renderValues = lastValidValues
		}
		renderValues.Valid = renderActive

		wasDragging := editState.Dragging != ""
		hud.Draw(c, sw, sh, &tmpl, renderValues, editMode, editState, in)
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
	overlayCmd        *exec.Cmd
	overlayOn         bool
	debugInfo         bool
	fpsLimit          int
	fpsDropdownOpen   bool
	fpsDropdownScroll int
	hudEditMode       bool
	hideFPS           bool

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

var overlayMu sync.Mutex

func startOverlay() {
	overlayMu.Lock()
	if overlayOn {
		overlayMu.Unlock()

		return
	}
	overlayOn = true
	overlayMu.Unlock()

	go func() {
		cmd := exec.Command(os.Args[0], overlayChildArg, strconv.Itoa(os.Getpid()))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		overlayMu.Lock()
		err := cmd.Start()
		if err != nil {
			overlayOn = false
			overlayMu.Unlock()

			return
		}
		overlayCmd = cmd
		overlayMu.Unlock()

		cmd.Wait()

		overlayMu.Lock()

		if overlayCmd == cmd {
			overlayCmd = nil
			overlayOn = false
		}
		overlayMu.Unlock()
	}()
}

func stopOverlay() {
	overlayMu.Lock()
	cmd := overlayCmd
	overlayCmd = nil
	overlayOn = false
	overlayMu.Unlock()

	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
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
	winW    = 520
	winH    = 592
	margin  = 24
	headerH = 40
)

func HeaderDragRegion(curW int) native.Rect {
	return native.Rect{X: 0, Y: 0, W: curW - headerH*2, H: headerH}
}

func drawHeader(c *native.Canvas, in *native.Input, w *native.Window, curW int) {
	c.FillRect(native.Rect{X: 0, Y: 0, W: curW, H: headerH}, colorPanel)
	_, th := c.TextSize(i18n.T("app.title"), 14)
	c.TextBold(16, (headerH+th)/2-2, colorText, 14, i18n.T("app.title"))

	minimizeRect := native.Rect{X: curW - headerH*2, Y: 0, W: headerH, H: headerH}
	minHover := minimizeRect.Contains(in.MouseX, in.MouseY)
	minBg := colorPanel
	if minHover {
		minBg = colorPanelHover
	}
	c.FillRect(minimizeRect, minBg)
	lineW := 12
	lx := minimizeRect.X + (minimizeRect.W-lineW)/2
	ly := minimizeRect.Y + minimizeRect.H/2 + 3
	c.FillRect(native.Rect{X: lx, Y: ly, W: lineW, H: 2}, colorText)
	if minHover && in.Released {
		w.Hide()
	}

	closeRect := native.Rect{X: curW - headerH, Y: 0, W: headerH, H: headerH}
	hover := closeRect.Contains(in.MouseX, in.MouseY)
	closeBg := colorPanel
	if hover {
		closeBg = colorDanger
	}
	c.FillRect(closeRect, closeBg)
	xw, xh := c.TextSize("x", 15)
	c.Text(closeRect.X+(closeRect.W-xw)/2, closeRect.Y+(closeRect.H+xh)/2-2, colorText, 15, "x")
	if hover && in.Released {
		w.Close()
	}
}

func launcherFrame(c *native.Canvas, in *native.Input, w *native.Window) bool {
	originalIn := in
	curW, curH := c.Size()
	c.FillRect(native.Rect{X: 0, Y: 0, W: curW, H: curH}, colorBg)

	y := margin
	if !native.HasNativeTitleBar {
		w.SetDragRegion(HeaderDragRegion(curW))
		drawHeader(c, in, w, curW)
		y = headerH + margin
	}

	contentW := curW - 2*margin
	if contentW < 100 {
		contentW = 100
	}

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

	const hotkeyW = 220
	const rowGap = 16
	fpsW := contentW - hotkeyW - rowGap
	if fpsW < 100 {
		fpsW = 100
	}
	fpsRect := native.Rect{X: margin + hotkeyW + rowGap, Y: 0, W: fpsW, H: 34}

	c.Text(margin, y+14, colorTextDim, 13, "Хоткей вкл/выкл оверлея")
	c.Text(fpsRect.X, y+14, colorTextDim, 13, i18n.T("settings.fps_limit"))
	y += 14 + 10

	hkLabel := hotkeyLabel
	if capturingHotkey {
		hkLabel = "Нажмите комбинацию клавиш…"
	}
	if native.Button(c, in, native.Rect{X: margin, Y: y, W: hotkeyW, H: 34}, hkLabel, colorPanel, colorPanelHover, colorText, 13) {
		capturingHotkey = true
	}

	fpsRect.Y = y
	if native.SelectBox(c, in, fpsRect, fpsLabelFor(fpsLimit), fpsDropdownOpen, colorPanel, colorPanelHover, colorText, 13) {
		fpsDropdownOpen = !fpsDropdownOpen
		if fpsDropdownOpen {
			fpsDropdownScroll = 0
		}
	}
	if fpsDropdownOpen {
		listBounds := native.SelectListBounds(fpsRect, len(fpsOptions), curH)
		if listBounds.Contains(in.MouseX, in.MouseY) {
			masked := *in
			masked.MouseDown, masked.Pressed, masked.Released, masked.ScrollDelta = false, false, false, 0
			in = &masked
		}
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
		uw, uh := c.TextSize(updText, 11)
		updY := buildY - uh - 6
		updRect := native.Rect{X: margin + contentW - uw - 6, Y: updY - 3, W: uw + 6, H: uh + 6}

		hovered := updRect.Contains(in.MouseX, in.MouseY)
		w.SetHandCursor(hovered)

		updCol := colorAccent
		if hovered {
			updCol = colorAccentHover
			if in.Released {
				openURL(releasesURL)
			}
		}
		c.Text(margin+contentW-uw, updY, updCol, 11, updText)
	}

	if fpsDropdownOpen {
		newIdx, selected := native.SelectList(c, originalIn, fpsRect, fpsOptionLabels(), fpsIndex(fpsLimit), &fpsDropdownScroll, curH, colorPanel, colorPanelHover, colorText, colorTextDim, 13)
		switch {
		case selected:
			fpsLimit = fpsOptions[newIdx]
			saveState()
			fpsDropdownOpen = false

		case originalIn.Released && !fpsRect.Contains(originalIn.MouseX, originalIn.MouseY) &&
			!native.SelectListBounds(fpsRect, len(fpsOptions), curH).Contains(originalIn.MouseX, originalIn.MouseY):
			fpsDropdownOpen = false
		}
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
	if !native.HasNativeTitleBar {
		w.SetDragRegion(HeaderDragRegion(winW))
	}
	w.GrabHotkey(hotkeyKeysym, hotkeyMods)
	defer w.UngrabHotkey()

	w.EnableTray("WTRTO", i18n.T("tray.show"), i18n.T("tray.exit"),
		func() { w.Show(); w.Focus() },
		func() { w.Close() },
	)
	defer w.DisableTray()

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

		return launcherFrame(c, in, w)
	})

	stopOverlay()
}
