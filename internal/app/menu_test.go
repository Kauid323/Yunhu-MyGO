package app

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// TestMenuOpensAndCloses checks the menu at the foot of the rail: the
// button opens it, and a press outside it or Escape closes it, which the
// popup it is built on does.
func TestMenuOpensAndCloses(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(2)
	tt := dark(ui.NewTester(a.view, 900, 600))

	if tt.HasText("退出账号") {
		t.Fatal("the menu shows before it is asked for")
	}

	// The button is the last of the rail, above its bottom padding.
	tt.ClickAt(28, 570)
	if !a.more.open {
		t.Fatal("the button did not open the menu")
	}
	tt.Advance(200 * time.Millisecond)
	for _, want := range []string{"退出账号", "聊天记录管理", "收藏", "调色盘"} {
		if !tt.HasText(want) {
			t.Errorf("the open menu lacks %q; texts: %q", want, tt.Texts())
		}
	}

	// Choosing a row closes it, whether or not the row has a feature.
	if err := tt.Click("聊天记录管理"); err != nil {
		t.Fatal(err)
	}
	if a.more.open {
		t.Error("choosing a row left the menu open")
	}

	tt.ClickAt(28, 570)
	tt.ClickAt(500, 300)
	if a.more.open {
		t.Error("a press outside the menu left it open")
	}

	tt.ClickAt(28, 570)
	if !a.more.open {
		t.Fatal("the button did not open the menu again")
	}
	tt.Key(0, ui.KeyEscape)
	if a.more.open {
		t.Error("Escape left the menu open")
	}

	// It goes the way it came: nothing of it is left over the list once
	// it has.
	tt.Advance(300 * time.Millisecond)
	if got := tt.Image().RGBAAt(150, 500); got == (color.RGBA{0x30, 0x30, 0x30, 0xff}) {
		t.Error("the menu is still drawn after it closed")
	}
}

// TestUnreadIsOnTheChatIcon checks where the count of what waits is drawn:
// on the chat section, not on the account's picture.
func TestUnreadIsOnTheChatIcon(t *testing.T) {
	// The two windows differ in one thing only, the messages waiting, so
	// that whatever red the account's picture has cancels out.
	plain := openChat(2)
	busy := openChat(2)
	busy.convs = []yh.Conversation{{ChatID: "g1", ChatType: yh.ChatGroup, Name: "群", Unread: 7}}

	quiet := ui.NewTester(plain.view, 900, 600).Image()
	loud := ui.NewTester(busy.view, 900, 600).Image()

	// The account sits at the top of the rail, the chat section under it.
	const rail = 56
	account := image.Rect(0, 0, rail, 56)
	chat := image.Rect(0, 58, rail, 98)

	if n := redPixels(loud, account) - redPixels(quiet, account); n > 4 {
		t.Errorf("the account's picture gained %d red pixels: the count is still on it", n)
	}
	if n := redPixels(loud, chat) - redPixels(quiet, chat); n < 8 {
		t.Errorf("the chat section gained only %d red pixels: it carries no count", n)
	}
}

// redPixels counts the pixels of r that are red enough to be the badge.
func redPixels(img *image.RGBA, r image.Rectangle) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R > 190 && c.G < 110 && c.B < 110 {
				n++
			}
		}
	}
	return n
}

// TestMenuArrives checks that the menu comes rather than being there the
// frame it opens: that frame has it still on its way, and a moment later
// it is where it stays. The tester's clock stands still between frames, so
// the frame the click opens in is the first of the way.
func TestMenuArrives(t *testing.T) {
	a := openChat(2)
	tt := dark(ui.NewTester(a.view, 900, 600))

	tt.ClickAt(28, 570)
	coming := tt.Image()
	tt.Advance(200 * time.Millisecond)
	arrived := tt.Image()

	// A point well inside the panel: the list behind it at first, the
	// menu's own color once it is there.
	const x, y = 150, 500
	menu := color.RGBA{0x30, 0x30, 0x30, 0xff}
	if got := coming.RGBAAt(x, y); got == menu {
		t.Errorf("the menu was drawn whole in the frame it opened: %v", got)
	}
	if got := arrived.RGBAAt(x, y); got != menu {
		t.Errorf("the menu did not arrive: %v at %d,%d", got, x, y)
	}
	if sameImage(coming, arrived) {
		t.Error("the menu appeared at once, with no entrance")
	}
}

// sameImage reports whether two frames are the same picture.
func sameImage(a, b *image.RGBA) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	return bytes.Equal(a.Pix, b.Pix)
}
