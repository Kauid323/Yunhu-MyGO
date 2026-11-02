package app

import (
	"math"
	"strings"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// A picture is drawn in a box of this many DIPs, and asked of the host at
// a quarter of the size it was sent in, or at this many pixels when the
// message does not say what size that was. A sticker is a small picture
// with nothing behind it.
const (
	pictureBoxDIP = 240
	pictureBoxPx  = 480
	// pictureShrink is how much smaller each side is asked for than the
	// picture was sent: a quarter of a 2560×1530 screenshot is 640×382,
	// which is still more than the box shows, and a sixteenth of the
	// pixels to decode and keep.
	pictureShrink = 4

	// stickerDIP is the side of the square a sticker is drawn in, and
	// stickerPx what it is asked of the host, for a screen with twice the
	// pixels.
	stickerDIP = 120
	stickerPx  = 240
)

// picture draws a message that shows a picture: the host serves it at the
// size it is drawn, so an upload of several megabytes costs a few KB of
// memory. The caption is what the message says besides the picture.
//
// Nothing is drawn behind it: a photograph with a bubble around it is a
// photograph in a frame, and a chat shows one as it is. A press on it
// opens it in a window of its own, at the size it was sent in.
func (a *App) picture(c *ui.Context, m *yh.Message, url, caption string) {
	w, h := picturePixels(m)
	box := ui.Column(c).Gap(4).MaxWidth(pictureBoxDIP)
	box.Children(func() {
		if bmp := a.images.getBox(url, w, h, a); bmp != nil {
			dw, dh := ui.ImageSize(bmp)
			// Drawn for a screen with twice the pixels, within the box.
			k := float32(2)
			if dw*k > pictureBoxDIP {
				k = pictureBoxDIP / dw
			}
			if dh*k > pictureBoxDIP {
				k = pictureBoxDIP / dh
			}
			shot := ui.Box(c).Cursor(ui.CursorPointer).Children(func() {
				ui.Image(c, bmp).Width(dw * k).Height(dh * k).Radius(6)
			})
			if shot.Clicked() {
				pw, ph := previewPixels(m)
				a.showPicture(url, pw, ph)
			}
		} else {
			// On its way: the room it will take, without a frame.
			ui.Column(c).Width(pictureBoxDIP * 0.62).Height(pictureBoxDIP * 0.44).
				Radius(6).Background(qq.inputBg).Justify(ui.Center).AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, i18n.T("msg.imageLoading")).FontSize(11).TextColor(qq.textFaint)
				})
		}

		// A picture may carry words, and a message of the account's own
		// its state.
		if caption = strings.TrimSpace(caption); caption != "" && caption != url {
			ui.Text(c, caption).FontSize(12.5).LineHeight(1.4).TextColor(qq.text).Wrap()
		}
		a.bubbleState(c, m, qq.text, 0)
	})
}

// sticker draws a message that shows a sticker: the picture on its own, as
// a chat draws one, without a bubble and without a border.
func (a *App) sticker(c *ui.Context, m *yh.Message, url string) {
	bmp := a.images.getBox(url, stickerPx, stickerPx, a)
	if bmp == nil {
		// On its way: the room it will take, without a frame.
		ui.Box(c).Size(stickerDIP*0.8, stickerDIP*0.8).Radius(8).
			Background(qq.inputBg).Opacity(0.5)
		return
	}
	w, h := ui.ImageSize(bmp)
	k := float32(2)
	if w*k > stickerDIP {
		k = stickerDIP / w
	}
	if h*k > stickerDIP {
		k = stickerDIP / h
	}
	ui.Image(c, bmp).Width(w * k).Height(h * k)
	a.bubbleState(c, m, qq.text, 0)
}

// stickerHost is where a sticker's picture is served from: a message names
// it as a path on that host rather than as a whole address.
const stickerHost = "https://chat-img.jwznb.com"

// stickerURL is the picture a sticker message shows, or "" when it names
// none. The message carries it as an address, or as the path of one.
// stickerURLFor is the address a picture the sticker host keeps is fetched
// from: the API names it as a path, and a message and a saved sticker name it
// the same way.
func stickerURLFor(url string) string {
	url = strings.TrimSpace(url)
	switch {
	case url == "":
		return ""
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		return url
	case strings.HasPrefix(url, "//"):
		return "https:" + url
	}
	return stickerHost + "/" + strings.TrimLeft(url, "/")
}

func stickerURL(m *yh.Message) string {
	return stickerURLFor(firstNonEmpty(m.Content.Image, m.Content.ImageURL))
}

// picturePixels is the size a picture is asked for: a quarter of each side
// of the one sent, so that what is decoded is a sixteenth of its pixels.
// A message that does not say how big its picture was is asked for at the
// box's own size, and one so small that a quarter of it would not fill the
// box is asked for as it was.
func picturePixels(m *yh.Message) (w, h int) {
	w, h = m.Content.Width, m.Content.Height
	if w <= 0 || h <= 0 {
		return pictureBoxPx, pictureBoxPx
	}
	w, h = w/pictureShrink, h/pictureShrink
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

// pictureURL is the picture a message shows, or "" for one that shows
// none: an image message carries one, and a message of markup may have
// one in it.
func pictureURL(m *yh.Message) string {
	switch m.ContentType {
	case yh.ContentImage:
		return firstNonEmpty(m.Content.ImageURL, m.Content.Image)
	case yh.ContentHTML:
		return htmlImage(m.Content.Text)
	}
	return ""
}

// avatar draws a round picture, loading it from url when it is not here yet.
func (a *App) avatar(c *ui.Context, url, name string, size float32) {
	if name == "" {
		name = "?"
	}
	var bmp *ui.Bitmap
	if url != "" {
		bmp = a.images.get(url, avatarPixels(size), a)
	}
	ui.Avatar(c, name, bmp).Size(size, size)
}

// avatarPixels is how many pixels wide a picture is asked for: twice the
// DIPs it is drawn at, for a screen with twice the pixels, within what the
// host serves.
func avatarPixels(size float32) int {
	px := int(math.Ceil(float64(size) * 2))
	if px < 32 {
		px = 32
	}
	if px > 256 {
		px = 256
	}
	return px
}
