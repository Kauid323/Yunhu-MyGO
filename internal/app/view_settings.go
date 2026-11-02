package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The settings window is drawn as the chat window is: the appearance and
// the language are taken from the desktop every frame, so that the two
// windows never disagree about them.

// settingsMenuW is the strip of sections down the left of the window, and
// settingsRowH one row of it.
const (
	settingsMenuW = 200
	settingsRowH  = 36
)

// settingsFrame draws the settings window: the menu down its left side, the
// page of the section that is picked beside it.
func (a *App) settingsFrame(c *ui.Context, s *settings) {
	qq = skin(a.dark(c))
	c.SetTheme(theme(qq.dark))
	i18n.Set(c.Locale())

	// The layer a drop-down is drawn in, made every frame rather than by
	// the drop-down that needs it: a panel can only animate in when it is
	// born into a parent that was there the frame before. See view.
	ui.Overlay(c, func() {})

	bar := c.TitleBar()
	ui.Row(c).Fill().Background(qq.setBg).Children(func() {
		a.settingsMenu(c, s, bar)
		a.settingsPage(c, s, bar)
	})
}

// settingsMenu is the strip of sections down the left of the window, and
// the room over them that the window's buttons take, which drags the window
// as a title bar does.
func (a *App) settingsMenu(c *ui.Context, s *settings, bar ui.TitleBar) {
	ui.Column(c).Width(settingsMenuW).FillHeight().Background(qq.setSidebar).
		Padding(0, 8, 8, 8).Gap(2).Children(func() {
		// The gap under the strip is counted in: the rows start four
		// points under the buttons, where the menu of the window this one
		// was copied from starts.
		ui.Box(c).Height(bar.Height + 2).DragWindow()
		for i, sec := range settingsSections {
			if sec.apart {
				ui.Box(c).Height(10)
			}
			a.settingsMenuRow(c, s, i, sec)
		}
	})
}

// settingsMenuRow is one section of the menu, lit while its page is the one
// shown.
func (a *App) settingsMenuRow(c *ui.Context, s *settings, i int, sec settingsSection) {
	row := ui.Row(c).Height(settingsRowH).Radius(8).Padding(0, 12).Gap(9).
		AlignItems(ui.Center).Cursor(ui.CursorPointer)
	switch {
	case s.section == i:
		row.Background(qq.setSidebarActive)
	case row.Hovered():
		row.Background(qq.setSidebarHover)
	}
	row.Children(func() {
		ui.Icon(c, sec.icon).FontSize(16).TextColor(qq.setText)
		ui.Text(c, i18n.T(sec.label)).FontSize(13).TextColor(qq.setText).
			Grow(1).MinWidth(0).SingleLine()
	})
	if row.Clicked() {
		s.section = i
		// A drop-down of the page that is left goes with it.
		s.closeSelects()
	}
}

// settingsPage is the section that is picked: its word across the top, and
// what it sets under it, which scrolls when the window is too short for it.
func (a *App) settingsPage(c *ui.Context, s *settings, bar ui.TitleBar) {
	ui.Column(c).Grow(1).FillHeight().Children(func() {
		ui.Box(c).Height(bar.Height + 4).DragWindow()
		// The word of the page is inset as the rows of its cards are, so
		// that it sits over the first of them rather than over the card.
		a.settingsTitle(c, s)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(0, 20, 30, 20).Children(func() {
				a.settingsBody(c, s)
			})
		})
	})
}

// settingsBody draws the page of the section that is picked, or the page it
// holds under its own. A section with no page yet draws nothing, which is what
// its absence here means.
func (a *App) settingsBody(c *ui.Context, s *settings) {
	if s.sub != 0 {
		switch s.sub {
		case storageSpacePage:
			a.storageSpacePage(c, s)
		case hotkeysMorePage:
			a.hotkeysMorePage(c, s)
		}
		return
	}
	switch s.section {
	case 0:
		a.generalPage(c, s)
	case 1:
		a.noticesPage(c, s)
	case 2:
		a.storagePage(c, s)
	case 3:
		a.hotkeysPage(c, s)
	case 5:
		a.signInPage(c, s)
	case 7:
		a.palettePage(c, s)
	case 8:
		a.aboutPage(c, s)
	}
}

// settingsTitle is the word across the top of the page: the section's, or the
// page under it, with the way back to the section beside it.
func (a *App) settingsTitle(c *ui.Context, s *settings) {
	ui.Row(c).Height(24).Gap(6).AlignItems(ui.Center).Padding(0, 37, 0, 20).
		Children(func() {
			if s.sub != 0 {
				back := ui.Box(c).Size(18, 18).Radius(4).Center().
					Cursor(ui.CursorPointer)
				if back.Hovered() {
					back.Background(qq.setSidebarHover)
				}
				back.Children(func() {
					ui.Icon(c, iconChevronLeft).FontSize(14).TextColor(qq.setText)
				})
				if back.Clicked() {
					s.sub = 0
					s.closeSelects()
				}
			}
			ui.Text(c, i18n.T(a.settingsTitleKey(s))).FontSize(17).Bold().
				TextColor(qq.setText).SingleLine()
		})
}

// settingsTitleKey is the key of the word over the page.
func (a *App) settingsTitleKey(s *settings) string {
	if s.sub == storageSpacePage {
		return "settings.storage.space"
	}
	if s.sub == hotkeysMorePage {
		return "settings.hotkeys.more"
	}
	return settingsSections[s.section].label
}
