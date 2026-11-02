package app

import (
	"math"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// A picture opens in a window of its own, sized to it: a photograph is
// looked at, not squinted at between two lines of chat.

// viewer is the window a picture is open in, and what it shows.
type viewer struct {
	win *mygo.Window
	url string
	// w and h are the pixels the picture is asked of the host in: the
	// size it was sent in, within what a screen shows. winW and winH are
	// the size the window was made in, which its first frame is laid out
	// with before there is a frame to measure.
	w, h       int
	winW, winH int

	// zoom is how much larger the picture is drawn than the window holds
	// it, and offX and offY how far it is moved from the middle of the
	// window: Ctrl and the wheel change the first, a drag the second.
	zoom       float32
	offX, offY float32
	// viewW, viewH and drawnW, drawnH are what the window shows and what
	// the picture takes of it, as the last frame laid them out: they are
	// what a zoom is kept inside.
	viewW, viewH   float32
	drawnW, drawnH float32
	// drag says a press is moving the picture, and where the pointer was.
	drag         bool
	dragX, dragY float32
}

// How far a picture zooms: a quarter of the window's size up to eight
// times it.
const (
	minZoom = 0.25
	maxZoom = 8
)

// showPicture opens a picture in a window of its own, or shows it in the
// one that is already open: the window is a place to look at a picture,
// not one per picture.
func (a *App) showPicture(url string, w, h int) {
	if url == "" {
		return
	}
	if v := a.viewer; v != nil {
		v.url, v.w, v.h = url, w, h
		v.zoom, v.offX, v.offY = 1, 0, 0
		if v.win != nil && !v.win.IsDestroyed() {
			v.win.SetTitle(i18n.T("picture.title"))
			v.win.Invalidate()
			v.win.Focus()
		}
		return
	}
	v := &viewer{url: url, w: w, h: h, zoom: 1}
	a.viewer = v
	if a.win == nil {
		// No window to open one beside, as in the tests.
		return
	}
	// The window is made after the frame that asked for it: one made
	// while the view is being built is made in the middle of a paint.
	a.update(func() { a.openViewer(v) })
}

// openViewer makes the window a picture opens in, on the main thread and
// outside a frame.
func (a *App) openViewer(v *viewer) {
	width, height := viewerSize(v.w, v.h)
	v.winW, v.winH = width, height
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:           i18n.T("picture.title"),
		Width:           width,
		Height:          height,
		MinWidth:        260,
		MinHeight:       200,
		BackgroundColor: viewerCSS,
		Content:         ui.View(func(c *ui.Context) { a.viewerFrame(c, v) }),
	})
	win.OnClosed(func() {
		a.update(func() {
			if a.viewer == v {
				a.viewer = nil
			}
		})
	})
	v.win = win
}

// viewerFrame draws the picture window: the picture as large as the window
// holds it, on the dark the window is made of, and takes the input that
// looks at it closer.
func (a *App) viewerFrame(c *ui.Context, v *viewer) {
	qq = skin(a.dark(c))
	c.SetTheme(theme(qq.dark))
	i18n.Set(c.Locale())

	box := ui.Box(c).Fill().Background(qq.viewerBg)
	box.Children(func() {
		bmp := a.images.getBox(v.url, v.w, v.h, a)
		if bmp == nil {
			ui.Spinner(c).Margin(ui.Auto)
			return
		}
		v.picture(c, bmp, box.Bounds())
	})
	box.HandleInput(v.input)
}

// picture draws the picture where the zoom and the drag have put it.
func (v *viewer) picture(c *ui.Context, bmp *ui.Bitmap, r ui.Rect) {
	bw, bh := ui.ImageSize(bmp)
	if r.W <= 0 || r.H <= 0 {
		// The first frame, before there is one to measure.
		r = ui.Rect{W: float32(v.winW), H: float32(v.winH)}
		if r.W <= 0 || r.H <= 0 {
			r = ui.Rect{W: float32(bw), H: float32(bh)}
		}
	}
	// A window that was never zoomed draws the picture at the size the
	// window holds it.
	if v.zoom <= 0 {
		v.zoom = 1
	}
	// fit is how large a pixel of the picture is drawn in DIPs when all of
	// it is in the window.
	fit := min(float32(r.W)/float32(bw), float32(r.H)/float32(bh))
	k := fit * v.zoom
	v.viewW, v.viewH = r.W, r.H
	v.drawnW, v.drawnH = float32(bw)*k, float32(bh)*k
	v.clampOff()
	ui.Image(c, bmp).
		Absolute().
		Left(r.W/2 + v.offX - v.drawnW/2).
		Top(r.H/2 + v.offY - v.drawnH/2).
		Width(v.drawnW).Height(v.drawnH)
}

// input takes what looks at a picture closer: Ctrl and the wheel zoom
// around the pointer, a pinch zooms as the fingers say, and a drag moves
// it about.
func (v *viewer) input(ev ui.InputEvent) bool {
	switch ev.Kind {
	case ui.InputScroll:
		if ev.Mods&ui.Ctrl == 0 {
			return false
		}
		// A notch is worth a tenth of the picture or so, a touchpad's
		// move much less.
		v.zoomAt(ev.X, ev.Y, float32(math.Pow(1.0025, -float64(ev.DY))))
		return true
	case ui.InputZoom:
		v.zoomAt(ev.X, ev.Y, ev.Scale)
		return true
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		v.drag, v.dragX, v.dragY = true, ev.X, ev.Y
		return true
	case ui.InputPointerMove:
		if !v.drag {
			return false
		}
		v.offX += ev.X - v.dragX
		v.offY += ev.Y - v.dragY
		v.dragX, v.dragY = ev.X, ev.Y
		v.clampOff()
		return true
	case ui.InputPointerUp:
		if !v.drag {
			return false
		}
		v.drag = false
		return true
	}
	return false
}

// zoomAt zooms by f around the pointer at x, y, which keeps the point of
// the picture under it where it is.
func (v *viewer) zoomAt(x, y, f float32) {
	if f <= 0 {
		return
	}
	z := min(max(v.zoom*f, minZoom), maxZoom)
	f = z / v.zoom
	if f == 1 {
		return
	}
	cx, cy := v.viewW/2, v.viewH/2
	v.offX = (x - cx) - ((x-cx)-v.offX)*f
	v.offY = (y - cy) - ((y-cy)-v.offY)*f
	v.zoom = z
	// What the picture takes of the window grows with it, so that the
	// clamp below is kept to the size it is drawn at now rather than at
	// the size of the last frame.
	v.drawnW, v.drawnH = v.drawnW*f, v.drawnH*f
	v.clampOff()
}

// clampOff keeps the picture in the window: it moves no further than what
// of it is out of view, so that a drag cannot lose it.
func (v *viewer) clampOff() {
	limX := max(0, (v.drawnW-v.viewW)/2)
	limY := max(0, (v.drawnH-v.viewH)/2)
	v.offX = min(max(v.offX, -limX), limX)
	v.offY = min(max(v.offY, -limY), limY)
}

// viewerSize is the window a picture opens in: the picture's own shape,
// within what fits on a screen, and never smaller than a window worth
// looking through.
func viewerSize(w, h int) (int, int) {
	const maxW, maxH = 1100, 820
	const minW, minH = 320, 240
	if w <= 0 || h <= 0 {
		return 720, 540
	}
	k := math.Min(1, math.Min(float64(maxW)/float64(w), float64(maxH)/float64(h)))
	width := int(math.Max(minW, math.Round(float64(w)*k)))
	height := int(math.Max(minH, math.Round(float64(h)*k)))
	return width, height
}

// previewPixels is the size a picture is asked of the host for its window:
// the size it was sent in, so that it is sharp at the size it is drawn,
// within what a screen shows and what the cache can hold.
func previewPixels(m *yh.Message) (int, int) {
	const most = 1600
	w, h := m.Content.Width, m.Content.Height
	if w <= 0 || h <= 0 {
		return most, most
	}
	if w > most || h > most {
		k := math.Min(float64(most)/float64(w), float64(most)/float64(h))
		w, h = int(math.Round(float64(w)*k)), int(math.Round(float64(h)*k))
	}
	return w, h
}
