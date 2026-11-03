package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// ask is a question the window puts to the reader over what it shows: a
// title, one line to fill, and two answers. It is what a thing that needs
// a name is asked with, rather than a window of its own.
type ask struct {
	open  bool
	title string // a key of the dictionary
	hint  string // the key of what the field is for
	value string
	do    func(string)
}

// newChatMenu is the menu of the button beside the search box: the things
// that are made rather than found.
//
// Adding a friend has no row of its own that works: the API has no route
// that asks another account to be one, so the row is drawn dim rather than
// left out, as the client this one copies draws what it will not do.
func newChatMenu(a *App) menu {
	return menu{
		width: 150,
		below: true,
		rows: []menuItem{
			{icon: iconGroupPlus, label: "conv.newGroup", do: a.askGroup},
			{icon: iconUserPlus, label: "conv.addFriend", dim: true},
			{icon: iconBot, label: "conv.newBot", do: a.askBot},
		},
	}
}

// askGroup asks what a new chat is to be called.
func (a *App) askGroup() {
	a.ask = ask{open: true, title: "conv.newGroup", hint: "conv.groupName",
		do: a.createGroup}
}

// askBot asks what a new bot is to be called.
func (a *App) askBot() {
	a.ask = ask{open: true, title: "conv.newBot", hint: "conv.botName",
		do: a.createBot}
}

// createGroup makes the chat and loads the list again, so that it is in it.
func (a *App) createGroup(name string) {
	go func() {
		_, err := a.client.CreateGroup(name, "")
		a.update(func() {
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.note = say("conv.groupMade")
		})
		if err == nil {
			a.refreshConversations()
		}
	}()
}

// createBot makes the bot. A bot is not a chat, so the list does not
// change: what changes is that the account has one more.
func (a *App) createBot(name string) {
	go func() {
		err := a.client.CreateBot(name, "")
		a.update(func() {
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.note = say("conv.botMade")
		})
	}()
}

// askView draws the question while it is being asked.
func (a *App) askView(c *ui.Context) {
	if !a.ask.open {
		return
	}
	scrim := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).
		Background(ui.RGBA(0, 0, 0, 0.45)).Center()
	if scrim.Clicked() {
		a.ask.open = false
	}
	scrim.Children(func() {
		card := ui.Column(c).Width(300).Padding(16).Gap(12).Radius(10).
			Background(qq.panelBg).Border(1, qq.border).
			Shadow(0, 16, 40, 0, ui.RGBA(0, 0, 0, 0.45))
		card.Children(func() {
			ui.Text(c, i18n.T(a.ask.title)).FontSize(13.5).TextColor(qq.text)
			in := ui.TextInput(c, &a.ask.value).AutoFocus().
				Placeholder(i18n.T(a.ask.hint)).
				Background(qq.inputBg).Border(1, qq.border).Radius(6).
				FontSize(12.5).TextColor(qq.text)
			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, i18n.T("ask.cancel")).Radius(6).
					Padding(4, 12).FontSize(12).Clicked() {
					a.ask.open = false
				}
				if ui.Button(c, i18n.T("ask.ok")).Radius(6).
					Padding(4, 12).FontSize(12).Clicked() || in.Submitted() {
					name := a.ask.value
					do := a.ask.do
					a.ask = ask{}
					if name != "" && do != nil {
						do(name)
					}
				}
			})
		})
	})
}
