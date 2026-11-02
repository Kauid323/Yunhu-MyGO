package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The settings are one window: the sections down its left side, the one
// that is picked beside them. A section is not a window of its own — a
// window per section would be a window per switch — and the window is
// opened from the menu at the foot of the rail, as the picture window is
// opened from a picture.
//
// What the window shows is the app's prefs, not a copy of them: a choice is
// made once, and both the window that offers it and the window that follows
// it read the same value (prefs.go).

// settingsW and settingsH are the window the settings open in: wide enough
// for a card of rows with the menu beside it.
const (
	settingsW = 840
	settingsH = 720
)

// settingsSection is one row of the window's menu: its icon and its word.
type settingsSection struct {
	icon  *ui.SVG
	label string // a key of the dictionary
	// apart puts a gap above the row: the window's own settings are one
	// group, and the account's another.
	apart bool
}

// settingsSections is that menu, in the order it is drawn. Every row is
// drawn: the ones whose page is still to be written show a page with
// nothing on it, which is what their absence from settingsBody means.
var settingsSections = []settingsSection{
	{icon: iconSliders, label: "settings.nav.general"},
	{icon: iconBell, label: "settings.nav.notices"},
	{icon: iconDisk, label: "settings.nav.storage"},
	{icon: iconKeyboard, label: "settings.nav.shortcuts"},
	{icon: iconPeople, label: "settings.nav.permissions"},
	{icon: iconSettings, label: "settings.nav.signIn", apart: true},
	{icon: iconLink, label: "settings.nav.binding"},
	{icon: iconPalette, label: "settings.nav.palette"},
	{icon: iconInfo, label: "settings.nav.about"},
}

// The two layouts the window can be in: the one with the chat beside the
// list, and the narrow panel that opens a chat in a window of its own.
const (
	modeEfficient = iota
	modeClassic
)

// The appearances the 通用 page offers. They are kept as they are picked,
// but the window still takes the desktop's: an appearance of its own would
// be a second place to look for the one the desktop already has.
const (
	lookDay = iota
	lookNight
	lookSystem
)

// The values the drop-downs of the pages take, as the last part of their
// keys of the dictionary: a setting kept as a key is said in the language
// the window is in when it is drawn, not in the one it was picked in.
var (
	sendKeys    = []string{"enter", "ctrlEnter"}
	windowWays  = []string{"merge", "apart"}
	commands    = []string{"expression", "command"}
	closeWays   = []string{"tray", "quit"}
	noticeKinds = []string{"windows", "qq", "off"}
	trayKinds   = []string{"face", "box"}
	muteWays    = []string{"notify", "quiet", "blocked"}
)

// settings is the settings window: which page it is on, and which of that
// page's drop-downs is open.
type settings struct {
	win *mygo.Window
	// section is the row of the menu whose page is shown, and sub the page
	// inside it that has one of its own: the storage space under the
	// storage, the more hotkeys under the hotkeys. Zero is the section's
	// own page.
	section, sub int

	// Which of the pages' drop-downs is open, one field each: a press
	// anywhere else closes the one that is, and a press on another closes
	// it and opens that one. openGroup is the chat whose way of arriving is
	// open, which is a drop-down per row of a list.
	openSend, openDbl, openNew, openCmd, openClose bool
	openNotice, openTray, openMute                 bool
	openGroup                                      string
}

// OpenSettings opens the settings window, or brings the one already open
// forward with the section it was left on: there is one window, whatever it
// was last showing.
func (a *App) OpenSettings() {
	s := a.settings
	if s == nil {
		s = &settings{}
		a.settings = s
	} else if s.win != nil && !s.win.IsDestroyed() {
		s.win.Focus()
		return
	}
	if a.win == nil {
		// No window to open one beside, as in the tests.
		return
	}
	// The window is made after the frame that asked for it: one made
	// while a view is being built is made in the middle of a paint.
	a.update(func() { a.openSettingsWindow(s) })
}

// openSettingsWindow makes the window, on the main thread and outside a
// frame.
func (a *App) openSettingsWindow(s *settings) {
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:     i18n.T("settings.title"),
		Width:     settingsW,
		Height:    settingsH,
		MinWidth:  720,
		MinHeight: 520,
		// The window keeps the desktop's own buttons and draws under
		// them: the menu goes up to the top of the window, as the
		// sections of the window it was copied from do.
		TitleBarStyle: mygo.TitleBarHidden,
		Content:       ui.View(func(c *ui.Context) { a.settingsFrame(c, s) }),
	})
	win.OnClosed(func() {
		a.update(func() {
			if a.settings == s {
				a.settings = nil
			}
		})
	})
	s.win = win
}

// closeSelects closes every drop-down of the page: one is closed when the
// section it belongs to is left, or another of them is opened.
func (s *settings) closeSelects() {
	s.openSend, s.openDbl, s.openNew, s.openCmd, s.openClose = false, false, false, false, false
	s.openNotice, s.openTray, s.openMute = false, false, false
	s.openGroup = ""
}
