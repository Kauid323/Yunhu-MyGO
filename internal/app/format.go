package app

import (
	"errors"
	"strings"
	"time"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The state as words: what a message says, what a chat is called, when it
// was last talked in. Nothing here draws.

// line is a piece of writing the window shows: either words of the
// dictionary, with what fills their placeholders, or words that came from
// the server, which are shown as they are.
//
// It is kept as a key rather than as the finished sentence so that it
// follows a change of language: the desktop can switch one while "signing
// back in" is on the screen.
type line struct {
	key  string
	args []any
	raw  string
}

// say is a line of the dictionary.
func say(key string, args ...any) line { return line{key: key, args: args} }

// words is a line that is not the dictionary's, such as the server's own
// account of what went wrong.
func words(s string) line { return line{raw: s} }

// String is the line as it reads now.
func (l line) String() string {
	if l.key == "" {
		return l.raw
	}
	return i18n.T(l.key, l.args...)
}

// empty reports whether there is no line at all.
func (l line) empty() bool { return l.key == "" && l.raw == "" }

// messageText is what a message shows: its text, and a line of detail for
// the kinds that are not text.
func messageText(m *yh.Message) (body, detail string) {
	c := m.Content
	switch m.ContentType {
	case yh.ContentImage:
		return c.Text, ""
	case yh.ContentHTML:
		// Markup is read for its words: there is no page here to show it,
		// and its tags are not worth reading.
		return htmlText(c.Text), ""
	case yh.ContentFile:
		size := ""
		if c.FileSize > 0 {
			size = i18n.T("msg.fileSize", float64(c.FileSize)/1024)
		}
		return firstNonEmpty(c.Text, i18n.T("msg.file", c.FileName)),
			strings.TrimSpace(size + " " + c.FileURL)
	case yh.ContentAudio:
		return firstNonEmpty(c.Text, i18n.T("msg.voice", c.AudioTime)), c.AudioURL
	case yh.ContentSticker:
		return firstNonEmpty(c.Text, i18n.T("msg.sticker")), ""
	case yh.ContentForm:
		return firstNonEmpty(c.Text, i18n.T("msg.form")), ""
	}
	if c.Text != "" {
		// The quote of a message that answers another is not part of what
		// it says: it has a block of its own above the words, which the
		// bubble draws.
		return c.Text, ""
	}
	if c.PostTitle != "" {
		return c.PostTitle, c.PostContent
	}
	if c.ImageURL != "" || c.Image != "" {
		return i18n.T("msg.image"), firstNonEmpty(c.ImageURL, c.Image)
	}
	if c.VideoURL != "" {
		return i18n.T("msg.video"), c.VideoURL
	}
	if c.AudioURL != "" {
		return i18n.T("msg.voiceShort"), c.AudioURL
	}
	return preview(m.ContentType, c), c.QuoteText
}

// preview is a message in one line, as a conversation's list shows it. The
// words are the dictionary's, which is why the preview is made here rather
// than by the client: a preview is what the window says, not what the
// server sent.
func preview(typ int, c yh.Content) string {
	switch typ {
	case yh.ContentImage:
		return i18n.T("msg.image")
	case yh.ContentFile:
		if c.FileName != "" {
			return i18n.T("msg.file", c.FileName)
		}
		return i18n.T("msg.fileBare")
	case yh.ContentAudio:
		return i18n.T("msg.voiceShort")
	case yh.ContentSticker:
		return i18n.T("msg.sticker")
	case yh.ContentForm:
		return i18n.T("msg.form")
	case yh.ContentCall:
		return i18n.T("msg.call")
	}
	switch {
	case c.VideoURL != "":
		return i18n.T("msg.video")
	case c.ImageURL != "" || c.Image != "":
		return i18n.T("msg.image")
	case c.AudioURL != "":
		return i18n.T("msg.voiceShort")
	case c.Tip != "":
		return c.Tip
	}
	text := c.Text
	if text == "" {
		text = c.PostTitle
	}
	return clip(strings.ReplaceAll(text, "\n", " "))
}

// clip shortens a preview to one line's worth, counting characters rather
// than bytes: half a character is not a preview.
func clip(s string) string {
	const limit = 80
	if r := []rune(s); len(r) > limit {
		return string(r[:limit]) + "…"
	}
	return s
}

// errorLine is what went wrong, in the words of the dictionary. A failure
// the server explained itself keeps the server's words: they are its
// account of the request, not a line of ours to translate.
func errorLine(err error) line {
	var api *yh.APIError
	switch {
	case errors.As(err, &api):
		if api.Msg == "" {
			return say("err.code", api.Code)
		}
		return words(api.Error())
	case errors.Is(err, yh.ErrNetwork):
		return say("err.network")
	case errors.Is(err, yh.ErrResponse):
		return say("err.response")
	case errors.Is(err, yh.ErrStatus):
		return say("err.status")
	case errors.Is(err, yh.ErrNoToken):
		return say("err.noToken")
	case errors.Is(err, yh.ErrNoAccount):
		return say("err.account")
	case errors.Is(err, yh.ErrNoCaptcha):
		return say("err.captcha")
	}
	return words(err.Error())
}

// kindName names a chat type.
func kindName(chatType int) string {
	switch chatType {
	case yh.ChatUser:
		return i18n.T("chat.kind.user")
	case yh.ChatGroup:
		return i18n.T("chat.kind.group")
	case yh.ChatBot:
		return i18n.T("chat.kind.bot")
	}
	return i18n.T("chat.kind.other")
}

// clock formats a conversation's time, as a chat list does: the hour for
// today, the day for this week, the date after that. The layouts are the
// dictionary's, so that a language writes the date the way it writes
// dates.
func clock(t time.Time) string { return dayLine(t) }

// dayLine is the day a time falls on, as a chat writes it. The days are
// told apart by the calendar rather than by hours elapsed: a message sent
// at eleven last night is yesterday's at one in the morning, whatever the
// two hours between them.
func dayLine(t time.Time) string {
	if !real(t) {
		return ""
	}
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch {
	case !t.Before(midnight):
		return t.Format(i18n.T("time.clock"))
	case !t.Before(midnight.AddDate(0, 0, -1)):
		return i18n.T("time.yesterday")
	case !t.Before(midnight.AddDate(0, 0, -6)):
		return i18n.T("time.weekday." + weekday(t.Weekday()))
	case t.Year() == now.Year():
		return t.Format(i18n.T("time.monthDay"))
	}
	return t.Format(i18n.T("time.date"))
}

// whenLine is a time as the history writes it between stretches of
// talking: the hour for today, and the day it was for the days around it.
func whenLine(t time.Time) string {
	if !real(t) {
		return ""
	}
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	hm := t.Format(i18n.T("time.clock"))
	switch {
	case !t.Before(midnight):
		return i18n.T("time.stamp.today", hm)
	case !t.Before(midnight.AddDate(0, 0, -1)):
		return i18n.T("time.stamp.yesterday", hm)
	case !t.Before(midnight.AddDate(0, 0, -6)):
		return i18n.T("time.stamp.weekday", i18n.T("time.weekday."+weekday(t.Weekday())), hm)
	}
	date := t.Format(i18n.T("time.monthDay"))
	if t.Year() != now.Year() {
		date = t.Format(i18n.T("time.date"))
	}
	return i18n.T("time.stamp.date", date, hm)
}

// weekday is the key of a day of the week, which the dictionary names.
func weekday(d time.Weekday) string {
	return [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}[int(d)]
}

// connState names how the live connection stands, in the language in use.
func connState(s yh.State) string {
	switch s {
	case yh.StateOnline:
		return i18n.T("conn.online")
	case yh.StateRetrying:
		return i18n.T("conn.reconnecting")
	}
	return i18n.T("conn.offline")
}

// totalUnread is how many messages wait across all conversations.
func totalUnread(convs []yh.Conversation) int {
	n := 0
	for _, conv := range convs {
		n += conv.Unread
	}
	return n
}

// shortCount shortens a count the way a chat list does: 99+ after 99.
func shortCount(n int) string {
	if n > 99 {
		return "99+"
	}
	return i18n.Num(n)
}
