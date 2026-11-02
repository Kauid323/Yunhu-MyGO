package app

import (
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// TestChatOpensAtTheFirstUnread checks where a chat with messages waiting
// opens: at the first of them, not at the newest one, and with the line
// across the history over it.
func TestChatOpensAtTheFirstUnread(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.unreadOnOpen = 40
	a.setMessages(chatMessages("A", 60))
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()

	if got := len(a.msgs) - a.unreadFrom; got != 40 {
		t.Fatalf("the chat holds %d unread, want 40", got)
	}
	first, _ := a.msgList.Visible()
	if first > a.unreadFrom {
		t.Errorf("the chat opened at row %d, past the first unread at %d", first, a.unreadFrom)
	}
	if !tt.HasText(i18n.T("chat.newLine")) {
		t.Errorf("no line marks the unread messages; texts: %q", tt.Texts())
	}
}

// TestUnreadCountsDownToTheEnd checks the button over the list: it counts
// what is still below the view, counts off as the reader goes down through
// it, and is gone once the newest message is in view.
func TestUnreadCountsDownToTheEnd(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.unreadOnOpen = 40
	a.setMessages(chatMessages("A", 60))
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()

	label := func(n int) string { return i18n.T("chat.newMessages", i18n.Num(n)) }
	if got := a.unreadBelow(); got != 40 {
		t.Fatalf("the button counts %d, want 40", got)
	}
	if !tt.HasText(label(40)) {
		t.Errorf("the button is not drawn; texts: %q", tt.Texts())
	}

	// Going down through them counts them off.
	tt.Scroll(520, 400, 0, 300)
	tt.Frame()
	if got := a.unreadBelow(); got >= 40 {
		t.Errorf("scrolling down left %d of 40 unread", got)
	}

	// At the end, nothing is below: the button is gone.
	tt.Scroll(520, 400, 0, 100000)
	tt.Frame()
	if got := a.unreadBelow(); got != 0 {
		t.Errorf("%d unread are counted at the end of the list", got)
	}
	if tt.HasText(label(1)) || tt.HasText(label(40)) {
		t.Errorf("the button is still drawn at the end; texts: %q", tt.Texts())
	}
}

// TestButtonGoesToTheUnread checks that pressing the button scrolls back to
// the first message that has not been read.
func TestButtonGoesToTheUnread(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.unreadOnOpen = 40
	a.setMessages(chatMessages("A", 60))
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()

	// Back to the top of the history, where the unread ones are far below.
	tt.Scroll(520, 400, 0, -100000)
	tt.Frame()
	if first, _ := a.msgList.Visible(); first != 0 {
		t.Fatalf("the list is at row %d, not at its top", first)
	}

	if err := tt.Click(i18n.T("chat.newMessages", i18n.Num(40))); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	first, _ := a.msgList.Visible()
	if first > a.unreadFrom {
		t.Errorf("the button left the list at row %d, above nothing unread at %d", first, a.unreadFrom)
	}
}

// TestStretchesOfTalking checks where the history writes a time: at the
// first message, after a silence, and on another day.
func TestStretchesOfTalking(t *testing.T) {
	a := openChat(0)
	now := time.Now()
	at := func(d time.Duration) *yh.Message {
		return &yh.Message{ContentType: yh.ContentText, TimestampMS: uint64(now.Add(d).UnixMilli())}
	}
	a.msgs = []*yh.Message{
		at(-2 * time.Hour),     // 0: the first of the history
		at(-119 * time.Minute), // 1: a minute later, the same stretch
		at(-60 * time.Minute),  // 2: an hour later, a stretch of its own
		at(-59 * time.Minute),  // 3: the same stretch
		at(-58 * time.Minute),  // 4: the same stretch
		{},                     // 5: no time at all
		at(-30 * time.Minute),  // 6: a stretch of its own
	}
	want := []bool{true, false, true, false, false, false, true}
	for i, w := range want {
		if got := a.startsStretch(i); got != w {
			t.Errorf("startsStretch(%d) = %v, want %v", i, got, w)
		}
	}

	// A message of the day before opens a stretch whatever the silence.
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	a.msgs = []*yh.Message{
		{TimestampMS: uint64(midnight.Add(-10 * time.Minute).UnixMilli())},
		{TimestampMS: uint64(midnight.Add(time.Minute).UnixMilli())},
	}
	if !a.startsStretch(1) {
		t.Error("a message sent after midnight did not open a stretch")
	}
}

// TestWhenLine checks the time written between stretches: the hour for
// today, the day for the days before, the date after that.
func TestWhenLine(t *testing.T) {
	defer i18n.Set("zh-CN")

	now := time.Now()
	hm := now.Format(i18n.T("time.clock"))
	if got := whenLine(now); got != hm {
		t.Errorf("today reads %q, want %q", got, hm)
	}

	// The day before, by the calendar rather than by hours elapsed.
	yesterday := now.AddDate(0, 0, -1)
	if got, want := whenLine(yesterday), i18n.T("time.stamp.yesterday", yesterday.Format(i18n.T("time.clock"))); got != want {
		t.Errorf("yesterday reads %q, want %q", got, want)
	}

	// Three days back is a weekday, and a month back a date.
	three := now.AddDate(0, 0, -3)
	if got := whenLine(three); got != i18n.T("time.stamp.weekday", i18n.T("time.weekday."+weekday(three.Weekday())), three.Format(i18n.T("time.clock"))) {
		t.Errorf("three days back reads %q", got)
	}
	old := now.AddDate(0, -1, 0)
	if got := whenLine(old); got != i18n.T("time.stamp.date", old.Format(i18n.T("time.monthDay")), old.Format(i18n.T("time.clock"))) {
		t.Errorf("a month back reads %q", got)
	}

	// A message the server gave no time for has none.
	if got := whenLine(time.Time{}); got != "" {
		t.Errorf("a message without a time reads %q", got)
	}
}

// TestArrivalsAreReadAtTheEnd checks that a message which lands while the
// reader is at the end of the list is read as it lands, and that one which
// lands while they are back in the history opens a stretch of unread ones
// at itself.
func TestArrivalsAreReadAtTheEnd(t *testing.T) {
	a := openChat(30)
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if !a.msgList.AtEnd() {
		t.Fatal("the chat did not open at its end")
	}
	a.unreadFrom = len(a.msgs)

	a.addMessage(&yh.Message{MsgID: "n1", ContentType: yh.ContentText,
		Content: yh.Content{Text: "在看着"}, TimestampMS: uint64(time.Now().UnixMilli())})
	if a.unread() {
		t.Error("a message that landed at the end opened a stretch of unread ones")
	}

	// Back in the history, the next one waits.
	tt.Scroll(520, 400, 0, -600)
	tt.Frame()
	a.addMessage(&yh.Message{MsgID: "n2", ContentType: yh.ContentText,
		Content: yh.Content{Text: "走开了"}, TimestampMS: uint64(time.Now().UnixMilli())})
	if !a.unread() {
		t.Fatal("a message that landed while the reader was away is not unread")
	}
	if a.unreadFrom != len(a.msgs)-1 {
		t.Errorf("the unread stretch starts at %d, want the message that landed", a.unreadFrom)
	}
	if got := a.unreadBelow(); got != 1 {
		t.Errorf("the button counts %d, want the one that landed", got)
	}
}

// TestReadNeverComesBack checks that the count only goes down: once the
// reader has been through the messages that came in while they were away,
// going back up through the history does not count them again — coming
// into the chat at all is reading what is there.
func TestReadNeverComesBack(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.client.Base = "http://127.0.0.1:1"
	a.hasMore = false
	a.unreadOnOpen = 40
	a.setMessages(chatMessages("A", 60))
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if got := a.unreadBelow(); got != 40 {
		t.Fatalf("the button counts %d, want 40", got)
	}

	// Down to the end: everything has been looked at.
	tt.Scroll(520, 400, 0, 100000)
	tt.Frame()
	if got := a.unreadBelow(); got != 0 {
		t.Fatalf("%d unread are counted at the end of the list", got)
	}

	// And back up through the whole history: none of it is unread again.
	tt.Scroll(520, 400, 0, -100000)
	tt.Frame()
	if got := a.unreadBelow(); got != 0 {
		t.Errorf("going back up counted %d unread again", got)
	}
	if tt.HasText(i18n.T("chat.newMessages", i18n.Num(1))) {
		t.Errorf("the button came back; texts: %q", tt.Texts())
	}
}

// TestPictureOpensInAWindow checks that a press on a picture in the
// history asks for a window of its own, at the size the picture was sent
// in. The window itself needs a desktop, so what is checked here is the
// state the press leaves.
func TestPictureOpensInAWindow(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.msgs = []*yh.Message{{
		MsgID: "p1", ChatID: "g1",
		Sender:      yh.Sender{ChatID: "200", Name: "小明"},
		ContentType: yh.ContentImage,
		Content:     yh.Content{ImageURL: picURL, Text: "截图看一下", Width: 960, Height: 600},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}}
	seedPicture(t, a, picURL)
	// A picture of 960×600 is asked for at a quarter of each side, which
	// is what the history draws it from.
	seed(t, a, picURL, 240, 150)
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()

	// The picture sits over its caption.
	r, ok := tt.Find("截图看一下")
	if !ok {
		t.Fatalf("the caption is not drawn; texts: %q", tt.Texts())
	}
	tt.ClickAt(r.X+r.W/2, r.Y-20)

	if a.viewer == nil {
		t.Fatal("a press on the picture opened nothing")
	}
	if a.viewer.url != picURL {
		t.Errorf("the window shows %q, want the picture pressed", a.viewer.url)
	}
	if a.viewer.w != 960 || a.viewer.h != 600 {
		t.Errorf("the picture is asked for at %dx%d, want the size it was sent in", a.viewer.w, a.viewer.h)
	}
}

// TestViewerSize checks the window a picture opens in: its own shape,
// within what fits on a screen, and no smaller than a window worth looking
// through.
func TestViewerSize(t *testing.T) {
	cases := []struct{ w, h, wantW, wantH int }{
		{0, 0, 720, 540},        // a picture that did not say how big it is
		{960, 600, 960, 600},    // one that fits as it is
		{4000, 3000, 1093, 820}, // one too large for the screen
		{100, 100, 320, 240},    // one too small to open a window at
	}
	for _, c := range cases {
		if gotW, gotH := viewerSize(c.w, c.h); gotW != c.wantW || gotH != c.wantH {
			t.Errorf("viewerSize(%d, %d) = %dx%d, want %dx%d", c.w, c.h, gotW, gotH, c.wantW, c.wantH)
		}
	}
}

// TestViewerFrameShowsThePicture checks the window's own drawing: the
// picture at the largest size the window holds, on the dark it is made of.
func TestViewerFrameShowsThePicture(t *testing.T) {
	a := openChat(0)
	v := &viewer{url: picURL, w: 480, h: 300}
	seed(t, a, picURL, 480, 300)

	tt := ui.NewTester(func(c *ui.Context) { a.viewerFrame(c, v) }, 600, 400)
	// The first frame has no bounds to fit the picture into; the second
	// is the window as it stays.
	tt.Frame()
	tt.Frame()

	// Fitted in a 600×400 window, a 480×300 picture takes 600×375 of it.
	if n := purplePixels(tt.Image(), image.Rect(0, 0, 600, 400)); n < 200000 {
		t.Errorf("only %d pixels of the picture were drawn", n)
	}
	// The window is dark in both appearances.
	if c := tt.Image().RGBAAt(10, 10); c.R > 40 || c.G > 40 || c.B > 40 {
		t.Errorf("the window is not dark: %v", c)
	}
}

// TestStickerURL checks how a sticker's picture is named: a whole address,
// or the path of one on the sticker host.
func TestStickerURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"expression/a.png", "https://chat-img.jwznb.com/expression/a.png"},
		{"/expression/a.png", "https://chat-img.jwznb.com/expression/a.png"},
		{"https://chat-img.jwznb.com/a.png", "https://chat-img.jwznb.com/a.png"},
		{"//chat-img.jwznb.com/a.png", "https://chat-img.jwznb.com/a.png"},
	}
	for _, c := range cases {
		if got := stickerURL(&yh.Message{Content: yh.Content{Image: c.in}}); got != c.want {
			t.Errorf("stickerURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestStickerIsDrawnAsAPicture checks that a sticker is drawn, and not read
// out as the word for one: a chat shows the picture it was sent.
func TestStickerIsDrawnAsAPicture(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := openChat(0)
	a.msgs = []*yh.Message{{
		MsgID: "s1", ChatID: "g1",
		Sender:      yh.Sender{ChatID: "200", Name: "小明"},
		ContentType: yh.ContentSticker,
		Content:     yh.Content{Image: stickerPath},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}}
	seedSticker(t, a, stickerPath)

	tt := dark(ui.NewTester(a.view, 1000, 700))
	if tt.HasText(i18n.T("msg.sticker")) {
		t.Errorf("the sticker is shown as words; texts: %q", tt.Texts())
	}
	// The picture is the one seeded: its pixels are in the message area,
	// which is right of the conversation list and below the header.
	chat := image.Rect(300, 60, 1000, 640)
	if n := purplePixels(tt.Image(), chat); n < 2000 {
		t.Errorf("only %d pixels of the sticker were drawn", n)
	}
}

// purplePixels counts the pixels of r that came from a picture seeded by
// seed, which is a gradient of violet.
func purplePixels(img *image.RGBA, r image.Rectangle) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.B > 120 && c.B < 200 && c.R < 200 && c.G < 160 {
				n++
			}
		}
	}
	return n
}
