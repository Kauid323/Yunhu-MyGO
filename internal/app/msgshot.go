package app

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// A picture of messages is drawn as the chat draws it — the bubbles, the
// faces and the markdown are the ones the history shows, because they are
// the same code — but in memory rather than in a window, at as many pixels
// to a DIP as the settings ask for. ui.Render lays the card out and draws
// it at that resolution, so the words in the picture are drawn that large
// rather than drawn small and stretched.
//
// The card is drawn a slice at a time: each slice lifts the card by the
// height of the slices before it, and the slices are stacked into one
// picture. A slice is a screenful, so that no single drawing is a hundred
// megabytes.
const (
	// shotW is how wide the card is, in DIPs: wide enough for a sentence
	// to sit on a line, narrow enough to be read on a phone. It is the
	// width the card is laid out at whatever the resolution.
	shotW = 460
	// shotRoom is how much of the card one slice holds, in DIPs.
	shotRoom = 900
	// shotSlices is as many slices as one picture is made of, which is a
	// history far longer than anyone would put in a picture.
	shotSlices = 60
	// shotPad is the margin around the card.
	shotPad = 18
	// shotWaitMax is how long a slice waits for the pictures in it.
	shotWaitMax = 2500 * time.Millisecond
)

// shotScalePx is how many pixels a DIP of the card takes at each resolution
// the settings offer. A picture is made at shotDefaultScale when nothing is
// picked.
var (
	shotScalePx      = map[string]int{"x1": 1, "x2": 2, "x3": 3}
	shotDefaultScale = "x2"
)

// shotPixels is how many pixels a DIP of the card takes in a picture.
func shotPixels(p prefs) int {
	if n, ok := shotScalePx[p.ShotDPI]; ok {
		return n
	}
	return shotScalePx[shotDefaultScale]
}

// shotEmpty is the colour of the room around the card, which is cut off
// when the picture is taken: nothing the card draws is ever this.
var shotEmpty = ui.RGB(1, 0, 1)

// errShotTall is a card running past the last slice, which a history of
// thousands of messages would do.
var errShotTall = errors.New("the messages do not fit in a picture")

// shotWin is one picture being made: the messages in it, who they were
// said to, when, how finely it is drawn, and the appearance and the
// language of the window it came from.
type shotWin struct {
	msgs []*yh.Message
	chat string
	when time.Time
	px   int
	dark bool
	lang string
}

// shotPart is one slice of the picture, encoded, and how many rows and
// columns of the picture it holds.
type shotPart struct {
	data       []byte
	rows, cols int
}

// messageImage draws msgs as the chat shows them and writes the picture
// where the reader says.
func (a *App) messageImage(msgs []*yh.Message) {
	if len(msgs) == 0 || a.shotBusy {
		return
	}
	s := &shotWin{
		msgs: append([]*yh.Message(nil), msgs...),
		chat: a.currentTitle(),
		when: time.Now(),
		px:   shotPixels(a.prefs),
		// The picture takes the appearance and the language the window is
		// in: it is drawn away from the window, which has neither of its
		// own.
		dark: qq.dark,
		lang: i18n.Lang(),
	}
	name, dir := shotFile(s.chat, s.when), a.filesDir()
	// The dialog blocks until it is answered, so it is asked from a
	// goroutine — with the window and the words it needs held here, as the
	// page they come from is not to be read from there.
	parent := a.win
	title, filter := i18n.T("shot.save"), i18n.T("shot.png")
	a.shotBusy = true

	go func() {
		defer a.update(func() { a.shotBusy = false })
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Parent:      parent,
			Title:       title,
			DefaultPath: filepath.Join(dir, name),
			Filters:     []mygo.FileFilter{{Name: filter, Extensions: []string{"png"}}},
		})
		if err != nil || path == "" {
			// A dialog that was closed asks for nothing.
			return
		}
		data, err := a.shotPNG(s)
		if err != nil {
			a.update(func() { a.note = shotError(err) })
			return
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			a.update(func() { a.note = errorLine(err) })
			return
		}
		a.update(func() { a.note = say("shot.saved", filepath.Base(path)) })
	}()
}

// shotError says what went wrong with a picture in the words of the
// reader's language.
func shotError(err error) line {
	if errors.Is(err, errShotTall) {
		return say("shot.tooTall")
	}
	return errorLine(err)
}

// shotPNG is the picture itself: the card drawn a slice at a time, each
// slice taken once the pictures in it have come.
func (a *App) shotPNG(s *shotWin) ([]byte, error) {
	var parts []shotPart
	up := float32(0)
	for i := 0; i < shotSlices; i++ {
		// The first drawing of a slice is what asks for the pictures in
		// it; the one after the wait has them.
		a.shotDraw(s, up)
		a.shotSettle()
		part, used, last, err := shotSlice(a.shotDraw(s, up))
		if err != nil {
			return nil, err
		}
		if part != nil {
			parts = append(parts, *part)
		}
		if last {
			if len(parts) == 0 {
				return nil, errShotTall
			}
			return shotStack(parts)
		}
		up += used
	}
	return nil, errShotTall
}

// shotDraw draws the card lifted by up DIPs, in memory and at the
// resolution of the picture.
func (a *App) shotDraw(s *shotWin, up float32) *image.RGBA {
	var img *image.RGBA
	// The skin and the language the card is drawn in are the window's, so
	// the drawing waits for the window's thread rather than taking them
	// while a frame of the window is being drawn with them.
	a.onMain(func() {
		img = ui.Render(func(c *ui.Context) { a.shotView(c, s, up) },
			shotW, shotRoom, float32(s.px))
	})
	return img
}

// onMain runs fn on the window's thread and waits for it to finish.
func (a *App) onMain(fn func()) {
	if a.win == nil {
		fn()
		return
	}
	done := make(chan struct{})
	a.update(func() {
		defer close(done)
		fn()
	})
	<-done
}

// shotSettle gives the pictures the card asked for the moment they take to
// come: they are fetched while the card is drawn, and the drawing after
// this one has them rather than the boxes they will fill.
func (a *App) shotSettle() {
	// The asking happens as the frame is drawn, which is just before this,
	// so the first look is given a moment to see that it was asked.
	time.Sleep(80 * time.Millisecond)
	until := time.Now().Add(shotWaitMax)
	for time.Now().Before(until) {
		if !a.images.busy() {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
}

// shotView draws the card: who the words were said to, the messages as the
// history has them, and a line saying where the picture came from. up is
// how much of the card is above the slice.
func (a *App) shotView(c *ui.Context, s *shotWin, up float32) {
	// The card is drawn in the appearance and the language of the window
	// it came from, as the history is.
	qq = skin(s.dark)
	c.SetTheme(theme(s.dark))
	i18n.Set(s.lang)

	ui.Column(c).Fill().Background(shotEmpty).Children(func() {
		card := ui.Column(c).Width(shotW).Background(qq.chatBg).
			Padding(shotPad).Gap(11).Translate(0, -up)
		card.Children(func() {
			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, s.chat).FontSize(13.5).TextColor(qq.text).SingleLine()
				ui.Text(c, s.when.Format("2006/01/02 15:04")).FontSize(10).
					TextColor(qq.textFaint).SingleLine()
			})
			ui.Box(c).Height(1).Background(qq.menuLine)
			for _, m := range s.msgs {
				a.shotRow(c, m)
			}
			ui.Box(c).Height(1).Background(qq.menuLine)
			ui.Text(c, i18n.T("shot.mark")).FontSize(9.5).
				TextColor(qq.textFaint).Margin(ui.Auto).SingleLine()
		})
	})
}

// shotRow is one message in the card: the account's on the right, anyone
// else's on the left with their face and name, and a line for what is not
// a message at all, as the history draws them.
func (a *App) shotRow(c *ui.Context, m *yh.Message) {
	switch {
	case m.RecalledAt > 0:
		ui.Text(c, i18n.T("chat.recalled")).FontSize(11).
			TextColor(qq.textFaint).Margin(ui.Auto)
		return
	case m.ContentType == yh.ContentTip || m.ContentType == yh.ContentCall:
		text := m.Content.Tip
		if text == "" {
			text = preview(m.ContentType, m.Content)
		}
		ui.Text(c, text).FontSize(11).TextColor(qq.textFaint).Margin(ui.Auto)
		return
	}

	if m.Mine(a.me.ID) {
		ui.Row(c).Children(func() {
			ui.Spacer(c)
			a.bubble(c, m)
		})
		return
	}
	ui.Row(c).Gap(9).AlignItems(ui.Start).Children(func() {
		a.messageAvatar(c, m.Sender, 30)
		ui.Column(c).Gap(4).AlignItems(ui.Start).Children(func() {
			name := m.Sender.Name
			if name == "" {
				name = m.Sender.ChatID
			}
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, name).FontSize(10.5).TextColor(qq.textMuted).SingleLine()
				ui.Text(c, stampText(m.Time())).FontSize(10).
					TextColor(qq.textFaint).SingleLine()
			})
			a.bubble(c, m)
		})
	})
}

// shotSlice cuts one drawing down to the card in it and encodes it. used is
// how much of the card the slice holds, in DIPs, and last says the card
// does not go on past it — either because the card ends here or because the
// slice is already below its end.
func shotSlice(img *image.RGBA) (part *shotPart, used float32, last bool, err error) {
	b := img.Bounds()
	end, last := shotEnd(img)
	if end < 0 {
		// The slice is entirely below the card: there is nothing here.
		return nil, 0, true, nil
	}
	cut := image.NewRGBA(image.Rect(0, 0, b.Dx(), end+1))
	draw.Draw(cut, cut.Bounds(), img, image.Pt(0, 0), draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, cut); err != nil {
		return nil, 0, false, err
	}
	return &shotPart{data: buf.Bytes(), rows: end + 1, cols: b.Dx()},
		float32(end+1) * shotW / float32(b.Dx()), last, nil
}

// shotEnd is the last row of the card in a slice: the last one that is not
// the colour of the room around it. last is false only when that row is the
// slice's own last, as then the card ran past the slice and goes on.
func shotEnd(img *image.RGBA) (end int, last bool) {
	b := img.Bounds()
	row := make([]byte, b.Dx()*4)
	for i := 0; i < b.Dx(); i++ {
		row[i*4], row[i*4+1], row[i*4+2], row[i*4+3] = shotEmpty.R, shotEmpty.G, shotEmpty.B, 255
	}
	for y := b.Dy() - 1; y >= 0; y-- {
		line := img.Pix[y*img.Stride : y*img.Stride+b.Dx()*4]
		if !bytes.Equal(line, row) {
			return y, y < b.Dy()-1
		}
	}
	return -1, true
}

// shotStack puts the slices one under the other, in the order they were
// drawn, and encodes the picture. A slice is decoded as it is drawn into
// the picture and let go after, so that only one of them is held whole.
func shotStack(parts []shotPart) ([]byte, error) {
	width, height := 0, 0
	for _, p := range parts {
		if p.cols > width {
			width = p.cols
		}
		height += p.rows
	}
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	y := 0
	for _, p := range parts {
		img, err := png.Decode(bytes.NewReader(p.data))
		if err != nil {
			return nil, err
		}
		b := img.Bounds()
		draw.Draw(out, image.Rect(0, y, b.Dx(), y+b.Dy()), img, b.Min, draw.Src)
		y += p.rows
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// shotFile is what the picture is called: the chat it came from and the
// moment, with the characters a file name cannot hold left out.
func shotFile(chat string, when time.Time) string {
	name := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) {
			return '-'
		}
		return r
	}, strings.TrimSpace(chat))
	if name == "" {
		name = "mygohu"
	}
	return name + "-" + when.Format("20060102-150405") + ".png"
}
