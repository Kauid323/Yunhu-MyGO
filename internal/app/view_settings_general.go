package app

import (
	"github.com/egoist/mygo/ui"
)

// generalPage is the 通用 page: how the window is laid out and lit, then how
// a chat is written in and opened. Every control on it writes a value of
// the app's prefs (prefs.go), so that what is picked here is what the
// window does, now and on the next run.
func (a *App) generalPage(c *ui.Context, s *settings) {
	gap(c, 30)
	sectionTitle(c, "settings.general.mode")
	a.modeCard(c, s)

	gap(c, 24)
	sectionTitle(c, "settings.general.appearance")
	a.lookCard(c, s)

	gap(c, 12)
	linkText(c, "settings.general.colours", nil)

	gap(c, 24)
	sectionTitle(c, "settings.chat")
	a.chatCard(c, s)

	gap(c, 24)
	sectionTitle(c, "settings.files")
	a.filesCard(c, s)

	gap(c, 24)
	sectionTitle(c, "settings.panel")
	a.panelCard(c, s)

	gap(c, 24)
	sectionTitle(c, "settings.browse")
	a.browseCard(c, s)
}

// modeCard is the two layouts the window can be in, each drawn as the window
// it would give. Picking one lays the window out that way at once: the
// classic panel is a different window, not a different page.
func (a *App) modeCard(c *ui.Context, s *settings) {
	settingCard(c, 10, 10, func() {
		ui.Row(c).Gap(12).Children(func() {
			choiceBox(c, shotWideW, a.prefs.Mode == modeEfficient, "settings.mode.efficient",
				func() { a.pickPrefs(func(p *prefs) { p.Mode = modeEfficient }) },
				func(w, h float32) { windowShot(c, w, h, dayShot, false) })
			choiceBox(c, shotWideW, a.prefs.Mode == modeClassic, "settings.mode.classic",
				func() { a.pickPrefs(func(p *prefs) { p.Mode = modeClassic }) },
				func(w, h float32) { windowShot(c, w, h, dayShot, true) })
		})
	})
}

// lookCard is the three appearances the window can be in: the light one, the
// dark one, and the one the desktop is in, which is drawn as both.
func (a *App) lookCard(c *ui.Context, s *settings) {
	settingCard(c, 10, 10, func() {
		ui.Row(c).Gap(10).Children(func() {
			choiceBox(c, shotTallW, a.prefs.Look == lookDay, "settings.look.day",
				func() { a.pickPrefs(func(p *prefs) { p.Look = lookDay }) },
				func(w, h float32) { windowShot(c, w, h, dayShot, false) })
			choiceBox(c, shotTallW, a.prefs.Look == lookNight, "settings.look.night",
				func() { a.pickPrefs(func(p *prefs) { p.Look = lookNight }) },
				func(w, h float32) { windowShot(c, w, h, nightShot, false) })
			choiceBox(c, shotTallW, a.prefs.Look == lookSystem, "settings.look.system",
				func() { a.pickPrefs(func(p *prefs) { p.Look = lookSystem }) },
				func(w, h float32) { followSystemShot(c, w, h) })
		})
	})
}

// followSystemShot is the window the desktop decides: the light one, and the
// dark one over its right half, each cut off where the other begins.
func followSystemShot(c *ui.Context, w, h float32) {
	half := w / 2
	ui.Row(c).Size(w, h).Clip().Children(func() {
		light := ui.Box(c).Width(half).Height(h).ClipX()
		light.Children(func() { windowShot(c, w, h, dayShot, false) })
		dark := ui.Box(c).Width(w - half).Height(h).ClipX()
		dark.Children(func() {
			// The same window, moved left by the half that was cut off.
			ui.Box(c).Absolute().Left(-half).Size(w, h).Children(func() {
				windowShot(c, w, h, nightShot, false)
			})
		})
	})
}

// chatCard is how a message is sent and how a chat is opened.
func (a *App) chatCard(c *ui.Context, s *settings) {
	settingCard(c, 16, 0, func() {
		settingRow(c, false, "settings.chat.sendKey", "", func() {
			selectBox(c, &s.openSend, &a.prefs.SendKey, sendKeys,
				func() { a.savePrefs() })
		})
		// Opening a chat in a window of its own is what the narrow layout
		// has no room for, hence the note.
		settingRow(c, true, "settings.chat.dblClick", "settings.chat.dblClickNote", func() {
			selectBox(c, &s.openDbl, &a.prefs.DblClick, sendKeys,
				func() { a.savePrefs() })
		})
		settingRow(c, true, "settings.chat.newWindow", "", func() {
			selectBox(c, &s.openNew, &a.prefs.NewWindow, windowWays,
				func() { a.savePrefs() })
		})
		settingRow(c, true, "settings.chat.suggest", "", func() {
			switchBox(c, &a.prefs.Suggest, a.savePrefs)
		})
		settingRow(c, true, "settings.chat.slash", "", func() {
			switchBox(c, &a.prefs.Slash, a.savePrefs)
		})
		settingRow(c, true, "settings.chat.command", "", func() {
			selectBox(c, &s.openCmd, &a.prefs.Command, commands,
				func() { a.savePrefs() })
		})
	})
}

// filesCard is what a document is opened with, and whether a page is opened
// in the desktop's own browser. Where the files are kept is the storage
// page's, where the folder itself is chosen.
func (a *App) filesCard(c *ui.Context, s *settings) {
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.files.local", "settings.files.localNote",
			&a.prefs.LocalViewer)
		a.switchRow(c, true, "settings.files.online", "settings.files.onlineNote",
			&a.prefs.OnlineViewer)
	})
}

// panelCard is the window itself: it hides at the edge of the screen when the
// pointer leaves it, it has a button in the taskbar, and closing it does one
// of two things.
func (a *App) panelCard(c *ui.Context, s *settings) {
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.panel.dock", "", &a.prefs.Dock)
		a.switchRow(c, true, "settings.panel.taskbar", "", &a.prefs.Taskbar)
		a.selectRow(c, true, "settings.panel.close", "", &s.openClose,
			&a.prefs.CloseTo, closeWays)
	})
}

// browseCard is what a page in a message is opened with.
func (a *App) browseCard(c *ui.Context, s *settings) {
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.browse.system", "", &a.prefs.DefaultBrowser)
	})
}
