package app

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// chatApp is an app with a conversation list and a chat open, and no window:
// what the two layouts are built from.
func chatApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德", AvatarURL: ""}
	a.convs = []yh.Conversation{
		{ChatID: "g1", ChatType: yh.ChatGroup, Name: "mygo 交流群", Content: "今天写点什么", Unread: 3},
		{ChatID: "u2", ChatType: yh.ChatUser, Name: "小明", Content: "晚上一起吃饭吗"},
	}
	a.current, a.currentType = "g1", yh.ChatGroup
	a.msgs = []*yh.Message{{
		MsgID: "m1", ChatID: "g1", Sender: yh.Sender{ChatID: "200", Name: "小明"},
		ContentType: yh.ContentText, Content: yh.Content{Text: "在写了"},
	}}
	return a
}

// TestClassicLayout checks that the layout the settings pick is the one the
// window is laid out in: the narrow panel has no chat in it, and the wide one
// has.
func TestClassicLayout(t *testing.T) {
	i18n.Set("zh-CN")
	defer i18n.Set("zh-CN")

	a := chatApp(t)
	a.prefs.Mode = modeClassic
	tt := ui.NewTester(a.view, classicW, classicH)
	if !tt.HasText("阿德") || !tt.HasText("mygo 交流群") {
		t.Fatalf("the panel lacks the account or the chats; texts: %q", tt.Texts())
	}
	if tt.HasText(i18n.T("chat.send")) {
		t.Errorf("the panel shows a chat, which the classic layout has no room for")
	}

	// The wide layout is another window, with a list of its own: the same
	// list state cannot be shown by two.
	wide := chatApp(t)
	wide.prefs.Mode = modeEfficient
	tt = ui.NewTester(wide.view, efficientW, efficientH)
	if !tt.HasText(i18n.T("chat.send")) {
		t.Errorf("the wide layout shows no chat; texts: %q", tt.Texts())
	}
}

// TestPrefsAreKept checks that a choice made in the settings is kept for the
// next run: it is written where the session is, and read back by an app that
// is made again.
func TestPrefsAreKept(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())

	a := New()
	a.pickPrefs(func(p *prefs) { p.Mode = modeClassic })
	a.pickPrefs(func(p *prefs) { p.SendKey = "ctrlEnter" })
	a.pickPrefs(func(p *prefs) { p.Slash = false })

	back := New()
	back.prefs = loadPrefs()
	if back.prefs.Mode != modeClassic {
		t.Errorf("the layout came back as %d, want the classic one", back.prefs.Mode)
	}
	if back.prefs.SendKey != "ctrlEnter" {
		t.Errorf("the send key came back as %q", back.prefs.SendKey)
	}
	if back.prefs.Slash {
		t.Errorf("the slash commands came back on")
	}
	// And what was never touched is what a window that has never been set
	// opens with.
	if back.prefs.Look != lookSystem || back.prefs.NewWindow != "merge" {
		t.Errorf("the settings that were not set changed: %+v", back.prefs)
	}
}
