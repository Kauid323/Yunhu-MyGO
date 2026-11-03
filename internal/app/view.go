package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// view builds the window from the app's state: the sign-in screen until
// there is an account, the chat screen after.
//
// It also takes the appearance and the language of the frame from the
// desktop. Both are handed over afresh on every frame, so switching the
// system's dark mode or its language switches the window's too, without a
// setting of its own.
func (a *App) view(c *ui.Context) {
	// The pictures this frame asks for are the ones it draws, and the cache
	// keeps them while it does: what is on screen is not dropped to make
	// room for what is on screen.
	a.images.beginFrame()
	// The sign-in panel is drawn in the colours of the client it copies,
	// in either appearance: it is looked at before there is an account to
	// have an appearance with.
	if a.winLogin || a.me == nil {
		qq = skin(true)
		c.SetTheme(theme(true))
	} else {
		qq = skin(a.dark(c))
		c.SetTheme(theme(qq.dark))
	}
	i18n.Set(c.Locale())

	// The layer the popups are drawn in is made here, every frame, rather
	// than by the first popup to need it: a panel can only animate in
	// when it is born into a parent that was there the frame before, and
	// a layer made along with it is not. The card of a user is drawn in
	// it, over whatever window is showing the history.
	ui.Overlay(c, func() {
		a.userCard(c)
		a.convMenuView(c)
		a.msgMenuView(c)
		a.askView(c)
		a.groupCardPanel(c)
	})

	ui.Column(c).Fill().Background(qq.chatBg).Children(func() {
		if a.winLogin || a.me == nil {
			a.loginView(c)
			return
		}
		a.shell(c)
	})
}

// shell is the chat screen, laid out the way the settings say: the narrow
// classic panel, or the rail, the conversation list, the open chat and, on
// the right, what is known about the conversation.
func (a *App) shell(c *ui.Context) {
	if a.prefs.Mode == modeClassic {
		a.classicShell(c)
		return
	}
	ui.Row(c).Fill().Children(func() {
		a.rail(c, true)
		a.convPanel(c, convPanelW)
		// A chat that was opened in a window of its own is not shown here
		// as well: two views of one chat would be one too many, and the
		// window is the one that was asked for.
		if a.chatWin != nil {
			ui.Column(c).Grow(1).FillHeight().Center().Background(qq.chatBg).
				Children(func() {
					ui.Text(c, i18n.T("chat.inWindow")).FontSize(12).
						TextColor(qq.textFaint)
				})
			return
		}
		a.chatPane(c)
		if a.infoOpen {
			a.infoPanel(c)
		}
	})
}
