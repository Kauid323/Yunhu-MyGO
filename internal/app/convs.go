package app

import (
	"sort"
	"strings"

	"mygohu/internal/yh"
)

// refreshConversations loads the conversation list.
func (a *App) refreshConversations() {
	go func() {
		list, _, err := a.client.Conversations("")
		if err != nil {
			a.update(func() { a.note = errorLine(err) })
			return
		}
		a.update(func() {
			a.convs = list
			a.sortConversations()
			a.note = line{}
		})
		a.update(func() {
			if a.current == "" && len(a.convs) > 0 {
				a.open(a.convs[0])
			}
		})
	}()
}

// sortConversations puts the conversation talked in last first.
func (a *App) sortConversations() {
	sort.SliceStable(a.convs, func(i, j int) bool {
		return a.convs[i].Time().After(a.convs[j].Time())
	})
}

// bump moves a conversation up with its new message.
// bump moves a chat to where what was said in it puts it: the line the list
// shows is written here, from the message itself, because a pushed message
// is not a line of a list and says nothing of who said it.
//
// unread says whether the count goes up: a message that arrived in the chat
// on screen has been read by the time it is drawn, but the chat still moves
// to the top of the list and still says what was said.
func (a *App) bump(m *yh.Message, unread bool) {
	me := ""
	if a.me != nil {
		me = a.me.ID
	}
	for i := range a.convs {
		c := &a.convs[i]
		if c.ChatID != m.ChatID {
			continue
		}
		c.Content = said(m, c.ChatType, me)
		c.SendTime = uint64(m.Time().Unix())
		c.TimestampMS = m.TimestampMS
		c.At = false
		for _, id := range m.Content.MentionedIDs {
			if id == me && me != "" {
				c.At = true
			}
		}
		if unread {
			c.Unread++
		}
		a.sortConversations()
		return
	}
	a.refreshConversations()
}

// currentConv returns the conversation on screen.
func (a *App) currentConv() *yh.Conversation {
	for i := range a.convs {
		if a.convs[i].ChatID == a.current {
			return &a.convs[i]
		}
	}
	return nil
}

// filteredConvs are the conversations the search keeps, with the chats that
// were pinned at the top of them: a pinned chat is one that is looked at
// first, whatever was said in the others.
func (a *App) filteredConvs() []yh.Conversation {
	q := strings.ToLower(strings.TrimSpace(a.filter))
	var out []yh.Conversation
	for _, c := range a.convs {
		if q != "" && !strings.Contains(strings.ToLower(c.Title()), q) &&
			!strings.Contains(strings.ToLower(c.Content), q) {
			continue
		}
		out = append(out, c)
	}
	if len(a.prefs.Pinned) > 0 {
		sort.SliceStable(out, func(i, j int) bool {
			return a.prefs.Pinned[out[i].ChatID] && !a.prefs.Pinned[out[j].ChatID]
		})
	}
	return out
}

// chatTypeOf is what kind of chat an id is, from the list: 0 when the list
// does not know it.
func (a *App) chatTypeOf(chatID string) int {
	for _, c := range a.convs {
		if c.ChatID == chatID {
			return c.ChatType
		}
	}
	return 0
}
