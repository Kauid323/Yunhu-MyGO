package app

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// gapLine is how long a silence between two messages has to be before the
// time is worth a line of its own, as a chat shows it: five minutes, and a
// new day whatever the silence.
const gapLine = 5 * time.Minute

// messageList is the conversation's history, with the pill over its foot
// that says how many messages are still below what has been seen. Older
// messages come in by themselves as the reader goes back through it.
func (a *App) messageList(c *ui.Context) {
	// The pill is placed against the list rather than against the window,
	// so that it sits over the foot of the history whatever is below it.
	ui.Box(c).Grow(1).Children(func() {
		a.messages(c)
		a.jumpToUnread(c)
		a.pickBar(c)
	})
}

// messages is the history itself, scrolled to where reading left off.
func (a *App) messages(c *ui.Context) {
	// Ask for the scroll while the list is being built: a scroll asked
	// for outside a frame waits for the next one, which shows the top
	// until something else redraws.
	switch {
	case a.jumpToEnd:
		a.msgList.ScrollToEnd()
		a.jumpToEnd = false
	case a.jumpTo >= 0:
		a.msgList.ScrollTo(a.jumpTo, ui.Start)
		a.jumpTo = -1
	}

	// The list keeps its place by the message at its top, so the ones
	// that come in above do not move what is being read.
	a.msgList.Key = func(i int) any { return a.msgs[i].MsgID }

	rows := len(a.msgs)
	ui.List(c, &a.msgList, rows, func(i int) {
		a.messageRow(c, a.msgs[i], i)
	}).Grow(1).Padding(14, 18).Gap(12).Justify(ui.End).Children(func() {
		switch {
		case a.current == "":
			ui.Text(c, i18n.T("chat.empty")).FontSize(12.5).TextColor(qq.textFaint).
				Margin(ui.Auto)
		case a.loading && len(a.msgs) == 0:
			ui.Spinner(c).Margin(ui.Auto)
		case len(a.msgs) == 0:
			ui.Text(c, i18n.T("chat.noMessages")).FontSize(12.5).TextColor(qq.textFaint).
				Margin(ui.Auto)
		}
	})

	// Reading back through the history loads more of it, going down through
	// it reads it, and coming forward through a history that was opened in
	// the middle of itself loads what follows it.
	first, last := a.msgList.Visible()
	a.markRead(first, last)
	if a.wantMore(first, last, rows) {
		a.loadedAt, a.moreAt = first, time.Now()
		a.loadOlder()
	}
	if a.wantNewer(last, rows) {
		a.loadNewer()
	}
}

// moreWait is how soon after a page is asked for another may be: a list
// showing its top asks from the same row every frame, and a page that came
// back empty must not be asked for on each of them.
const moreWait = 400 * time.Millisecond

// wantMore reports whether the history should grow: four fifths of what is
// here is behind the reader by the time the top fifth shows, and the rest
// follows before they reach the first message.
//
// The row the last page was asked for at guards against asking for one on
// every frame: the list must have gone further back since, as it does when
// the reader scrolls. Without it a list that failed to keep its place, and
// so kept showing the top, would page the whole history in.
func (a *App) wantMore(first, last, rows int) bool {
	if !a.hasMore || a.loadingMore || !farBack(first, last, rows) {
		return false
	}
	if !a.moreAt.IsZero() && time.Since(a.moreAt) < moreWait {
		return false
	}
	return first < a.loadedAt
}

// farBack reports whether the reader has gone back through four fifths of
// the history, so that the top fifth of it is what shows: the point at
// which more of it is asked for.
func farBack(first, last, rows int) bool {
	return rows > 0 && last >= first && first <= rows/5
}

// messageRow is one message: the account's on the right, others' on the
// left with their avatar and name. A stretch of talking is opened by its
// time, and the messages that came in while the reader was away are
// opened by the line across the history.
func (a *App) messageRow(c *ui.Context, m *yh.Message, i int) {
	if a.startsStretch(i) {
		timeLine(c, m.Time())
	}
	if i == a.unreadFrom && a.unread() {
		newLine(c)
	}

	if m.RecalledAt > 0 {
		ui.Text(c, i18n.T("chat.recalled")).FontSize(11).TextColor(qq.textFaint).Margin(ui.Auto)
		return
	}
	if m.ContentType == yh.ContentTip || m.ContentType == yh.ContentCall {
		text := m.Content.Tip
		if text == "" {
			text = preview(m.ContentType, m.Content)
		}
		ui.Text(c, text).FontSize(11).TextColor(qq.textFaint).Margin(ui.Auto)
		return
	}

	if m.Mine(a.me.ID) {
		row := ui.Row(c)
		row.Children(func() {
			a.pickBox(c, m)
			ui.Spacer(c)
			a.bubble(c, m)
		})
		a.rightClickMessage(m, row)
		a.rowPress(m, row)
		a.flashRow(c, m, row)
		return
	}

	// The face opens what is known about whoever said it. The time of a
	// message is asked for by resting on the message, and the column is
	// only as wide as the name and the bubble in it: the room beside a
	// short message is not the message, so passing through it asks for
	// nothing. The column hugs them rather than filling the row for the
	// same reason a bubble is not stretched by the time appearing in the
	// line above it.
	var col ui.Element
	row := ui.Row(c).Gap(9).AlignItems(ui.Start)
	row.Children(func() {
		a.pickBox(c, m)
		a.messageAvatar(c, m.Sender, 32)
		col = ui.Column(c).MaxWidth(540).Gap(4).AlignItems(ui.Start)
		col.Children(func() {
			name := m.Sender.Name
			if name == "" {
				name = m.Sender.ChatID
			}
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, name).FontSize(10.5).TextColor(qq.textMuted).SingleLine()
				if a.stampOn(c, m) {
					ui.Text(c, stampText(m.Time())).FontSize(10).
						TextColor(qq.textFaint).SingleLine()
				}
			})
			a.bubble(c, m)
		})
	})
	// The time of a message and the menu of it are for reading, not for
	// picking: while the history is being picked from, a press chooses.
	if !a.picking {
		a.hoverStamp(c, m, col)
		a.rightClickMessage(m, row)
	}
	a.rowPress(m, row)
	a.flashRow(c, m, row)
}

// messageAvatar is the face of whoever said a message, which opens the card of
// them beside it.
func (a *App) messageAvatar(c *ui.Context, s yh.Sender, size float32) {
	box := ui.Box(c).Size(size, size).Radius(size / 2).Cursor(ui.CursorPointer)
	box.Children(func() { a.avatar(c, s.AvatarURL, s.Name, size) })
	if box.Clicked() {
		a.openUserCard(s.ChatID, s.Name, s.AvatarURL, box.Bounds())
	}
}

// stampWait is how long the pointer has to rest on a message before its time
// is shown: the time answers a question asked by staying, so that a pointer
// crossing the history on its way somewhere else says nothing.
const stampWait = 3 * time.Second

// stampOn reports whether a message is showing its time. The wait is ended by
// the frame asked for when the pointer settled, and that frame draws the time
// in the same pass rather than one after it.
func (a *App) stampOn(c *ui.Context, m *yh.Message) bool {
	if a.hoverID != m.MsgID || a.hoverAt.IsZero() {
		return false
	}
	if a.stampID == m.MsgID {
		return true
	}
	if c.Now().Sub(a.hoverAt) < stampWait {
		return false
	}
	a.stampID = m.MsgID
	return true
}

// hoverStamp notes where the pointer is. Settling on a message starts the
// wait, and leaving takes the time with it: a time left behind on a message
// nobody is looking at would be a time shown to the wrong reader.
func (a *App) hoverStamp(c *ui.Context, m *yh.Message, box ui.Element) {
	if box.Hovered() {
		if a.hoverID == m.MsgID {
			return
		}
		a.hoverID, a.hoverAt, a.stampID = m.MsgID, c.Now(), ""
		// The margin is for the frame that is asked for, which may come a
		// hair before the wait is up.
		c.After(stampWait + 20*time.Millisecond)
		return
	}
	if a.hoverID == m.MsgID {
		a.hoverID, a.hoverAt, a.stampID = "", time.Time{}, ""
	}
}

// stampText is a time as a message writes it when the pointer rests on it: the
// day and the hour, to the second, as the chat this window copies does.
func stampText(t time.Time) string {
	if !real(t) {
		return ""
	}
	return i18n.T("time.stamp.full", t.Format("01-02"), t.Format("15:04:05"))
}

// startsStretch reports whether a message opens a stretch of talking, and
// so carries the time of it: the first of the history, one sent long
// enough after the one above, and one sent on another day.
func (a *App) startsStretch(i int) bool {
	if i <= 0 {
		return i == 0 && len(a.msgs) > 0
	}
	if i >= len(a.msgs) {
		return false
	}
	prev, cur := a.msgs[i-1].Time(), a.msgs[i].Time()
	if !real(cur) {
		return false
	}
	if !real(prev) {
		return true
	}
	if cur.Sub(prev) > gapLine {
		return true
	}
	y1, m1, d1 := prev.Date()
	y2, m2, d2 := cur.Date()
	return y1 != y2 || m1 != m2 || d1 != d2
}

// real reports whether a time is one the server gave: a message without
// one is shown without a time rather than in 1970.
func real(t time.Time) bool { return !t.IsZero() && t.Year() >= 2000 }

// timeLine is the time a stretch of talking began, in the middle of the
// history, as a chat shows it: a pill on the ground, so that it reads as a
// mark of the history rather than as something someone said.
func timeLine(c *ui.Context, t time.Time) {
	s := whenLine(t)
	if s == "" {
		return
	}
	ui.Row(c).Justify(ui.Center).Padding(2, 0, 3, 0).Children(func() {
		ui.Box(c).Padding(2, 9).Radius(9).Background(qq.timePill).
			Children(func() {
				ui.Text(c, s).FontSize(10.5).TextColor(qq.textFaint)
			})
	})
}

// newLine marks where the messages that came in while the reader was away
// begin, as a chat does.
func newLine(c *ui.Context) {
	ui.Row(c).Gap(10).AlignItems(ui.Center).Padding(2, 0, 4, 0).Children(func() {
		ui.Box(c).Grow(1).Height(1).Background(qq.accent).Opacity(0.5)
		ui.Text(c, i18n.T("chat.newLine")).FontSize(10.5).TextColor(qq.accent)
		ui.Box(c).Grow(1).Height(1).Background(qq.accent).Opacity(0.5)
	})
}

// jumpToUnread is the pill over the bottom right of the history, as QQ
// has it: a down arrow and the number of messages still below what has
// been seen. It sits over the foot of the list rather than its top,
// because what it points at is further down. The number drops by one for
// each row that comes into view, so scrolling through the unread ones
// counts them off rather than holding a total that never changes. Pressing
// it jumps to the very bottom.
func (a *App) jumpToUnread(c *ui.Context) {
	if a.picking {
		return
	}
	n := a.unreadBelow()
	if n <= 0 {
		return
	}
	box := ui.Row(c).Absolute().Bottom(14).Right(16).Height(28).Padding(0, 14).
		Radius(14).Gap(5).AlignItems(ui.Center).Background(qq.menuBg).
		Border(1, qq.menuLine).Shadow(0, 2, 8, 0, ui.RGBA(0, 0, 0, 0.16)).
		Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.menuHover)
	}
	box.Children(func() {
		ui.Icon(c, iconChevronsDown).FontSize(13).TextColor(qq.accent)
		text := i18n.Num(n)
		if n > 99 {
			text = "99+"
		}
		ui.Text(c, text).FontSize(12).TextColor(qq.accent).SingleLine()
	})
	if box.Clicked() {
		if a.hasNewer {
			a.hasNewer = false
			a.jumpToEnd = true
			a.loadMessages(a.current, a.currentType)
		} else {
			a.msgList.ScrollToEnd()
			a.jumpToEnd = true
			a.readFrom = len(a.msgs)
		}
	}
}

// bubble draws a message's content in a bubble. Both sides are the same
// gray: which is whose is told by the side it sits on, as a chat tells it,
// rather than by a color of its own. A picture and a sticker are drawn
// bare, as a chat draws them: a bubble around a photograph is a frame
// around it.
func (a *App) bubble(c *ui.Context, m *yh.Message) {
	bg, fg := qq.bubbleOther, qq.text

	text, detail := messageText(m)
	if m.ContentType == yh.ContentSticker {
		if url := stickerURL(m); url != "" {
			a.sticker(c, m, url)
			return
		}
	}
	// A picture that is still on its way has no address to be fetched from
	// yet: it shows the room it will take rather than a word for a message
	// nobody can read.
	if m.Pending && m.ContentType == yh.ContentImage {
		a.pendingPicture(c, m)
		return
	}
	if m.Pending && m.ContentType == yh.ContentVideo {
		a.pendingVideo(c, m)
		return
	}
	if url := pictureURL(m); url != "" {
		a.picture(c, m, url, text)
		return
	}
	if m.ContentType == yh.ContentVideo || m.Content.VideoURL != "" {
		a.videoBubble(c, m, bg, fg)
		return
	}
	if text == "" && detail == "" {
		text = i18n.T("chat.unknown")
	}
	box := ui.Column(c).Padding(8, 12).Radius(8).Background(bg).Gap(3).MaxWidth(500)
	box.Children(func() {
		// A message that answers another shows what it answers above what
		// it says, as a chat shows a quote.
		if q := strings.TrimSpace(m.Content.QuoteText); q != "" {
			quoteBlock(c, q)
		}
		switch {
		case m.ContentType == yh.ContentMarkdown:
			// Markdown is drawn in the styles it asks for, not read as
			// stars and hashes.
			a.markdown(c, m.Content.Text, fg)
		case text != "":
			ui.Text(c, text).FontSize(13).LineHeight(1.45).TextColor(fg).
				Selectable().Wrap()
		}
		if detail != "" {
			ui.Text(c, detail).FontSize(10.5).TextColor(fg).Opacity(0.75).
				Selectable().SingleLine()
		}
		a.bubbleState(c, m, fg, 0)
	})
	if m.EditedAt > 0 {
		ui.Text(c, i18n.T("chat.edited")).FontSize(9.5).TextColor(qq.textFaint)
	}
}

// videoBubble is a clip in the history: its name and size beside the play
// mark, rather than the long address it is served from. There is no player
// in the window, so the row keeps the address for whoever opens it and
// shows what a reader can recognise.
func (a *App) videoBubble(c *ui.Context, m *yh.Message, bg, fg ui.Color) {
	box := ui.Column(c).Padding(8, 12).Radius(8).Background(bg).Gap(3).MaxWidth(500)
	box.Children(func() {
		if q := strings.TrimSpace(m.Content.QuoteText); q != "" {
			quoteBlock(c, q)
		}
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, iconVideo).FontSize(21).TextColor(fg)
			ui.Column(c).Gap(2).MaxWidth(260).Children(func() {
				name := m.Content.FileName
				if name == "" {
					name = i18n.T("msg.video")
				}
				ui.Text(c, name).FontSize(12.5).TextColor(fg).
					MinWidth(0).SingleLine()
				if m.Content.FileSize > 0 {
					ui.Text(c, i18n.T("msg.fileSize", float64(m.Content.FileSize)/1024)).
						FontSize(10.5).TextColor(fg).Opacity(0.75).SingleLine()
				}
			})
		})
		a.bubbleState(c, m, fg, 0)
	})
}

// quoteBlock is the message a message answers, above what it says: who said
// it, and what they said. It is a step off the bubble, and comes forward with
// a shadow under the pointer, as the chat this window copies does.
// quoteBlock is what a message answers, above what it says. It has no
// ground of its own: a quote sits in the bubble it is in, told apart from
// the answer by the line down its left and the dimmer words, so that the
// bubble reads as one thing rather than as a box inside a box.
func quoteBlock(c *ui.Context, text string) {
	name, body := splitQuote(text)
	box := ui.Column(c).Gap(2).Padding(1, 0, 1, 10).
		BorderWidth(0, 0, 0, 2).BorderColor(qq.quoteBar)
	box.Children(func() {
		if name != "" {
			ui.Text(c, name).FontSize(10.5).TextColor(qq.textMuted).SingleLine()
		}
		ui.Text(c, body).FontSize(12).TextColor(qq.textMuted).SingleLine()
	})
}

// splitQuote cuts a quoted message into who said it and what they said: the
// server writes it as "name: text", and one that names nobody is all text.
func splitQuote(s string) (name, body string) {
	for _, sep := range []string{"：", ": "} {
		if i := strings.Index(s, sep); i > 0 && i < 24 {
			return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+len(sep):])
		}
	}
	return "", strings.TrimSpace(s)
}

// bubbleState draws what became of a message of the account's own: still
// on its way, or failed. pad is the side padding it sits in, as a bubble
// of a picture has none.
func (a *App) bubbleState(c *ui.Context, m *yh.Message, fg ui.Color, pad float32) {
	text := ""
	switch {
	case m.Pending:
		text = i18n.T("chat.pending")
	case m.Failed != "":
		text = i18n.T("chat.failed", m.Failed)
	default:
		return
	}
	t := ui.Text(c, text).FontSize(10).TextColor(fg).Wrap()
	if m.Pending {
		t.Opacity(0.6)
	} else {
		t.TextColor(qq.error)
	}
	if pad > 0 {
		t.Padding(0, pad, 4, pad)
	}
}

// firstNonEmpty is the first of the values that is not empty.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// flashRow marks the row a jump landed on: the list scrolls to what was
// found, and the tint says which row it stopped at, fading over the second
// after it.
func (a *App) flashRow(c *ui.Context, m *yh.Message, row ui.Element) {
	k := a.flash(m.MsgID)
	if k <= 0 {
		return
	}
	c.After(16 * time.Millisecond)
	row.Background(ui.RGBA(45, 119, 229, 0.16*k)).Radius(8)
}
