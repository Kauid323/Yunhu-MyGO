package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// msgMenuState is the menu a right press on a message opens: the message it
// is about, and the place on the window that was pressed.
type msgMenuState struct {
	menu
	msg  *yh.Message
	x, y float32
}

// openMsgMenu opens the menu of a message where it was pressed.
func (a *App) openMsgMenu(m *yh.Message, at ui.Rect) {
	if m == nil || m.MsgID == "" {
		return
	}
	x, y := at.X+40, at.Y+10
	if px, py, ok := cursorPoint(); ok && a.win != nil && !a.win.IsDestroyed() {
		b := a.win.Bounds()
		x, y = float32(px-b.X), float32(py-b.Y)
	}
	a.msgMenu = msgMenuState{menu: a.msgMenuRows(m), msg: m, x: x, y: y}
	a.msgMenu.open = true
}

// msgMenuView draws that menu while it is open, hung off the place that was
// pressed.
func (a *App) msgMenuView(c *ui.Context) {
	if !a.msgMenu.open {
		return
	}
	anchor := ui.Box(c).Absolute().Left(a.msgMenu.x).Top(a.msgMenu.y).Size(1, 1)
	showMenu(c, anchor, &a.msgMenu.menu)
}

// msgMenuRows is what the menu of a message holds. A message of the account's
// own can be taken back; one that is somebody else's cannot, and the row is
// not drawn for it at all rather than drawn and refused.
//
// The one row that is dim is the one this client cannot do: nothing in the
// API favourites a message. It is shown dim rather than left out, as the
// client this one copies shows what it will not do at the moment.
func (a *App) msgMenuRows(m *yh.Message) menu {
	rows := []menuItem{
		{icon: iconCopy, label: "msg.copy", do: func() { a.copyMessage(m) }},
		{icon: iconForward, label: "msg.forward", do: func() { a.msgForwardRows(m) }},
		{icon: iconSaveImage, label: "msg.makeImage", do: func() { a.messageImage([]*yh.Message{m}) }},
		{icon: iconStar, label: "msg.favorite", dim: true},
		{icon: iconCheckCircle, label: "msg.select", do: func() { a.startPicking(m.MsgID) }},
		{icon: iconQuote, label: "msg.quote", do: func() { a.quote = m }},
	}
	if m.Mine(a.me.ID) && m.RecalledAt == 0 {
		rows = append(rows, menuItem{icon: iconRecall, label: "msg.recall",
			do: func() { a.recallMessages([]*yh.Message{m}) }})
	}
	rows = append(rows, menuItem{icon: iconTrash, label: "msg.delete",
		do: func() { a.deleteMessage(m) }})
	return menu{width: 150, point: true, rows: rows}
}

// msgForwardRows puts the chats a message can be sent on to in place of the
// menu, as the mute options replace it in the menu of a chat: a forward is
// two presses, not a window of its own.
func (a *App) msgForwardRows(m *yh.Message) {
	rows := make([]menuItem, 0, 9)
	for i, conv := range a.convs {
		if i == 8 {
			break
		}
		c := conv
		rows = append(rows, menuItem{
			icon: iconChat,
			text: c.Title(),
			do:   func() { a.forwardMessage(m, c) },
		})
	}
	if len(rows) == 0 {
		rows = append(rows, menuItem{icon: iconForward, label: "msg.noChat", dim: true})
	}
	rows = append(rows, menuItem{
		icon:  iconChevronLeft,
		label: "conv.back",
		do:    func() { a.msgMenu.rows = a.msgMenuRows(m).rows },
	})
	a.msgMenu.rows = rows
}

// copyMessage puts what a message says on the clipboard.
func (a *App) copyMessage(m *yh.Message) {
	text, detail := messageText(m)
	if text == "" {
		text = detail
	}
	if text == "" {
		text = i18n.T("search.image")
	}
	copyText(text)
}

// forwardMessage sends a message on to another chat.
func (a *App) forwardMessage(m *yh.Message, to yh.Conversation) {
	from := yh.Target{ChatID: to.ChatID, ChatType: to.ChatType}
	chatType := a.currentType
	go func() {
		err := a.client.ForwardMessage(m.MsgID, chatType, []yh.Target{from})
		a.update(func() {
			if err != nil {
				a.note = errorLine(err)
			}
		})
	}()
}

// deleteMessage takes a message off the history: the route deletes it for
// this account, and the row goes with it.
func (a *App) deleteMessage(m *yh.Message) {
	chatID, chatType := a.current, a.currentType
	go func() {
		err := a.client.DeleteMessage(chatID, chatType, m.MsgID)
		a.update(func() {
			if err != nil {
				a.note = errorLine(err)
				return
			}
			kept := a.msgs[:0]
			for _, x := range a.msgs {
				if x.MsgID != m.MsgID {
					kept = append(kept, x)
				}
			}
			a.msgs = kept
		})
	}()
}

// rightClickMessage opens the menu of a message when the row it is drawn in
// is pressed with the secondary button. While the history is being picked
// from there is no menu: a press chooses, and the bar is where the actions
// are.
func (a *App) rightClickMessage(m *yh.Message, row ui.Element) {
	if a.picking {
		return
	}
	if row.RightClicked() {
		a.openMsgMenu(m, row.Bounds())
	}
}

// quoteTextOf is how a message is written when it is quoted: who said it,
// and what they said, as the server writes a quote and as the block in the
// bubble reads it back.
func quoteTextOf(m *yh.Message) string {
	text, _ := messageText(m)
	if text == "" {
		text = i18n.T("search.image")
	}
	name := m.Sender.Name
	if name == "" {
		name = m.Sender.ChatID
	}
	return name + ": " + text
}
