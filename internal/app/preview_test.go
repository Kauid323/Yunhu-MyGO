package app

import (
	"image"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

// zoomed is a picture open in a 600×400 window at twice the size that fits
// in it: 1200×750 of picture, of which half or more is out of view.
func zoomed() *viewer {
	return &viewer{
		url: picURL, w: 480, h: 300,
		winW: 600, winH: 400,
		zoom:  2,
		viewW: 600, viewH: 400,
		drawnW: 1200, drawnH: 750,
	}
}

// TestWheelZoomsThePicture checks that Ctrl and the wheel zoom the picture,
// and that a wheel on its own is left to whatever scrolls: a picture in its
// own window has nothing to scroll.
func TestWheelZoomsThePicture(t *testing.T) {
	v := &viewer{zoom: 1, viewW: 600, viewH: 400, drawnW: 600, drawnH: 375}

	if v.input(ui.InputEvent{Kind: ui.InputScroll, X: 300, Y: 200, DY: -120}) {
		t.Error("a wheel on its own was taken as a zoom")
	}
	if v.zoom != 1 {
		t.Errorf("a wheel on its own zoomed to %v", v.zoom)
	}

	if !v.input(ui.InputEvent{Kind: ui.InputScroll, Mods: ui.Ctrl, X: 300, Y: 200, DY: -120}) {
		t.Fatal("Ctrl and the wheel up was not taken")
	}
	if v.zoom <= 1 {
		t.Errorf("Ctrl and the wheel up zoomed to %v, want more than 1", v.zoom)
	}

	// And the other way, by as much as it came.
	if !v.input(ui.InputEvent{Kind: ui.InputScroll, Mods: ui.Ctrl, X: 300, Y: 200, DY: 120}) {
		t.Fatal("Ctrl and the wheel down was not taken")
	}
	if math.Abs(float64(v.zoom-1)) > 1e-4 {
		t.Errorf("the wheel down left the zoom at %v, want it back at 1", v.zoom)
	}
}

// TestZoomStaysWithinBounds checks that a picture zooms no further than a
// quarter of its size and no larger than eight times it.
func TestZoomStaysWithinBounds(t *testing.T) {
	v := zoomed()
	for i := 0; i < 20; i++ {
		v.zoomAt(300, 200, 2)
	}
	if v.zoom != maxZoom {
		t.Errorf("zooming in stopped at %v, want %v", v.zoom, maxZoom)
	}
	for i := 0; i < 40; i++ {
		v.zoomAt(300, 200, 0.5)
	}
	if v.zoom != minZoom {
		t.Errorf("zooming out stopped at %v, want %v", v.zoom, minZoom)
	}
}

// TestZoomKeepsThePointUnderThePointer checks that what is under the
// pointer stays under it as the picture zooms: the point being looked at is
// the point being zoomed into.
func TestZoomKeepsThePointUnderThePointer(t *testing.T) {
	v := zoomed()
	// Where the middle of the picture is drawn, and how large a pixel of
	// it is: the picture is 480 wide over 1200 drawn.
	cx, cy := v.viewW/2+v.offX, v.viewH/2+v.offY
	k := v.drawnW / 480
	// The picture pixel under the pointer, at a pointer away from the
	// middle in both directions.
	x, y := float32(420), float32(260)
	u, w := (x-cx)/k, (y-cy)/k

	v.zoomAt(x, y, 1.5)

	cx, cy = v.viewW/2+v.offX, v.viewH/2+v.offY
	k = v.drawnW / 480
	if gotX, gotY := cx+u*k, cy+w*k; math.Abs(float64(gotX-x)) > 0.01 || math.Abs(float64(gotY-y)) > 0.01 {
		t.Errorf("the point under the pointer moved to %v, %v, want %v, %v", gotX, gotY, x, y)
	}
}

// TestDragMovesThePicture checks that a press moves the picture about, that
// letting go ends the move, and that the picture cannot be dragged out of
// the window.
func TestDragMovesThePicture(t *testing.T) {
	v := zoomed()

	if !v.input(ui.InputEvent{Kind: ui.InputPointerDown, X: 100, Y: 100}) {
		t.Fatal("a press on the picture was not taken")
	}
	v.input(ui.InputEvent{Kind: ui.InputPointerMove, X: 140, Y: 130})
	if v.offX != 40 || v.offY != 30 {
		t.Errorf("the picture moved to %v, %v, want 40, 30", v.offX, v.offY)
	}

	// Far past the edge, it stops where what is out of view ends.
	v.input(ui.InputEvent{Kind: ui.InputPointerMove, X: 5000, Y: 5000})
	if want := (v.drawnW - v.viewW) / 2; v.offX != want {
		t.Errorf("the picture was dragged to %v, want no further than %v", v.offX, want)
	}

	v.input(ui.InputEvent{Kind: ui.InputPointerUp})
	if v.drag {
		t.Error("the picture is still being moved after the pointer was let go")
	}
	if v.input(ui.InputEvent{Kind: ui.InputPointerMove, X: 0, Y: 0}) {
		t.Error("a move with no press was taken")
	}
}

// TestASmallPictureDoesNotMove checks that a picture which fits in the
// window is not dragged about: there is nothing of it out of view to bring
// in.
func TestASmallPictureDoesNotMove(t *testing.T) {
	v := &viewer{zoom: 1, viewW: 600, viewH: 400, drawnW: 300, drawnH: 188}
	v.input(ui.InputEvent{Kind: ui.InputPointerDown, X: 300, Y: 200})
	v.input(ui.InputEvent{Kind: ui.InputPointerMove, X: 420, Y: 300})
	if v.offX != 0 || v.offY != 0 {
		t.Errorf("a picture that fits was moved to %v, %v", v.offX, v.offY)
	}
}

// TestZoomChangesWhatIsDrawn checks the drawing itself: a picture zoomed
// out covers less of the window than one that fits, and one zoomed in more.
func TestZoomChangesWhatIsDrawn(t *testing.T) {
	a := openChat(0)
	seed(t, a, picURL, 480, 300)

	draw := func(zoom float32) int {
		v := &viewer{url: picURL, w: 480, h: 300, winW: 600, winH: 400, zoom: zoom}
		tt := ui.NewTester(func(c *ui.Context) { a.viewerFrame(c, v) }, 600, 400)
		// The first frame has no bounds to fit the picture into.
		tt.Frame()
		tt.Frame()
		return purplePixels(tt.Image(), image.Rect(0, 0, 600, 400))
	}

	small, whole, big := draw(0.5), draw(1), draw(2)
	if small >= whole {
		t.Errorf("a picture zoomed out covers %d pixels, not fewer than the %d it fits in", small, whole)
	}
	if big < whole {
		t.Errorf("a picture zoomed in covers %d pixels, fewer than the %d it fits in", big, whole)
	}
	// At twice the size it more than fills the window.
	if big < 600*400*9/10 {
		t.Errorf("a picture at twice the window's size covers only %d of its %d pixels", big, 600*400)
	}
}
