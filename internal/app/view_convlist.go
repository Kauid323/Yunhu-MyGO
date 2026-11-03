package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// convPanelW is how wide the conversation list is beside the chat. The
// narrow panel gives it no width of its own: it takes what is left.
const convPanelW = 250

// convPanel is the conversation list: a search box, then the chats. A width
// of zero makes it take what the window leaves, which is what the narrow
// panel does: there the list is beside the same rail and has no width of its
// own to keep.
func (a *App) convPanel(c *ui.Context, width float32) {
	panel := ui.Column(c).FillHeight().Background(qq.listBg).
		BorderWidth(0, 1, 0, 0).BorderColor(qq.border)
	if width > 0 {
		panel.Width(width)
	} else {
		panel.Grow(1).MinWidth(0)
	}
	panel.Children(func() {
		ui.Row(c).Padding(12, 10, 10, 10).Gap(8).AlignItems(ui.Center).Children(func() {
			search := ui.Row(c).Grow(1).Gap(5).Padding(0, 8).Height(28).Radius(7).
				Background(qq.searchBg).AlignItems(ui.Center)
			search.Children(func() {
				ui.Icon(c, iconSearch).FontSize(12).TextColor(qq.textFaint)
				ui.TextInput(c, &a.filter).Placeholder(i18n.T("conv.search")).Grow(1).FontSize(12).
					Background(qq.searchBg).Border(0, qq.searchBg)
			})
			plus := ui.Box(c).Size(26, 26).Radius(7).Center().
				Cursor(ui.CursorPointer).Background(qq.searchBg)
			if plus.Hovered() {
				plus.Background(qq.listSelected)
			}
			plus.Children(func() {
				ui.Icon(c, iconPlus).FontSize(14).TextColor(qq.textMuted)
			})
			if plus.Clicked() {
				a.convPlus.open = !a.convPlus.open
			}
			showMenu(c, plus, &a.convPlus)
		})

		convs := a.filteredConvs()
		a.convList.Key = func(i int) any { return convs[i].ChatID }
		ui.List(c, &a.convList, len(convs), func(i int) {
			a.convRow(c, convs[i])
		}).Grow(1).Padding(0, 8, 8, 8).Gap(1).Children(func() {
			if len(convs) == 0 {
				ui.Text(c, i18n.T("conv.empty")).FontSize(12).
					TextColor(qq.textFaint).Padding(16)
			}
		})

		// What the last load of the list had to say: how many came back,
		// or why none did.
		if !a.note.empty() {
			ui.Text(c, a.note.String()).FontSize(9.5).TextColor(qq.textFaint).
				Padding(0, 12, 8, 12).SingleLine()
		}
	})
}

// convRow is one conversation in the list.
func (a *App) convRow(c *ui.Context, conv yh.Conversation) {
	selected := conv.ChatID == a.current
	row := ui.Row(c).Padding(8, 9).Gap(9).Radius(8).
		Cursor(ui.CursorPointer).AlignItems(ui.Center)
	switch {
	case selected:
		row.Background(qq.listSelected)
	case row.Hovered():
		row.Background(qq.listHover)
	}
	row.Children(func() {
		a.avatar(c, conv.AvatarURL, conv.Title(), 38)
		ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, conv.Title()).FontSize(12.5).FontWeight(500).TextColor(qq.text).
					Grow(1).MinWidth(0).SingleLine()
				ui.Text(c, clock(conv.Time())).FontSize(9.5).TextColor(qq.textFaint)
			})
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				preview := conv.Content
				if conv.At {
					preview = i18n.T("conv.mentioned", preview)
				}
				if preview == "" {
					preview = kindName(conv.ChatType)
				}
				ui.Text(c, preview).FontSize(11).TextColor(qq.textMuted).
					Grow(1).MinWidth(0).SingleLine()
				if conv.Unread > 0 {
					bg := qq.badge
					if conv.DoNotDisturb {
						bg = qq.badgeMuted
					}
					ui.Box(c).Padding(1, 5).Radius(8).Background(bg).Children(func() {
						ui.Text(c, shortCount(conv.Unread)).FontSize(9).Bold().
							TextColor(qq.accentText)
					})
				}
			})
		})
	})
	if row.Clicked() {
		a.open(conv)
	}
	// A chat opened in a window of its own, which the settings offer: the
	// classic layout has no room for one and does it on a single press.
	if row.DoubleClicked() && a.prefs.Mode == modeEfficient {
		a.open(conv)
		a.openChatWindow()
	}
	// A right press asks what can be done with the chat rather than
	// opening it, and the menu comes out of where it was pressed.
	if row.RightClicked() {
		a.openConvMenu(conv, row.Bounds())
	}
}
