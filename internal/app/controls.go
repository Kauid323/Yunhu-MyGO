package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The parts every settings page is made of: the card a group of settings
// sits on, the rows inside it, the word over a group, and the pictures a
// page offers to pick from. They are here rather than in a view_*.go so
// that every page is drawn by the same code and comes out the same, as the
// menu of the rail is: one file, one set of calls, whichever page uses it.

// The heights of a row of a card: one line, and a line with a note under
// it. The card insets its rows by its own padding, and the line between two
// of them goes as far as the rows do.
const (
	settingRowH     = 44
	settingRowNoteH = 64
)

// settingCard draws the ground a group of settings sits on, holding what fn
// builds.
func settingCard(c *ui.Context, padX, padY float32, fn func()) {
	ui.Column(c).Radius(8).Background(qq.setCard).Padding(padY, padX).Children(fn)
}

// sectionTitle is the word over a group of settings, with the room under it
// that the group's card takes.
func sectionTitle(c *ui.Context, key string) {
	ui.Row(c).Height(18).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.T(key)).FontSize(12.5).FontWeight(500).
			TextColor(qq.setText).SingleLine()
	})
	ui.Box(c).Height(10)
}

// gap is room between two parts of a page.
func gap(c *ui.Context, h float32) { ui.Box(c).Height(h) }

// linkText is a line that does something when it is pressed, as the way to
// the colour themes is: drawn as a link, in the colour the window links in.
// A line whose feature is still to be written takes no do, and is drawn as
// the others are.
func linkText(c *ui.Context, key string, do func()) {
	t := ui.Text(c, i18n.T(key)).FontSize(12.5).TextColor(qq.setLink).
		Cursor(ui.CursorPointer)
	if do != nil && t.Clicked() {
		do()
	}
}

// settingRow is one row of a card: the name of what it sets, the note under
// the name if it has one, and the control that sets it at the right.
// divider draws the line above the row, which every row but the first has:
// a card has no line at either end. The note is a key of the dictionary.
func settingRow(c *ui.Context, divider bool, key, note string, control func()) {
	settingRowText(c, divider, key, i18n.T(note), control)
}

// settingRowText is settingRow with the note as it is written, for a line
// that is a value rather than a word of the dictionary: the path of a folder,
// say, which is the same in every language.
func settingRowText(c *ui.Context, divider bool, key, note string, control func()) {
	if divider {
		ui.Box(c).Height(1).Background(qq.setLine)
	}
	row := ui.Row(c).Height(settingRowH).Gap(12).AlignItems(ui.Center)
	if note != "" {
		row.Height(settingRowNoteH)
	}
	row.Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, i18n.T(key)).FontSize(13.5).TextColor(qq.setText).SingleLine()
			if note != "" {
				ui.Text(c, note).FontSize(11.5).
					TextColor(qq.setTextMuted).SingleLine()
			}
		})
		control()
	})
}

// switchRow is a row whose control is a switch over one of the settings.
func (a *App) switchRow(c *ui.Context, divider bool, key, note string, v *bool) {
	settingRow(c, divider, key, note, func() { switchBox(c, v, a.savePrefs) })
}

// selectRow is a row whose control is a drop-down over one of the settings.
func (a *App) selectRow(c *ui.Context, divider bool, key, note string, open *bool, value *string, options []string) {
	settingRow(c, divider, key, note, func() {
		selectBox(c, open, value, options, a.savePrefs)
	})
}

// The pictures a page offers: the box one is drawn in, which is as wide as
// its word under it, and the frame the picture sits in, which the frame's
// line takes two points off on every side.
const (
	shotH     = 70
	shotWideW = 104
	shotTallW = 100
)

// choiceBox is one of the pictures a page offers: the window it would give,
// and its word under it. The one that is taken is framed in the accent;
// pressing either takes it.
func choiceBox(c *ui.Context, w float32, selected bool, label string, do func(), shot func(w, h float32)) {
	col := ui.Column(c).Width(w).Gap(7).Cursor(ui.CursorPointer)
	col.Children(func() {
		frame := ui.Box(c).Width(w).Height(shotH).Radius(6).Border(2, ui.Transparent)
		switch {
		case selected:
			frame.Border(2, qq.setAccent)
		case frame.Hovered():
			frame.Border(2, qq.setField)
		}
		frame.Children(func() { shot(w-4, shotH-4) })
		// The word takes the room of a line whether or not it has one
		// under it, so that a card of pictures is as tall as its pictures
		// and the line under them.
		ui.Row(c).Height(17).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T(label)).FontSize(11.5).TextColor(qq.setText).
				Width(w).TextAlign(ui.Center).SingleLine()
		})
	})
	if col.Clicked() {
		do()
	}
}

// windowShot draws the window a choice would give, as small as a thumbnail
// holds it: the rail, the list and the chat beside it, or only the list for
// the layout that has no chat. It is a picture of the window, so it is drawn
// in the appearance it shows rather than in this window's colors.
func windowShot(c *ui.Context, w, h float32, p shotSkin, narrow bool) {
	ui.Box(c).Size(w, h).Radius(3).Background(p.ground).Children(func() {
		if narrow {
			narrowShot(c, w, h, p)
			return
		}
		wideShot(c, w, h, p)
	})
}

// wideShot is the window with the list and the chat side by side: the bar
// across its top, the rail down its left, the chats beside it and the
// messages that were said in the one that is open.
func wideShot(c *ui.Context, w, h float32, p shotSkin) {
	win := ui.Box(c).Absolute().Left(2).Top(2).Size(w-4, h-4).Radius(2).
		Background(p.window)
	win.Children(func() {
		shotBar(c, w-4, p)
		// The rail, with the sections on it.
		ui.Box(c).Absolute().Left(0).Top(6).Size(8, h-10).
			Background(p.bar).Children(func() {
			for i := 0; i < 4; i++ {
				ui.Box(c).Absolute().Left(2.5).Top(float32(3+i*6)).Size(3, 3).
					Radius(1).Background(p.mark)
			}
		})
		// The chats, each a face and the two lines under it.
		for i := 0; i < 5; i++ {
			top := float32(9 + i*9)
			ui.Box(c).Absolute().Left(11).Top(top).Size(3.5, 3.5).
				Radius(2).Background(p.mark)
			ui.Box(c).Absolute().Left(17).Top(top).Size(14, 2).
				Radius(1).Background(p.mark)
			ui.Box(c).Absolute().Left(17).Top(top+4).Size(9, 1.6).
				Radius(1).Background(p.mark)
		}
		// The chat that is open: what was typed in it, and the messages
		// said back, which sit on the right as the window puts them.
		ui.Box(c).Absolute().Left(w*0.42).Top(9).Size(w*0.2, 3.5).
			Radius(2).Background(p.mark)
		for i := 0; i < 3; i++ {
			bw := float32(18)
			if i == 1 {
				bw = 13
			}
			ui.Box(c).Absolute().Right(9).Top(float32(17+i*7)).Size(bw, 4).
				Radius(2).Background(p.bubble)
			ui.Box(c).Absolute().Right(5).Top(float32(18+i*7)).Size(2.5, 2.5).
				Radius(2).Background(p.mark)
		}
		ui.Box(c).Absolute().Left(w*0.42).Bottom(4).Size(w*0.48, 3.5).
			Radius(2).Background(p.mark)
	})
}

// narrowShot is the window that is only a list, which stands in the middle
// of the desktop rather than filling it.
func narrowShot(c *ui.Context, w, h float32, p shotSkin) {
	win := ui.Box(c).Absolute().Left(w*0.29).Top(3).Size(w*0.42, h-6).
		Radius(2).Background(p.window)
	win.Children(func() {
		ui.Box(c).Absolute().Left(3).Top(3).Size(4, 4).Radius(1).
			Background(p.mark)
		ui.Box(c).Absolute().Left(10).Top(3.5).Size(12, 3).Radius(2).
			Background(p.mark)
		for i := 0; i < 6; i++ {
			top := float32(12 + i*8)
			ui.Box(c).Absolute().Left(4).Top(top).Size(3, 3).Radius(2).
				Background(p.mark)
			ui.Box(c).Absolute().Left(9).Top(top).Size(14, 1.8).Radius(1).
				Background(p.mark)
			ui.Box(c).Absolute().Left(9).Top(top+3.4).Size(8, 1.4).
				Radius(1).Background(p.mark)
		}
	})
}

// shotBar is the bar across the top of the window a thumbnail shows, with
// the buttons of the window at its right.
func shotBar(c *ui.Context, w float32, p shotSkin) {
	ui.Box(c).Absolute().Left(0).Top(0).Size(w, 6).Background(p.bar).
		Children(func() {
			for i := 0; i < 3; i++ {
				ui.Box(c).Absolute().Right(float32(4+i*3)).Top(2.5).Size(2, 1).
					Radius(0.5).Background(p.mark)
			}
			ui.Box(c).Absolute().Left(3).Top(2.5).Size(10, 1).
				Radius(0.5).Background(p.mark)
		})
}
