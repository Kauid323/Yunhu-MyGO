package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The notices page: what a message that arrives is announced by — a
// notification on the desktop, a face or a box in the tray, a sound — and
// which chats are worth announcing at all.
func (a *App) noticesPage(c *ui.Context, s *settings) {
	gap(c, 30)
	sectionTitle(c, "settings.notices.popup")
	settingCard(c, 16, 0, func() {
		a.selectRow(c, false, "settings.notices.kind", "", &s.openNotice,
			&a.prefs.NoticeKind, noticeKinds)
		a.switchRow(c, true, "settings.notices.detail", "settings.notices.detailNote",
			&a.prefs.NoticeDetail)
		a.switchRow(c, true, "settings.notices.friend", "", &a.prefs.NotifyFriend)
		a.switchRow(c, true, "settings.notices.friendReq", "settings.notices.reqNote",
			&a.prefs.NotifyFriendQ)
		a.switchRow(c, true, "settings.notices.group", "", &a.prefs.NotifyGroup)
		a.switchRow(c, true, "settings.notices.groupReq", "settings.notices.groupReqNote",
			&a.prefs.NotifyGroupQ)
	})

	gap(c, 24)
	sectionTitle(c, "settings.notices.tray")
	settingCard(c, 10, 10, func() {
		ui.Row(c).Gap(12).Children(func() {
			choiceBox(c, shotWideW, a.prefs.TrayKind == "face", "settings.notices.trayFace",
				func() { a.pickPrefs(func(p *prefs) { p.TrayKind = "face" }) },
				func(w, h float32) { trayShot(c, w, h, false) })
			choiceBox(c, shotWideW, a.prefs.TrayKind == "box", "settings.notices.trayBox",
				func() { a.pickPrefs(func(p *prefs) { p.TrayKind = "box" }) },
				func(w, h float32) { trayShot(c, w, h, true) })
		})
	})
	gap(c, 12)
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.notices.friend", "", &a.prefs.TrayFriend)
		a.switchRow(c, true, "settings.notices.group", "", &a.prefs.TrayGroup)
		a.switchRow(c, true, "settings.notices.friendReq", "", &a.prefs.TrayFriendQ)
		a.switchRow(c, true, "settings.notices.groupReq", "", &a.prefs.TrayGroupQ)
	})

	gap(c, 24)
	sectionTitle(c, "settings.notices.shake")
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.notices.shakeWindow", "", &a.prefs.Shake)
	})

	gap(c, 24)
	sectionTitle(c, "settings.notices.sound")
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.notices.soundNew", "", &a.prefs.SoundNew)
		a.switchRow(c, true, "settings.notices.friend", "settings.notices.soundMsg",
			&a.prefs.SoundFriend)
		a.switchRow(c, true, "settings.notices.group", "settings.notices.soundMsg",
			&a.prefs.SoundGroup)
		a.switchRow(c, true, "settings.notices.friendReq", "settings.notices.soundSystem",
			&a.prefs.SoundFriendQ)
		a.switchRow(c, true, "settings.notices.groupReq", "settings.notices.soundSystem",
			&a.prefs.SoundGroupQ)
	})

	gap(c, 24)
	sectionTitle(c, "settings.notices.groups")
	a.groupCard(c, s)
}

// groupCard is how each chat is taken: it arrives with a word, without one, or
// not at all. The list is the conversations that are loaded, which is what the
// window has to go on.
func (a *App) groupCard(c *ui.Context, s *settings) {
	settingCard(c, 16, 0, func() {
		ui.Row(c).Height(44).Gap(8).AlignItems(ui.Center).Children(func() {
			box := ui.Row(c).Width(220).Gap(6).Padding(0, 10).Height(26).Radius(13).
				Background(qq.setFieldHover).AlignItems(ui.Center)
			box.Children(func() {
				ui.Icon(c, iconSearch).FontSize(12).TextColor(qq.setTextMuted)
				ui.TextInput(c, &a.filter).Placeholder(i18n.T("conv.search")).Grow(1).
					FontSize(12).Background(qq.setFieldHover).
					Border(0, qq.setFieldHover)
			})
			ui.Spacer(c)
		})
		convs := a.filteredConvs()
		if len(convs) == 0 {
			ui.Text(c, i18n.T("conv.empty")).FontSize(12).
				TextColor(qq.setTextMuted).Padding(8)
			return
		}
		ui.Scroll(c).MaxHeight(320).Children(func() {
			ui.Column(c).Children(func() {
				for i, conv := range convs {
					a.muteRow(c, s, conv, i > 0)
				}
			})
		})
	})
}

// muteRow is one chat of that list: its face, its name, and the way it
// arrives.
func (a *App) muteRow(c *ui.Context, s *settings, conv yh.Conversation, divider bool) {
	if divider {
		ui.Box(c).Height(1).Background(qq.setLine)
	}
	ui.Row(c).Height(48).Gap(10).AlignItems(ui.Center).Children(func() {
		a.avatar(c, conv.AvatarURL, conv.Title(), 28)
		ui.Text(c, conv.Title()).FontSize(13).TextColor(qq.setText).
			Grow(1).MinWidth(0).SingleLine()
		open := s.openGroup == conv.ChatID
		way := a.muted(conv.ChatID)
		box := ui.Row(c).Width(180).Height(24).Radius(6).Padding(0, 8).Gap(6).
			AlignItems(ui.Center).Cursor(ui.CursorPointer).Border(1, qq.setField)
		if open {
			box.Background(qq.setFieldHover).Border(1, qq.setFieldOpen)
		}
		box.Children(func() {
			ui.Text(c, i18n.T("settings.value."+way)).FontSize(12.5).
				TextColor(qq.setText).Grow(1).MinWidth(0).SingleLine()
			ui.Icon(c, iconChevron).FontSize(11).TextColor(qq.setTextMuted)
		})
		if box.Clicked() {
			if open {
				s.openGroup = ""
			} else {
				s.openGroup = conv.ChatID
			}
		}
		if !open {
			return
		}
		ui.PopoverBase(c, box, &open, func(panel ui.Element) {
			panel.AttachTo(box, ui.AnchorTopLeft, ui.AnchorTopLeft).
				Width(180).Padding(2).Radius(8).
				Background(qq.setPopup).Border(1, qq.setFieldOpen).
				Shadow(0, 8, 20, 0, ui.RGBA(0, 0, 0, 0.35)).
				Transition(selectArrives)
			for _, opt := range muteWays {
				item := ui.Row(c).Height(selectRowH).Padding(0, 10).Gap(6).Radius(6).
					AlignItems(ui.Center).Cursor(ui.CursorPointer)
				if way == opt {
					item.Background(qq.setPopupHover)
				} else if item.Hovered() {
					item.Background(qq.setPopupHover)
				}
				item.Children(func() {
					ui.Text(c, i18n.T("settings.value."+opt)).FontSize(12.5).
						TextColor(qq.setText).Grow(1).MinWidth(0).SingleLine()
					if way == opt {
						ui.Icon(c, iconCheck).FontSize(12).TextColor(qq.setText)
					}
				})
				if item.Clicked() {
					s.openGroup = ""
					a.setMuted(conv.ChatID, opt)
				}
			}
		})
	})
}

// trayShot is the picture of a tray notice: the face that flashes, or the box
// of messages, as small as a thumbnail holds them.
func trayShot(c *ui.Context, w, h float32, box bool) {
	ui.Box(c).Size(w, h).Radius(3).Background(dayShot.ground).Children(func() {
		card := ui.Box(c).Absolute().Left(3).Top(3).Size(w-6, h-6).Radius(3).
			Background(dayShot.window)
		card.Children(func() {
			if !box {
				// A face with a badge, and two lines beside it.
				ui.Box(c).Absolute().Left(5).Top(9).Size(14, 14).Radius(7).
					Background(dayShot.mark)
				ui.Box(c).Absolute().Left(24).Top(9).Size(30, 2).Radius(1).
					Background(dayShot.mark)
				ui.Box(c).Absolute().Left(24).Top(14).Size(20, 2).Radius(1).
					Background(dayShot.mark)
				ui.Box(c).Absolute().Right(5).Top(8).Size(5, 5).Radius(3).
					Background(qq.badge)
				return
			}
			// The box: a bar across its top, then a row of faces.
			ui.Box(c).Absolute().Left(0).Top(0).Size(w-6, 8).
				Background(dayShot.bar)
			for i := 0; i < 3; i++ {
				ui.Box(c).Absolute().Left(float32(5+i*12)).Top(14).Size(8, 8).
					Radius(4).Background(dayShot.mark)
			}
			ui.Box(c).Absolute().Left(5).Top(26).Size(w-16, 2).Radius(1).
				Background(dayShot.mark)
		})
	})
}
