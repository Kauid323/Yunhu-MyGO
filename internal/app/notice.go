package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// What arrives while the reader is not looking at it is told to the
// desktop as a notification, which opens that chat when it is pressed. A
// burst of messages is one notification rather than one each: they are
// gathered for noticeWait, and the chat that spoke last is the one the
// notification is about.

// noticeWait is how long a burst of messages is gathered for.
const noticeWait = 500 * time.Millisecond

// notice is what a chat has to say for itself since it last spoke.
type notice struct {
	chatID   string
	chatType int
	title    string
	count    int
	body     string
	at       time.Time
}

// tellArrival gathers a message that came in while the reader was not
// looking at the chat it is for, and tells the desktop about it once the
// burst is over. Which arrivals are worth telling about is the settings':
// the page that offers it also says what each chat is taken as, and a chat
// that is quiet or blocked says nothing.
func (a *App) tellArrival(m *yh.Message) {
	if !a.notices || m == nil || a.prefs.NoticeKind == "off" {
		return
	}
	if m.Sender.ChatType == yh.ChatGroup && !a.prefs.NotifyGroup {
		return
	}
	if m.Sender.ChatType != yh.ChatGroup && !a.prefs.NotifyFriend {
		return
	}
	if a.muted(m.ChatID) != "notify" {
		return
	}
	if a.pending == nil {
		a.pending = map[string]*notice{}
	}
	n := a.pending[m.ChatID]
	if n == nil {
		n = &notice{chatID: m.ChatID, chatType: m.Sender.ChatType, title: a.chatTitle(m)}
		a.pending[m.ChatID] = n
	}
	n.count++
	// A notification is read at a glance, so it shows a preview's worth of
	// the message rather than all of it: the system cuts a long one
	// wherever it reaches. The settings say whether it says that much at
	// all.
	if a.prefs.NoticeDetail {
		n.body = clip(preview(m.ContentType, m.Content))
	} else {
		n.body = i18n.T("notice.quiet")
	}
	n.at = time.Now()
	a.armNotice()
}

// armNotice starts the wait over: the notification comes once nothing has
// arrived for noticeWait, so a chat that is being typed into quickly is one
// notification rather than twenty.
func (a *App) armNotice() {
	if a.timer != nil {
		a.timer.Stop()
	}
	// The wait runs off the drawing thread, and comes back to it to read
	// and write what it gathers.
	a.timer = time.AfterFunc(noticeWait, func() { a.update(a.flushNotices) })
}

// flushNotices tells the desktop about the burst that just ended.
func (a *App) flushNotices() {
	var last *notice
	for _, n := range a.pending {
		if last == nil || n.at.After(last.at) {
			last = n
		}
	}
	a.pending = nil
	if last != nil {
		a.showNotice(last)
	}
}

// showNotice shows one notification for a chat. Another for the same chat
// replaces it rather than stacking a second.
func (a *App) showNotice(n *notice) {
	shown := *n
	if n.count > 1 {
		shown.body = i18n.T("notice.more", n.count, n.body)
	}
	if a.notify != nil {
		// A test reads what would be shown rather than the desktop
		// showing it.
		a.notify(&shown)
		return
	}
	note := mygo.NewNotification(mygo.NotificationOptions{
		ID:    noticeID(shown.chatID, shown.chatType),
		Title: shown.title,
		Body:  shown.body,
		Group: shown.chatID,
	})
	// A desktop that refuses them is a desktop the window still shows the
	// count on: nothing to do about it here.
	_ = note.Show()
}

// chatTitle is what a chat is called, for a notification: its name in the
// list when the list has it, and else who spoke, as the list may not be
// loaded yet.
func (a *App) chatTitle(m *yh.Message) string {
	for i := range a.convs {
		if a.convs[i].ChatID == m.ChatID {
			return a.convs[i].Title()
		}
	}
	if m.Sender.ChatType == yh.ChatUser && m.Sender.Name != "" {
		return m.Sender.Name
	}
	if m.ChatID != "" {
		return m.ChatID
	}
	return i18n.T("app.name")
}

// onNoticeClick opens the chat a notification was about, and brings the
// window forward: a notification is a way into the chat it is about.
func (a *App) onNoticeClick(id string) {
	chatID, chatType, ok := noticeChat(id)
	if !ok {
		return
	}
	a.update(func() {
		a.openChat(chatID, chatType)
		if a.win != nil {
			a.win.Show()
			a.win.Focus()
		}
	})
}

// openChat opens a chat by what a notification knows of it: the list may
// not have it yet, as a notification can be pressed before the list loads.
func (a *App) openChat(chatID string, chatType int) {
	if chatID == "" || a.current == chatID {
		return
	}
	conv := yh.Conversation{ChatID: chatID, ChatType: chatType}
	for i := range a.convs {
		if a.convs[i].ChatID == chatID {
			conv = a.convs[i]
			break
		}
	}
	a.open(conv)
}

// noticeID is the notification of a chat, which says which chat it is
// about when it is pressed.
func noticeID(chatID string, chatType int) string {
	return "yh:" + chatID + ":" + strconv.Itoa(chatType)
}

// noticeChat reads what a notification's id says about the chat it is
// about.
func noticeChat(id string) (chatID string, chatType int, ok bool) {
	rest, found := strings.CutPrefix(id, "yh:")
	if !found {
		return "", 0, false
	}
	i := strings.LastIndexByte(rest, ':')
	if i <= 0 {
		return "", 0, false
	}
	t, err := strconv.Atoi(rest[i+1:])
	if err != nil {
		return "", 0, false
	}
	return rest[:i], t, true
}
