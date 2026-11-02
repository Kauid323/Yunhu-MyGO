package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The classic layout is the narrow panel the desktop client has always had:
// the account across its top, and under it the same strip of sections and
// the same conversation list the wide layout has beside the chat. A chat is
// not in it — it opens in a window of its own (chatwin.go) — which is what
// lets the panel be narrow enough to slide to the edge of the screen when it
// is not being used (dock.go).
const (
	classicW = 320
	classicH = 640
)

// classicShell draws the panel: the account across its top, then the rail and
// the chats, which are the ones the wide layout uses — there is one rail and
// one list in this app, wherever they are shown.
func (a *App) classicShell(c *ui.Context) {
	ui.Column(c).Fill().Background(qq.chatBg).Children(func() {
		a.classicAccount(c)
		ui.Row(c).Grow(1).MinHeight(0).Children(func() {
			a.rail(c, false)
			a.convPanel(c, 0)
		})
	})
	// The panel keeps out of the way of what is being done elsewhere.
	a.dockFrame(c)
}

// classicAccount is who is signed in, across the top of the panel: the face,
// the name, and what the account says of itself.
func (a *App) classicAccount(c *ui.Context) {
	ui.Row(c).Padding(16, 14, 12, 14).Gap(14).AlignItems(ui.Center).Children(func() {
		a.avatar(c, a.me.AvatarURL, a.me.Name, 64)
		ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
			ui.Text(c, a.me.Name).FontSize(15).FontWeight(600).
				TextColor(qq.text).SingleLine()
			intro := a.me.Introduction
			if intro == "" {
				intro = i18n.T("classic.noIntro")
			}
			ui.Text(c, intro).FontSize(11).TextColor(qq.textMuted).SingleLine()
		})
	})
}
