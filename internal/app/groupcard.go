package app

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The card of a group is the same component as the card of a person: a panel
// in the layer over the window, opened beside what was pressed, with a
// backdrop that closes it. What it holds is what the group says of itself,
// which is asked for when it is first opened.
const groupCardW = 320

// openGroupCard shows the card of the group the chat is with, beside at, or
// puts away the one that is open.
func (a *App) openGroupCard(at ui.Rect) {
	if a.current == "" || a.currentType != yh.ChatGroup {
		return
	}
	if a.groupOpen && a.groupID == a.current {
		a.groupOpen = false
		return
	}
	a.groupOpen, a.groupID, a.groupAt = true, a.current, at
	if a.group == nil || a.group.ID != a.current {
		a.group = nil
		a.loadGroup(a.current)
	}
}

// loadGroup asks what a group says of itself, once: the card shows what the
// conversation already knows while it comes.
func (a *App) loadGroup(id string) {
	go func() {
		g, err := a.client.GroupInfo(id)
		if err != nil {
			return
		}
		a.update(func() {
			if a.current == id {
				a.group = g
			}
		})
	}()
}

// groupCard draws the card in the layer over the window, at the title it
// belongs to.
func (a *App) groupCardPanel(c *ui.Context) {
	if !a.groupOpen {
		return
	}
	back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0)
	if back.Clicked() {
		a.groupOpen = false
		return
	}
	// Under the title, and kept in the window: a card that opened past an
	// edge would be a card half read.
	x := a.groupAt.X
	y := a.groupAt.Y + 8
	if a.win != nil {
		b := a.win.Bounds()
		if x+groupCardW > float32(b.Width)-8 {
			x = float32(b.Width) - groupCardW - 8
		}
		if y+420 > float32(b.Height)-8 {
			y = float32(b.Height) - 428
		}
	}
	panel := ui.Box(c).Absolute().Left(max(x, 8)).Top(max(y, 8)).
		Width(groupCardW).Padding(16).Radius(10).Background(qq.menuBg).
		Border(1, qq.menuLine).Shadow(0, 10, 26, 0, ui.RGBA(0, 0, 0, 0.40)).
		TransformOrigin(0, 0).Transition(cardArrives)
	panel.Children(func() { a.groupCardBody(c) })
}

// groupCardBody is what the card holds: the group's face and name, how many
// are in it, what it is filed under, and the lines it says of itself.
func (a *App) groupCardBody(c *ui.Context) {
	g := a.group
	if g != nil && g.ID != a.current {
		g = nil
	}
	conv := a.currentConv()

	name, avatar, id := "", "", a.current
	if conv != nil {
		name, avatar = conv.Title(), conv.AvatarURL
	}
	headcount, intro, category := 0, "", ""
	if g != nil {
		if g.Name != "" {
			name = g.Name
		}
		if g.AvatarURL != "" {
			avatar = g.AvatarURL
		}
		id = g.ID
		headcount, intro, category = g.Headcount, g.Intro, g.Category
	}

	ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
		// The face at the size it was uploaded, not a thumbnail of it:
		// the card is opened on purpose, and pressed to look at it.
		face := ui.Box(c).Size(64, 64).Cursor(ui.CursorPointer)
		face.Children(func() {
			var bmp *ui.Bitmap
			if avatar != "" {
				bmp = a.images.getBox(avatar, 0, 0, a)
			}
			ui.Avatar(c, name, bmp).Size(64, 64)
		})
		if face.Clicked() {
			a.showPicture(avatar, 0, 0)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
			ui.Text(c, name).FontSize(16).FontWeight(600).TextColor(qq.text).
				Wrap()
			ui.Text(c, membersLine(id, headcount)).FontSize(11).
				TextColor(qq.textMuted).SingleLine()
			if category != "" {
				chip := ui.Box(c).Padding(2, 7).Radius(4).Background(qq.inputBg)
				chip.Children(func() {
					ui.Text(c, category).FontSize(10.5).TextColor(qq.textMuted).
						SingleLine()
				})
			}
		})
	})

	ui.Box(c).Height(10)
	remark := ""
	if conv != nil {
		remark = conv.Remark
	}
	a.groupCardRow(c, "group.remark", remark, "group.setRemark", false)
	a.groupCardRow(c, "group.intro", intro, "", true)
	a.groupCardRow(c, "group.category", category, "", false)

	ui.Box(c).Height(12)
	ui.Row(c).Gap(10).Children(func() {
		if ui.Button(c, i18n.T("group.share")).Grow(1).Radius(6).
			Padding(6, 0).FontSize(12.5).Clicked() {
			copyText(id)
		}
		if ui.Button(c, i18n.T("group.message")).Grow(1).Radius(6).
			Background(qq.accent).Padding(6, 0).FontSize(12.5).Clicked() {
			a.groupOpen = false
		}
	})
}

// groupCardRow is one line of the card: what it is, and what the group says
// of it. A line with nothing to say shows what it would take to fill it.
func (a *App) groupCardRow(c *ui.Context, label, value, empty string, chevron bool) {
	if value == "" && empty == "" {
		return
	}
	ui.Row(c).Height(28).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, i18n.T(label)).FontSize(11.5).TextColor(qq.textMuted).
			Width(58)
		shown := value
		col := qq.text
		if shown == "" {
			shown, col = i18n.T(empty), qq.textFaint
		}
		ui.Text(c, shown).FontSize(12).TextColor(col).Grow(1).MinWidth(0).
			SingleLine()
		if chevron {
			ui.Icon(c, iconChevronRight).FontSize(11).TextColor(qq.textFaint)
		}
	})
}

// membersLine is who a group is, as the card writes it: its id, and how many
// are in it when that is known.
func membersLine(id string, headcount int) string {
	if headcount <= 0 {
		return id
	}
	return i18n.T("group.members", id, i18n.Num(headcount))
}

// groupTitle is what the header of a group chat says: its name, and how many
// are in it.
// groupTitle is what the header of a group chat says: its name, and how
// many are in it. The brackets are the locale's, not ASCII: a Chinese
// header in half-width brackets reads as something pasted in.
func (a *App) groupTitle(id, name string) string {
	// The count belongs to the chat it was asked for: what is kept is
	// the last group that answered, which may not be this one yet.
	if a.group == nil || a.group.ID != id || a.group.Headcount <= 0 {
		return name
	}
	return i18n.T("chat.groupTitle", name, strconv.Itoa(a.group.Headcount))
}

// cardArrives is how a card comes and goes: it grows into place from the
// corner it is hung by, rather than being there the moment it is asked for.
var cardArrives = ui.ElementTransition{
	Duration: 170 * time.Millisecond,
	Ease:     ui.EaseOut,
	Enter:    &ui.Motion{Y: -10, Opacity: 0, Scale: 0.96},
	Exit:     &ui.Motion{Y: -6, Opacity: 0, Scale: 0.98},
}
