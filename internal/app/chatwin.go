package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// A chat opened from the classic panel is a window of its own, as the
// desktop client opens one: the panel is a list, and a chat is read beside
// it rather than in it. What the window draws is the chat pane of the
// efficient layout and the panel of what is known about the chat — there is
// one chat screen in this app, and this is the same one in another window.

const (
	chatWinW = 720
	chatWinH = 640
)

// chatWin is the window a chat is open in, if one is.
type chatWin struct {
	win *mygo.Window
}

// openChatWindow shows the open chat in a window of its own, or brings the
// window that already shows a chat forward and shows this one in it: one
// window, as the panel is one panel.
func (a *App) openChatWindow() {
	if a.current == "" {
		return
	}
	if w := a.chatWin; w != nil && w.win != nil && !w.win.IsDestroyed() {
		w.win.SetTitle(a.chatWindowTitle())
		w.win.Invalidate()
		w.win.Focus()
		return
	}
	w := &chatWin{}
	a.chatWin = w
	if a.win == nil {
		// No window to open one beside, as in the tests.
		return
	}
	// The window is made after the frame that asked for it: one made while
	// a view is being built is made in the middle of a paint.
	a.update(func() { a.openChatWinWindow(w) })
}

// openChatWinWindow makes the window, on the main thread and outside a
// frame.
func (a *App) openChatWinWindow(w *chatWin) {
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:     a.chatWindowTitle(),
		Width:     chatWinW,
		Height:    chatWinH,
		MinWidth:  480,
		MinHeight: 400,
		Content:   ui.View(func(c *ui.Context) { a.chatWinFrame(c) }),
	})
	win.OnClosed(func() {
		a.update(func() {
			if a.chatWin == w {
				a.chatWin = nil
			}
		})
	})
	w.win = win
}

// chatTitle is what the window a chat is open in is called.
func (a *App) chatWindowTitle() string {
	if conv := a.currentConv(); conv != nil {
		return conv.Title()
	}
	return i18n.T("app.name")
}

// chatWinFrame draws the window a chat is open in: the chat pane, and what
// is known about the chat beside it when it is asked for.
func (a *App) chatWinFrame(c *ui.Context) {
	qq = skin(a.dark(c))
	c.SetTheme(theme(qq.dark))
	i18n.Set(c.Locale())

	ui.Overlay(c, func() { a.userCard(c) })
	ui.Row(c).Fill().Background(qq.chatBg).Children(func() {
		a.chatPane(c)
		if a.infoOpen {
			a.infoPanel(c)
		}
	})
}

// syncChatWindow keeps the window a chat is open in up to date: its title
// follows the chat, and it goes when the layout that opened it does.
func (a *App) syncChatWindow() {
	w := a.chatWin
	if w == nil || w.win == nil || w.win.IsDestroyed() {
		return
	}
	w.win.SetTitle(a.chatWindowTitle())
	w.win.Invalidate()
}

// closeChatWindow closes the window a chat is open in, if there is one.
func (a *App) closeChatWindow() {
	w := a.chatWin
	if w == nil {
		return
	}
	a.chatWin = nil
	if w.win != nil && !w.win.IsDestroyed() {
		w.win.Close()
	}
}
