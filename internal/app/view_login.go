package app

import (
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// loginView is the sign-in screen: a password, a phone number and the code
// texted to it, or a token.
func (a *App) loginView(c *ui.Context) {
	ui.Column(c).Fill().Center().Background(qq.chatBg).Padding(24).Children(func() {
		ui.Column(c).Width(340).Padding(30, 28).Gap(14).Radius(12).
			Background(qq.panelBg).Border(1, qq.border).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(34, 34).Radius(9).Center().Background(qq.accent).Children(func() {
					a.mark(c)
				})
				ui.Column(c).Gap(1).Children(func() {
					ui.Text(c, i18n.T("app.name")).FontSize(17).Bold().TextColor(qq.text)
					ui.Text(c, i18n.T("login.tagline")).FontSize(10.5).
						TextColor(qq.textFaint).Wrap()
				})
			})
			ui.Box(c).Height(1).Background(qq.border)

			a.loginTabs(c)

			switch a.loginMode {
			case modeToken:
				ui.Text(c, i18n.T("login.token")).FontSize(11).TextColor(qq.textMuted)
				in := ui.TextInput(c, &a.token).Placeholder("nj104901-xxxx-xxxx-xxxx").AutoFocus().
					Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5)
				if in.Submitted() && !a.busy {
					a.loginToken()
				}
			case modePhone:
				a.loginPhoneFields(c)
			default:
				ui.Text(c, i18n.T("login.email")).FontSize(11).TextColor(qq.textMuted)
				email := ui.TextInput(c, &a.email).Placeholder("you@example.com").AutoFocus().
					Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5)
				ui.Text(c, i18n.T("login.password")).FontSize(11).TextColor(qq.textMuted)
				pass := ui.TextInput(c, &a.password).Placeholder(i18n.T("login.passwordHint")).Password().
					Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5)
				if (email.Submitted() || pass.Submitted()) && !a.busy {
					a.login()
				}
			}

			if !a.hint.empty() {
				ui.Text(c, a.hint.String()).FontSize(11.5).TextColor(qq.accent).Wrap()
			}
			if !a.errText.empty() {
				ui.Text(c, a.errText.String()).FontSize(11.5).TextColor(qq.error).Wrap()
			}

			ui.Row(c).Gap(10).AlignItems(ui.Center).Margin(2, 0, 0, 0).Children(func() {
				button := ui.PrimaryButton(c, a.submitLabel()).Disabled(a.busy).
					Background(qq.accent).Radius(6).Padding(7, 20).FontSize(12.5)
				switch a.loginMode {
				case modeToken:
					button.OnClick(a.loginToken)
				case modePhone:
					button.OnClick(a.loginWithPhone)
				default:
					button.OnClick(a.login)
				}
				if a.busy {
					ui.Spinner(c)
				}
				ui.Spacer(c)
			})

			ui.Text(c, i18n.T("login.footnote")).FontSize(10).
				TextColor(qq.textFaint).Wrap()
		})
	})
}

// loginTabs is the three ways in, one of them taken.
func (a *App) loginTabs(c *ui.Context) {
	keys := [...]string{"login.tabPassword", "login.tabPhone", "login.tabToken"}
	ui.Row(c).Gap(16).Children(func() {
		for i, key := range keys {
			t := ui.Text(c, i18n.T(key)).FontSize(12).Cursor(ui.CursorPointer)
			if a.loginMode == i {
				t.Bold().TextColor(qq.accent)
			} else {
				t.TextColor(qq.textMuted)
			}
			if t.Clicked() {
				a.setLoginMode(i)
			}
		}
	})
}

// setLoginMode takes another way in. A picture to read out is asked for as
// the phone way is taken, as a code is not texted without one.
func (a *App) setLoginMode(mode int) {
	if a.loginMode == mode {
		return
	}
	a.loginMode = mode
	a.errText, a.hint = line{}, line{}
	if mode == modePhone && a.captchaID == "" {
		a.fetchCaptcha()
	}
}

// submitLabel is what the button at the foot of the card says, which the
// way in decides.
func (a *App) submitLabel() string {
	if a.loginMode == modeToken {
		return i18n.T("login.submitToken")
	}
	return i18n.T("login.submit")
}

// loginPhoneFields is the number, the picture that has to be read out
// before a code is texted to it, and the code itself.
func (a *App) loginPhoneFields(c *ui.Context) {
	ui.Text(c, i18n.T("login.phone")).FontSize(11).TextColor(qq.textMuted)
	phone := ui.TextInput(c, &a.phone).Placeholder(i18n.T("login.phoneHint")).AutoFocus().
		Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5)

	ui.Text(c, i18n.T("login.captcha")).FontSize(11).TextColor(qq.textMuted)
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		said := ui.TextInput(c, &a.phoneCode).Placeholder(i18n.T("login.captchaHint")).
			Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5).Grow(1)
		a.captchaPicture(c)
		if said.Submitted() && !a.busy {
			a.sendSmsCode()
		}
	})

	ui.Text(c, i18n.T("login.code")).FontSize(11).TextColor(qq.textMuted)
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		code := ui.TextInput(c, &a.smsCode).Placeholder(i18n.T("login.codeHint")).
			Background(qq.inputBg).Border(1, qq.border).Radius(6).FontSize(12.5).Grow(1)
		label, left := a.smsButton()
		if left > 0 {
			// The countdown is on the button: it counts itself down.
			c.After(time.Second)
		}
		b := ui.Button(c, label).Disabled(left > 0 || a.busy).Radius(6).
			Padding(5, 10).FontSize(11.5)
		if b.Clicked() {
			a.sendSmsCode()
		}
		if (phone.Submitted() || code.Submitted()) && !a.busy {
			a.loginWithPhone()
		}
	})
}

// captchaPicture is the picture that has to be read out. Pressed, another
// is asked for: one is read out once, and a wrong reading is worth another
// picture rather than the same one again.
func (a *App) captchaPicture(c *ui.Context) {
	// The box is the slot with room for its line: a picture the size of
	// the box would be drawn over the line.
	shot := ui.Box(c).Size(captchaW+2, captchaH+2).Radius(4).Background(qq.inputBg).
		Border(1, qq.border).Center().Cursor(ui.CursorPointer)
	shot.Children(func() {
		switch {
		case a.captcha != nil:
			// The picture is drawn as large as the slot holds it, in its
			// own shape: the API's pictures are wider than they are tall,
			// and one drawn in a slot of another shape is squashed.
			bw, bh := ui.ImageSize(a.captcha)
			k := min(captchaW/bw, captchaH/bh)
			ui.Image(c, a.captcha).Width(bw * k).Height(bh * k).Radius(3)
		case a.captchaGo:
			ui.Spinner(c)
		default:
			ui.Text(c, i18n.T("login.captchaLoad")).FontSize(10).TextColor(qq.textFaint)
		}
	})
	if shot.Clicked() {
		a.fetchCaptcha()
	}
}

// The slot a picture to read out is drawn in: wide enough for four
// characters to be read without squinting, and short enough to sit beside
// the field it goes with.
const (
	captchaW = 132
	captchaH = 30
)

// smsButton is what the button that asks for a code says, and how long
// until it can be pressed again: a code is texted once a minute.
func (a *App) smsButton() (string, int) {
	if left := a.smsLeft(); left > 0 {
		return i18n.T("login.resend", i18n.Num(left)), left
	}
	return i18n.T("login.sendCode"), 0
}

// mark is the app's mark, in the square at the top of the card. It is as
// large as fits: one character fills the square, and a word of more is
// drawn smaller rather than spilling out of it, as the dictionary decides
// what the mark is.
func (a *App) mark(c *ui.Context) {
	mark := i18n.T("app.mark")
	size := float32(17)
	if n := utf8.RuneCountInString(mark); n > 1 {
		// A letter is about six tenths of its size wide, and the square
		// has about three tenths of its own to spare.
		size = 34 * 0.9 / (0.62 * float32(n))
		if size > 17 {
			size = 17
		}
	}
	ui.Text(c, mark).FontSize(size).Bold().TextColor(qq.accentText).SingleLine()
}
