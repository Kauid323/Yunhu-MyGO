package app

import (
	"strings"
	"testing"
	"time"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// arrival is a message that came in for a chat that is not open.
func arrival(chatID, text string) *yh.Message {
	return &yh.Message{
		MsgID:       "m-" + chatID + "-" + text,
		ChatID:      chatID,
		Sender:      yh.Sender{ChatID: "200", Name: "小明", ChatType: yh.ChatGroup},
		ContentType: yh.ContentText,
		Content:     yh.Content{Text: text},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}
}

// taking is an app that the desktop tells of nothing, and that a test
// reads the notifications of instead.
func taking() (*App, chan notice) {
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德"}
	a.notices = true
	got := make(chan notice, 8)
	a.notify = func(n *notice) { got <- *n }
	return a, got
}

// TestNoticeIDNamesTheChat checks what a notification's id says about the
// chat it is about, which is what a press on it opens.
func TestNoticeIDNamesTheChat(t *testing.T) {
	id := noticeID("g1", yh.ChatGroup)
	chatID, chatType, ok := noticeChat(id)
	if !ok || chatID != "g1" || chatType != yh.ChatGroup {
		t.Fatalf("noticeChat(%q) = %q, %d, %v", id, chatID, chatType, ok)
	}

	// A chat whose id has a colon in it still reads back as itself: the
	// chat type is the last of them.
	if got, _, ok := noticeChat(noticeID("a:b", 2)); !ok || got != "a:b" {
		t.Errorf("a chat id with a colon reads back as %q, %v", got, ok)
	}

	for _, bad := range []string{"", "g1:1", "yh:", "yh:g1", "yh:g1:x"} {
		if _, _, ok := noticeChat(bad); ok {
			t.Errorf("noticeChat(%q) read a chat out of it", bad)
		}
	}
}

// TestABurstIsOneNotification checks the wait that gathers a burst: a chat
// that is being typed into quickly is one notification rather than one per
// message, and the wait starts over with each message, so it comes when
// the burst is over rather than in the middle of it.
func TestABurstIsOneNotification(t *testing.T) {
	defer i18n.Set("zh-CN")

	a, got := taking()
	a.tellArrival(arrival("g1", "在吗"))
	// A second message before the wait is over restarts it: the
	// notification must not come at the moment the first one alone would
	// have brought it.
	time.Sleep(noticeWait * 3 / 5)
	a.tellArrival(arrival("g1", "看下这个"))

	select {
	case n := <-got:
		t.Fatalf("a notification came in the middle of the burst: %+v", n)
	case <-time.After(noticeWait * 3 / 5):
	}

	select {
	case n := <-got:
		if n.count != 2 {
			t.Errorf("the notification counts %d messages, want 2", n.count)
		}
		if want := i18n.T("notice.more", 2, "看下这个"); n.body != want {
			t.Errorf("the notification reads %q, want %q", n.body, want)
		}
		if n.chatID != "g1" {
			t.Errorf("the notification is about %q, want the chat that spoke", n.chatID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no notification came for the burst")
	}
}

// TestOneMessageIsToldAsItself checks that a single message is not counted
// as a burst of one.
func TestOneMessageIsToldAsItself(t *testing.T) {
	defer i18n.Set("zh-CN")

	a, got := taking()
	a.tellArrival(arrival("g1", "在吗"))
	a.flushNotices()

	select {
	case n := <-got:
		if n.count != 1 {
			t.Errorf("the notification counts %d messages, want 1", n.count)
		}
		if n.body != "在吗" {
			t.Errorf("the notification reads %q, want the message itself", n.body)
		}
	case <-time.After(time.Second):
		t.Fatal("nothing was told of a message that came in")
	}
}

// TestTheChatThatSpokeLastIsTheOneTold checks a burst from several chats:
// they are one notification, about the chat that spoke last, rather than
// one window per chat.
func TestTheChatThatSpokeLastIsTheOneTold(t *testing.T) {
	a, got := taking()
	a.tellArrival(arrival("g1", "甲"))
	time.Sleep(2 * time.Millisecond)
	a.tellArrival(arrival("g2", "乙"))
	a.flushNotices()

	select {
	case n := <-got:
		if n.chatID != "g2" {
			t.Errorf("the notification is about %q, want the chat that spoke last", n.chatID)
		}
	case <-time.After(time.Second):
		t.Fatal("nothing was told of the burst")
	}
	// What was gathered is told once, and not again.
	a.flushNotices()
	select {
	case n := <-got:
		t.Errorf("the burst was told twice: %+v", n)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestNothingIsToldWithoutADesktop checks that a desktop which takes no
// notifications is not gathered for, and is told of nothing.
func TestNothingIsToldWithoutADesktop(t *testing.T) {
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德"}
	got := make(chan notice, 1)
	a.notify = func(n *notice) { got <- *n }

	a.tellArrival(arrival("g1", "在吗"))
	if a.pending != nil {
		t.Errorf("%d messages were gathered for a desktop that takes none", len(a.pending))
	}
	a.flushNotices()
	select {
	case n := <-got:
		t.Errorf("a desktop that takes none was told: %+v", n)
	default:
	}
}

// TestTheOpenChatIsNotToldOf checks that a message that lands in the chat
// on screen is not told of: looking at the chat is reading it. The window
// has to be behind another one, or down in the taskbar, for the desktop to
// be told, and a window that is not there at all is not away.
func TestTheOpenChatIsNotToldOf(t *testing.T) {
	a, got := taking()
	a.current, a.currentType = "g1", yh.ChatGroup

	a.onEvent(yh.Event{Cmd: "push_message", Message: arrival("g1", "在吗")})
	if len(a.msgs) != 1 {
		t.Fatalf("the message is not in the chat: %d messages", len(a.msgs))
	}
	select {
	case n := <-got:
		t.Errorf("a message on screen was told of: %+v", n)
	case <-time.After(50 * time.Millisecond):
	}

	// Another chat's message is told of, as the chat is not on screen.
	a.onEvent(yh.Event{Cmd: "push_message", Message: arrival("g2", "看下这个")})
	a.flushNotices()
	select {
	case n := <-got:
		if n.chatID != "g2" {
			t.Errorf("the notification is about %q, want the chat that is not on screen", n.chatID)
		}
	case <-time.After(time.Second):
		t.Fatal("a message for a chat that is not open was not told of")
	}
}

// TestANoticeOpensItsChat checks that a press on a notification opens the
// chat it is about, whether or not the list of conversations has loaded.
func TestANoticeOpensItsChat(t *testing.T) {
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德"}
	a.client.Base = "http://127.0.0.1:1"
	a.convs = []yh.Conversation{
		{ChatID: "g1", ChatType: yh.ChatGroup, Name: "群一"},
		{ChatID: "g2", ChatType: yh.ChatGroup, Name: "群二"},
	}
	a.current, a.currentType = "g1", yh.ChatGroup

	a.onNoticeClick(noticeID("g2", yh.ChatGroup))
	if a.current != "g2" {
		t.Errorf("a press opened %q, want the chat the notification was about", a.current)
	}
	if a.currentType != yh.ChatGroup {
		t.Errorf("the chat was opened as type %d", a.currentType)
	}

	// A notification that names no chat opens nothing.
	a.onNoticeClick("nonsense")
	if a.current != "g2" {
		t.Errorf("a press on nothing opened %q", a.current)
	}

	// A chat the list does not hold yet is still opened: a notification
	// can be pressed before the list loads.
	a.onNoticeClick(noticeID("g9", yh.ChatGroup))
	if a.current != "g9" {
		t.Errorf("a chat outside the list was not opened: %q", a.current)
	}
}

// TestAToldMessageIsOneLine checks what a notification shows of a message:
// a line's worth of it rather than all of it, as a notification is read at
// a glance.
func TestAToldMessageIsOneLine(t *testing.T) {
	a, got := taking()
	a.tellArrival(arrival("g1", strings.Repeat("字", 200)))
	a.flushNotices()

	select {
	case n := <-got:
		if r := []rune(n.body); len(r) > 90 {
			t.Errorf("the notification shows %d characters of the message", len(r))
		}
		if !strings.HasSuffix(n.body, "…") {
			t.Errorf("a long message is told as %q, want it cut with an ellipsis", n.body)
		}
	case <-time.After(time.Second):
		t.Fatal("nothing was told of the message")
	}
}
