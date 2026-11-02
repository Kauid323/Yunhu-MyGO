package app

import (
	"github.com/egoist/mygo/ui"
)

// railW is how wide the strip of sections is.
const railW = 56

// rail is the strip of sections on the far left: the account, then the
// sections, then the menu button at its foot. The account is left off a
// panel that shows it across its top, where it would be the same face twice
// in one window.
func (a *App) rail(c *ui.Context, account bool) {
	ui.Column(c).Width(railW).FillHeight().Background(qq.rail).
		AlignItems(ui.Center).Padding(14, 0, 10, 0).Gap(2).Children(func() {
		if account {
			a.railAvatar(c)
			ui.Box(c).Height(6)
		}

		a.railButton(c, iconChat, 0)
		a.railButton(c, iconContacts, 1)
		a.railButton(c, iconStar, 2)
		a.railButton(c, iconFolder, 3)
		a.railButton(c, iconGrid, 4)

		ui.Spacer(c)

		a.railButton(c, iconBell, 5)
		a.railMore(c)
	})
}

// railAvatar is the account at the top of the rail. It shows what is known
// about the account, and says nothing about the messages: those wait on
// the chat section, where they are read.
func (a *App) railAvatar(c *ui.Context) {
	box := ui.Box(c).Size(38, 38).Cursor(ui.CursorPointer)
	box.Children(func() {
		a.avatar(c, a.me.AvatarURL, a.me.Name, 36)
	})
	if box.Clicked() {
		a.infoOpen = !a.infoOpen
	}
}

// railButton is one section of the rail, lit while it is the open one. The
// chat section carries how many messages wait across all the chats, which
// is what the number on it means.
func (a *App) railButton(c *ui.Context, ic *ui.SVG, section int) {
	active := a.nav == section
	box := ui.Box(c).Size(40, 40).Radius(9).Center().Cursor(ui.CursorPointer)
	switch {
	case active:
		box.Background(qq.railActive)
	case box.Hovered():
		box.Background(qq.railHover)
	}
	box.Children(func() {
		col := qq.railIcon
		if active {
			col = qq.railIconActive
		}
		ui.Icon(c, ic).FontSize(20).TextColor(col)
		if n := totalUnread(a.convs); section == 0 && n > 0 {
			ui.Box(c).Absolute().Top(1).Right(1).Padding(1, 4).Radius(8).
				Background(qq.badge).Children(func() {
				ui.Text(c, shortCount(n)).FontSize(8.5).Bold().
					TextColor(qq.accentText)
			})
		}
	})
	if box.Clicked() {
		a.nav = section
	}
}

// railMore is the button at the foot of the rail: three lines, which open
// the menu of everything the window does that has no section of its own.
func (a *App) railMore(c *ui.Context) {
	box := ui.Box(c).Size(40, 40).Radius(9).Center().Cursor(ui.CursorPointer)
	if box.Hovered() || a.more.open {
		box.Background(qq.railHover)
	}
	box.Children(func() {
		ui.Icon(c, iconMenu).FontSize(18).TextColor(qq.railIcon)
	})
	if box.Clicked() {
		a.more.open = !a.more.open
	}
	showMenu(c, box, &a.more)
}
