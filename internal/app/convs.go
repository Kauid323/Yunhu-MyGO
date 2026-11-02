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
func (a *App) bump(m *yh.Message) {
	for i := range a.convs {
		c := &a.convs[i]
		if c.ChatID != m.ChatID {
			continue
		}
		c.Content = preview(m.ContentType, m.Content)
		c.SendTime = uint64(m.Time().Unix())
		c.TimestampMS = m.TimestampMS
		c.Unread++
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

// filteredConvs are the conversations the search keeps.
func (a *App) filteredConvs() []yh.Conversation {
	q := strings.ToLower(strings.TrimSpace(a.filter))
	if q == "" {
		return a.convs
	}
	var out []yh.Conversation
	for _, c := range a.convs {
		if strings.Contains(strings.ToLower(c.Title()), q) || strings.Contains(strings.ToLower(c.Content), q) {
			out = append(out, c)
		}
	}
	return out
}
