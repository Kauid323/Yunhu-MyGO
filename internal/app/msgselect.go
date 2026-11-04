package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// Picking is reading the history to choose from it: a message pressed in it
// is chosen rather than acted on, and the bar over the list says how many
// are chosen and what can be done with them. It is a mode rather than a
// menu, because what is picked outlives the press that picked it.
func (a *App) startPicking(first string) {
	a.picking = true
	a.picked = map[string]bool{}
	if first != "" {
		a.picked[first] = true
	}
}

// stopPicking puts the history back to being read.
func (a *App) stopPicking() {
	a.picking = false
	a.picked = nil
}

// togglePick chooses a message, or lets it go.
func (a *App) togglePick(id string) {
	if a.picked == nil {
		a.picked = map[string]bool{}
	}
	if a.picked[id] {
		delete(a.picked, id)
		return
	}
	a.picked[id] = true
}

// pickAll chooses every message the history holds, or lets them all go when
// they are all chosen already.
func (a *App) pickAll() {
	if len(a.picked) == len(a.msgs) {
		a.picked = map[string]bool{}
		return
	}
	a.picked = make(map[string]bool, len(a.msgs))
	for _, m := range a.msgs {
		a.picked[m.MsgID] = true
	}
}

// pickedMessages is what was chosen, oldest first: what the actions of the
// bar work on, in the order the history holds them.
func (a *App) pickedMessages() []*yh.Message {
	out := make([]*yh.Message, 0, len(a.picked))
	for _, m := range a.msgs {
		if a.picked[m.MsgID] {
			out = append(out, m)
		}
	}
	return out
}

// pickForward opens the list of chats a message can be sent on, hung off the
// button that asked for it, as the same list in the menu of a message is.
func (a *App) pickForward(at ui.Rect) {
	chosen := a.pickedMessages()
	if len(chosen) == 0 {
		return
	}
	rows := make([]menuItem, 0, 9)
	for i, conv := range a.convs {
		if i == 8 {
			break
		}
		c := conv
		rows = append(rows, menuItem{
			icon: iconChat,
			text: c.Title(),
			do:   func() { a.forwardPicked(chosen, c) },
		})
	}
	if len(rows) == 0 {
		rows = append(rows, menuItem{icon: iconForward, label: "msg.noChat", dim: true})
	}
	x, y := at.X+40, at.Y+10
	if px, py, ok := cursorPoint(); ok && a.win != nil && !a.win.IsDestroyed() {
		b := a.win.Bounds()
		x, y = float32(px-b.X), float32(py-b.Y)
	}
	a.msgMenu = msgMenuState{menu: menu{width: 150, point: true, rows: rows}, x: x, y: y}
	a.msgMenu.open = true
}

// forwardPicked sends the chosen messages on to one chat, and leaves the
// mode they were chosen in.
func (a *App) forwardPicked(msgs []*yh.Message, to yh.Conversation) {
	a.stopPicking()
	target := yh.Target{ChatID: to.ChatID, ChatType: to.ChatType}
	chatType := a.currentType
	go func() {
		var failed error
		for _, m := range msgs {
			if err := a.client.ForwardMessage(m.MsgID, chatType, []yh.Target{target}); err != nil {
				failed = err
			}
		}
		if failed != nil {
			a.update(func() { a.note = errorLine(failed) })
		}
	}()
}

// pickDelete takes the chosen messages off the history, as the menu of one
// message does.
func (a *App) pickDelete() {
	chosen := a.pickedMessages()
	if len(chosen) == 0 {
		return
	}
	a.stopPicking()
	chatID, chatType := a.current, a.currentType
	go func() {
		gone := make(map[string]bool, len(chosen))
		var failed error
		for _, m := range chosen {
			if err := a.client.DeleteMessage(chatID, chatType, m.MsgID); err != nil {
				failed = err
				continue
			}
			gone[m.MsgID] = true
		}
		a.update(func() {
			if len(gone) > 0 {
				kept := a.msgs[:0]
				for _, m := range a.msgs {
					if !gone[m.MsgID] {
						kept = append(kept, m)
					}
				}
				a.msgs = kept
			}
			if failed != nil {
				a.note = errorLine(failed)
			}
		})
	}()
}

// recallMessages takes messages back for everyone in the chat, and marks them
// so at once: the row keeps its place and says that it was recalled, as the
// chat on the other side shows it. Only messages of the account's own can be
// taken back, and the server is what says whether one is still in time.
func (a *App) recallMessages(msgs []*yh.Message) {
	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		if m.Mine(a.me.ID) && m.RecalledAt == 0 {
			ids = append(ids, m.MsgID)
		}
	}
	if len(ids) == 0 {
		a.note = say("msg.recallNone")
		return
	}
	chatID, chatType := a.current, a.currentType
	a.stopPicking()
	go func() {
		err := a.client.RecallMessages(chatID, chatType, ids)
		a.update(func() {
			if err != nil {
				a.note = errorLine(err)
				return
			}
			recalled := make(map[string]bool, len(ids))
			for _, id := range ids {
				recalled[id] = true
			}
			for _, m := range a.msgs {
				if recalled[m.MsgID] {
					m.RecalledAt = uint64(time.Now().Unix())
				}
			}
		})
	}()
}

// pickRecall takes back the chosen messages, as the menu of one message does.
func (a *App) pickRecall() {
	chosen := a.pickedMessages()
	if len(chosen) == 0 {
		return
	}
	a.recallMessages(chosen)
}
