package app

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// composer is where a message is typed: the tools along its top, the text
// in the middle and the send button in its corner. It lies flat on the
// chat, told apart from it by a line, as QQ's does, rather than standing
// on it as a card.
func (a *App) composer(c *ui.Context) {
	ui.Column(c).Padding(4, 14, 8, 14).Gap(0).Background(qq.chatBg).
		BorderWidth(1, 0, 0, 0).BorderColor(qq.border).Children(func() {
		ready := strings.TrimSpace(a.input) != "" && a.current != ""

		tools := ui.Row(c).Gap(0).AlignItems(ui.Center)
		tools.Children(func() {
			// The stickers are the one tool that does something yet: a
			// press opens the account's saved ones over the row.
			a.toolAction(c, iconEmoji, a.stickerOpen, a.toggleStickers)
			for _, ic := range []*ui.SVG{iconScissors, iconFolder, iconImage, iconMic, iconAt, iconHistory} {
				a.toolIcon(c, ic)
			}
		})
		a.stickersPanel(c, tools)

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
			send := ui.PrimaryButton(c, i18n.T("chat.send")).Disabled(!ready).
				Background(qq.accent).Radius(5).Padding(3, 16).FontSize(12)
			if send.Clicked() && ready {
				a.send()
			}
		})
	})
}

// toolIcon is one of the composer's tools.
func (a *App) toolIcon(c *ui.Context, ic *ui.SVG) {
	box := ui.Box(c).Size(28, 28).Radius(5).Center().Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.inputBg)
	}
	box.Children(func() {
		ui.Icon(c, ic).FontSize(15).TextColor(qq.textMuted)
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
