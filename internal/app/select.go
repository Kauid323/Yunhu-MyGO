package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// A drop-down is how a setting that takes one of a few values is set: the
// row shows the value it holds, and pressing it opens the rest over it, the
// one that is held marked. It is a component and not part of a page, so that
// every drop-down in the window is the same size and opens the same way.
const (
	selectW    = 148
	selectH    = 24
	selectRowH = 28
)

// selectArrives is how the list comes: it grows into its size as it fades
// in, from where the row it replaces was, so that it reads as the row
// opening rather than as a panel appearing over it.
var selectArrives = ui.ElementTransition{
	Duration: 120 * time.Millisecond,
	Ease:     ui.EaseOut,
	Enter:    &ui.Motion{Y: -4, Opacity: 0, Scale: 0.98},
	Exit:     &ui.Motion{Opacity: 0},
}

// selectBox draws a drop-down holding *value, one of options, which are the
// last part of their keys of the dictionary. open is whether its list is
// shown: the list closes itself on a press outside it or on Escape, and the
// value it is given is taken, which changed hears about. It is called for a
// choice made here and not for one made elsewhere, so that a setting is
// kept once rather than every frame.
func selectBox(c *ui.Context, open *bool, value *string, options []string, changed func()) {
	row := ui.Row(c).Width(selectW).Height(selectH).Radius(6).
		Padding(0, 8).Gap(6).AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Border(1, qq.setField)
	switch {
	case *open:
		row.Background(qq.setFieldHover).Border(1, qq.setFieldOpen)
	case row.Hovered():
		row.Background(qq.setFieldHover)
	}
	row.Children(func() {
		ui.Text(c, i18n.T("settings.value."+*value)).FontSize(13).
			TextColor(qq.setText).Grow(1).MinWidth(0).SingleLine()
		ui.Icon(c, iconChevron).FontSize(11).TextColor(qq.setTextMuted)
	})
	if row.Clicked() {
		*open = !*open
	}
	if !*open {
		return
	}
	// The list covers the row it belongs to, its first option where the
	// row was: the row is what the pointer was on, and the list is what it
	// is choosing in.
	ui.PopoverBase(c, row, open, func(panel ui.Element) {
		panel.AttachTo(row, ui.AnchorTopLeft, ui.AnchorTopLeft).
			Width(selectW).Padding(2).Radius(8).
			Background(qq.setPopup).Border(1, qq.setFieldOpen).
			Shadow(0, 8, 20, 0, ui.RGBA(0, 0, 0, 0.35)).
			Transition(selectArrives)
		for _, opt := range options {
			selectItem(c, open, value, opt, changed)
		}
	})
}

// selectItem is one option of an open list: its word, and the mark on the
// one that is held.
func selectItem(c *ui.Context, open *bool, value *string, opt string, changed func()) {
	row := ui.Row(c).Height(selectRowH).Padding(0, 10).Gap(6).Radius(6).
		AlignItems(ui.Center).Cursor(ui.CursorPointer)
	if *value == opt {
		row.Background(qq.setPopupHover)
	} else if row.Hovered() {
		row.Background(qq.setPopupHover)
	}
	row.Children(func() {
		ui.Text(c, i18n.T("settings.value."+opt)).FontSize(13).
			TextColor(qq.setText).Grow(1).MinWidth(0).SingleLine()
		if *value == opt {
			ui.Icon(c, iconCheck).FontSize(12).TextColor(qq.setText)
		}
	})
	if row.Clicked() {
		*value = opt
		*open = false
		if changed != nil {
			changed()
		}
	}
}
