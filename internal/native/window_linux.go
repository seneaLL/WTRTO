//go:build linux

package native

/*
#cgo pkg-config: freetype2 cairo
#cgo LDFLAGS: -lX11 -lXext -lXft
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <X11/keysym.h>
#include <X11/extensions/shape.h>
#include <X11/Xft/Xft.h>
#include <cairo/cairo.h>
#include <cairo/cairo-xlib.h>
#include <stdlib.h>
#include <string.h>

static int wtrto_error_handler(Display *d, XErrorEvent *e) {
    return 0;
}

static Display *wtrto_open() {
    XSetErrorHandler(wtrto_error_handler);

    return XOpenDisplay(NULL);
}

static Window wtrto_create_window(Display *d, int x, int y, int w, int h,
                                   int transparent, int decorated, int override_redirect,
                                   XVisualInfo *vinfo_out, Colormap *cmap_out) {
    int screen = DefaultScreen(d);
    XVisualInfo vinfo;
    int have_argb = 0;

    if (transparent) {
        have_argb = XMatchVisualInfo(d, screen, 32, TrueColor, &vinfo);
    }
    if (!have_argb) {
        vinfo.visual = DefaultVisual(d, screen);
        vinfo.depth = DefaultDepth(d, screen);
    }

    XSetWindowAttributes attrs;
    attrs.colormap = XCreateColormap(d, RootWindow(d, screen), vinfo.visual, AllocNone);
    attrs.background_pixel = 0;
    attrs.border_pixel = 0;
    attrs.override_redirect = override_redirect ? True : False;
    attrs.event_mask = ExposureMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask | StructureNotifyMask | KeyPressMask;

    Window win = XCreateWindow(d, RootWindow(d, screen), x, y, w, h, 0,
        vinfo.depth, InputOutput, vinfo.visual,
        CWColormap | CWBackPixel | CWBorderPixel | CWOverrideRedirect | CWEventMask,
        &attrs);

    XSelectInput(d, win, ExposureMask | ButtonPressMask | ButtonReleaseMask | PointerMotionMask | StructureNotifyMask | KeyPressMask);

    if (!decorated) {
        Atom motifHints = XInternAtom(d, "_MOTIF_WM_HINTS", False);
        struct { unsigned long flags, functions, decorations; long input_mode; unsigned long status; } hints;
        hints.flags = 2;
        hints.decorations = 0;
        XChangeProperty(d, win, motifHints, motifHints, 32, PropModeReplace, (unsigned char*)&hints, 5);
    }

    *vinfo_out = vinfo;
    *cmap_out = attrs.colormap;

    return win;
}

static void wtrto_set_min_size(Display *d, Window win, int minw, int minh) {
    XSizeHints hints;
    memset(&hints, 0, sizeof(hints));
    hints.flags = PMinSize;
    hints.min_width = minw;
    hints.min_height = minh;
    XSetWMNormalHints(d, win, &hints);
}

static void wtrto_set_always_on_top(Display *d, Window win) {
    Atom wmState = XInternAtom(d, "_NET_WM_STATE", False);
    Atom above = XInternAtom(d, "_NET_WM_STATE_ABOVE", False);
    XChangeProperty(d, win, wmState, XA_ATOM, 32, PropModeReplace, (unsigned char*)&above, 1);
}

static void wtrto_set_click_through(Display *d, Window win, int enable) {
    int event_base, error_base;
    if (!XShapeQueryExtension(d, &event_base, &error_base)) {
        return;
    }
    if (enable) {
        Region region = XCreateRegion();
        XShapeCombineRegion(d, win, ShapeInput, 0, 0, region, ShapeSet);
        XDestroyRegion(region);
    } else {
        XShapeCombineMask(d, win, ShapeInput, 0, 0, None, ShapeSet);
    }
}

static void wtrto_set_input_hint(Display *d, Window win) {
    XWMHints hints;
    hints.flags = InputHint;
    hints.input = True;
    XSetWMHints(d, win, &hints);
}

static void wtrto_focus(Display *d, Window win) {
    XSetInputFocus(d, win, RevertToParent, CurrentTime);
}

static Atom wtrto_wm_delete_atom(Display *d, Window win) {
    Atom wmDelete = XInternAtom(d, "WM_DELETE_WINDOW", False);
    XSetWMProtocols(d, win, &wmDelete, 1);

    return wmDelete;
}

static void wtrto_grab_key(Display *d, Window root, int keycode, unsigned int mods) {
    XGrabKey(d, keycode, mods, root, True, GrabModeAsync, GrabModeAsync);
}

static void wtrto_ungrab_key(Display *d, Window root, int keycode, unsigned int mods) {
    XUngrabKey(d, keycode, mods, root);
}

static KeySym wtrto_lookup_string(XKeyEvent *ev, char *buf, int buflen) {
    KeySym ks = NoSymbol;
    XLookupString(ev, buf, buflen, &ks, NULL);

    return ks;
}

static unsigned long wtrto_active_window_pid(Display *d, Window root) {
    Atom activeAtom = XInternAtom(d, "_NET_ACTIVE_WINDOW", False);
    Atom pidAtom = XInternAtom(d, "_NET_WM_PID", False);
    Atom actualType;
    int actualFormat;
    unsigned long nItems, bytesAfter;
    unsigned char *prop = NULL;
    unsigned long pid = 0;

    if (XGetWindowProperty(d, root, activeAtom, 0, 1, False, XA_WINDOW,
                            &actualType, &actualFormat, &nItems, &bytesAfter, &prop) != Success || !prop) {
        return 0;
    }
    Window active = *(Window *)prop;
    XFree(prop);

    if (active == None) {
        return 0;
    }

    prop = NULL;
    if (XGetWindowProperty(d, active, pidAtom, 0, 1, False, XA_CARDINAL,
                            &actualType, &actualFormat, &nItems, &bytesAfter, &prop) == Success && prop) {
        pid = *(unsigned long *)prop;
        XFree(prop);
    }

    return pid;
}

static XftFont *wtrto_open_font(Display *d, int screen, int size, int bold) {
    char name[64];
    if (bold) {
        snprintf(name, sizeof(name), "Noto Sans:bold-%d", size);
    } else {
        snprintf(name, sizeof(name), "Noto Sans-%d", size);
    }

    return XftFontOpenName(d, screen, name);
}
*/
import "C"

import (
	"time"
	"unsafe"
)

type Window struct {
	display  *C.Display
	win      C.Window
	visual   *C.Visual
	depth    C.int
	colormap C.Colormap
	screen   C.int

	pixmap C.Pixmap
	gc     C.GC

	cairoSurface *C.cairo_surface_t
	cairoCtx     *C.cairo_t

	xftDraw *C.XftDraw
	fonts   map[fontKey]*C.XftFont

	w, h        int
	resizable   bool
	shouldClose bool
	wmDelete    C.Atom
	fps         int

	root          C.Window
	hotkeyKeycode C.int
	hotkeyMods    C.uint
}

func NewWindow(opts WindowOptions) (*Window, error) {
	d := C.wtrto_open()
	if d == nil {
		return nil, errNoDisplay
	}

	screen := C.XDefaultScreen(d)

	var vinfo C.XVisualInfo
	var cmap C.Colormap

	transparent := C.int(0)
	if opts.Transparent {
		transparent = 1
	}
	decorated := C.int(0)
	if opts.Decorated {
		decorated = 1
	}
	overrideRedirect := C.int(0)
	if opts.ClickThrough {
		overrideRedirect = 1
	}

	win := C.wtrto_create_window(d, C.int(opts.X), C.int(opts.Y), C.int(opts.W), C.int(opts.H),
		transparent, decorated, overrideRedirect, &vinfo, &cmap)

	if opts.AlwaysOnTop {
		C.wtrto_set_always_on_top(d, win)
	}
	if opts.Decorated {
		C.wtrto_set_min_size(d, win, 360, 360)
	}

	title := C.CString(opts.Title)
	defer C.free(unsafe.Pointer(title))
	C.XStoreName(d, win, title)

	C.wtrto_set_input_hint(d, win)

	wmDelete := C.wtrto_wm_delete_atom(d, win)

	C.XMapWindow(d, win)
	C.XFlush(d)

	if opts.ClickThrough {
		C.wtrto_set_click_through(d, win, 1)
	}

	w := &Window{
		display:   d,
		win:       win,
		visual:    vinfo.visual,
		depth:     vinfo.depth,
		colormap:  cmap,
		screen:    screen,
		fonts:     make(map[fontKey]*C.XftFont),
		w:         opts.W,
		h:         opts.H,
		resizable: opts.Decorated,
		wmDelete:  wmDelete,
		fps:       30,
		root:      C.XDefaultRootWindow(d),
	}
	w.createBuffers(opts.W, opts.H)

	return w, nil
}

func (w *Window) createBuffers(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	w.pixmap = C.XCreatePixmap(w.display, C.Drawable(w.win), C.uint(width), C.uint(height), C.uint(w.depth))
	w.gc = C.XCreateGC(w.display, C.Drawable(w.pixmap), 0, nil)
	w.xftDraw = C.XftDrawCreate(w.display, C.Drawable(w.pixmap), w.visual, w.colormap)
	w.cairoSurface = C.cairo_xlib_surface_create(w.display, C.Drawable(w.pixmap), w.visual, C.int(width), C.int(height))
	w.cairoCtx = C.cairo_create(w.cairoSurface)
	w.w, w.h = width, height
}

func (w *Window) destroyBuffers() {
	if w.cairoCtx != nil {
		C.cairo_destroy(w.cairoCtx)
	}
	if w.cairoSurface != nil {
		C.cairo_surface_destroy(w.cairoSurface)
	}
	if w.xftDraw != nil {
		C.XftDrawDestroy(w.xftDraw)
	}
	if w.gc != nil {
		C.XFreeGC(w.display, w.gc)
	}
	if w.pixmap != 0 {
		C.XFreePixmap(w.display, w.pixmap)
	}
}

func (w *Window) resize(width, height int) {
	if width == w.w && height == w.h {
		return
	}
	w.destroyBuffers()
	w.createBuffers(width, height)
}

type fontKey struct {
	size int
	bold bool
}

func (w *Window) font(size int, bold bool) *C.XftFont {
	key := fontKey{size: size, bold: bold}
	if f, ok := w.fonts[key]; ok {
		return f
	}
	boldC := C.int(0)
	if bold {
		boldC = 1
	}
	f := C.wtrto_open_font(w.display, w.screen, C.int(size), boldC)
	w.fonts[key] = f

	return f
}

func (w *Window) Size() (int, int) {
	return w.w, w.h
}

func (w *Window) Pos() (int, int) {
	var x, y C.int
	var child C.Window
	C.XTranslateCoordinates(w.display, w.win, C.XDefaultRootWindow(w.display), 0, 0, &x, &y, &child)

	return int(x), int(y)
}

func (w *Window) SetPos(x, y int) {
	C.XMoveWindow(w.display, w.win, C.int(x), C.int(y))
	C.XFlush(w.display)
}

func (w *Window) SetClickThrough(enable bool) {
	v := C.int(0)
	if enable {
		v = 1
	}
	C.wtrto_set_click_through(w.display, w.win, v)
	C.XFlush(w.display)
}

func (w *Window) Close() {
	w.shouldClose = true
}

func (w *Window) Hide() {
	C.XUnmapWindow(w.display, w.win)
	C.XFlush(w.display)
}

func (w *Window) Show() {
	C.XMapWindow(w.display, w.win)
	C.XFlush(w.display)
}

func (w *Window) ActiveWindowPID() int32 {
	return int32(C.wtrto_active_window_pid(w.display, w.root))
}

func IsModifierKeySym(ks uint) bool {
	switch ks {
	case 0xffe1, 0xffe2, 0xffe3, 0xffe4, 0xffe5, 0xffe9, 0xffea, 0xffeb, 0xffec:
		return true
	}

	return false
}

func (w *Window) Focus() {
	C.wtrto_focus(w.display, w.win)
	C.XFlush(w.display)
}

var hotkeyIgnoreMods = [...]C.uint{0, C.uint(ModLock), 0x10, C.uint(ModLock) | 0x10}

func (w *Window) GrabHotkey(keysym uint, mods uint) bool {
	w.UngrabHotkey()
	keycode := C.XKeysymToKeycode(w.display, C.KeySym(keysym))
	if keycode == 0 {
		return false
	}
	for _, ignore := range hotkeyIgnoreMods {
		C.wtrto_grab_key(w.display, w.root, C.int(keycode), C.uint(mods)|ignore)
	}
	C.XFlush(w.display)
	w.hotkeyKeycode = C.int(keycode)
	w.hotkeyMods = C.uint(mods)

	return true
}

func (w *Window) UngrabHotkey() {
	if w.hotkeyKeycode == 0 {
		return
	}
	for _, ignore := range hotkeyIgnoreMods {
		C.wtrto_ungrab_key(w.display, w.root, w.hotkeyKeycode, w.hotkeyMods|ignore)
	}
	C.XFlush(w.display)
	w.hotkeyKeycode = 0
}

func (w *Window) pollEvents(in *Input) {
	in.Pressed = false
	in.Released = false
	in.KeyEvent = false
	in.KeyRune = 0
	in.KeyCtrlLetter = 0
	in.KeyBackspace = false
	in.KeyEnter = false
	in.KeyEscape = false
	in.KeyDelete = false
	in.KeyLeft = false
	in.KeyRight = false
	in.HotkeyTriggered = false

	for C.XPending(w.display) > 0 {
		var ev C.XEvent
		C.XNextEvent(w.display, &ev)

		switch *(*C.int)(unsafe.Pointer(&ev)) {
		case C.MotionNotify:
			me := (*C.XMotionEvent)(unsafe.Pointer(&ev))
			in.MouseX = int(me.x)
			in.MouseY = int(me.y)
			in.KeyMods = uint(me.state)
		case C.ButtonPress:
			be := (*C.XButtonEvent)(unsafe.Pointer(&ev))
			in.KeyMods = uint(be.state)
			if be.button == C.Button1 {
				in.MouseDown = true
				in.Pressed = true
				in.MouseX = int(be.x)
				in.MouseY = int(be.y)
			}
		case C.ButtonRelease:
			be := (*C.XButtonEvent)(unsafe.Pointer(&ev))
			in.KeyMods = uint(be.state)
			if be.button == C.Button1 {
				in.MouseDown = false
				in.Released = true
				in.MouseX = int(be.x)
				in.MouseY = int(be.y)
			}
		case C.KeyPress:
			ke := (*C.XKeyEvent)(unsafe.Pointer(&ev))
			if ke.window == w.root && w.hotkeyKeycode != 0 && ke.keycode == C.uint(w.hotkeyKeycode) {
				in.HotkeyTriggered = true
				break
			}
			var buf [8]C.char
			ks := C.wtrto_lookup_string(ke, &buf[0], 8)
			in.KeyEvent = true
			in.KeySym = uint(ks)
			in.KeyMods = uint(ke.state)
			if buf[0] != 0 {
				code := byte(buf[0])
				switch {
				case code >= 32 && code < 127:
					in.KeyRune = rune(code)

				case code >= 1 && code <= 26:
					in.KeyCtrlLetter = rune(code) + 'a' - 1
				}
			}
			switch ks {
			case C.XK_BackSpace:
				in.KeyBackspace = true
			case C.XK_Return, C.XK_KP_Enter:
				in.KeyEnter = true
			case C.XK_Escape:
				in.KeyEscape = true
			case C.XK_Delete:
				in.KeyDelete = true
			case C.XK_Left:
				in.KeyLeft = true
			case C.XK_Right:
				in.KeyRight = true
			}
		case C.ConfigureNotify:
			ce := (*C.XConfigureEvent)(unsafe.Pointer(&ev))
			if w.resizable {
				w.resize(int(ce.width), int(ce.height))
			}
		case C.ClientMessage:
			cm := (*C.XClientMessageEvent)(unsafe.Pointer(&ev))
			data := (*[5]C.long)(unsafe.Pointer(&cm.data))
			if C.Atom(data[0]) == w.wmDelete {
				w.shouldClose = true
			}
		}
	}
}

func (w *Window) SetFPS(fps int) {
	if fps <= 0 {
		fps = 30
	}
	w.fps = fps
}

func (w *Window) Run(frame FrameFunc) {
	var in Input
	canvas := &Canvas{win: w}

	for !w.shouldClose {
		start := time.Now()

		w.pollEvents(&in)

		C.XSetForeground(w.display, w.gc, 0)
		C.XFillRectangle(w.display, C.Drawable(w.pixmap), w.gc, 0, 0, C.uint(w.w), C.uint(w.h))

		if !frame(canvas, &in) {
			break
		}

		C.cairo_surface_flush(w.cairoSurface)
		C.XCopyArea(w.display, C.Drawable(w.pixmap), C.Drawable(w.win), w.gc, 0, 0, C.uint(w.w), C.uint(w.h), 0, 0)
		C.XFlush(w.display)

		frameDur := time.Second / time.Duration(w.fps)
		if elapsed := time.Since(start); elapsed < frameDur {
			time.Sleep(frameDur - elapsed)
		}
	}
}

var errNoDisplay = &nativeError{"cannot open X11 display"}

type nativeError struct{ msg string }

func (e *nativeError) Error() string { return e.msg }
