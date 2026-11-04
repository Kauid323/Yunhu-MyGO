package app

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// A group message may name people with an @: the name is typed into the
// text, and the account it stands for goes with the message, which is what
// makes the chat on the other side mark the name rather than show it as
// words. The panel over the composer is how a name is put in without typing
// it, and who it stands for.
const (
	mentionPanelW = 236
	mentionRowH   = 36
	mentionRows   = 7
	// mentionPage is how many of a group are asked for. A group of a
	// thousand is not a list to scroll: the panel is for naming the people
	// in the conversation, and what is beyond the first page is found by
	// typing the name.
	mentionPage = 100
)

// named is somebody a message names: the account, and the name that stands
// for it in the text.
type named struct {
	ID   string
	Name string
}

// loadMembers asks the group that is open who is in it, once per group: the
// panel is asked for again only when another group is opened.
func (a *App) loadMembers() {
	if a.mentionGo || len(a.members) > 0 || a.current == "" {
		return
	}
	group := a.current
	a.mentionGo = true
	go func() {
		list, err := a.client.Members(group, 1, mentionPage)
		a.update(func() {
			a.mentionGo = false
			if a.current != group {
				return
			}
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.members = list
		})
	}()
}

// toggleMention shows the panel of the group's members, or puts it away.
func (a *App) toggleMention() {
	a.mentionOpen = !a.mentionOpen
	if a.mentionOpen {
		a.loadMembers()
	}
}

// mentionPanel draws that panel over the row of tools while it is open.
func (a *App) mentionPanel(c *ui.Context, anchor ui.Element) {
	if !a.mentionOpen {
		return
	}
	ui.PopoverBase(c, anchor, &a.mentionOpen, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorBottomLeft).
			Left(-8).Width(mentionPanelW).Padding(8).Radius(10).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 8, 22, 0, ui.RGBA(0, 0, 0, 0.30)).
			TransformOrigin(0, 1).Transition(menuArrives)
		a.mentionPeople(c)
	})
}

// mentionPeople is the panel: a field to narrow the group by, and the
// members it leaves, a row each.
func (a *App) mentionPeople(c *ui.Context) {
	ui.TextInput(c, &a.mentionFind).Placeholder(i18n.T("mention.find")).
		FontSize(12).Padding(5, 8).Radius(6).Background(qq.inputBg).
		Border(0, qq.inputBg).Margin(0, 0, 6, 0)
	if a.mentionGo {
		ui.Spinner(c).Margin(ui.Auto)
		return
	}
	people := a.membersMatching()
	if len(people) == 0 {
		text := i18n.T("mention.empty")
		if len(a.members) == 0 {
			text = i18n.T("mention.none")
		}
		ui.Text(c, text).FontSize(11.5).TextColor(qq.textMuted).Padding(8)
		return
	}
	shown := min(len(people), mentionRows)
	ui.List(c, &a.mentionList, len(people), func(i int) {
		a.mentionRow(c, people[i])
	}).Height(float32(shown * mentionRowH)).Padding(0).Gap(0)
}

// membersMatching is the members the panel shows: all of them, or the ones
// whose name holds what has been typed into it.
func (a *App) membersMatching() []yh.Member {
	find := strings.TrimSpace(strings.ToLower(a.mentionFind))
	if find == "" {
		return a.members
	}
	out := make([]yh.Member, 0, len(a.members))
	for _, m := range a.members {
		if strings.Contains(strings.ToLower(m.Name), find) {
			out = append(out, m)
		}
	}
	return out
}

// mentionRow is one member of the panel.
func (a *App) mentionRow(c *ui.Context, m yh.Member) {
	box := ui.Row(c).Height(mentionRowH).Padding(4, 6).Gap(8).
		AlignItems(ui.Center).Radius(6).Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.menuHover)
	}
	box.Children(func() {
		a.avatar(c, m.AvatarURL, m.Name, 24)
		ui.Text(c, m.Name).FontSize(12.5).TextColor(qq.text).Grow(1).
			MinWidth(0).SingleLine()
		if m.Permission >= 100 {
			ui.Text(c, i18n.T("mention.owner")).FontSize(10).TextColor(qq.textFaint)
		}
	})
	if box.Clicked() {
		a.mention(m)
	}
}

// mention puts a member's name into the text, and remembers the account it
// stands for: the name is what the reader sees, and the account is what the
// message carries.
func (a *App) mention(m yh.Member) {
	a.mentionOpen = false
	name := strings.TrimSpace(m.Name)
	if name == "" {
		return
	}
	if a.input != "" && !strings.HasSuffix(a.input, " ") {
		a.input += " "
	}
	a.input += "@" + name + " "
	for _, n := range a.named {
		if n.ID == m.ID {
			return
		}
	}
	a.named = append(a.named, named{ID: m.ID, Name: name})
}

// namedIn is the accounts a text still names: a name that was put in and
// then typed over is not one the message is about any more, so what is sent
// is worked out from the text rather than from what was picked.
func (a *App) namedIn(text string) []string {
	if len(a.named) == 0 {
		return nil
	}
	out := make([]string, 0, len(a.named))
	for _, n := range a.named {
		if strings.Contains(text, "@"+n.Name) {
			out = append(out, n.ID)
		}
	}
	return out
}

// forgetNamed drops the names the text no longer holds, so that a name
// typed, sent and deleted is not carried into the next message.
func (a *App) forgetNamed(text string) {
	if len(a.named) == 0 {
		return
	}
	kept := a.named[:0]
	for _, n := range a.named {
		if strings.Contains(text, "@"+n.Name) {
			kept = append(kept, n)
		}
	}
	a.named = kept
}
