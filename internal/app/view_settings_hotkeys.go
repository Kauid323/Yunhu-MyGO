package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The hotkeys page: the keys the window answers to. They are the app's own —
// a key that is pressed while the window is in front — and not keys the
// desktop gives it from anywhere, which is what the panel this one copies
// asks the system for. What each key is set to is kept in the settings.
var (
	// hotkeyPage is the keys of the page itself, in the order it draws
	// them: the name of what it does, the key it comes with, and the note
	// under it.
	hotkeyPage = []hotkey{
		{key: "send", def: "Enter", label: "settings.hotkeys.send"},
		{key: "unread", def: "Ctrl + Alt + Z", label: "settings.hotkeys.unread"},
		{key: "voice", def: "Win + Alt", label: "settings.hotkeys.voice",
			note: "settings.hotkeys.voiceNote"},
		{key: "shot", def: "Ctrl + Alt + A", label: "settings.hotkeys.shot"},
		{key: "record", def: "Ctrl + Alt + R", label: "settings.hotkeys.record"},
		{key: "all", def: "Ctrl + Alt + X", label: "settings.hotkeys.all"},
		{key: "mini", def: "Ctrl + Alt + E", label: "settings.hotkeys.mini"},
	}

	// hotkeyMore is the page under it, of the keys that are not worth the
	// room of the first one.
	hotkeyMore = []hotkey{
		{key: "ocr", def: "Ctrl + Alt + O", label: "settings.hotkeys.ocr"},
		{key: "translate", def: "Ctrl + Alt + F", label: "settings.hotkeys.translate"},
		{key: "longshot", def: "Ctrl + Alt + C", label: "settings.hotkeys.longshot"},
		{key: "lock", def: "Ctrl + Alt + L", label: "settings.hotkeys.lock"},
		{key: "chats", def: "Ctrl + 1", label: "settings.hotkeys.chats"},
		{key: "contacts", def: "Ctrl + 2", label: "settings.hotkeys.contacts"},
		{key: "search", def: "Ctrl + F", label: "settings.hotkeys.search"},
	}
)

// hotkey is one row of those pages.
type hotkey struct {
	key   string // the name the settings keep it by
	def   string // the key it comes with
	label string // a key of the dictionary
	note  string
}

// hotkeysPage is the page of the keys.
func (a *App) hotkeysPage(c *ui.Context, s *settings) {
	gap(c, 30)
	settingCard(c, 16, 0, func() {
		for i, h := range hotkeyPage {
			a.hotkeyRow(c, i > 0, h)
		}
		// The way back to the keys the page came with.
		ui.Row(c).Height(40).AlignItems(ui.Center).Justify(ui.End).Children(func() {
			ui.Button(c, i18n.T("settings.hotkeys.reset")).Radius(6).
				Padding(3, 12).FontSize(12).OnClick(a.resetHotkeys)
		})
	})

	gap(c, 16)
	settingCard(c, 16, 0, func() {
		row := ui.Row(c).Height(settingRowH).Gap(10).Padding(0, 0).Radius(6).
			AlignItems(ui.Center).Cursor(ui.CursorPointer)
		if row.Hovered() {
			row.Background(qq.setFieldHover)
		}
		row.Children(func() {
			ui.Text(c, i18n.T("settings.hotkeys.more")).FontSize(13.5).
				TextColor(qq.setText).Grow(1).SingleLine()
			ui.Icon(c, iconChevronRight).FontSize(12).TextColor(qq.setTextMuted)
		})
		if row.Clicked() {
			s.sub = hotkeysMorePage
		}
	})

	gap(c, 24)
	sectionTitle(c, "settings.hotkeys.conflict")
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.hotkeys.conflictShot", "", &a.prefs.Shake)
	})
}

// hotkeysMorePage is the keys under the page of the keys.
func (a *App) hotkeysMorePage(c *ui.Context, s *settings) {
	gap(c, 30)
	settingCard(c, 16, 0, func() {
		for i, h := range hotkeyMore {
			a.hotkeyRow(c, i > 0, h)
		}
		a.switchRow(c, true, "settings.hotkeys.ctrlW", "", &a.prefs.Slash)
		ui.Row(c).Height(40).AlignItems(ui.Center).Justify(ui.End).Children(func() {
			ui.Button(c, i18n.T("settings.hotkeys.reset")).Radius(6).
				Padding(3, 12).FontSize(12).OnClick(a.resetHotkeys)
		})
	})
}

// hotkeyRow is one key: what it does, and the key itself with the way to take
// it away.
func (a *App) hotkeyRow(c *ui.Context, divider bool, h hotkey) {
	if divider {
		ui.Box(c).Height(1).Background(qq.setLine)
	}
	ui.Row(c).Height(settingRowH).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
			ui.Text(c, i18n.T(h.label)).FontSize(13.5).TextColor(qq.setText).SingleLine()
			if h.note != "" {
				ui.Text(c, i18n.T(h.note)).FontSize(11).
					TextColor(qq.setTextMuted).SingleLine()
			}
		})
		key := a.hotkey(h)
		chip := ui.Row(c).Width(140).Height(24).Radius(6).Padding(0, 8).
			AlignItems(ui.Center).Justify(ui.Center).Border(1, qq.setField)
		chip.Children(func() {
			text := key
			col := qq.setText
			if text == "" {
				text = i18n.T("settings.hotkeys.none")
				col = qq.setTextMuted
			}
			ui.Text(c, text).FontSize(12).TextColor(col).SingleLine()
		})
		clear := ui.Box(c).Size(20, 20).Radius(10).Center().
			Cursor(ui.CursorPointer).Background(qq.setFieldHover)
		if clear.Hovered() {
			clear.Background(qq.setFieldOpen)
		}
		clear.Children(func() {
			ui.Text(c, "×").FontSize(12).TextColor(qq.setTextMuted)
		})
		if clear.Clicked() {
			a.pickPrefs(func(p *prefs) { a.setHotkey(p, h, "") })
		}
	})
}

// hotkey is the key a row stands for: the one it was set to, or the one it
// comes with.
func (a *App) hotkey(h hotkey) string {
	if key, ok := a.prefs.Hotkeys[h.key]; ok {
		return key
	}
	return h.def
}

// setHotkey keeps a key for what a row does, or forgets it: a key that is
// empty is one that does nothing.
func (a *App) setHotkey(p *prefs, h hotkey, key string) {
	if p.Hotkeys == nil {
		p.Hotkeys = map[string]string{}
	}
	if key == "" {
		p.Hotkeys[h.key] = ""
		return
	}
	p.Hotkeys[h.key] = key
}

// resetHotkeys puts every key back to the one it comes with.
func (a *App) resetHotkeys() {
	a.pickPrefs(func(p *prefs) { p.Hotkeys = map[string]string{} })
}
