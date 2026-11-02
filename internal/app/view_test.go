package app

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// TestLoginView builds the sign-in screen without a window. The words are
// asked of the dictionary rather than written here, so that renaming the
// app does not fail the test.
func TestLoginView(t *testing.T) {
	i18n.Set("zh-CN")
	defer i18n.Set("zh-CN")

	a := New()
	tt := ui.NewTester(a.view, 900, 640)
	for _, want := range []string{i18n.T("app.name"), "邮箱", "密码", "登录"} {
		if !tt.HasText(want) {
			t.Errorf("sign-in screen lacks %q; texts: %q", want, tt.Texts())
		}
	}
	if err := tt.Click(i18n.T("login.tabToken")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("登录 Token") {
		t.Errorf("token mode did not show; texts: %q", tt.Texts())
	}

	// And the phone way: the number, the picture to read out, and the
	// code texted to it.
	if err := tt.Click(i18n.T("login.tabPhone")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, want := range []string{i18n.T("login.phone"), i18n.T("login.captcha"), i18n.T("login.code")} {
		if !tt.HasText(want) {
			t.Errorf("phone mode lacks %q; texts: %q", want, tt.Texts())
		}
	}
}

// TestChatView builds the chat screen with a conversation open.
func TestChatView(t *testing.T) {
	i18n.Set("zh-CN")
	defer i18n.Set("zh-CN")

	a := New()
	a.me = &yh.User{ID: "100", Name: "我"}
	a.convs = []yh.Conversation{{
		ChatID:   "g1",
		ChatType: yh.ChatGroup,
		Name:     "mygo 交流群",
		Content:  "今天写点什么",
		SendTime: uint64(time.Now().Unix()),
		Unread:   3,
	}}
	a.current = "g1"
	a.currentType = yh.ChatGroup
	a.msgs = []*yh.Message{
		{
			MsgID: "m1", ChatID: "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: "你好，mygo湖"},
			TimestampMS: uint64(time.Now().UnixMilli()),
		},
		{
			MsgID: "m2", ChatID: "g1",
			Sender:      yh.Sender{ChatID: "100", Name: "我"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: "在写了"},
			TimestampMS: uint64(time.Now().UnixMilli()),
		},
	}
	// The panel of what is known about the chat is shown when it is asked
	// for, so it is asked for here.
	a.infoOpen = true
	tt := ui.NewTester(a.view, 1000, 720)
	for _, want := range []string{"mygo 交流群", "你好，mygo湖", "在写了", "发送", "会话信息"} {
		if !tt.HasText(want) {
			t.Errorf("chat screen lacks %q; texts: %q", want, tt.Texts())
		}
	}
}

// chatMessages makes n messages named by prefix.
func chatMessages(prefix string, n int) []*yh.Message {
	base := time.Now().Add(-time.Hour)
	out := make([]*yh.Message, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &yh.Message{
			MsgID:       fmt.Sprintf("%s-%d", prefix, i),
			ChatID:      "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: fmt.Sprintf("%s %d", prefix, i)},
			TimestampMS: uint64(base.Add(time.Duration(i) * time.Minute).UnixMilli()),
		})
	}
	return out
}

// TestChatOpensAtEnd checks that opening a chat shows its newest message.
// The chat opened first is scrolled up, so the list has stopped following
// its end: the next chat must still open at its end, which only the scroll
// asked for while the list is built can do.
func TestChatOpensAtEnd(t *testing.T) {
	a := New()
	a.me = &yh.User{ID: "100", Name: "我"}
	a.convs = []yh.Conversation{
		{ChatID: "g1", ChatType: yh.ChatGroup, Name: "群一"},
		{ChatID: "g2", ChatType: yh.ChatGroup, Name: "群二"},
	}
	a.current, a.currentType = "g1", yh.ChatGroup

	tt := ui.NewTester(a.view, 1000, 700)
	a.setMessages(chatMessages("A", 60))
	tt.Frame()
	if !a.msgList.AtEnd() {
		t.Fatalf("the first chat did not open at its end")
	}

	// Read back through the history: the list stops following its end.
	a.msgList.ScrollTo(0, ui.Start)
	tt.Frame()
	if a.msgList.AtEnd() {
		t.Fatalf("scrolling to the start left the list at its end")
	}

	// Open another chat, which loads empty and then fills.
	a.current, a.currentType = "g2", yh.ChatGroup
	a.msgs, a.hasMore = nil, false
	tt.Frame()
	a.setMessages(chatMessages("B", 60))
	tt.Frame()

	if !a.msgList.AtEnd() {
		t.Errorf("the second chat opened at its start, not its end")
	}
	if !tt.HasText("B 59") {
		t.Errorf("the newest message of the second chat is not in view; texts: %q", tt.Texts())
	}
	if tt.HasText("B 0") {
		t.Errorf("the oldest message of the second chat is in view")
	}
}

// TestInfoPanelIsNotOpenedByAChat checks that picking a conversation leaves
// the panel of what is known about it as it was: it is opened when it is
// asked for, not every time another chat is opened.
func TestInfoPanelIsNotOpenedByAChat(t *testing.T) {
	a := openChat(0)
	a.infoOpen = false
	// The chat's messages are asked for as it opens: the address is one
	// that refuses at once, so the test stays offline.
	a.client.Base = "http://127.0.0.1:1"

	a.open(yh.Conversation{ChatID: "g9", ChatType: yh.ChatGroup, Name: "另一个群"})
	if a.infoOpen {
		t.Error("opening a chat opened the info panel")
	}

	// And one that was open stays open.
	a.infoOpen = true
	a.open(yh.Conversation{ChatID: "g10", ChatType: yh.ChatGroup, Name: "又一个群"})
	if !a.infoOpen {
		t.Error("opening a chat closed the info panel")
	}
}

// TestBothSidesShareTheBubbleColour checks that a message of the account's
// own is drawn on the same gray as one from anyone else: which is whose is
// told by the side it sits on, as a chat tells it, not by a color of its
// own.
func TestBothSidesShareTheBubbleColour(t *testing.T) {
	const text = "一样的底色"
	bubble := func(mine bool) color.RGBA {
		a := openChat(0)
		sender := yh.Sender{ChatID: "200", Name: "小明"}
		if mine {
			sender = yh.Sender{ChatID: "100", Name: "阿德"}
		}
		a.msgs = []*yh.Message{{
			MsgID: "m1", ChatID: "g1", Sender: sender,
			ContentType: yh.ContentText, Content: yh.Content{Text: text},
			TimestampMS: uint64(time.Now().UnixMilli()),
		}}
		tt := dark(ui.NewTester(a.view, 1000, 700))
		r, ok := tt.Find(text)
		if !ok {
			t.Fatalf("the message is not drawn; texts: %q", tt.Texts())
		}
		// The bubble holds the text with its padding around it: a few
		// points left of the text is the bubble itself.
		return tt.Image().RGBAAt(int(r.X)-5, int(r.Y+r.H/2))
	}

	theirs, mine := bubble(false), bubble(true)
	if theirs != mine {
		t.Errorf("another's message is drawn on %v, the account's own on %v", theirs, mine)
	}
	// And both are the one bubble color of the dark appearance.
	if want := (color.RGBA{R: 0x2e, G: 0x2e, B: 0x2e, A: 0xff}); theirs != want {
		t.Errorf("the bubble is drawn on %v, want %v", theirs, want)
	}
}
