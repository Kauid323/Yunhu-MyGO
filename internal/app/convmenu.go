package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// convMenuState is the menu a right press on a chat opens: the chat it is
// about, and the place on the window that was pressed. A menu with no
// button to come out of comes out of that place.
type convMenuState struct {
	menu
	chatID string
	x, y   float32
}

// openConvMenu opens the menu of a chat where it was pressed. The pointer
// is asked for rather than the row: the menu is about the place, and the
// row it was pressed on is only what it is about.
func (a *App) openConvMenu(conv yh.Conversation, at ui.Rect) {
	x, y := at.X+24, at.Y+8
	if px, py, ok := cursorPoint(); ok && a.win != nil && !a.win.IsDestroyed() {
		b := a.win.Bounds()
		x, y = float32(px-b.X), float32(py-b.Y)
	}
	a.convMenu = convMenuState{
		menu:   a.convMenuRows(conv),
		chatID: conv.ChatID,
		x:      x,
		y:      y,
	}
	a.convMenu.open = true
}

// convMenuView draws that menu while it is open, hung off the place that
// was pressed.
func (a *App) convMenuView(c *ui.Context) {
	if !a.convMenu.open {
		return
	}
	// A box the size of nothing, where the press was: the menu hangs off
	// it as it hangs off a button.
	anchor := ui.Box(c).Absolute().Left(a.convMenu.x).Top(a.convMenu.y).
		Size(1, 1)
	showMenu(c, anchor, &a.convMenu.menu)
}

// convMenuRows is what the menu of a chat holds.
func (a *App) convMenuRows(conv yh.Conversation) menu {
	pinned := a.prefs.Pinned[conv.ChatID]
	pinKey := "conv.pin"
	if pinned {
		pinKey = "conv.unpin"
	}
	return menu{
		width: 186,
		point: true,
		rows: []menuItem{
			{icon: iconPin, label: pinKey, do: func() { a.pinConv(conv, !pinned) }},
			{icon: iconCopy, label: "conv.copyId", do: func() { copyText(conv.ChatID) }},
			{icon: iconCheckCircle, label: "conv.markRead", do: func() { a.markConvRead(conv) }},
			{icon: iconWindow, label: "conv.openWindow", do: func() { a.openConvWindow(conv) }},
			{icon: iconBell, label: "conv.mute", more: true, do: func() { a.convMuteRows(conv) }},
			{icon: iconTrash, label: "conv.remove", do: func() { a.removeConv(conv) }},
		},
	}
}

// convMuteRows puts what a chat's messages are taken as in place of the
// menu: the same three ways the notices page offers, for the one chat.
//
// What the menu holds is changed rather than the menu itself: the popup
// watches the field that says it is open, and a menu made anew is a menu
// whose openness nothing is watching any more.
func (a *App) convMuteRows(conv yh.Conversation) {
	now := a.muted(conv.ChatID)
	rows := make([]menuItem, 0, 4)
	for _, way := range []string{"notify", "quiet", "blocked"} {
		w := way
		it := menuItem{
			icon:  iconBell,
			label: "settings.value." + w,
			do: func() {
				a.setMuted(conv.ChatID, w)
				a.convMenu.open = false
			},
		}
		if now == w {
			it.icon = iconCheck
		}
		rows = append(rows, it)
	}
	rows = append(rows, menuItem{
		icon:  iconChevronLeft,
		label: "conv.back",
		do:    func() { a.convMenu.rows = a.convMenuRows(conv).rows },
	})
	a.convMenu.rows = rows
}

// pinConv puts a chat at the top of the list, or takes it back to where the
// server would have it.
func (a *App) pinConv(conv yh.Conversation, on bool) {
	a.pickPrefs(func(p *prefs) {
		if p.Pinned == nil {
			p.Pinned = map[string]bool{}
		}
		if !on {
			delete(p.Pinned, conv.ChatID)
			return
		}
		p.Pinned[conv.ChatID] = true
	})
	// The list is drawn again: a chat that moved is not seen to move
	// until something draws it.
	a.update(func() {})
}

// markConvRead says a chat has been read, and takes the count off it.
func (a *App) markConvRead(conv yh.Conversation) {
	go func() { _ = a.client.MarkRead(conv.ChatID) }()
	a.update(func() {
		for i := range a.convs {
			if a.convs[i].ChatID != conv.ChatID {
				continue
			}
			a.convs[i].Unread = 0
			a.convs[i].At = false
		}
	})
}

// openConvWindow opens a chat in a window of its own: the classic layout
// has no room for a chat and does it on a press, and the other one puts it
// there too when it is asked.
func (a *App) openConvWindow(conv yh.Conversation) {
	a.open(conv)
	if a.prefs.Mode == modeEfficient {
		a.openChatWindow()
	}
}

// removeConv takes a chat off the list. The API has no way to leave one, so
// it comes back when the list is loaded again and the chat says something:
// this hides it, it does not leave it.
func (a *App) removeConv(conv yh.Conversation) {
	a.update(func() {
		kept := a.convs[:0]
		for _, c := range a.convs {
			if c.ChatID != conv.ChatID {
				kept = append(kept, c)
			}
		}
		a.convs = kept
		if a.current != conv.ChatID {
			return
		}
		a.current, a.currentType = "", 0
		a.msgs = nil
		a.closeChatWindow()
	})
}

// copyText puts a string on the clipboard.
func copyText(s string) {
	_ = mygo.Clipboard.Write(transfer.TextData(s))
}
