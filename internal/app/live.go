package app

import (
	"context"
	"time"

	"mygohu/internal/yh"
)

// connect keeps a live connection, so messages arrive as they are sent.
// The account's id comes as an argument rather than from a.me: a.me is set
// on the drawing thread, and this runs on another one.
func (a *App) connect(userID string) {
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.conn = a.client.NewConn(userID, a.onEvent, func(s yh.State) {
		a.update(func() { a.live = s })
	})
	go a.conn.Run(ctx)
}

// onEvent applies something the server pushed.
func (a *App) onEvent(ev yh.Event) {
	a.update(func() {
		switch ev.Cmd {
		case "push_message":
			m := ev.Message
			if m == nil || m.MsgID == "" {
				return
			}
			if a.me != nil && m.Sender.ChatID == a.me.ID {
				return
			}
			// Looking at the chat is reading it, so nothing is told of
			// what is already on screen — unless the window is behind
			// another one or down in the taskbar, which is when a
			// desktop notification earns its keep. The list is told
			// either way: what was said moves the chat and changes the
			// line it shows, whether or not the chat is the one on
			// screen.
			here := m.ChatID == a.current
			if here {
				a.addMessage(m)
				go func() { _ = a.client.MarkRead(m.ChatID) }()
			}
			if a.me != nil {
				a.bump(m, !here)
				if !here || a.away() {
					a.tellArrival(m)
				}
			}
		case "edit_message":
			if m := ev.Message; m != nil {
				for _, old := range a.msgs {
					if old.MsgID == m.MsgID {
						if m.Content.Text != "" {
							old.Content.Text = m.Content.Text
						}
						old.EditedAt = m.EditedAt
						break
					}
				}
			}
		case "blocked_message":
			if m := ev.Message; m != nil {
				for _, old := range a.msgs {
					if old.MsgID == m.MsgID {
						old.RecalledAt = m.RecalledAt
						if old.RecalledAt == 0 {
							old.RecalledAt = uint64(time.Now().UnixMilli())
						}
						break
					}
				}
			}
		case "stream_message":
			s := ev.Stream
			for _, old := range a.msgs {
				if old.MsgID == s.MsgID {
					old.Content.Text += s.Content
					break
				}
			}
		}
	})
}

// away reports whether the window is somewhere the reader is not looking:
// behind another window, or down in the taskbar.
func (a *App) away() bool {
	return a.win != nil && (!a.win.IsFocused() || a.win.IsMinimized())
}

// addMessage adds a message to the open chat, unless it is there already.
// A message that arrives while the reader is at the end of the list is
// read as it lands, and opens no stretch of unread ones; one that arrives
// while they are away from the end opens one at itself.
func (a *App) addMessage(m *yh.Message) {
	for _, old := range a.msgs {
		if old.MsgID == m.MsgID {
			return
		}
	}
	read := a.msgList.AtEnd()
	a.msgs = append(a.msgs, m)
	a.trimMessages()
	if read {
		a.unreadFrom = len(a.msgs)
		a.readFrom = len(a.msgs)
	}
}
