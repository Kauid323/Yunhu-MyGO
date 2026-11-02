package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// infoPanel is what is known about the open conversation. It is shown when
// it is asked for, and stays as it is while other conversations are
// opened: it is a panel of the chat screen, not of one chat.
func (a *App) infoPanel(c *ui.Context) {
	conv := a.currentConv()
	ui.Column(c).Width(200).FillHeight().Background(qq.panelBg).
		BorderWidth(0, 0, 0, 1).BorderColor(qq.border).Children(func() {
		ui.Row(c).Height(52).Padding(0, 12).Gap(6).AlignItems(ui.Center).
			BorderWidth(0, 0, 1, 0).BorderColor(qq.border).Children(func() {
			ui.Text(c, i18n.T("chat.info")).FontSize(12.5).FontWeight(600).
				TextColor(qq.text).Grow(1)
			close := ui.Box(c).Size(24, 24).Radius(5).Center().
				Cursor(ui.CursorPointer)
			if close.Hovered() {
				close.Background(qq.inputBg)
			}
			close.Children(func() {
				ui.Text(c, "×").FontSize(15).TextColor(qq.textMuted)
			})
			if close.Clicked() {
				a.infoOpen = false
			}
		})

		ui.Column(c).Padding(16, 14).Gap(6).AlignItems(ui.Center).Children(func() {
			if conv == nil {
				ui.Text(c, i18n.T("conv.empty")).FontSize(11.5).TextColor(qq.textFaint)
				return
			}
			a.avatar(c, conv.AvatarURL, conv.Title(), 56)
			ui.Text(c, conv.Title()).FontSize(13).FontWeight(600).TextColor(qq.text).
				SingleLine().Margin(2, 0, 0, 0)
			ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Padding(1, 6).Radius(4).Background(qq.inputBg).Children(func() {
					ui.Text(c, kindName(conv.ChatType)).FontSize(9).TextColor(qq.textMuted)
				})
				if conv.VIP {
					ui.Box(c).Padding(1, 6).Radius(4).Background(qq.inputBg).Children(func() {
						ui.Text(c, "VIP").FontSize(9).TextColor(qq.tagOwner)
					})
				}
			})
		})

		ui.Box(c).Height(1).Background(qq.border)

		if conv != nil {
			ui.Column(c).Padding(12, 14).Gap(9).Children(func() {
				a.infoRow(c, i18n.T("chat.id"), conv.ChatID)
				a.infoRow(c, i18n.T("chat.unread"), i18n.Num(conv.Unread))
				if conv.DoNotDisturb {
					a.infoRow(c, i18n.T("chat.notice"), i18n.T("chat.muted"))
				}
				a.infoRow(c, i18n.T("chat.connection"), connState(a.live))
			})
		}
	})
}

// infoRow is one labeled fact about the conversation.
func (a *App) infoRow(c *ui.Context, label, value string) {
	ui.Column(c).Gap(2).Children(func() {
		ui.Text(c, label).FontSize(10).TextColor(qq.textFaint)
		ui.Text(c, value).FontSize(11.5).TextColor(qq.textMuted).Selectable().Wrap()
	})
}
