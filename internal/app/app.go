// Package app is mygo湖's window: a native chat client for Yunhu.
//
// Two layers live here, told apart by the name of the file:
//
//   - The state and the actions that change it, in the files named after
//     what they do: app.go (the window and its state), session.go
//     (signing in), live.go (what the server pushes), convs.go and
//     chat.go (the conversation list and the open chat).
//   - The drawing, in the files named view_*.go. A drawing file reads the
//     state and never writes it, and never asks the network for anything:
//     what it needs is either here already, or it calls an action, which
//     goes and gets it and comes back to the drawing thread through
//     update.
//
// The network itself is not here at all: it is internal/yh, the Yunhu
// API's client, its protobuf codec and its websocket. Nothing in this
// package writes HTTP, and nothing in internal/yh draws.
package app

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

const (
	screenLogin = iota
	screenChat
)

// msgPage is how many messages a page asks for, and msgFirst how many the
// first page of a chat loads.
const (
	msgPage  = 30
	msgFirst = 50
)

// efficientW and efficientH are the window the chat is read in: the rail,
// the list and the chat beside each other. The classic layout has its own
// size (view_classic.go).
const (
	efficientW = 1000
	efficientH = 720
)

// The ways the sign-in screen signs in.
const (
	modePassword = iota
	modePhone
	modeToken
)

// appVersion is what this build is, as the about page writes it: the client
// this one is written for has a version of its own, which says nothing about
// this one.
const appVersion = "0.1.0"

// App is the window's state.
type App struct {
	win    *mygo.Window
	client *yh.Client
	conn   *yh.Conn
	cancel context.CancelFunc

	screen   int
	nav      int
	infoOpen bool
	// loginMode is which way the sign-in screen signs in: a password, a
	// phone number and the code texted to it, or a token.
	loginMode int
	email     string
	password  string
	token     string
	// phone is the number being signed in with, phoneCode is what the
	// picture it comes with says, and smsCode is what was texted to it.
	// captcha is that picture, which is read out before a code is texted.
	phone     string
	phoneCode string
	smsCode   string
	captchaID string
	captcha   *ui.Bitmap
	captchaGo bool
	// smsAt is when a code was last texted to the number: another is not
	// texted until smsWait has passed.
	smsAt time.Time
	busy  bool
	// hint is what went right, as errText is what went wrong: a code was
	// texted, and the screen says so.
	hint    line
	errText line

	me          *yh.User
	convs       []yh.Conversation
	filter      string
	current     string
	currentType int
	msgs        []*yh.Message
	msgList     ui.ListState
	convList    ui.ListState
	input       string
	// note is what went wrong with the last thing asked for, if
	// anything did: the list under the conversations says it.
	note        line
	live        yh.State
	loading     bool
	loadingMore bool
	hasMore     bool
	// loadedAt is the top row the last page was asked for at, and moreAt
	// when it was: the two of them keep a list that is at its top from
	// asking for a page on every frame, and keep it from never asking
	// again.
	loadedAt  int
	moreAt    time.Time
	jumpToEnd bool
	// unreadFrom is the first message of the open chat that came in while
	// the reader was away: the line across the history goes over it, and
	// it stays there until another chat is opened. It is len(msgs) when
	// nothing came in, and it is kept as the history grows, so that the
	// line stays over the message it was put over.
	unreadFrom int
	// readFrom is the first message the reader has not been past yet: what
	// the button over the list counts. It only moves forward — coming back
	// up through the history does not make messages unread again.
	readFrom int
	// unreadOnOpen is how many of a chat's messages were unread when it
	// was opened, which its page has not come back with yet.
	unreadOnOpen int
	// jumpTo is the row the list is scrolled to as it is built, or -1 for
	// none: a jump is asked for while the list is built, not after.
	jumpTo int
	// viewer is the window a picture was opened in, if one is open.
	viewer *viewer
	// settings is the settings window, if it is open, and what it shows.
	settings *settings
	// chatWin is the window a chat is open in when the classic layout has
	// no room for it, if it is open.
	chatWin *chatWin
	// prefs is what the settings window sets and the window does about it,
	// and dock is where the classic panel is on its way to and from the
	// edge of the screen.
	prefs prefs
	dock  dock
	// stickers are the account's saved stickers, which the panel over the
	// composer shows, and whether that panel is open: the list is asked for
	// once, when it is first shown.
	stickers    []yh.Expression
	stickerOpen bool
	stickerGo   bool
	// space is what the app takes up on this machine, as the storage page
	// reports it, and whether it is being measured.
	space   space
	spaceGo bool
	// card is the card of a user that a face in the history opened, of
	// whom the id, name and face are known before what is asked for about
	// them comes back.
	card       *yh.Profile
	cardOpen   bool
	cardID     string
	cardName   string
	cardAvatar string
	// cardAt is the face the card was opened beside, in the window's points.
	cardAt ui.Rect
	// notices says the desktop takes notifications, and pending is what
	// arrived for the chats that are not open and has not been told about
	// yet: timer is the wait for the end of the burst.
	notices bool
	pending map[string]*notice
	timer   *time.Timer
	// notify shows a notification: the desktop's own, unless a test takes
	// it instead to read what would be shown.
	notify func(*notice)
	images *imageCache
	more   menu
}

// New makes the app, before the window exists.
func New() *App {
	a := &App{
		client:   yh.New(),
		screen:   screenLogin,
		infoOpen: false,
		images:   newImageCache(),
		live:     yh.StateOffline,
		// No page has been asked for yet, so any top row is further back
		// than the last one was.
		loadedAt: 1 << 30,
		jumpTo:   -1,
		// The saved settings are read when the window is made, not here:
		// an app made by a test is made without touching the files of the
		// person running it.
		prefs: defaultPrefs(),
	}
	// A chat reads from its end: FollowEnd keeps it there, and scrolling
	// away only pauses that until the end comes back in view.
	a.msgList.FollowEnd = true
	a.more = moreMenu(a)
	return a
}

// Start creates the window and resumes the saved session.
func (a *App) Start() {
	a.prefs = loadPrefs()
	width, height, minW := efficientW, efficientH, 280
	if a.prefs.Mode == modeClassic {
		width, height = classicW, classicH
	}
	a.win = mygo.NewWindow(mygo.WindowOptions{
		Title:  i18n.T("app.name"),
		Width:  width,
		Height: height,
		// The classic panel is narrow and the efficient one is not: the
		// window has to be allowed to be either.
		MinWidth:  minW,
		MinHeight: 400,
		Content:   ui.View(a.view),
	})
	// A notification opens the chat it is about, on the window that is
	// already here.
	if mygo.NotificationsSupported() {
		a.notices = true
		mygo.App.OnNotificationClick(a.onNoticeClick)
	}
	a.resume()
}

// update runs fn on the main thread, then draws the window again.
func (a *App) update(fn func()) {
	if a.win == nil {
		fn()
		return
	}
	a.win.Update(fn)
	// The windows that show what the same state holds are drawn again when
	// something lands in it: a picture, a chat read on its own, the
	// settings.
	if v := a.viewer; v != nil && v.win != nil {
		v.win.Invalidate()
	}
	if w := a.chatWin; w != nil && w.win != nil {
		w.win.Invalidate()
	}
	if s := a.settings; s != nil && s.win != nil {
		s.win.Invalidate()
	}
}

// imageReferer is the Referer the image host wants: it answers 403 to a
// picture asked for from anywhere else, so pictures are asked for as the
// account's own requests are, from the chat API's host.
func (a *App) imageReferer() string {
	if a.client != nil && a.client.Base != "" {
		return a.client.Base
	}
	return yh.DefaultBase
}

// Capture writes what the window shows to path as a PNG. It draws the view
// itself rather than the screen, so it works over a remote session where a
// screen grab comes out black.
func (a *App) Capture(path string) error {
	if a.win == nil {
		return errors.New("the window is not open")
	}
	return a.capture(a.win, path)
}

// CaptureSettings writes what the settings window shows to path, as Capture
// does for the window of the chat: it is the only way to look at that
// window over a remote session.
func (a *App) CaptureSettings(path string) error {
	if a.settings == nil || a.settings.win == nil {
		return errors.New("the settings window is not open")
	}
	// The window draws a frame of its own first: a capture of a window
	// that has not drawn since the other one did comes out with the other
	// one's glyphs.
	w := a.settings.win
	w.Invalidate()
	time.Sleep(150 * time.Millisecond)
	return a.capture(w, path)
}

// capture draws a window to a file.
func (a *App) capture(w *mygo.Window, path string) error {
	data, err := w.CapturePage()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
