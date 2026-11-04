package app

import (
	"strings"
	"time"

	"mygohu/internal/yh"
)

// open shows a conversation, and loads its messages. It leaves the
// conversation's information panel as it is: what is known about a chat is
// asked for, not shown every time another one is picked.
//
// Where the chat is shown is the layout's business: the efficient one has
// it beside the list, and the classic one has no room for it and opens it in
// a window of its own.
func (a *App) open(conv yh.Conversation) {
	if a.current == conv.ChatID {
		// Picking the chat that is already open brings the window it is
		// in forward, which is what the classic panel has to do: the
		// chat is not in the panel. In wide mode it deselects the chat
		// and enters the clear screen state ("清屏").
		if a.prefs.Mode == modeClassic {
			a.openChatWindow()
			return
		}
		a.clearChat()
		return
	}
	a.current = conv.ChatID
	a.currentType = conv.ChatType
	// A group says how many are in it, which the header shows: it is
	// asked for as the chat is opened, and the header fills in when it
	// comes.
	if conv.ChatType == yh.ChatGroup {
		a.loadGroup(conv.ChatID)
	} else {
		a.group = nil
	}
	a.msgs = nil
	a.loading = true
	a.loadingMore = false
	a.hasMore = false
	a.hasNewer, a.loadingNewer, a.newerFrom = false, false, ""
	a.loadedAt = 1 << 30
	a.input = ""
	// Who is in the group is the group's, not this chat's: another chat is
	// asked for its own list, and the names picked in this one go with the
	// text that was cleared.
	a.members, a.mentionFind, a.mentionOpen, a.mentionGo = nil, "", false, false
	a.named = nil
	// A format was chosen for a text in the chat that was open, and that
	// text is gone: the next one starts as words again.
	a.sendType = yh.ContentText
	a.formatMenu.open = false
	// Messages picked in one chat are not picked in another: the bar that
	// acts on them goes with the history they were chosen from.
	a.picking, a.picked = false, nil
	// A message of the chat that was open is not the message the pointer
	// is on in this one, whatever ids they happen to share.
	a.hoverID, a.hoverAt, a.stampID = "", time.Time{}, ""
	// A message of the chat that was open is not the one being answered
	// in this one.
	a.quote = nil
	// What was unread is remembered before the list is told it is read:
	// it is where the chat opens, and where the line across the history
	// goes.
	unread := 0
	for i := range a.convs {
		if a.convs[i].ChatID == conv.ChatID {
			unread = a.convs[i].Unread
			a.convs[i].Unread = 0
		}
	}
	a.unreadOnOpen = unread
	a.unreadFrom = 0
	a.readFrom = 0
	go func() { _ = a.client.MarkRead(conv.ChatID) }()
	a.loadMessages(conv.ChatID, conv.ChatType)
	if a.prefs.Mode == modeClassic {
		// The panel has no room for a chat: it is read in a window of its
		// own, which follows the chat that is open.
		a.openChatWindow()
	}
}

// clearChat deselects the open conversation and returns to the empty screen.
func (a *App) clearChat() {
	a.current = ""
	a.currentType = 0
	a.group = nil
	a.msgs = nil
	a.loading = false
	a.loadingMore = false
	a.hasMore = false
	a.hasNewer = false
	a.loadingNewer = false
	a.newerFrom = ""
	a.input = ""
	a.members = nil
	a.mentionFind = ""
	a.mentionOpen = false
	a.mentionGo = false
	a.named = nil
	a.formatMenu.open = false
	a.picking = false
	a.picked = nil
	a.hoverID, a.hoverAt, a.stampID = "", time.Time{}, ""
	a.quote = nil
	a.unreadFrom = 0
	a.readFrom = 0
	a.infoOpen = false
}

// loadMessages loads the latest messages of a chat, or the page a jump is
// for when one is waiting: a chat opened at something that was found opens
// at the page that thing is on, and the newest page is never asked for.
func (a *App) loadMessages(chatID string, chatType int) {
	if a.jumpID != "" {
		a.loadAround()
		return
	}
	go func() {
		msgs, err := a.client.Messages(chatID, chatType, msgFirst)
		a.update(func() {
			if a.current != chatID {
				return
			}
			a.loading = false
			if err != nil {
				a.note = errorLine(err)
				return
			}
			a.setMessages(msgs)
		})
	}()
}

// setMessages shows a chat's first page, opened where reading left off: at
// the first message that came in while the reader was away, or at the
// newest one when there was nothing unread.
func (a *App) setMessages(msgs []*yh.Message) {
	a.msgs = msgs
	a.hasMore = len(msgs) >= msgFirst
	// The newest page ends at the newest message: there is nothing below it
	// to be asked for.
	a.hasNewer, a.newerFrom = false, ""
	a.loadedAt = 1 << 30
	a.unreadFrom = len(msgs) - a.unreadOnOpen
	if a.unreadFrom < 0 {
		a.unreadFrom = 0
	}
	a.readFrom = a.unreadFrom
	a.unreadOnOpen = 0
	// The scroll is asked for as the list is built, so that the same frame
	// lays out where the chat opens rather than waiting for the next one:
	// at the first message that was unread, or at the newest one when
	// there was none.
	a.jumpTo = -1
	if a.findJump() {
		return
	}
	if a.unread() {
		a.jumpTo = a.unreadFrom
		return
	}
	a.jumpToEnd = true
}

// loadOlder puts the messages before the oldest one on screen above it, and
// leaves the list where it was scrolled.
func (a *App) loadOlder() {
	if a.loadingMore || !a.hasMore || a.current == "" || len(a.msgs) == 0 {
		return
	}
	chatID, chatType := a.current, a.currentType
	before := a.msgs[0].MsgID
	a.loadingMore = true
	go func() {
		older, err := a.client.MessagesBefore(chatID, chatType, msgPage, before)
		a.update(func() {
			a.loadingMore = false
			if err != nil {
				a.note = errorLine(err)
				// A page that did not come is asked for again, and the row
				// it was asked for at says nothing about the next try: the
				// reader may be sitting at the top, where that row is the
				// only one there is.
				a.loadedAt = 1 << 30
				return
			}
			if a.current != chatID {
				return
			}
			if len(older) == 0 {
				a.hasMore = false
				return
			}
			fresh := a.withoutKnown(older)
			if len(fresh) == 0 {
				// The page held nothing new: asking again would ask for
				// the same messages for ever.
				a.hasMore = false
				return
			}
			a.prepend(fresh)
			a.hasMore = len(older) >= msgPage
			a.findJump()
		})
	}()
}

// prepend puts a page of older messages above the history, and moves what the
// list is showing down with it: the unread ones are that many rows further
// down now, and so is the row the last page was asked for at. Without that
// last one the top of the history would be a row nothing is ever asked for
// at again.
func (a *App) prepend(fresh []*yh.Message) {
	a.msgs = append(fresh, a.msgs...)
	a.unreadFrom += len(fresh)
	a.readFrom += len(fresh)
	a.loadedAt += len(fresh)
}

// withoutKnown drops from a page the messages already on screen, so that a
// chat paged twice shows each message once.
func (a *App) withoutKnown(page []*yh.Message) []*yh.Message {
	known := make(map[string]bool, len(a.msgs))
	for _, m := range a.msgs {
		known[m.MsgID] = true
	}
	out := page[:0]
	for _, m := range page {
		if !known[m.MsgID] {
			out = append(out, m)
		}
	}
	return out
}

// reload reloads the open conversation.
func (a *App) reload() {
	if a.current == "" {
		a.refreshConversations()
		return
	}
	a.loading = true
	a.loadMessages(a.current, a.currentType)
}

// unread reports whether anything came in while the reader was away, which
// the line across the history marks.
func (a *App) unread() bool { return a.unreadFrom < len(a.msgs) }

// unreadBelow is how many messages the reader has not been past yet: what
// the pill over the foot of the list counts. It only goes down. Going back
// up through the history does not count anything again, as coming into the
// chat at all is reading what is there.
func (a *App) unreadBelow() int {
	if a.msgList.AtEnd() {
		return 0
	}
	if n := len(a.msgs) - a.readFrom; n > 0 {
		return n
	}
	return 0
}

// markRead records how far down the history the reader has been: a row is
// read once it has come into view, so the number on the pill over the foot
// of the list drops by one for each message scrolled past. Counting by the
// first visible row instead would hold the total until a whole screenful
// had gone by, and leave messages that are plainly on screen unread.
func (a *App) markRead(first, last int) {
	if a.readFrom >= len(a.msgs) {
		return
	}
	if a.msgList.AtEnd() {
		a.readFrom = len(a.msgs)
		return
	}
	if last < first {
		return
	}
	if seen := last + 1; seen > a.readFrom {
		a.readFrom = min(seen, len(a.msgs))
	}
}

// send sends what is in the composer.
func (a *App) send() {
	text := strings.TrimSpace(a.input)
	if text == "" || a.current == "" || a.me == nil {
		return
	}
	a.input = ""
	chatID, chatType := a.current, a.currentType
	// What the text is written in is a choice for this message: it goes
	// back to words once the text has gone, as the choice was for it.
	kind := yh.SendKind(a.sendType)
	a.sendType = yh.ContentText
	// Who the text names goes with it: a name that was put in and then
	// typed over is not one the message is about any more.
	mentioned := a.namedIn(text)
	a.forgetNamed(text)
	// The message being answered, if there is one: it is sent with the
	// text, and the field stops answering it once it has gone.
	quoted := a.quote
	a.quote = nil
	pending := &yh.Message{
		MsgID:       "local-" + yh.NewID(),
		Sender:      yh.Sender{ChatID: a.me.ID, Name: a.me.Name, AvatarURL: a.me.AvatarURL, ChatType: yh.ChatUser},
		ContentType: yh.SendKind(kind),
		Content:     yh.Content{Text: text},
		TimestampMS: uint64(time.Now().UnixMilli()),
		Pending:     true,
	}
	if quoted != nil {
		pending.Content.QuoteText = quoteTextOf(quoted)
	}
	a.msgs = append(a.msgs, pending)
	a.trimMessages()
	a.jumpToEnd = true
	a.readFrom = len(a.msgs)

	go func() {
		var sent *yh.Message
		var err error
		if quoted != nil {
			sent, err = a.client.SendQuoted(chatID, chatType, text, quoted.MsgID,
				pending.Content.QuoteText, mentioned, kind)
		} else {
			sent, err = a.client.SendText(chatID, chatType, text, mentioned, kind)
		}
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
		if err == nil {
			a.update(func() { a.touchConv(chatID, text) })
		}
	}()
}

// touchConv puts what was just sent at the top of the conversation list,
// where the client puts it: the chat that was written in is the newest one,
// and its line under the name says what was sent.
func (a *App) touchConv(chatID, preview string) {
	for i := range a.convs {
		if a.convs[i].ChatID != chatID {
			continue
		}
		a.convs[i].Content = preview
		a.convs[i].SendTime = uint64(time.Now().Unix())
		a.sortConversations()
		return
	}
}

// jumpToMessage scrolls the history to a message the search found, which
// the chat may have scrolled past long ago. A message that is here already
// is scrolled to; one that is not is asked for by the page it is on, which
// is one request rather than the pages back to it, and the row it lands on
// is marked for a moment so that it is found by the eye as well as by the
// list.
func (a *App) jumpToMessage(id string, seq uint64) {
	if id == "" {
		return
	}
	a.jumpID, a.jumpSeq = id, seq
	a.findJump()
}

// findJump looks for the message a jump was asked for, and says whether it
// found it: a message that is not in the history is asked for by the page it
// is on when where it sits is known, and by the pages before the history one
// at a time when it is not.
func (a *App) findJump() bool {
	if a.jumpID == "" {
		return false
	}
	for i, m := range a.msgs {
		if m.MsgID != a.jumpID {
			continue
		}
		a.jumpTo = i
		a.flashID, a.flashAt = a.jumpID, time.Now()
		a.jumpID, a.jumpSeq = "", 0
		return true
	}
	if a.loadingAround {
		return false
	}
	if a.jumpSeq > 0 {
		a.loadAround()
		return false
	}
	if a.hasMore && !a.loadingMore {
		a.loadOlder()
	}
	return false
}

// loadAround asks for the page a message is on, and shows it as the history:
// a result from months ago is one request away rather than a page at a time.
// What is above the page is asked for as the reader goes back through it,
// and what is below it as they come forward, so that the chat is whole again
// in both directions.
func (a *App) loadAround() {
	id, seq := a.jumpID, a.jumpSeq
	chatID, chatType := a.current, a.currentType
	if id == "" || chatID == "" || a.loadingAround {
		return
	}
	a.loadingAround = true
	a.loading = true
	go func() {
		msgs, err := a.client.MessagesAround(chatID, chatType, msgPage, id, seq)
		a.update(func() {
			a.loadingAround = false
			if a.current != chatID {
				return
			}
			a.loading = false
			if err != nil {
				a.note = errorLine(err)
				// Where the message sits was not taken, so it is looked
				// for the long way, one page at a time; a chat with no
				// history to look through opens at its end instead.
				a.jumpSeq = 0
				if len(a.msgs) == 0 {
					a.jumpID = ""
					a.loadMessages(chatID, chatType)
					return
				}
				a.findJump()
				return
			}
			a.setAround(msgs, id)
		})
	}()
}

// setAround shows the page a jump landed on as the history, scrolled to the
// message itself.
func (a *App) setAround(msgs []*yh.Message, id string) {
	a.msgs = msgs
	// A page that came back full has more after it; one that did not has
	// reached the newest message.
	a.hasMore = len(msgs) > 0
	a.hasNewer = len(msgs) > msgPage
	a.loadedAt = 1 << 30
	a.moreAt, a.newerFrom = time.Time{}, ""
	a.jumpToEnd = false
	a.jumpTo = -1
	for i, m := range msgs {
		if m.MsgID == id {
			a.jumpTo = i
			a.flashID, a.flashAt = id, time.Now()
			break
		}
	}
	// Nothing here is unread in the way the button over the list counts:
	// the reader came to this page on purpose, and what is below it is
	// history they chose to leave.
	a.unreadFrom = len(msgs)
	a.readFrom = len(msgs)
	a.unreadOnOpen = 0
	a.jumpID, a.jumpSeq = "", 0
	if a.jumpTo < 0 {
		// The page did not hold what was asked for: the chat opens at its
		// end, as it does without a jump.
		a.loadMessages(a.current, a.currentType)
	}
}

// loadNewer puts the messages after the last one on screen below it: the
// history was opened in the middle of itself, and what follows the page is
// what makes it whole again as the reader comes forward.
func (a *App) loadNewer() {
	if a.loadingNewer || !a.hasNewer || a.current == "" || len(a.msgs) == 0 {
		return
	}
	chatID, chatType := a.current, a.currentType
	last := a.msgs[len(a.msgs)-1]
	a.loadingNewer, a.newerFrom = true, last.MsgID
	go func() {
		after, err := a.client.MessagesAfter(chatID, chatType, last.MsgSeq)
		a.update(func() {
			a.loadingNewer = false
			if a.current != chatID {
				return
			}
			if err != nil {
				a.note = errorLine(err)
				return
			}
			fresh := a.withoutKnown(after)
			if len(fresh) == 0 {
				// The page held nothing new: asking again would ask for
				// the same messages for ever.
				a.hasNewer = false
				return
			}
			a.msgs = append(a.msgs, fresh...)
			a.hasNewer = len(after) >= msgAfter
			a.trimMessages()
		})
	}()
}

// wantNewer reports whether the history should grow downwards: the reader
// has come to the last fifth of what is here, and there is more of it below
// because the chat was opened in the middle. The last message is what says
// whether asking would add anything, so a list resting at its end asks once
// rather than on every frame.
func (a *App) wantNewer(last, rows int) bool {
	if !a.hasNewer || a.loadingNewer || a.current == "" || len(a.msgs) == 0 {
		return false
	}
	if rows == 0 || last < 0 || last < rows-rows/5 {
		return false
	}
	return a.msgs[len(a.msgs)-1].MsgID != a.newerFrom
}

// flash reports how strongly a message is marked for having been jumped to,
// which fades over the second after it: a row that was found is a row that
// has to be seen.
func (a *App) flash(id string) float32 {
	if id == "" || id != a.flashID {
		return 0
	}
	k := 1 - float32(time.Since(a.flashAt))/float32(1200*time.Millisecond)
	if k <= 0 {
		a.flashID = ""
		return 0
	}
	return k
}

// openConvByID opens the chat with an id, whether or not it is in the list:
// a search finds messages of the chat that is open, which the list has.
func (a *App) openConvByID(id string, chatType int) {
	for _, conv := range a.convs {
		if conv.ChatID == id {
			a.open(conv)
			return
		}
	}
	a.open(yh.Conversation{ChatID: id, ChatType: chatType})
}

// msgKeep is how many messages a chat keeps in memory. A chat that has been
// read for hours would otherwise hold every message it ever loaded, and the
// pictures in them, which is the largest thing this client holds. What is
// dropped is the oldest: the pages before it are still on the server, and
// scrolling up asks for them again.
const msgKeep = 900

// trimMessages drops the oldest messages once a chat holds more than it
// keeps, and moves what is counted in rows along with them.
func (a *App) trimMessages() {
	if len(a.msgs) <= msgKeep {
		return
	}
	drop := len(a.msgs) - msgKeep
	kept := make([]*yh.Message, msgKeep)
	copy(kept, a.msgs[drop:])
	a.msgs = kept

	a.unreadFrom = max(a.unreadFrom-drop, 0)
	a.readFrom = max(a.readFrom-drop, 0)
	a.loadedAt = max(a.loadedAt-drop, 0)
	if a.jumpTo >= 0 {
		a.jumpTo = max(a.jumpTo-drop, 0)
	}
	// What was dropped is older than what is kept, so the pages before it
	// are still there to be asked for.
	a.hasMore = true
}

// sendFormatMenu is the menu that says what the next text is written in.
// Words, markdown and markup are the three the route takes: the words of a
// markdown message are drawn in the styles they ask for, and markup is read
// for its words here. The choice is for one message — send() puts it back
// to words as soon as the text has gone.
func sendFormatMenu(a *App) menu {
	return menu{
		width: 150,
		rows: []menuItem{
			{icon: iconText, label: "format.text",
				do: func() { a.sendType = yh.ContentText }},
			{icon: iconMarkdown, label: "format.markdown",
				do: func() { a.sendType = yh.ContentMarkdown }},
			{icon: iconHTML, label: "format.html",
				do: func() { a.sendType = yh.ContentHTML }},
		},
	}
}

// sendOptionsMenu is the menu behind the send button's arrow: which key
// sends the text. The tick is the icon of the row that is chosen, and the
// other row keeps the space so the two words line up as QQ has them. It is
// made afresh each time the arrow is pressed, so the tick is always on the
// key that is in force now.
func sendOptionsMenu(a *App) menu {
	choose := func(key string) func() {
		return func() {
			a.prefs.SendKey = key
			a.savePrefs()
		}
	}
	rows := []menuItem{
		{label: "send.enter", do: choose("enter")},
		{label: "send.ctrlEnter", do: choose("ctrlEnter")},
	}
	if a.prefs.SendKey == "ctrlEnter" {
		rows[1].icon = iconCheck
	} else {
		rows[0].icon = iconCheck
	}
	// Wide enough for the longer of the two lines: a menu that clips
	// "Ctrl + Enter" into an ellipsis is no way to choose a key.
	return menu{width: 248, above: true, rows: rows}
}
