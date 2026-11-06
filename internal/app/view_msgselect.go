package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// pickBarH is how tall the bar of the pick mode is. The row inside it grows
// to the whole of the content box: a row only as tall as the buttons in it
// would sit at the bar's top edge, and the words would read a few DIPs high.
const pickBarH = 34

// pickArrives is how the bar comes and goes: it rises the few DIPs from the
// composer's side as it fades in, so that entering the pick mode reads as a
// tool laid over the history rather than a strip appearing on it.
var pickArrives = ui.ElementTransition{
	Duration: 150 * time.Millisecond,
	Ease:     ui.EaseOut,
	Enter:    &ui.Motion{Y: 8, Opacity: 0, Scale: 0.97},
	Exit:     &ui.Motion{Y: 6, Opacity: 0},
}

// pickFills and pickTicks are how a message is marked: the circle takes the
// colour of a chosen one over a moment, and the tick lands in it with a pop
// of its own, so that choosing reads as something being stamped rather than
// as a colour changing between two frames.
var (
	pickFills = ui.ElementTransition{Duration: 130 * time.Millisecond, Colors: true}
	pickTicks = ui.ElementTransition{
		Duration: 160 * time.Millisecond,
		Ease:     ui.EaseOut,
		Enter:    &ui.Motion{Scale: 0.35, Opacity: 0},
		Exit:     &ui.Motion{Scale: 0.6, Opacity: 0},
	}
)

// The bar of the pick mode lies over the foot of the history: how many
// messages are chosen, and what can be done with them. It is over the list
// rather than under the composer, so that choosing does not push the chat
// around, as the button that goes to what is unread is over the same foot.
func (a *App) pickBar(c *ui.Context) {
	// The row is made whether or not the bar is in it: a bar born into a
	// row that the frame before did not show would appear rather than rise.
	ui.Row(c).Absolute().Left(0).Right(0).Bottom(0).Justify(ui.Center).
		Padding(0, 0, 12, 0).Children(func() {
		if !a.picking {
			return
		}
		// The bar hugs what it holds rather than filling the window: a
		// strip across the whole history would read as a part of the list
		// rather than as something laid over it.
		bar := ui.Box(c).Height(pickBarH).Padding(0, 6).Radius(pickBarH/2).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 4, 14, 0, ui.RGBA(0, 0, 0, 0.22)).
			Transition(pickArrives)
		bar.Children(func() {
			ui.Row(c).Grow(1).Gap(2).AlignItems(ui.Center).Children(func() {
				n := len(a.picked)
				text := i18n.T("select.none")
				if n > 0 {
					text = i18n.T("select.title", i18n.Num(n))
				}
				ui.Text(c, text).FontSize(11.5).TextColor(qq.textMuted).
					Padding(0, 8, 0, 6).SingleLine()
				a.pickButton(c, i18n.T("select.all"), a.pickAll, len(a.msgs) > 0)
				a.pickButton(c, i18n.T("select.forward"), func() { a.pickForward(bar.Bounds()) }, n > 0)
				a.pickButton(c, i18n.T("select.image"),
					func() { a.messageImage(a.pickedMessages()) }, n > 0)
				a.pickButton(c, i18n.T("select.recall"), a.pickRecall, n > 0)
				a.pickButton(c, i18n.T("select.delete"), a.pickDelete, n > 0)
				a.pickButton(c, i18n.T("select.cancel"), a.stopPicking, true)
			})
		})
	})
}

// pickButton is one action of that bar, drawn as a word rather than as an
// icon: the bar says what it does, and a row of glyphs would be a puzzle.
func (a *App) pickButton(c *ui.Context, label string, do func(), on bool) {
	col := qq.accent
	if !on {
		col = qq.textFaint
	}
	box := ui.Box(c).Height(26).Padding(0, 9).Radius(13).Center().
		Cursor(ui.CursorPointer)
	if on && box.Hovered() {
		box.Background(qq.menuHover)
	}
	box.Children(func() {
		ui.Text(c, label).FontSize(11.5).TextColor(col).SingleLine()
	})
	if on && box.Clicked() {
		do()
	}
}

// pickBox is the mark beside a message while the history is being picked
// from: a circle that fills with a tick when the message is chosen. It is
// centred on the message rather than held to the top of it, so that the
// mark belongs to the whole message and not to the name over it.
func (a *App) pickBox(c *ui.Context, m *yh.Message) {
	if !a.picking {
		return
	}
	on := a.picked[m.MsgID]
	box := ui.Box(c).Size(20, 20).Radius(10).Center().
		Cursor(ui.CursorPointer).Margin(0, 2, 0, 2).
		AlignSelf(ui.Center).Transition(pickFills)
	if on {
		box.Background(qq.accent).Border(0, qq.accent)
	} else {
		box.Border(1.5, qq.textFaint)
	}
	box.Children(func() {
		if on {
			ui.Icon(c, iconCheck).FontSize(12).
				TextColor(ui.RGBA(255, 255, 255, 1)).
				Transition(pickTicks)
		}
	})
	if box.Clicked() {
		a.togglePick(m.MsgID)
	}
}

// rowPress is what a press on a message does: it chooses the message while
// the history is being picked from, and nothing otherwise.
func (a *App) rowPress(m *yh.Message, row ui.Element) {
	if !a.picking {
		return
	}
	if row.Clicked() {
		a.togglePick(m.MsgID)
	}
}
