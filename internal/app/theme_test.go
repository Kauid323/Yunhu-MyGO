package app

import (
	"image/color"
	"regexp"
	"testing"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// TestWindowFollowsTheAppearance checks that the window takes the skin of
// the desktop rather than one of its own: #222222 in the dark appearance
// and #f5f5f5 in the light one, as the background the chat is drawn on.
func TestWindowFollowsTheAppearance(t *testing.T) {
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德"}
	tt := ui.NewTester(a.view, 900, 600)

	for _, c := range []struct {
		dark bool
		want color.RGBA
	}{
		{true, color.RGBA{0x22, 0x22, 0x22, 0xff}},
		{false, color.RGBA{0xf5, 0xf5, 0xf5, 0xff}},
	} {
		tt.SetDark(c.dark)
		tt.Frame()
		if got := tt.Image().RGBAAt(880, 200); got != c.want {
			t.Errorf("dark=%v: the chat's background is %v, want %v", c.dark, got, c.want)
		}
	}
}

// TestTheTwoSkinsAreToldApart checks that the dark and the light skin are
// not the same palette, which would leave one of the appearances drawn in
// the other's colors.
func TestTheTwoSkinsAreToldApart(t *testing.T) {
	if darkSkin.chatBg == lightSkin.chatBg {
		t.Error("both appearances are drawn on the same background")
	}
	if darkSkin.text == lightSkin.text {
		t.Error("both appearances write in the same color")
	}
	if !darkSkin.dark || lightSkin.dark {
		t.Error("a skin is marked for the wrong appearance")
	}
	if theme(true).Dark != true || theme(false).Dark != false {
		t.Error("the mygo theme does not follow the skin")
	}
}

// TestWindowFollowsTheLanguage checks that the window's own words are in
// the language of the desktop, and that they change when it does, without
// the window being made again.
func TestWindowFollowsTheLanguage(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := New()
	tt := ui.NewTester(a.view, 900, 640)

	tt.SetLocale("en-US")
	for _, want := range []string{"Sign in", "Email", "Password"} {
		if !tt.HasText(want) {
			t.Fatalf("an English desktop does not show %q; texts: %q", want, tt.Texts())
		}
	}
	if tt.HasText("邮箱") {
		t.Errorf("an English desktop still shows the Chinese words; texts: %q", tt.Texts())
	}

	tt.SetLocale("zh-CN")
	if !tt.HasText("邮箱") {
		t.Errorf("a Chinese desktop does not show its own words; texts: %q", tt.Texts())
	}
}

// keyLike is a string shaped like a key of the dictionary.
var keyLike = regexp.MustCompile(`^[a-z][a-zA-Z]*\.[a-zA-Z][a-zA-Z.]*$`)

// TestNoKeyIsShown checks that every key the window asks for has an entry:
// a key with none is drawn as it is, which shows a mistake rather than a
// word.
func TestNoKeyIsShown(t *testing.T) {
	defer i18n.Set("zh-CN")

	// The sign-in screen, with what the last attempt had to say on it.
	login := New()
	login.errText = say("login.expired")
	noKeys(t, "the sign-in screen", ui.NewTester(login.view, 900, 640))

	// The chat screen, with everything on it: the list, the panel of what
	// is known about the chat, the menu, and the note under the list.
	a := openChat(4)
	a.live = yh.StateOnline
	a.convs = []yh.Conversation{
		{ChatID: "g1", ChatType: yh.ChatGroup, Name: "群", Content: "在吗", Unread: 2},
		{ChatID: "u2", ChatType: yh.ChatUser, Name: "小明", Content: "在吗", At: true},
		{ChatID: "b3", ChatType: yh.ChatBot, Name: "助手", DoNotDisturb: true},
	}
	a.infoOpen = true
	a.more.open = true
	a.note = say("err.network")
	noKeys(t, "the chat screen", ui.NewTester(a.view, 1000, 700))
}

// noKeys checks that what a screen draws is words, in every language the
// window speaks.
func noKeys(t *testing.T, what string, tt *ui.Tester) {
	t.Helper()
	for _, lang := range []string{"zh-CN", "en-US"} {
		tt.SetLocale(lang)
		for _, s := range tt.Texts() {
			if keyLike.MatchString(s) {
				t.Errorf("%s in %s: the key %q has no entry", what, lang, s)
			}
		}
	}
}
