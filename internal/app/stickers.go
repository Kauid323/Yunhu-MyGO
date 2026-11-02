package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The panel over the composer is the account's saved stickers: the pictures it
// keeps to answer with. A message of one carries the picture's name and the
// sticker's id, which is what the server keeps so that it can be added back to
// the collection — a sticker sent without it arrives as a picture.
const (
	stickerPanelW = 296
	stickerCell   = 44
	stickerPerRow = 6
	stickerRows   = 5
)

// loadStickers asks for the account's stickers, once per run: they change when
// the account adds one, which this window does not do.
func (a *App) loadStickers() {
	if a.stickerGo || len(a.stickers) > 0 {
		return
	}
	a.stickerGo = true
	go func() {
		list, err := a.client.Expressions()
		a.update(func() {
			a.stickerGo = false
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.stickers = list
		})
	}()
}

// toggleStickers shows the panel of stickers, or puts it away, asking for them
// the first time it is shown.
func (a *App) toggleStickers() {
	a.stickerOpen = !a.stickerOpen
	if a.stickerOpen {
		a.loadStickers()
	}
}

// stickersPanel draws the panel over what it is anchored to while it is open.
func (a *App) stickersPanel(c *ui.Context, anchor ui.Element) {
	if !a.stickerOpen {
		return
	}
	ui.PopoverBase(c, anchor, &a.stickerOpen, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorBottomLeft).
			Left(-8).Width(stickerPanelW).Padding(8).Radius(10).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 8, 22, 0, ui.RGBA(0, 0, 0, 0.30)).
			TransformOrigin(0, 1).Transition(menuArrives)
		a.stickerGrid(c)
	})
}

// stickerGrid is the stickers, a row at a time, as many as the account has:
// the panel is as tall as five rows of them and scrolls to the rest, which
// are only fetched as they are shown.
func (a *App) stickerGrid(c *ui.Context) {
	switch {
	case a.stickerGo:
		ui.Spinner(c).Margin(ui.Auto)
		return
	case len(a.stickers) == 0:
		ui.Text(c, i18n.T("sticker.empty")).FontSize(11.5).
			TextColor(qq.textMuted).Padding(8)
		return
	}
	ui.Scroll(c).MaxHeight(stickerCell * stickerRows).Children(func() {
		ui.Column(c).Gap(2).Children(func() {
			for i := 0; i < len(a.stickers); i += stickerPerRow {
				row := a.stickers[i:min(i+stickerPerRow, len(a.stickers))]
				ui.Row(c).Gap(2).Children(func() {
					for _, ex := range row {
						a.stickerButton(c, ex)
					}
				})
			}
		})
	})
}

// stickerButton is one sticker of the panel: its picture, and a press that
// sends it to the chat that is open.
func (a *App) stickerButton(c *ui.Context, ex yh.Expression) {
	box := ui.Box(c).Size(stickerCell, stickerCell).Radius(6).Center().
		Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.menuHover)
	}
	box.Children(func() {
		if bmp := a.images.getBox(stickerURLFor(ex.URL), stickerPx, stickerPx, a); bmp != nil {
			w, h := ui.ImageSize(bmp)
			k := min(stickerCell/w, stickerCell/h)
			ui.Image(c, bmp).Width(w * k).Height(h * k)
		}
	})
	if box.Clicked() {
		a.sendSticker(ex)
	}
}

// sendSticker sends one of the account's stickers to the chat that is open,
// which shows it at once and marks it failed if the server does not take it,
// as a message of text is.
func (a *App) sendSticker(ex yh.Expression) {
	if a.current == "" || a.me == nil {
		return
	}
	a.stickerOpen = false
	chatID, chatType := a.current, a.currentType
	pending := &yh.Message{
		MsgID: "local-" + yh.NewID(),
		Sender: yh.Sender{
			ChatID: a.me.ID, Name: a.me.Name, AvatarURL: a.me.AvatarURL,
			ChatType: yh.ChatUser,
		},
		ContentType: yh.ContentSticker,
		Content:     yh.Content{Image: ex.URL},
		TimestampMS: uint64(time.Now().UnixMilli()),
		Pending:     true,
	}
	a.msgs = append(a.msgs, pending)
	a.jumpToEnd = true

	go func() {
		sent, err := a.client.SendSticker(chatID, chatType, ex.URL, ex.ID)
		a.update(func() {
			for i, m := range a.msgs {
				if m.MsgID != pending.MsgID {
					continue
				}
				if err != nil {
					m.Pending = false
					m.Failed = err.Error()
					a.note = errorLine(err)
					return
				}
				sent.Sender = pending.Sender
				a.msgs[i] = sent
				return
			}
		})
		if err != nil {
			return
		}
		// The list says what was said last, which is now a sticker.
		a.update(func() {
			for i := range a.convs {
				if a.convs[i].ChatID != chatID {
					continue
				}
				a.convs[i].Content = preview(yh.ContentSticker, yh.Content{Image: ex.URL})
				a.convs[i].SendTime = uint64(time.Now().Unix())
				a.sortConversations()
				return
			}
		})
	}()
}
