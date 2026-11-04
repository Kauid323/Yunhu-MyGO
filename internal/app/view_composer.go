package app

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// composer is where a message is typed: the tools along its top, the text
// in the middle and the send button in its corner. It lies flat on the
// chat, told apart from it by a line, as QQ's does, rather than standing
// on it as a card.
func (a *App) composer(c *ui.Context) {
	ui.Column(c).Padding(4, 14, 8, 14).Gap(0).Background(qq.chatBg).
		BorderWidth(1, 0, 0, 0).BorderColor(qq.border).Children(func() {
		ready := strings.TrimSpace(a.input) != "" && a.current != ""

		// What the next message answers, if it answers one: the words it
		// answers sit over the field, and go with the text when it is sent.
		if a.quote != nil {
			a.quoteBar(c)
		}

		tools := ui.Row(c).Gap(0).AlignItems(ui.Center)
		tools.Children(func() {
			// The tools that put something into the message: a sticker, a
			// picture, a clip, a file, and in a group the way to name
			// someone. The clip is its own tool because it goes to its own
			// bucket, even though a file dialog could pick one as well.
			a.toolAction(c, iconEmoji, a.stickerOpen, a.toggleStickers)
			a.toolAction(c, iconImage, false, a.pickImages)
			a.toolAction(c, iconVideo, false, a.pickVideos)
			a.toolAction(c, iconFolder, false, a.pickFiles)
			if a.currentType == yh.ChatGroup {
				a.toolAction(c, iconAt, a.mentionOpen, a.toggleMention)
			}
			// What the next text is written in: the button says which,
			// and the menu beside it chooses another. It is lit while
			// anything but words is chosen, so a markdown message is not
			// sent by surprise.
			showMenu(c, a.formatButton(c), &a.formatMenu)
			// The clock opens the chat's history: what was said in it,
			// searched. It sits at the far end of the row, apart from the
			// tools that put something into the message.
			ui.Spacer(c)
			a.toolAction(c, iconHistory, false, a.OpenSearch)
		})
		a.stickersPanel(c, tools)
		a.mentionPanel(c, tools)

		in := ui.TextArea(c, &a.input).Lines(1, 6).FontSize(13).
			Background(qq.chatBg).Border(0, qq.chatBg).AutoFocus()
		// Enter sends and Ctrl+Enter leaves a line, or the other way
		// round, as the settings say. The key is taken here rather than
		// left to the field, which would send whatever it was told to.
		in.HandleInput(func(ev ui.InputEvent) bool {
			// The "/" the settings offer opens the stickers, as the
			// panel this client copies does: the character itself is
			// not typed, it is the way in.
			if ev.Kind == ui.InputText && ev.Text == "/" && a.input == "" && a.prefs.Slash {
				a.toggleStickers()
				return true
			}
			if ev.Kind != ui.InputKeyDown || ev.Key != ui.KeyEnter {
				return false
			}
			sends := a.prefs.SendKey != "ctrlEnter"
			if held := ev.Mods&ui.Ctrl != 0; held == sends {
				// The other key: Enter leaves a line, or Ctrl+Enter
				// sends, and the field does what it always does.
				return false
			}
			if ready {
				a.send()
			}
			return true
		})

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spacer(c)
			a.sendButton(c, ready)
		})
	})
}

// toolAction is a tool that does something: lit while what it opens is open,
// as the stickers are while the panel over the row is.
func (a *App) toolAction(c *ui.Context, ic *ui.SVG, lit bool, do func()) {
	box := ui.Box(c).Size(28, 28).Radius(5).Center().Cursor(ui.CursorPointer)
	switch {
	case lit:
		box.Background(qq.inputBg)
	case box.Hovered():
		box.Background(qq.inputBg)
	}
	box.Children(func() {
		col := qq.textMuted
		if lit {
			col = qq.accent
		}
		ui.Icon(c, ic).FontSize(15).TextColor(col)
	})
	if box.Clicked() {
		do()
	}
}

// sendButton is the send control as QQ draws it: one blue rounded button,
// the word on the left and an arrow behind a hairline on the right. The
// word sends the text, the arrow opens the choice of which key sends it.
// It is one control rather than a button and a smaller one beside it, so
// both halves are the same height and there is nothing to line up.
func (a *App) sendButton(c *ui.Context, ready bool) {
	btn := ui.Box(c).Height(32).Radius(10).Background(qq.accent).
		Cursor(ui.CursorPointer)
	btn.Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			left := ui.Box(c).Height(32).Padding(0, 16).Center().
				Cursor(ui.CursorPointer)
			left.Children(func() {
				t := ui.Text(c, i18n.T("chat.send")).FontSize(12.5).
					TextColor(ui.RGBA(255, 255, 255, 1)).SingleLine()
				if !ready {
					t.Opacity(0.62)
				}
			})
			if ready && left.Clicked() {
				a.send()
			}
			ui.Box(c).Width(1).Height(15).
				Background(ui.RGBA(255, 255, 255, 0.32))
			// 32 high and 94 wide all together, as QQ's is: the word in
			// its padding, a hairline, then the arrow's own square.
			right := ui.Box(c).Width(36).Height(32).Center().
				Cursor(ui.CursorPointer)
			arrow := iconChevron
			if a.sendMenu.open {
				arrow = iconChevronUp
			}
			right.Children(func() {
				ui.Icon(c, arrow).FontSize(13).
					TextColor(ui.RGBA(255, 255, 255, 0.95))
			})
			if right.Clicked() {
				if a.sendMenu.open {
					a.sendMenu.open = false
				} else {
					// Made afresh so the tick is on the key in force
					// now, which the settings window may have changed.
					a.sendMenu = sendOptionsMenu(a)
					a.sendMenu.open = true
				}
			}
		})
	})
	showMenu(c, btn, &a.sendMenu)
}

// formatButton is the tool that says what the next text is written in. The
// word is the label rather than a glyph: three formats look much alike as
// small icons, and the one that will be used should not be a puzzle.
func (a *App) formatButton(c *ui.Context) ui.Element {
	kind := yh.SendKind(a.sendType)
	chosen := kind != yh.ContentText
	// The same 28 DIP as the icon tools, and Center() as they have it:
	// a box without it puts its row against the top edge, which reads as
	// the word sitting higher than the icons beside it.
	box := ui.Box(c).Height(28).Padding(0, 8).Radius(5).Center().
		Cursor(ui.CursorPointer)
	if chosen || box.Hovered() {
		box.Background(qq.inputBg)
	}
	ic := iconText
	switch kind {
	case yh.ContentMarkdown:
		ic = iconMarkdown
	case yh.ContentHTML:
		ic = iconHTML
	}
	box.Children(func() {
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			col := qq.textMuted
			if chosen {
				col = qq.accent
			}
			ui.Icon(c, ic).FontSize(13).TextColor(col)
			ui.Text(c, a.formatLabel()).FontSize(10.5).TextColor(col).SingleLine()
		})
	})
	if box.Clicked() {
		a.formatMenu.open = !a.formatMenu.open
	}
	return box
}

// formatLabel is the word the button says: what the next text will be sent
// as, words when the composer has not been asked for a format.
func (a *App) formatLabel() string {
	switch yh.SendKind(a.sendType) {
	case yh.ContentMarkdown:
		return i18n.T("format.markdown")
	case yh.ContentHTML:
		return i18n.T("format.html")
	default:
		return i18n.T("format.text")
	}
}

// quoteBar is the message being answered, over the field: the same line
// down the left as the quote in a bubble has, with the way to stop
// answering it at its right.
func (a *App) quoteBar(c *ui.Context) {
	m := a.quote
	ui.Row(c).Height(26).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Width(2).FillHeight().Background(qq.quoteBar)
		ui.Text(c, quoteTextOf(m)).FontSize(11).TextColor(qq.textMuted).
			Grow(1).MinWidth(0).SingleLine()
		box := ui.Box(c).Size(18, 18).Radius(4).Center().Cursor(ui.CursorPointer)
		if box.Hovered() {
			box.Background(qq.inputBg)
		}
		box.Children(func() {
			ui.Icon(c, iconClose).FontSize(10).TextColor(qq.textMuted)
		})
		if box.Clicked() {
			a.quote = nil
		}
	})
}
