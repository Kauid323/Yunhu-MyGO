package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The menu is the popup a button opens, as the one at the foot of the
// rail: a row of quick actions across its top, a line, then its rows. It
// is a component and not part of a screen, so that every menu in the
// window is drawn by the same code and comes out the same: showMenu puts
// one beside any element, above it where the window has no room below.
const (
	menuWidth = 180
	menuRowH  = 30
)

// menuArrives is how a menu comes and goes: it rises the few DIPs from
// where it opens, growing into its size as it fades in, so that it reads
// as coming out of the button rather than being there all along. The
// desktop asking for less motion fades it in place instead.
var menuArrives = ui.ElementTransition{
	Duration: 140 * time.Millisecond,
	Ease:     ui.EaseOut,
	Enter:    &ui.Motion{Y: 10, Opacity: 0, Scale: 0.94},
	Exit:     &ui.Motion{Y: 6, Opacity: 0, Scale: 0.97},
}

// menu is one popup menu: what it holds, and whether it is open.
type menu struct {
	open bool
	// width is how wide it is, and menuWidth when it is not said: a menu
	// in a panel 320 wide cannot be as wide as one in the window.
	width float32
	// below opens it under its button rather than above it, which is what
	// a menu at the top of a window needs: the rail's button is at the
	// foot, and its menu opens up and away from it.
	below bool
	// point hangs it off a place rather than off a button: a menu opened
	// by a right press has no button to come out of, and opens from the
	// point that was pressed.
	point bool
	quick []menuItem
	rows  []menuItem
}

// menuItem is a row of a menu, or one of the quick actions across its top.
// A row with no do is one whose feature is still to be written: it is
// drawn as the others are, and choosing it closes the menu.
type menuItem struct {
	icon  *ui.SVG
	label string // a key of the dictionary
	more  bool   // the row opens a menu of its own, and says so with a chevron
	// dim is a row that is drawn but cannot be taken: a feature the
	// client does not have yet is shown as one that is not there rather
	// than left out, as the client it copies dims what it cannot do.
	dim bool
	// text is a label that is not a key of the dictionary, as the name
	// of a chat is: a menu of chats says what they are called.
	text string
	do   func()
}

// showMenu draws m beside anchor while it is open, and closes it on a
// press outside it or on Escape, which the popup does itself. anchor is
// built before it, in the same frame.
func showMenu(c *ui.Context, anchor ui.Element, m *menu) {
	if !m.open {
		return
	}
	width := m.width
	if width == 0 {
		width = menuWidth
	}
	at, self := ui.AnchorBottomRight, ui.AnchorBottomLeft
	switch {
	case m.point:
		// Its top left on the place that was pressed, opening away from
		// it and folding back where the window has no room.
		at, self = ui.AnchorTopLeft, ui.AnchorTopLeft
	case m.below:
		// Hanging straight down from the button's right edge, so that a
		// menu at the top of a window opens over what is under it rather
		// than off the screen.
		at, self = ui.AnchorBottomRight, ui.AnchorTopRight
	}
	ui.PopoverBase(c, anchor, &m.open, func(panel ui.Element) {
		// The bottom left of the panel sits at the bottom right of the
		// button, so that the two end on the same line and the menu opens
		// up and away from the rail.
		panel.AttachTo(anchor, at, self).
			Width(width).Padding(6).Radius(10).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 8, 22, 0, ui.RGBA(0, 0, 0, 0.30)).
			TransformOrigin(0, 1).Transition(menuArrives)
		if !m.below && !m.point {
			panel.Left(8)
		}
		menuBody(c, m)
	})
}

// menuBody draws what a menu holds.
func menuBody(c *ui.Context, m *menu) {
	if len(m.quick) > 0 {
		ui.Row(c).Gap(0).Children(func() {
			for _, it := range m.quick {
				menuQuick(c, m, it)
			}
		})
		ui.Box(c).Height(1).Background(qq.menuLine).Margin(5, 4, 5, 4)
	}
	for _, it := range m.rows {
		menuRow(c, m, it)
	}
}

// menuQuick is one of the quick actions across the top of a menu: an icon
// with its word under it.
func menuQuick(c *ui.Context, m *menu, it menuItem) {
	box := ui.Column(c).Grow(1).Padding(7, 2).Gap(5).Center().Radius(7).
		Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.menuHover)
	}
	box.Children(func() {
		ui.Icon(c, it.icon).FontSize(23).TextColor(qq.menuText)
		ui.Text(c, i18n.T(it.label)).FontSize(12).TextColor(qq.menuText).SingleLine()
	})
	if box.Clicked() {
		m.open = false
		if it.do != nil {
			it.do()
		}
	}
}

// menuRow is one row of a menu: its icon, its word, and the chevron of a
// row that opens more.
func menuRow(c *ui.Context, m *menu, it menuItem) {
	row := ui.Row(c).Height(menuRowH).Padding(0, 8).Gap(11).Radius(6).
		AlignItems(ui.Center)
	if !it.dim {
		row.Cursor(ui.CursorPointer)
	}
	if row.Hovered() && !it.dim {
		row.Background(qq.menuHover)
	}
	row.Children(func() {
		ink := qq.menuText
		if it.dim {
			ink = qq.menuTextMuted
		}
		ui.Icon(c, it.icon).FontSize(19).TextColor(ink)
		label := it.text
		if label == "" {
			label = i18n.T(it.label)
		}
		ui.Text(c, label).FontSize(13.5).TextColor(ink).
			Grow(1).MinWidth(0).SingleLine()
		if it.more {
			ui.Icon(c, iconChevronRight).FontSize(12).TextColor(qq.menuTextMuted)
		}
	})
	if row.Clicked() && !it.dim {
		m.open = false
		if it.do != nil {
			it.do()
		}
	}
}

// moreMenu is the menu at the foot of the rail: what the window offers
// that has no section of its own. The rows are all drawn; the ones with no
// feature behind them yet do nothing, which is what their absence from
// here means.
func moreMenu(a *App) menu {
	return menu{
		quick: []menuItem{
			{icon: iconBookmark, label: "menu.favorites"},
			{icon: iconFolder, label: "menu.files", do: a.openFilesDir},
			{icon: iconPalette, label: "menu.palette"},
		},
		rows: []menuItem{
			{icon: iconHistoryLog, label: "menu.history"},
			{icon: iconUpdate, label: "menu.update"},
			{icon: iconHelp, label: "menu.help", do: a.openHelp},
			{icon: iconLock, label: "menu.lock"},
			{icon: iconSettings, label: "menu.settings", do: a.OpenSettings},
			{icon: iconPower, label: "menu.logout", do: a.logout},
		},
	}
}
