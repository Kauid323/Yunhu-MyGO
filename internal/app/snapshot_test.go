package app

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// TestSnapshot writes the screens as PNGs, to look at them.
func TestSnapshot(t *testing.T) {
	dir := filepath.Join("..", "..", "shots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The settings are read from the user's own directory as the app is
	// made, which a test must neither read nor write.
	t.Setenv("APPDATA", t.TempDir())

	// The two appearances of the sign-in screen, and of the chat.
	login := New()
	writePNG(t, filepath.Join(dir, "login.png"), ui.NewTester(login.view, 900, 620))
	writePNG(t, filepath.Join(dir, "login-dark.png"), dark(ui.NewTester(login.view, 900, 620)))

	// The phone way in, with a picture to read out and a code texted to
	// the number, the way it looks once a code has been asked for.
	phone := New()
	phone.loginMode = modePhone
	phone.phone = "13800000000"
	phone.captchaID = "snapshot"
	phone.captcha = fakeCaptcha(t)
	phone.smsAt = time.Now()
	writePNG(t, filepath.Join(dir, "login-phone.png"), ui.NewTester(phone.view, 900, 620))

	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德", AvatarURL: ""}
	a.live = yh.StateOnline
	a.convs = []yh.Conversation{
		{ChatID: "g1", ChatType: yh.ChatGroup, Name: "mygo 交流群", Content: "小明：今天写点什么", SendTime: uint64(time.Now().Unix()), Unread: 3},
		{ChatID: "u2", ChatType: yh.ChatUser, Name: "小明", Content: "晚上一起吃饭吗", SendTime: uint64(time.Now().Add(-time.Hour).Unix())},
		{ChatID: "b3", ChatType: yh.ChatBot, Name: "云湖助手", Content: "[图片]", SendTime: uint64(time.Now().Add(-26 * time.Hour).Unix()), At: true},
	}
	a.current = "g1"
	a.currentType = yh.ChatGroup
	a.note = line{}
	now := time.Now()
	a.msgs = []*yh.Message{
		{MsgID: "m1", ChatID: "g1", Sender: yh.Sender{ChatID: "200", Name: "小明"}, ContentType: yh.ContentText,
			Content: yh.Content{Text: "这个 mygo湖 的界面是纯 Go 画的？"}, TimestampMS: uint64(now.Add(-2 * time.Hour).UnixMilli())},
		{MsgID: "m2", ChatID: "g1", Sender: yh.Sender{ChatID: "100", Name: "阿德"}, ContentType: yh.ContentText,
			Content: yh.Content{Text: "对，用 mygo 的 native UI，没有 WebView，也没有前端构建。"}, TimestampMS: uint64(now.Add(-8 * time.Minute).UnixMilli())},
		{MsgID: "m3", ChatID: "g1", Sender: yh.Sender{ChatID: "300", Name: "小七"}, ContentType: yh.ContentText,
			Content: yh.Content{Text: "消息走的是云湖 API，protobuf 也是手写的编解码，一个二进制就够了。"}, TimestampMS: uint64(now.Add(-7 * time.Minute).UnixMilli())},
		{MsgID: "m4", ChatID: "g1", Sender: yh.Sender{ChatID: "100", Name: "阿德"}, ContentType: yh.ContentText,
			Content: yh.Content{Text: "试试发条消息 👋"}, TimestampMS: uint64(now.Add(-2 * time.Minute).UnixMilli())},
		{MsgID: "m5", ChatID: "g1", Sender: yh.Sender{ChatID: "200", Name: "小明"}, ContentType: yh.ContentImage,
			Content: yh.Content{ImageURL: picURL, Text: "截图看一下"}, TimestampMS: uint64(now.Add(-time.Minute).UnixMilli())},
		{MsgID: "m6", ChatID: "g1", Sender: yh.Sender{ChatID: "300", Name: "小七"}, ContentType: yh.ContentSticker,
			Content: yh.Content{Image: stickerPath}, TimestampMS: uint64(now.Add(-45 * time.Second).UnixMilli())},
		{MsgID: "m7", ChatID: "g1", Sender: yh.Sender{ChatID: "100", Name: "阿德"}, ContentType: yh.ContentMarkdown,
			Content: yh.Content{Text: mdSample}, TimestampMS: uint64(now.Add(-30 * time.Second).UnixMilli())},
	}
	a.input = ""
	seedPicture(t, a, picURL)
	seedSticker(t, a, stickerPath)
	writePNG(t, filepath.Join(dir, "chat.png"), ui.NewTester(a.view, 1000, 700))
	writePNG(t, filepath.Join(dir, "chat-dark.png"), dark(ui.NewTester(a.view, 1000, 700)))

	// The same chat with the newest two messages unread: the line across
	// the history, and the button over it once the reader goes back.
	a.unreadFrom = len(a.msgs) - 2
	writePNG(t, filepath.Join(dir, "chat-new.png"), dark(ui.NewTester(a.view, 1000, 700)))

	back := dark(ui.NewTester(a.view, 1000, 700))
	back.Frame()
	back.Scroll(520, 400, 0, -320)
	back.Frame()
	writePNG(t, filepath.Join(dir, "chat-unread.png"), back)
	a.unreadFrom = len(a.msgs)

	// The same screen with the panel of what is known about the chat open.
	a.infoOpen = true
	writePNG(t, filepath.Join(dir, "chat-info.png"), dark(ui.NewTester(a.view, 1000, 700)))
	a.infoOpen = false

	// The menu at the foot of the rail, open and arrived: the frame it
	// opens in has it still coming.
	a.more.open = true
	menu := dark(ui.NewTester(a.view, 1000, 700))
	menu.Advance(300 * time.Millisecond)
	writePNG(t, filepath.Join(dir, "menu.png"), menu)
	a.more.open = false

	// And the screen in another language, which the desktop can ask for.
	tt := ui.NewTester(a.view, 1000, 700)
	tt.SetLocale("en-US")
	writePNG(t, filepath.Join(dir, "chat-en.png"), tt)

	// A picture opened in a window of its own, which needs no window to
	// be drawn. The viewer is given the size its window is made in, which
	// is what its first frame is laid out with: openViewer does the same.
	seed(t, a, picURL, 480, 300)
	shot := &viewer{url: picURL, w: 480, h: 300, winW: 720, winH: 480, zoom: 1}
	writePNG(t, filepath.Join(dir, "viewer.png"),
		ui.NewTester(func(c *ui.Context) { a.viewerFrame(c, shot) }, 720, 480))

	// The settings window, in both appearances. It is a window of its own,
	// with the window's buttons drawn over its top right corner: the room
	// they take is given to the tester, as the desktop gives it.
	set := &settings{}
	settingsView := func(c *ui.Context) { a.settingsFrame(c, set) }
	for _, want := range []struct {
		name string
		dark bool
	}{
		{"settings.png", true},
		{"settings-light.png", false},
	} {
		tt := ui.NewTester(settingsView, settingsW, settingsH)
		tt.SetTitleBar(ui.TitleBar{Height: 32, Right: 138})
		tt.SetDark(want.dark)
		writePNG(t, filepath.Join(dir, want.name), tt)
	}

	// And with a drop-down open, which covers the row it belongs to.
	set.openDbl = true
	opened := dark(ui.NewTester(settingsView, settingsW, settingsH))
	opened.SetTitleBar(ui.TitleBar{Height: 32, Right: 138})
	opened.Advance(300 * time.Millisecond)
	writePNG(t, filepath.Join(dir, "settings-select.png"), opened)
	set.openDbl = false

	// Every page of the settings window, one each, named after its section.
	for i, name := range []string{
		"general", "notices", "storage", "hotkeys", "permissions",
		"signIn", "binding", "palette", "about",
	} {
		page := &settings{section: i}
		view := func(c *ui.Context) { a.settingsFrame(c, page) }
		tt := dark(ui.NewTester(view, settingsW, settingsH))
		tt.SetTitleBar(ui.TitleBar{Height: 32, Right: 138})
		writePNG(t, filepath.Join(dir, "settings-"+name+".png"), tt)
	}

	// A message that answers another, which shows what it answers above what
	// it says, and the card of a user beside the face that said it.
	a.msgs = append(a.msgs, &yh.Message{
		MsgID: "m8", ChatID: "g1", Sender: yh.Sender{ChatID: "200", Name: "小明"},
		ContentType: yh.ContentText,
		Content: yh.Content{
			Text:      "（蓝圆圈）这得问狗吧本人啦~ 不过真要下第三方客户端，记得看清楚权限再授权哦。",
			QuoteText: "yyyyt: @那狗吧 不会又是第三方客户端吧？",
		},
		TimestampMS: uint64(now.Add(-10 * time.Second).UnixMilli()),
	})
	a.cardOpen, a.cardID = true, "200"
	a.cardName, a.cardAvatar = "小明", ""
	a.cardAt = ui.Rect{X: 640, Y: 90, W: 32, H: 32}
	a.jumpToEnd = true
	a.card = &yh.Profile{
		ID: "200", Name: "小明", Remark: "虚构的", Introduction: "啊煤呐~啊煤呐~",
		Gender: 2, City: "北京城区",
	}
	writePNG(t, filepath.Join(dir, "chat-quote.png"), dark(ui.NewTester(a.view, 1000, 700)))
	a.cardOpen, a.card = false, nil
	a.msgs = a.msgs[:len(a.msgs)-1]

	// The panel of saved stickers over the composer, with a few of them
	// already fetched.
	a.stickers = []yh.Expression{
		{ID: "1", URL: "expression/one.png"},
		{ID: "2", URL: "expression/two.png"},
		{ID: "3", URL: "expression/three.png"},
	}
	for _, ex := range a.stickers {
		seed(t, a, stickerURLFor(ex.URL), stickerPx, stickerPx)
	}
	a.stickerOpen = true
	stickers := dark(ui.NewTester(a.view, 1000, 700))
	stickers.Advance(300 * time.Millisecond)
	writePNG(t, filepath.Join(dir, "chat-stickers.png"), stickers)
	a.stickerOpen = false
	a.stickers = nil

	// The classic layout, which is the same window laid narrow: the account
	// across its top, the search under it, and the chats beside the strip
	// of sections. A chat is not in it — it opens in a window of its own.
	classic := New()
	classic.me = a.me
	classic.convs = a.convs
	classic.prefs.Mode = modeClassic
	classic.note = line{}
	writePNG(t, filepath.Join(dir, "classic.png"),
		dark(ui.NewTester(classic.view, classicW, classicH)))

	// The window a chat is opened in, which is the chat pane of the wide
	// layout in a window of its own.
	writePNG(t, filepath.Join(dir, "chat-window.png"),
		dark(ui.NewTester(a.chatWinFrame, chatWinW, chatWinH)))
}

// dark is a tester showing the window in the dark appearance, as a desktop
// set to dark draws it.
func dark(tt *ui.Tester) *ui.Tester {
	tt.SetDark(true)
	return tt
}

// picURL is the picture the snapshot shows, served from the cache rather
// than from the network.
const picURL = "https://chat-img.jwznb.com/snapshot.png"

// stickerPath is a sticker as a message names it: the path of the picture
// on the sticker host, not a whole address.
const stickerPath = "expression/snapshot.png"

// mdSample is a message written in markdown, as the bots send them.
const mdSample = "### 今日更新\n" +
	"这次把 **消息列表** 和 *输入框* 都改了，`mygo湖` 还是纯 Go 的。\n" +
	"- **消息列表**：滚动到 80% 自动加载\n" +
	"- **输入框**：改成平铺\n" +
	"- ~~旧的卡片样式~~\n" +
	"> 位置不变才是关键\n" +
	"详见 [更新日志](https://yhchat.com/log)。"

// seedPicture puts a picture in the cache, so a snapshot draws it without
// asking the network for it.
func seedPicture(t *testing.T, a *App, url string) {
	t.Helper()
	seed(t, a, url, pictureBoxPx, pictureBoxPx)
}

// seedSticker puts a sticker's picture in the cache, as the message names
// it rather than as it is fetched.
func seedSticker(t *testing.T, a *App, path string) {
	t.Helper()
	seed(t, a, stickerURL(&yh.Message{Content: yh.Content{Image: path}}), stickerPx, stickerPx)
}

// seed puts a picture of w×h pixels in the cache under the key it is asked
// for at that size.
func seed(t *testing.T, a *App, url string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(60 + x/4), G: uint8(40 + y/3), B: 150, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	bmp, err := ui.DecodeBitmap(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	a.images.keep(fmt.Sprintf("%s@%dx%d", url, w, h), bmp)
}

func writePNG(t *testing.T, path string, tt *ui.Tester) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

// fakeCaptcha is a picture to read out, as the API sends one: four strokes
// where letters would be, in the 180×40 the API draws them in. It is only
// looked at here, so any picture does.
func fakeCaptcha(t *testing.T) *ui.Bitmap {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 180, 40))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	for _, x0 := range []int{16, 56, 96, 136} {
		for y := 8; y < 32; y++ {
			img.Set(x0+y%7, y, color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff})
			img.Set(x0+6+y%7, y, color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	bmp, err := ui.DecodeBitmap(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return bmp
}
