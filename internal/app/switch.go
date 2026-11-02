package app

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// A switch is how a setting that is either on or off is set: a track with a
// knob at one end of it. The knob slides to the other end and the track
// takes the accent as it goes, so that the change is seen rather than read,
// and a switch pressed twice in a row turns back from where the knob had
// got to instead of jumping to the other end.
const (
	switchW     = 28
	switchH     = 16
	switchKnob  = 10
	switchInset = 3
)

// switchMoves is how the knob and the track go from one end to the other: a
// short curve rather than a spring, so that the knob lands where it goes
// without passing it. The desktop asking for less motion puts it there at
// once, which ui does on its own.
var switchMoves = ui.Animation{Duration: 170 * time.Millisecond, Ease: ui.EaseOut}

// switchBox draws a switch that turns v on and off, and calls changed when
// it is turned: the setting it stands for is kept there, once a press,
// rather than every frame it is drawn.
func switchBox(c *ui.Context, v *bool, changed func()) {
	box := ui.Box(c).Size(switchW, switchH).Radius(switchH / 2).
		Cursor(ui.CursorPointer)
	// How far the knob has come, from 0 with the switch off to 1 with it
	// on: the frame after it is pressed it is part of the way, and the
	// frames after that take it the rest of it.
	t := box.AnimateTo("on", onOff(*v), switchMoves)
	box.Background(qq.setSwitchOff.Mix(qq.setSwitchOn, t))
	box.Children(func() {
		ui.Box(c).Absolute().
			Left(switchInset+t*(switchW-2*switchInset-switchKnob)).
			Top((switchH-switchKnob)/2).
			Size(switchKnob, switchKnob).Radius(switchKnob / 2).
			Background(qq.setSwitchKnob)
	})
	if box.Clicked() {
		*v = !*v
		if changed != nil {
			changed()
		}
	}
}

// onOff is a switch's state as the number it is drawn by.
func onOff(on bool) float32 {
	if on {
		return 1
	}
	return 0
}
