package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// A face in the history is a way into what is known about the person behind it:
// a press opens a card of them beside it, which is asked for when it is first
// opened and kept until another face is pressed.
//
// The card is drawn in the layer over the window rather than as a popup of the
// row it belongs to: a row of the history is built as the list lays it out, and
// a popup made there comes out in the wrong skin and once per row that names
// the same person.
const userCardW = 300

// openUserCard shows the card of a user beside the face at, or puts away the
// one that is open: the card is of one person at a time.
func (a *App) openUserCard(id, name, avatar string, at ui.Rect) {
	if a.cardOpen && a.cardID == id {
		a.cardOpen = false
		return
	}
	a.cardOpen, a.cardID = true, id
	a.cardName, a.cardAvatar = name, avatar
	a.cardAt = at
	if a.card == nil || a.card.ID != id {
		a.card = nil
		a.loadCard(id)
	}
}

// loadCard asks what is known about a user, once: the card shows what the
// message said of them while it comes.
func (a *App) loadCard(id string) {
	go func() {
		p, err := a.client.UserProfile(id)
		if err != nil {
			return
		}
		a.update(func() {
			if a.cardID == id {
				a.card = p
			}
		})
	}()
}

// userCard draws the card in the layer over the window, at the face it belongs
// to, with a backdrop that closes it on a press anywhere else.
func (a *App) userCard(c *ui.Context) {
	if !a.cardOpen {
		return
	}
	back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0)
	if back.Clicked() {
		a.cardOpen = false
		return
	}
	// Beside the face, and kept in the window: a card that opened past the
	// right edge would be a card half read.
	x := a.cardAt.X + 42
	if a.win != nil {
		if b := a.win.Bounds(); x+userCardW > float32(b.Width)-8 {
			x = float32(b.Width) - userCardW - 8
		}
	}
	panel := ui.Box(c).Absolute().Left(max(x, 8)).Top(max(a.cardAt.Y-8, 8)).
		Width(userCardW).Padding(16).Radius(10).Background(qq.menuBg).
		Border(1, qq.menuLine).Shadow(0, 10, 26, 0, ui.RGBA(0, 0, 0, 0.40))
	panel.Children(func() { a.cardBody(c, a.cardID) })
}

// cardBody is what the card holds: who they are, what the account noted about
// them, what they say of themselves, and the way into a chat with them.
func (a *App) cardBody(c *ui.Context, id string) {
	p := a.card
	name, avatar := a.cardName, a.cardAvatar
	if p != nil {
		name, avatar = firstNonEmpty(p.Name, name), firstNonEmpty(p.AvatarURL, avatar)
	}
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		a.avatar(c, avatar, name, 64)
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, name).FontSize(15).FontWeight(600).
					TextColor(qq.setText).SingleLine()
				if p != nil && p.VIP {
					ui.Box(c).Padding(0, 5).Radius(3).Background(qq.inputBg).
						Children(func() {
							ui.Text(c, "VIP").FontSize(8.5).TextColor(qq.tagOwner)
						})
				}
			})
			ui.Text(c, i18n.T("card.id", id)).FontSize(11).
				TextColor(qq.setTextMuted).SingleLine()
		})
	})

	if p != nil {
		ui.Box(c).Height(1).Background(qq.setLine).Margin(12, 0, 10, 0)
		ui.Column(c).Gap(9).Children(func() {
			if g := genderLine(p.Gender); g != "" {
				ui.Text(c, g).FontSize(12).TextColor(qq.setTextMuted)
			}
			a.cardRow(c, "card.remark", firstNonEmpty(p.Remark, i18n.T("card.none")))
			a.cardRow(c, "card.sign", firstNonEmpty(p.Introduction, i18n.T("card.none")))
			if place := firstNonEmpty(p.City, p.District); place != "" {
				a.cardRow(c, "card.place", place)
			}
		})
	} else {
		ui.Row(c).Height(56).Center().Children(func() { ui.Spinner(c) })
	}

	ui.Row(c).Gap(10).AlignItems(ui.Center).Margin(14, 0, 0, 0).Children(func() {
		ui.Spacer(c)
		// The way into a chat with them: a chat that is already in the list
		// is opened, and one that is not is a chat this client cannot start
		// — the API has no route for it, so the button says so.
		conv, ok := a.chatWith(id)
		send := ui.PrimaryButton(c, i18n.T("card.message")).Disabled(!ok).
			Background(qq.accent).Radius(6).Padding(5, 18).FontSize(12.5)
		if send.Clicked() && ok {
			a.cardOpen = false
			a.open(conv)
		}
	})
}

// cardRow is one labeled line of the card, with the value cut short rather than
// pushing the card wider.
func (a *App) cardRow(c *ui.Context, key, value string) {
	ui.Row(c).Gap(10).AlignItems(ui.Start).Children(func() {
		ui.Text(c, i18n.T(key)).FontSize(11.5).TextColor(qq.setTextMuted).
			Width(52).SingleLine()
		ui.Text(c, value).FontSize(12).TextColor(qq.setText).
			Grow(1).MinWidth(0).SingleLine()
	})
}

// chatWith is the conversation with a user, if the list has one.
func (a *App) chatWith(id string) (yh.Conversation, bool) {
	for _, conv := range a.convs {
		if conv.ChatID == id {
			return conv, true
		}
	}
	return yh.Conversation{}, false
}

// genderLine is what the card says of a user's gender, which is empty when they
// have not said.
func genderLine(g int) string {
	switch g {
	case 1:
		return "♂ " + i18n.T("card.man")
	case 2:
		return "♀ " + i18n.T("card.woman")
	case 3:
		return i18n.T("card.other")
	}
	return ""
}
