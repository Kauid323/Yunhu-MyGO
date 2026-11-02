package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// prefs is what the settings window sets and the window does about it: how
// it is laid out, how a message is sent, and where a chat opens. It is kept
// between runs beside the session, so that a choice is made once rather
// than every time the window is opened.
//
// The settings window reads and writes this and not a copy of it: a value
// that the window that offers it and the window that follows it hold apart
// is a value that disagrees with itself.
type prefs struct {
	Mode      int    `json:"mode"`      // modeEfficient or modeClassic
	Look      int    `json:"look"`      // lookDay, lookNight or lookSystem
	SendKey   string `json:"sendKey"`   // "enter" or "ctrlEnter"
	DblClick  string `json:"dblClick"`  // what opens a chat in a window of its own
	NewWindow string `json:"newWindow"` // "merge" or "apart"
	Command   string `json:"command"`   // "expression" or "command"
	Suggest   bool   `json:"suggest"`
	Slash     bool   `json:"slash"`
	// FilesDir and MsgDir are where the files that are received and the
	// messages are kept, or "" for the folder the desktop gives the app.
	FilesDir string `json:"filesDir"`
	MsgDir   string `json:"msgDir"`
	// The files card of the general page: the built-in viewer for local and
	// for online documents, and whether a page opens in the desktop's own
	// browser.
	LocalViewer    bool `json:"localViewer"`
	OnlineViewer   bool `json:"onlineViewer"`
	DefaultBrowser bool `json:"defaultBrowser"`
	// The main panel: it hides at the edge of the screen when the pointer
	// leaves it, it has a button in the taskbar, and closing it does what
	// CloseTo says: "tray" or "quit".
	Dock    bool   `json:"dock"`
	Taskbar bool   `json:"taskbar"`
	CloseTo string `json:"closeTo"`

	// The notices page: what a notification is drawn by, whether it says
	// what the message says, and which arrivals are worth one.
	NoticeKind    string `json:"noticeKind"` // "windows", "qq" or "off"
	NoticeDetail  bool   `json:"noticeDetail"`
	NotifyFriend  bool   `json:"notifyFriend"`
	NotifyGroup   bool   `json:"notifyGroup"`
	NotifyFriendQ bool   `json:"notifyFriendReq"`
	NotifyGroupQ  bool   `json:"notifyGroupReq"`
	// The tray: a flashing face or the message box, and which arrivals
	// flash it.
	TrayKind     string `json:"trayKind"` // "face" or "box"
	TrayFriend   bool   `json:"trayFriend"`
	TrayGroup    bool   `json:"trayGroup"`
	TrayFriendQ  bool   `json:"trayFriendReq"`
	TrayGroupQ   bool   `json:"trayGroupReq"`
	Shake        bool   `json:"shake"`
	SoundNew     bool   `json:"soundNew"`
	SoundFriend  bool   `json:"soundFriend"`
	SoundGroup   bool   `json:"soundGroup"`
	SoundFriendQ bool   `json:"soundFriendReq"`
	SoundGroupQ  bool   `json:"soundGroupReq"`
	// GroupMute is how each chat is taken: "notify", "quiet" or "blocked",
	// by chat id. A chat that is not here is taken as "notify".
	GroupMute map[string]string `json:"groupMute"`

	// The sign-in page.
	AutoStart  bool `json:"autoStart"`
	AutoLogin  bool `json:"autoLogin"`
	QuickLogin bool `json:"quickLogin"`

	// Hotkeys are the keys the window answers to, by what they do: a key
	// name as the page writes it, "" for none.
	Hotkeys map[string]string `json:"hotkeys"`
	// Theme is the colour theme the palette page has picked.
	Theme string `json:"theme"`
}

// defaultPrefs is what a window that has never been set opens with: the
// layout it has had, Enter sending, and the "/" commands on.
func defaultPrefs() prefs {
	return prefs{
		Look:      lookSystem,
		SendKey:   "enter",
		DblClick:  "enter",
		NewWindow: "merge",
		Command:   "expression",
		Slash:     true,
		// A window that has never been set draws the desktop's own
		// notifications, says what a message says, and hides at the edge
		// of the screen when the pointer leaves it — what the panel this
		// client copies does out of the box.
		NoticeKind:    "windows",
		NoticeDetail:  true,
		NotifyFriend:  true,
		NotifyGroup:   true,
		NotifyFriendQ: true,
		NotifyGroupQ:  true,
		TrayKind:      "face",
		TrayFriend:    true,
		TrayGroup:     true,
		TrayFriendQ:   true,
		TrayGroupQ:    true,
		Shake:         true,
		SoundNew:      true,
		SoundFriend:   true,
		SoundGroup:    true,
		SoundFriendQ:  true,
		SoundGroupQ:   true,
		Dock:          true,
		Taskbar:       true,
		CloseTo:       "tray",
		QuickLogin:    true,
		Hotkeys:       map[string]string{},
		GroupMute:     map[string]string{},
	}
}

// muted is how a chat is taken, by the notices page: "notify", "quiet" (it
// arrives without a word) or "blocked" (it is not shown at all).
func (a *App) muted(chatID string) string {
	if way, ok := a.prefs.GroupMute[chatID]; ok && way != "" {
		return way
	}
	return "notify"
}

// setMuted says how a chat is taken from now on.
func (a *App) setMuted(chatID, way string) {
	a.pickPrefs(func(p *prefs) {
		if p.GroupMute == nil {
			p.GroupMute = map[string]string{}
		}
		if way == "notify" {
			delete(p.GroupMute, chatID)
			return
		}
		p.GroupMute[chatID] = way
	})
}

// prefsPath is where the settings are kept: beside the session, which is
// the one directory the app writes to.
func prefsPath() (string, error) {
	session, err := yh.ConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(session), "settings.json"), nil
}

// loadPrefs reads the settings, or the defaults when there are none: a file
// that does not parse is one written by a version with other settings, and
// what it does not say keeps its default.
func loadPrefs() prefs {
	path, err := prefsPath()
	if err != nil {
		return defaultPrefs()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return defaultPrefs()
	}
	p := defaultPrefs()
	if err := json.Unmarshal(raw, &p); err != nil {
		return defaultPrefs()
	}
	return p
}

// savePrefs writes the settings for the next run. A write that fails leaves
// the window as it is: a setting that is not kept is worth less than one
// that stops the window.
func (a *App) savePrefs() {
	path, err := prefsPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	raw, err := json.MarshalIndent(a.prefs, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0o600)
}

// pickPrefs changes a setting, keeps it, and lets the window follow it at
// once: a layout that only took effect the next time the window opened
// would look like it did nothing.
func (a *App) pickPrefs(change func(*prefs)) {
	mode, look := a.prefs.Mode, a.prefs.Look
	change(&a.prefs)
	a.savePrefs()
	if a.prefs.Mode != mode {
		a.applyMode()
	}
	// Every window is drawn in the appearance the settings say, so every
	// window is drawn again when they change.
	if a.prefs.Look != look {
		a.update(func() {})
	}
}

// dark reports whether the window is in the dark appearance: the one the
// settings picked, or the desktop's while they say to follow it.
func (a *App) dark(c *ui.Context) bool {
	switch a.prefs.Look {
	case lookDay:
		return false
	case lookNight:
		return true
	}
	return c.Theme().Dark
}

// filesDir is the folder the files that are sent and received are kept in:
// the one the settings picked, or the one the desktop gives the app.
func (a *App) filesDir() string {
	if a.prefs.FilesDir != "" {
		return a.prefs.FilesDir
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Downloads")
}

// openFilesDir shows that folder in the file manager.
func (a *App) openFilesDir() {
	dir := a.filesDir()
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	_ = mygo.Shell.OpenPath(dir)
}

// chooseFilesDir asks the desktop for the folder the files are kept in.
func (a *App) chooseFilesDir() {
	a.chooseDir(a.filesDir(), func(p *prefs, dir string) { p.FilesDir = dir })
}

// helpURL is where the window's help goes: the API it is written against,
// which is the only documentation this client has.
const helpURL = "https://yh-api.yyyyt.top/api/v1"

// openHelp shows that documentation in the browser.
func (a *App) openHelp() {
	_ = mygo.Shell.OpenExternal(helpURL)
}

// applyMode lays the window out the way the settings say, and resizes it:
// the classic panel is narrow because a chat is not in it, and the
// efficient one holds the chat beside the list.
func (a *App) applyMode() {
	if a.win == nil || a.win.IsDestroyed() {
		return
	}
	if a.prefs.Mode == modeClassic {
		a.win.SetSize(classicW, classicH)
		a.win.Invalidate()
		return
	}
	// The chat goes back into the window it is beside, and the one it was
	// in is closed: two windows showing the same chat would be one too
	// many.
	a.closeChatWindow()
	// A panel that was away when the layout changed would stay away, off
	// the edge of the screen: it is put back where it was before it went.
	a.undock()
	a.win.SetSize(efficientW, efficientH)
	a.win.Invalidate()
}

// undock brings the panel back from the edge of the screen, if it is there,
// without sliding it: it is a layout change, not the pointer coming back.
func (a *App) undock() {
	d := &a.dock
	if !d.hidden && !d.sliding {
		return
	}
	if a.win != nil && !a.win.IsDestroyed() && d.side != edgeNone {
		b := a.win.Bounds()
		x, y := a.dockedAt(d.side, b, a.workArea(b), false)
		a.win.SetPosition(int(x+0.5), int(y+0.5))
	}
	d.hidden, d.sliding = false, false
	d.side = edgeNone
	d.away = time.Time{}
}
