package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The bar of the pick mode lies over the foot of the history: how many
// messages are chosen, and what can be done with them. It is over the list
// rather than under the composer, so that choosing does not push the chat
// around, as the button that goes to what is unread is over its top.
func (a *App) pickBar(c *ui.Context) {
	if !a.picking {
		return
	}
	ui.Row(c).Absolute().Left(0).Right(0).Bottom(0).Justify(ui.Center).
		Padding(0, 0, 12, 0).Children(func() {
		// The bar hugs what it holds rather than filling the window: a
		// strip across the whole history would read as a part of the list
		// rather than as something laid over it.
		bar := ui.Box(c).Height(34).Padding(0, 6).Radius(17).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 4, 14, 0, ui.RGBA(0, 0, 0, 0.22))
		bar.Children(func() {
			ui.Row(c).Gap(2).AlignItems(ui.Center).Children(func() {
				n := len(a.picked)
				text := i18n.T("select.none")
				if n > 0 {
					text = i18n.T("select.title", i18n.Num(n))
				}
				ui.Text(c, text).FontSize(11.5).TextColor(qq.textMuted).
					Padding(0, 8, 0, 6).SingleLine()
				a.pickButton(c, i18n.T("select.all"), a.pickAll, len(a.msgs) > 0)
				a.pickButton(c, i18n.T("select.forward"), func() { a.pickForward(bar.Bounds()) }, n > 0)
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
// from: a circle that fills with a tick when the message is chosen.
func (a *App) pickBox(c *ui.Context, m *yh.Message) {
	if !a.picking {
		return
	}
	on := a.picked[m.MsgID]
	box := ui.Box(c).Size(20, 20).Radius(10).Center().
		Cursor(ui.CursorPointer).Margin(0, 2, 0, 2)
	if on {
		box.Background(qq.accent).Border(0, qq.accent)
	} else {
		box.Border(1.5, qq.textFaint)
	}
	box.Children(func() {
		if on {
			ui.Icon(c, iconCheck).FontSize(12).TextColor(ui.RGBA(255, 255, 255, 1))
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
