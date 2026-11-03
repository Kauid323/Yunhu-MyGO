package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// chatPane is the open conversation: its header, its messages and the
// composer.
func (a *App) chatPane(c *ui.Context) {
	ui.Column(c).Grow(1).FillHeight().Background(qq.chatBg).Children(func() {
		a.chatHeader(c)
		a.messageList(c)
		a.composer(c)
	})
}

// chatHeader names the open conversation, with its actions on the right.
func (a *App) chatHeader(c *ui.Context) {
	conv := a.currentConv()
	ui.Row(c).Height(52).Padding(0, 16).Gap(2).AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).BorderColor(qq.border).Children(func() {
		// The title is a way into the card of a group: a press on it asks
		// the group what it says of itself, and shows it beside the name.
		// It is the whole of the header: what kind of chat it is and what
		// it is called in the API are not what a header is for.
		title := ui.Column(c).Grow(1).MinWidth(0).Gap(1).Cursor(ui.CursorPointer)
		title.Children(func() {
			name, sub := i18n.T("app.name"), connState(a.live)
			if conv != nil {
				name = conv.Title()
				if conv.ChatType == yh.ChatGroup {
					name = a.groupTitle(conv.ChatID, name)
				}
				sub = ""
			}
			ui.Text(c, name).FontSize(16.5).FontWeight(600).TextColor(qq.text).
				SingleLine()
			if sub != "" {
				ui.Text(c, sub).FontSize(10).TextColor(qq.textFaint).SingleLine()
			}
		})
		if title.Clicked() && conv != nil && conv.ChatType == yh.ChatGroup {
			a.openGroupCard(title.Bounds())
		}
		if conv != nil {
			for _, ic := range []*ui.SVG{iconPhone, iconVideo, iconShare, iconGrid} {
				a.headerIcon(c, ic)
			}
			// The folder of files and the help are the two of the header's
			// actions this client has a feature behind: the rest are drawn
			// and do nothing, which is what their doing nothing means.
			a.headerIconClickable(c, iconFolderOpen, false, a.openFilesDir)
			a.headerIconClickable(c, iconHelp, false, a.openHelp)
		}
		a.headerIconClickable(c, iconInfo, a.infoOpen, func() {
			a.infoOpen = !a.infoOpen
		})
		a.headerIconClickable(c, iconMore, false, func() {
			a.reload()
		})
	})
}

// headerIcon is one of the header's actions, dimmed until hovered.
func (a *App) headerIcon(c *ui.Context, ic *ui.SVG) {
	box := ui.Box(c).Size(28, 28).Radius(6).Center().Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.inputBg)
	}
	box.Children(func() {
		ui.Icon(c, ic).FontSize(16).TextColor(qq.textMuted)
	})
}

// headerIconClickable is a header action that does something.
func (a *App) headerIconClickable(c *ui.Context, ic *ui.SVG, lit bool, onClick func()) {
	box := ui.Box(c).Size(28, 28).Radius(6).Center().Cursor(ui.CursorPointer)
	switch {
	case lit:
		box.Background(qq.inputBg)
	case box.Hovered():
		box.Background(qq.inputBg)
	}
	box.Children(func() {
		col := qq.textMuted
		if lit {
			col = qq.accent
		}
		ui.Icon(c, ic).FontSize(16).TextColor(col)
	})
	if box.Clicked() {
		onClick()
	}
}
