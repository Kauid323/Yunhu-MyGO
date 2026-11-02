package app

import (
	"strings"
	"time"

	"mygohu/internal/yh"
)

// open shows a conversation, and loads its messages. It leaves the
// conversation's information panel as it is: what is known about a chat is
// asked for, not shown every time another one is picked.
//
// Where the chat is shown is the layout's business: the efficient one has
// it beside the list, and the classic one has no room for it and opens it in
// a window of its own.
func (a *App) open(conv yh.Conversation) {
	if a.current == conv.ChatID {
		// Picking the chat that is already open brings the window it is
		// in forward, which is what the classic panel has to do: the
		// chat is not in the panel.
		if a.prefs.Mode == modeClassic {
			a.openChatWindow()
		}
		return
	}
	a.current = conv.ChatID
	a.currentType = conv.ChatType
	a.msgs = nil
	a.loading = true
	a.loadingMore = false
	a.hasMore = false
	a.loadedAt = 1 << 30
	a.input = ""
	// What was unread is remembered before the list is told it is read:
	// it is where the chat opens, and where the line across the history
	// goes.
	unread := 0
	for i := range a.convs {
		if a.convs[i].ChatID == conv.ChatID {
			unread = a.convs[i].Unread
			a.convs[i].Unread = 0
		}
	}
	a.unreadOnOpen = unread
	a.unreadFrom = 0
	a.readFrom = 0
	go func() { _ = a.client.MarkRead(conv.ChatID) }()
	a.loadMessages(conv.ChatID, conv.ChatType)
	if a.prefs.Mode == modeClassic {
		// The panel has no room for a chat: it is read in a window of its
		// own, which follows the chat that is open.
		a.openChatWindow()
	}
}

// loadMessages loads the latest messages of a chat.
func (a *App) loadMessages(chatID string, chatType int) {
	go func() {
		msgs, err := a.client.Messages(chatID, chatType, msgFirst)
		a.update(func() {
			if a.current != chatID {
				return
			}
			a.loading = false
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.setMessages(msgs)
		})
	}()
}

// setMessages shows a chat's first page, opened where reading left off: at
// the first message that came in while the reader was away, or at the
// newest one when there was nothing unread.
func (a *App) setMessages(msgs []*yh.Message) {
	a.msgs = msgs
	a.hasMore = len(msgs) >= msgFirst
	a.loadedAt = 1 << 30
	a.unreadFrom = len(msgs) - a.unreadOnOpen
	if a.unreadFrom < 0 {
		a.unreadFrom = 0
	}
	a.readFrom = a.unreadFrom
	a.unreadOnOpen = 0
	// The scroll is asked for as the list is built, so that the same frame
	// lays out where the chat opens rather than waiting for the next one:
	// at the first message that was unread, or at the newest one when
	// there was none.
	a.jumpTo = -1
	if a.unread() {
		a.jumpTo = a.unreadFrom
		return
	}
	a.jumpToEnd = true
}

// loadOlder puts the messages before the oldest one on screen above it, and
// leaves the list where it was scrolled.
func (a *App) loadOlder() {
	if a.loadingMore || !a.hasMore || a.current == "" || len(a.msgs) == 0 {
		return
	}
	chatID, chatType := a.current, a.currentType
	before := a.msgs[0].MsgID
	a.loadingMore = true
	go func() {
		older, err := a.client.MessagesBefore(chatID, chatType, msgPage, before)
		a.update(func() {
			a.loadingMore = false
			if err != nil {
				a.note = errorLine(err)
				// A page that did not come is asked for again, and the row
				// it was asked for at says nothing about the next try: the
				// reader may be sitting at the top, where that row is the
				// only one there is.
				a.loadedAt = 1 << 30
				return
			}
			if a.current != chatID {
				return
			}
			if len(older) == 0 {
				a.hasMore = false
				return
			}
			fresh := a.withoutKnown(older)
			if len(fresh) == 0 {
				// The page held nothing new: asking again would ask for
				// the same messages for ever.
				a.hasMore = false
				return
			}
			a.prepend(fresh)
			a.hasMore = len(older) >= msgPage
		})
	}()
}

// prepend puts a page of older messages above the history, and moves what the
// list is showing down with it: the unread ones are that many rows further
// down now, and so is the row the last page was asked for at. Without that
// last one the top of the history would be a row nothing is ever asked for
// at again.
func (a *App) prepend(fresh []*yh.Message) {
	a.msgs = append(fresh, a.msgs...)
	a.unreadFrom += len(fresh)
	a.readFrom += len(fresh)
	a.loadedAt += len(fresh)
}

// withoutKnown drops from a page the messages already on screen, so that a
// chat paged twice shows each message once.
func (a *App) withoutKnown(page []*yh.Message) []*yh.Message {
	known := make(map[string]bool, len(a.msgs))
	for _, m := range a.msgs {
		known[m.MsgID] = true
	}
	out := page[:0]
	for _, m := range page {
		if !known[m.MsgID] {
			out = append(out, m)
		}
	}
	return out
}

// reload reloads the open conversation.
func (a *App) reload() {
	if a.current == "" {
		a.refreshConversations()
		return
	}
	a.loading = true
	a.loadMessages(a.current, a.currentType)
}

// unread reports whether anything came in while the reader was away, which
// the line across the history marks.
func (a *App) unread() bool { return a.unreadFrom < len(a.msgs) }

// unreadBelow is how many messages the reader has not been past yet: what
// the button over the list counts. It only goes down. Going back up
// through the history does not count anything again, as coming into the
// chat at all is reading what is there.
func (a *App) unreadBelow() int {
	if n := len(a.msgs) - a.readFrom; n > 0 {
		return n
	}
	return 0
}

// markRead records that the reader has been as far down as the top of the
// view: what is above it is read, and the count on the button drops by as
// much. At the end of the list everything is read, the newest message
// being in view there.
func (a *App) markRead(first, last int) {
	if a.readFrom >= len(a.msgs) {
		return
	}
	if a.msgList.AtEnd() {
		a.readFrom = len(a.msgs)
		return
	}
	if last < first {
		return
	}
	if first > a.readFrom {
		a.readFrom = first
	}
}

// send sends what is in the composer.
func (a *App) send() {
	text := strings.TrimSpace(a.input)
	if text == "" || a.current == "" || a.me == nil {
		return
	}
	a.input = ""
	chatID, chatType := a.current, a.currentType
	pending := &yh.Message{
		MsgID:       "local-" + yh.NewID(),
		Sender:      yh.Sender{ChatID: a.me.ID, Name: a.me.Name, AvatarURL: a.me.AvatarURL, ChatType: yh.ChatUser},
		ContentType: yh.ContentText,
		Content:     yh.Content{Text: text},
		TimestampMS: uint64(time.Now().UnixMilli()),
		Pending:     true,
	}
	a.msgs = append(a.msgs, pending)
	a.jumpToEnd = true

	go func() {
		sent, err := a.client.SendText(chatID, chatType, text)
		a.update(func() {
			for i, m := range a.msgs {
				if m.MsgID != pending.MsgID {
					continue
				}
				if err != nil {
					m.Pending = false
					m.Failed = err.Error()
					a.note = errorLine(err)
					return
				}
				sent.Sender = pending.Sender
				a.msgs[i] = sent
				return
			}
		})
		if err == nil {
			a.update(func() {
				for i := range a.convs {
					if a.convs[i].ChatID == chatID {
						a.convs[i].Content = text
						a.convs[i].SendTime = uint64(time.Now().Unix())
						a.sortConversations()
						return
					}
				}
			})
		}
	}()
}
