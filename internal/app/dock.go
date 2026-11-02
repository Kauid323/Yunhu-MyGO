package app

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The panel keeps out of the way. When the pointer has been away from it for
// a moment it slides to the edge of the screen it is nearest — one of the
// sides, or the top, which is where a panel dragged up there goes — leaving a
// sliver of itself in sight; when the pointer comes back to that sliver it
// slides out again, flush against the edge, so that the pointer is over it as
// it arrives.
//
// The pointer is asked for rather than waited on: the window is told nothing
// when the pointer leaves it — it only hears what happens on it — so a frame
// asks the desktop where the pointer is, a few times a second, which is cheap
// enough for a panel that is mostly idle.
const (
	dockWait   = 700 * time.Millisecond // how long the pointer is away before it goes
	dockSliver = 3                      // how much of the panel stays in sight
	dockReach  = 6                      // how near the edge the pointer may be to bring it back
	dockStep   = 12 * time.Millisecond  // a frame of a slide
	dockSlide  = 220 * time.Millisecond // how long a slide takes
	dockPoll   = 120 * time.Millisecond // how often the pointer is asked for
)

// edge is the edge of the screen the panel hides against.
type edge int

const (
	edgeNone edge = iota
	edgeLeft
	edgeRight
	edgeTop
)

// dock is where the panel is on its way to and from an edge.
type dock struct {
	// hidden says the panel is away, or on its way: the edge, and not the
	// place it was left at, is where it is looked for. side is the edge it
	// went to, which is where it comes back from.
	hidden bool
	side   edge
	// away is when the pointer was last seen outside the panel, and zero
	// while it is inside it.
	away time.Time
	// sliding is a slide in progress: from where it started, to where it
	// goes, and since when.
	sliding      bool
	fromX, fromY float32
	toX, toY     float32
	at           time.Time
}

// dockFrame slides the panel to and from the edge as the pointer comes and
// goes, and asks for the frames that watch for it. A window the settings say
// does not hide is left where it is put, and watches for nothing.
func (a *App) dockFrame(c *ui.Context) {
	if !a.prefs.Dock {
		return
	}
	win := a.win
	if win == nil || win.IsDestroyed() {
		return
	}
	px, py, ok := cursorPoint()
	if !ok {
		// No pointer to follow, as on a platform that does not report one.
		return
	}
	b := win.Bounds()
	scale := a.scale()
	x, y := float32(px)/scale, float32(py)/scale
	d := &a.dock

	if d.sliding {
		a.dockSlide(c)
		return
	}

	inside := a.inPanel(x, y, b)
	switch {
	case inside:
		d.away = time.Time{}
		if d.hidden {
			a.startDockSlide(c, b, false)
			return
		}
	case d.hidden:
		// It is away and the pointer is not at the edge it went to: it
		// stays away, which is the whole point of going.
	case d.away.IsZero():
		d.away = time.Now()
	case time.Since(d.away) >= dockWait:
		a.startDockSlide(c, b, true)
		return
	}
	// Another look later: the window hears nothing of a pointer that is not
	// on it.
	c.After(dockPoll)
}

// inPanel reports whether the pointer is on the panel, or — while the panel
// is away — at the edge it went to, where the sliver of it is.
func (a *App) inPanel(x, y float32, b mygo.Rectangle) bool {
	if !a.dock.hidden {
		return x >= float32(b.X) && x <= float32(b.X+b.Width) &&
			y >= float32(b.Y) && y <= float32(b.Y+b.Height)
	}
	area := a.workArea(b)
	reach := float32(dockReach)
	switch a.dock.side {
	case edgeTop:
		return y <= float32(area.Y)+reach &&
			x >= float32(b.X) && x <= float32(b.X+b.Width)
	case edgeRight:
		return x >= float32(area.X+area.Width)-reach &&
			y >= float32(b.Y) && y <= float32(b.Y+b.Height)
	case edgeLeft:
		return x <= float32(area.X)+reach &&
			y >= float32(b.Y) && y <= float32(b.Y+b.Height)
	}
	return false
}

// startDockSlide begins taking the panel away, or bringing it back.
func (a *App) startDockSlide(c *ui.Context, b mygo.Rectangle, hide bool) {
	d := &a.dock
	area := a.workArea(b)
	if hide || d.side == edgeNone {
		d.side = a.nearestEdge(b, area)
	}
	d.fromX, d.fromY = float32(b.X), float32(b.Y)
	d.toX, d.toY = a.dockedAt(d.side, b, area, hide)
	d.at = time.Now()
	d.sliding = true
	d.hidden = hide
	a.dockSlide(c)
}

// dockSlide moves the panel one frame of its way to where it is going, and
// asks for the frame after it: the next step of the slide while it is still
// on it, and the next look for the pointer once it has arrived. A slide that
// asked for neither would leave the panel where it landed, with nothing
// watching for the pointer that brings it back.
func (a *App) dockSlide(c *ui.Context) {
	d := &a.dock
	k := float32(time.Since(d.at)) / float32(dockSlide)
	if k >= 1 {
		k = 1
		d.sliding = false
	}
	e := ui.EaseOut(k)
	x := d.fromX + (d.toX-d.fromX)*e
	y := d.fromY + (d.toY-d.fromY)*e
	a.win.SetPosition(int(x+0.5), int(y+0.5))
	if d.sliding {
		c.After(dockStep)
		return
	}
	c.After(dockPoll)
}

// nearestEdge is the edge of the screen the panel is nearest: the two sides,
// or the top.
func (a *App) nearestEdge(b, area mygo.Rectangle) edge {
	cx := float32(b.X) + float32(b.Width)/2
	cy := float32(b.Y) + float32(b.Height)/2
	left := cx - float32(area.X)
	right := float32(area.X+area.Width) - cx
	top := cy - float32(area.Y)
	switch {
	case top <= left && top <= right:
		return edgeTop
	case right <= left:
		return edgeRight
	}
	return edgeLeft
}

// dockedAt is where the panel rests against e: flush with the edge when it is
// out, and all but a sliver of it off the screen when it is away.
func (a *App) dockedAt(e edge, b, area mygo.Rectangle, away bool) (float32, float32) {
	x, y := float32(b.X), float32(b.Y)
	w, h := float32(b.Width), float32(b.Height)
	sliver := float32(dockSliver)
	switch e {
	case edgeLeft:
		if away {
			x = float32(area.X) - w + sliver
		} else {
			x = float32(area.X)
		}
	case edgeRight:
		if away {
			x = float32(area.X+area.Width) - sliver
		} else {
			x = float32(area.X+area.Width) - w
		}
	case edgeTop:
		if away {
			y = float32(area.Y) - h + sliver
		} else {
			y = float32(area.Y)
		}
	}
	return x, y
}

// workArea is the part of the screen the panel may rest on: what the taskbar
// and the dock of the desktop leave.
func (a *App) workArea(b mygo.Rectangle) mygo.Rectangle {
	return mygo.Screen.DisplayNearestPoint(mygo.Point{
		X: b.X + b.Width/2,
		Y: b.Y + b.Height/2,
	}).WorkArea
}

// scale is how many pixels the desktop draws a point with: the pointer is
// measured in pixels and the window in points.
func (a *App) scale() float32 {
	if d := mygo.Screen.DisplayNearestPoint(mygo.Point{}); d.ScaleFactor > 0 {
		return float32(d.ScaleFactor)
	}
	return 1
}
