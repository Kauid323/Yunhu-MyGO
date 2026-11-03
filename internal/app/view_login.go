package app

import (
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The sign-in panel is the window of the client this one copies: 320 by
// 460, with its own buttons in the corner and no title bar of the
// desktop's. Everything on it is laid out from the middle.
const (
	loginW, loginH = 320, 460
	// loginChrome is the strip the panel's own buttons sit in.
	loginChrome = 44
	// loginFieldW is the width of everything the panel lines up: the
	// fields, the button, and the words under them.
	loginFieldW = 255
	loginFieldH = 40
	// loginFaceBig is the face on the panel that shows an account, and
	// loginFaceSmall the one above the form.
	loginFaceBig   = 100
	loginFaceSmall = 80
	// loginPlus is the circle that adds an account, as the client draws
	// it: 38 points across, its middle 86 to the right of the face's.
	// loginPlusIn is where its left edge is when it is still in the middle
	// of the face, and loginPlusOut where it rests.
	loginPlus    = 38
	loginPlusIn  = loginFaceBig/2 - loginPlus/2
	loginPlusOut = 117
)

// The colours of the panel. It is drawn in these in either appearance: a
// sign-in screen is looked at before there is an account to have an
// appearance with, and the client it copies is purple in both.
var (
	loginTop    = ui.Hex("#2a1935")
	loginMid    = ui.Hex("#211940")
	loginLow    = ui.Hex("#2e1b2e")
	loginBottom = ui.Hex("#321c2a")

	loginInk   = ui.Hex("#eae8eb")
	loginDim   = ui.Hex("#a29aa6")
	loginFaint = ui.Hex("#8b8494")
	loginRing  = ui.Hex("#8e8796")

	loginField = ui.RGBA(255, 255, 255, 0.10)
	loginSoft  = ui.RGBA(255, 255, 255, 0.06)
	loginEdge  = ui.RGBA(255, 255, 255, 0.14)

	loginAction = ui.Hex("#0066cc")
	loginOff    = ui.Hex("#1d3063")
	loginOffInk = ui.Hex("#8f96ad")
	loginLink   = ui.Hex("#2d77e5")

	loginTipBg = ui.Hex("#37373a")
)

// plusTime is how long the circle that adds an account takes to come out
// of the middle of the face.
const plusTime = 180 * time.Millisecond

// loginView draws the sign-in panel: the account that was here last, the
// form for another, or the network's settings.
func (a *App) loginView(c *ui.Context) {
	ui.Box(c).Fill().LinearGradient(ui.LinearGradient{
		Angle: 180,
		Stops: []ui.GradientStop{
			{Color: loginTop, At: 0},
			{Color: loginMid, At: 0.42},
			{Color: loginLow, At: 0.74},
			{Color: loginBottom, At: 1},
		},
	}).Children(func() {
		if a.loginProxy {
			a.proxyPage(c)
		} else {
			a.loginBody(c)
		}
		a.loginChrome(c)
	})
}

// loginBody is the page of the panel: the account that was here last, or
// the form for one.
func (a *App) loginBody(c *ui.Context) {
	ui.Column(c).Fill().AlignItems(ui.Center).Children(func() {
		switch a.loginMode {
		case modePassword, modePhone:
			a.loginForm(c)
		default:
			a.loginAccount(c)
		}
	})
}

// loginAccount is the panel as it opens: the mark, the face of the account
// that was here last, its name, and the button that goes back in with it.
func (a *App) loginAccount(c *ui.Context) {
	// The strip the panel's own buttons sit in is over the page, so the
	// page starts below it.
	ui.Box(c).Height(loginChrome)
	a.loginWordmark(c)

	face := ui.Box(c).Size(loginFaceBig, loginFaceBig).Margin(30, 0, 0, 0)
	face.Children(func() {
		a.loginFace(c, loginFaceBig, a.loginName(), a.loginAvatar())
		a.loginAdd(c, face)
	})

	row := ui.Row(c).Gap(5).AlignItems(ui.Center).Margin(26, 0, 0, 0).
		Cursor(ui.CursorPointer)
	row.Children(func() {
		ui.Text(c, a.loginName()).FontSize(14.5).TextColor(loginInk).SingleLine()
		ui.Icon(c, iconChevron).FontSize(11).TextColor(loginDim)
	})
	if row.Clicked() {
		a.loginAccounts = !a.loginAccounts
	}
	if a.loginAccounts {
		a.loginAccountList(c, row)
	}

	ui.Row(c).Margin(17, 0, 0, 0).Children(func() {
		loginRadio(c, a.prefs.AutoLogin, "login.auto", func() {
			a.pickPrefs(func(p *prefs) { p.AutoLogin = !p.AutoLogin })
		})
	})

	ui.Row(c).Margin(32, 0, 0, 0).Children(func() {
		loginButton(c, i18n.T("login.submit"), !a.busy, a.loginAgain)
	})

	ui.Spacer(c)
	ui.Row(c).Gap(12).AlignItems(ui.Center).Padding(0, 0, 31, 0).Children(func() {
		loginLinkText(c, "login.link.password", func() { a.setLoginMode(modePassword) })
		loginDivider(c)
		loginLinkText(c, "login.link.register", a.openRegister)
	})
}

// loginForm is the panel asking for an account: the password it is signed
// in with, or the number and the code texted to it. The phone way has a
// field more than the password way, so its face is smaller and its lines
// closer together, to leave the words at the foot their room.
func (a *App) loginForm(c *ui.Context) {
	// Below the strip the panel's own buttons sit in.
	ui.Box(c).Height(loginChrome)

	faceSize, first, gap, radioGap, buttonGap, agreeGap :=
		float32(loginFaceSmall), float32(24), float32(12), float32(16),
		float32(30), float32(15)
	if a.loginMode == modePhone {
		faceSize, first, gap, radioGap, buttonGap, agreeGap =
			64, 18, 8, 12, 24, 12
	}
	face := ui.Box(c).Size(faceSize, faceSize).Margin(4, 0, 0, 0)
	face.Children(func() {
		a.loginFace(c, faceSize, a.loginName(), a.loginAvatar())
	})

	if a.loginMode == modePhone {
		a.phoneFields(c, first, gap)
	} else {
		a.passwordFields(c, first, gap)
	}

	ui.Row(c).Width(loginFieldW).Margin(radioGap, 0, 0, 0).Children(func() {
		loginRadio(c, a.prefs.AutoLogin, "login.auto", func() {
			a.pickPrefs(func(p *prefs) { p.AutoLogin = !p.AutoLogin })
		})
		ui.Spacer(c)
		loginCheck(c, a.prefs.Remember, "login.remember", func() {
			a.pickPrefs(func(p *prefs) { p.Remember = !p.Remember })
		})
	})

	ready := a.formReady()
	ui.Row(c).Margin(buttonGap, 0, 0, 0).Children(func() {
		loginButton(c, i18n.T("login.submit"), ready && !a.busy, a.submit)
	})

	a.loginAgreement(c, agreeGap)

	ui.Spacer(c)
	ui.Row(c).Gap(12).AlignItems(ui.Center).Padding(0, 0, 31, 0).Children(func() {
		if a.loginMode == modePhone {
			loginLinkText(c, "login.link.password", func() { a.setLoginMode(modePassword) })
		} else {
			loginLinkText(c, "login.link.phone", func() { a.setLoginMode(modePhone) })
		}
		loginDivider(c)
		loginLinkText(c, "login.link.register", a.openRegister)
	})
}

// passwordFields is the account and the password, with the way back to the
// accounts that have been here before beside the first.
func (a *App) passwordFields(c *ui.Context, first, gap float32) {
	box := ui.Box(c).Width(loginFieldW).Height(loginFieldH).Margin(first, 0, 0, 0).
		Radius(6).Background(loginField)
	box.Children(func() {
		ui.TextInput(c, &a.email).Fill().AutoFocus().
			Background(ui.Transparent).Border(0, ui.Transparent).FontSize(13.5).
			TextColor(loginInk)
		ui.Icon(c, iconChevron).FontSize(11).TextColor(loginDim).
			Absolute().Right(14).Top(14)
	})
	if box.Clicked() {
		a.loginAccounts = !a.loginAccounts
	}
	if a.loginAccounts {
		a.loginAccountList(c, box)
	}

	ui.Box(c).Width(loginFieldW).Height(loginFieldH).Margin(gap, 0, 0, 0).
		Radius(6).Background(loginField).Children(func() {
		ui.TextInput(c, &a.password).Fill().Password().
			Background(ui.Transparent).Border(0, ui.Transparent).FontSize(13.5).
			TextColor(loginInk)
	})
}

// phoneFields is the number, the picture that has to be read out before a
// code is texted to it, and the code itself.
func (a *App) phoneFields(c *ui.Context, first, gap float32) {
	ui.Box(c).Width(loginFieldW).Height(loginFieldH).Margin(first, 0, 0, 0).
		Radius(6).Background(loginField).Children(func() {
		ui.TextInput(c, &a.phone).Fill().AutoFocus().
			Placeholder(i18n.T("login.phoneHint")).
			Background(ui.Transparent).Border(0, ui.Transparent).FontSize(13.5).
			TextColor(loginInk)
	})

	ui.Row(c).Width(loginFieldW).Gap(10).Margin(gap, 0, 0, 0).Children(func() {
		ui.Box(c).Grow(1).Height(loginFieldH).Radius(6).Background(loginField).
			Children(func() {
				ui.TextInput(c, &a.phoneCode).Fill().
					Placeholder(i18n.T("login.captchaHint")).
					Background(ui.Transparent).Border(0, ui.Transparent).FontSize(13.5).
					TextColor(loginInk)
			})
		a.loginCaptcha(c)
	})

	ui.Row(c).Width(loginFieldW).Gap(10).Margin(gap, 0, 0, 0).Children(func() {
		ui.Box(c).Grow(1).Height(loginFieldH).Radius(6).Background(loginField).
			Children(func() {
				ui.TextInput(c, &a.smsCode).Fill().
					Placeholder(i18n.T("login.codeHint")).
					Background(ui.Transparent).Border(0, ui.Transparent).FontSize(13.5).
					TextColor(loginInk)
			})
		label, left := a.smsButton()
		if left > 0 {
			// The countdown is on the button: it counts itself down.
			c.After(time.Second)
		}
		box := ui.Box(c).Width(96).Height(loginFieldH).Radius(6).Center().
			Background(loginSoft)
		if left == 0 && !a.busy {
			box.Cursor(ui.CursorPointer)
			box.Background(loginAction)
		}
		box.Children(func() {
			ui.Text(c, label).FontSize(11.5).TextColor(loginInk)
		})
		if box.Clicked() && left == 0 && !a.busy {
			a.sendSmsCode()
		}
	})
}

// loginCaptcha is the picture that has to be read out. Pressed, another is
// asked for: one is read out once, and a wrong reading is worth another
// picture rather than the same one again.
func (a *App) loginCaptcha(c *ui.Context) {
	shot := ui.Box(c).Width(96).Height(loginFieldH).Radius(6).
		Background(loginSoft).Center().Cursor(ui.CursorPointer)
	shot.Children(func() {
		switch {
		case a.captcha != nil:
			bw, bh := ui.ImageSize(a.captcha)
			k := min(float32(88)/bw, float32(34)/bh)
			ui.Image(c, a.captcha).Width(bw * k).Height(bh * k).Radius(3)
		case a.captchaGo:
			ui.Spinner(c)
		default:
			ui.Text(c, i18n.T("login.captchaLoad")).FontSize(10).
				TextColor(loginDim)
		}
	})
	if shot.Clicked() {
		a.fetchCaptcha()
	}
}

// loginAgreement is the line under the button: what is agreed to by going
// in, with the two pages that say what that is.
func (a *App) loginAgreement(c *ui.Context, gap float32) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Margin(gap, 0, 0, 0).Children(func() {
		loginRadio(c, a.prefs.Terms, "", func() {
			a.pickPrefs(func(p *prefs) { p.Terms = !p.Terms })
		})
		ui.Text(c, i18n.T("login.agreed")).FontSize(11).TextColor(loginDim)
		loginLinkText(c, "login.terms", a.openTerms)
		ui.Text(c, i18n.T("login.and")).FontSize(11).TextColor(loginDim)
		loginLinkText(c, "login.privacy", a.openPrivacy)
	})
}

// loginWordmark is the panel's mark: the app's name, the first part of it
// in the blue the client's mark is drawn in and the rest in its pink, so
// that it reads as a mark rather than as a line of text.
func (a *App) loginWordmark(c *ui.Context) {
	name := i18n.T("app.name")
	head, tail := name, ""
	if n := utf8.RuneCountInString(name); n > 1 {
		r := []rune(name)
		head, tail = string(r[:n-1]), string(r[n-1:])
	}
	ui.Row(c).Height(26).AlignItems(ui.Center).Children(func() {
		ui.Text(c, head).FontSize(26).FontWeight(600).
			TextColor(ui.Hex("#7eddfe")).SingleLine()
		if tail != "" {
			ui.Text(c, tail).FontSize(26).FontWeight(600).
				TextColor(ui.Hex("#daaaff")).SingleLine()
		}
	})
}

// loginFace is the account's face on the panel: the picture when there is
// one, and the mark on white when there is nobody here yet.
func (a *App) loginFace(c *ui.Context, size float32, name, url string) {
	box := ui.Box(c).Size(size, size).Radius(size / 2).
		Background(ui.Hex("#ffffff")).Center()
	box.Children(func() {
		if url != "" {
			bmp := a.images.get(url, avatarPixels(size), a)
			if bmp != nil {
				ui.Image(c, bmp).Width(size).Height(size).Radius(size / 2)
				return
			}
		}
		mark := i18n.T("app.mark")
		n := utf8.RuneCountInString(mark)
		s := size * 0.36
		if n > 1 {
			s = size * 0.9 / (0.62 * float32(n))
		}
		ui.Text(c, mark).FontSize(s).Bold().TextColor(loginTop).SingleLine()
	})
}

// loginAdd is the circle that adds an account. It comes out of the middle
// of the face when the pointer settles on it, and goes back in when the
// pointer leaves: a second account is asked for by reaching for the face,
// not by a button that is always there.
func (a *App) loginAdd(c *ui.Context, face ui.Element) {
	// The circle is only seen once it has slid past the face, so what is
	// drawn is clipped where the face's middle is: the box the circle
	// lives in starts there, and its left edge is the face's middle.
	area := ui.Box(c).Absolute().Left(loginFaceBig / 2).Top(0).
		Width(loginPlusOut - loginFaceBig/2 + loginPlus).Height(loginFaceBig).
		ClipX().PassThrough()

	k := a.plusSlide(c, a.plusHover)
	var plus ui.Element
	area.Children(func() {
		if k <= 0.01 {
			return
		}
		// The circle starts on the face's middle, and its rest is as far
		// from that middle as the client draws it.
		x := -loginPlus/2 + (loginPlusOut-loginFaceBig/2+loginPlus/2)*k
		d := loginPlus * (0.55 + 0.45*k)
		box := ui.Box(c).Absolute().Left(x+(loginPlus-d)/2).
			Top((loginFaceBig-d)/2).Size(d, d).Radius(d/2).
			Border(max(1, 2*k), ui.RGBA(255, 255, 255, k)).
			Background(ui.RGBA(55, 55, 55, 0.94)).Center().
			Cursor(ui.CursorPointer).Opacity(k)
		box.Children(func() {
			ui.Icon(c, iconPlus).FontSize(27).TextColor(ui.Hex("#8f8f90"))
		})
		if box.Clicked() {
			a.loginAnother()
		}
		plus = box
	})

	if k > 0.45 {
		tip := ui.Box(c).Absolute().Left(101).Top(3).Height(20).Padding(0, 8).
			Radius(6).Background(loginTipBg).Center().Opacity(k)
		tip.Children(func() {
			ui.Text(c, i18n.T("login.multi")).FontSize(11).TextColor(loginInk).
				SingleLine()
		})
	}

	// Where the pointer is is noted for the frame that follows: an element
	// knows where it was in the frame before this one.
	a.plusHover = face.Hovered() || plus.Hovered()
}

// plusSlide is how far the circle that adds an account has come out: 0 in
// the middle of the face, 1 at its rest. The pointer is noted after the
// face is drawn, so the frame that follows is the one that moves it.
func (a *App) plusSlide(c *ui.Context, on bool) float32 {
	target := float32(0)
	if on {
		target = 1
	}
	if a.plusOn != on {
		a.plusOn, a.plusFrom, a.plusAt = on, a.plusK, c.Now()
	}
	if a.plusK == target {
		return target
	}
	if a.plusAt.IsZero() {
		a.plusFrom, a.plusAt = a.plusK, c.Now()
	}
	k := float32(c.Now().Sub(a.plusAt)) / float32(plusTime)
	if k >= 1 {
		a.plusK, a.plusAt = target, time.Time{}
		return target
	}
	a.plusK = a.plusFrom + (target-a.plusFrom)*ui.EaseOut(k)
	c.After(16 * time.Millisecond)
	return a.plusK
}

// loginChrome is the strip at the top of the panel: the way back from a
// page of its own, the menu, and the button that closes it. It is the
// panel's title bar, and a drag on it moves the window.
func (a *App) loginChrome(c *ui.Context) {
	bar := ui.Row(c).Absolute().Left(0).Top(0).Right(0).Height(loginChrome).
		Padding(0, 10, 0, 0).AlignItems(ui.Center).DragWindow()
	bar.Children(func() {
		if a.loginProxy {
			back := ui.Box(c).Size(28, 28).Radius(6).Center().
				Cursor(ui.CursorPointer)
			if back.Hovered() {
				back.Background(loginSoft)
			}
			back.Children(func() {
				ui.Icon(c, iconChevronLeft).FontSize(14).TextColor(loginDim)
			})
			if back.Clicked() {
				a.loginProxy, a.proxyOpen = false, false
			}
			ui.Text(c, i18n.T("login.proxy.title")).FontSize(14.5).
				TextColor(loginInk).Grow(1).TextAlign(ui.Center)
			ui.Box(c).Size(28, 28)
		} else {
			ui.Spacer(c)
			more := a.loginChromeButton(c, iconMenu, nil)
			if more.Clicked() {
				a.loginMenu.open = !a.loginMenu.open
			}
			// The panel's menu is the window's menu: the same component
			// the rail opens, so that it comes and goes the same way.
			showMenu(c, more, &a.loginMenu)
		}
		a.loginChromeButton(c, iconClose, func() {
			if a.win != nil && !a.win.IsDestroyed() {
				a.win.Close()
			}
		})
	})
}

// loginChromeButton is one of the buttons in the corner of the panel.
func (a *App) loginChromeButton(c *ui.Context, ic *ui.SVG, do func()) ui.Element {
	box := ui.Box(c).Size(28, 28).Radius(6).Center().Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(loginSoft)
	}
	box.Children(func() {
		ui.Icon(c, ic).FontSize(15).TextColor(loginDim)
	})
	if box.Clicked() && do != nil {
		do()
	}
	return box
}

// loginAccountList is the list of accounts that have been here: the one
// that is kept, and the way to another. It hangs under whatever asked for
// it, as the window's menus hang off their buttons.
func (a *App) loginAccountList(c *ui.Context, anchor ui.Element) {
	ui.PopoverBase(c, anchor, &a.loginAccounts, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorBottomLeft, ui.AnchorTopLeft).Top(6).
			Width(loginFieldW).Padding(4).Radius(8).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 8, 22, 0, ui.RGBA(0, 0, 0, 0.30)).
			Transition(menuArrives)
		panel.Children(func() {
			if a.remembered != nil {
				row := ui.Row(c).Height(38).Padding(0, 8).Gap(8).
					AlignItems(ui.Center).Radius(6).Cursor(ui.CursorPointer)
				if row.Hovered() {
					row.Background(qq.menuHover)
				}
				row.Children(func() {
					a.loginFace(c, 24, a.loginName(), a.loginAvatar())
					ui.Text(c, a.loginName()).FontSize(12.5).TextColor(qq.menuText).
						SingleLine()
				})
				if row.Clicked() {
					a.loginAccounts = false
					a.setLoginMode(modeAccount)
				}
			}
			row := ui.Row(c).Height(32).Padding(0, 8).Gap(8).AlignItems(ui.Center).
				Radius(6).Cursor(ui.CursorPointer)
			if row.Hovered() {
				row.Background(qq.menuHover)
			}
			row.Children(func() {
				ui.Icon(c, iconPlus).FontSize(12).TextColor(qq.menuTextMuted)
				ui.Text(c, i18n.T("login.addAccount")).FontSize(12).
					TextColor(qq.menuText).SingleLine()
			})
			if row.Clicked() {
				a.loginAccounts = false
				a.loginAnother()
			}
		})
	})
}

// proxyPage is the network's settings: how the requests leave this
// machine, which is what a network that needs a proxy is set here for.
func (a *App) proxyPage(c *ui.Context) {
	ui.Column(c).Fill().Padding(0, 28, 0, 28).Children(func() {
		ui.Text(c, i18n.T("login.proxy.type")).FontSize(13).
			TextColor(loginInk).Margin(66, 0, 0, 0)

		box := ui.Box(c).FillWidth().Height(36).Margin(10, 0, 0, 0).Radius(6).
			Background(loginField).Cursor(ui.CursorPointer)
		box.Children(func() {
			ui.Text(c, i18n.T("login.proxy."+a.proxyKind())).FontSize(13).
				TextColor(loginInk).Absolute().Left(12).Top(9)
			ui.Icon(c, iconChevron).FontSize(12).TextColor(loginDim).
				Absolute().Right(12).Top(11)
			if a.proxyOpen {
				// The kinds are listed over what is under the box rather
				// than pushing it down: a list that moves the page as it
				// opens is a list that moves the thing it came out of.
				list := ui.Box(c).Absolute().Top(40).Left(0).Right(0).
					Padding(4).Radius(6).Background(qq.menuBg).
					Border(1, qq.menuLine).
					Shadow(0, 8, 22, 0, ui.RGBA(0, 0, 0, 0.30))
				list.Children(func() {
					for _, kind := range []string{"none", "http", "socks5"} {
						row := ui.Row(c).Height(31).Padding(0, 10).
							AlignItems(ui.Center).Radius(4).
							Cursor(ui.CursorPointer)
						if row.Hovered() {
							row.Background(qq.menuHover)
						}
						row.Children(func() {
							ui.Text(c, i18n.T("login.proxy."+kind)).FontSize(13).
								TextColor(qq.menuText).Grow(1)
							if a.proxyKind() == kind {
								ui.Icon(c, iconCheck).FontSize(12).
									TextColor(qq.menuTextMuted)
							}
						})
						if row.Clicked() {
							a.pickPrefs(func(p *prefs) {
								if kind == "none" {
									p.Proxy = ""
									return
								}
								p.Proxy = kind
							})
							a.proxyOpen = false
						}
					}
				})
			}
		})
		if box.Clicked() {
			a.proxyOpen = !a.proxyOpen
		}

		if a.prefs.Proxy != "" {
			ui.Text(c, i18n.T("login.proxy.addr")).FontSize(13).
				TextColor(loginInk).Margin(18, 0, 0, 0)
			ui.Box(c).FillWidth().Height(36).Margin(10, 0, 0, 0).Radius(6).
				Background(loginField).Children(func() {
				ui.TextInput(c, &a.prefs.ProxyAt).Fill().
					Placeholder("127.0.0.1:7890").
					Background(ui.Transparent).Border(0, ui.Transparent).
					FontSize(13).TextColor(loginInk)
			})
		}

		ui.Spacer(c)
		ui.Row(c).Gap(12).Justify(ui.End).Padding(0, 0, 28, 0).Children(func() {
			if loginSmallButton(c, i18n.T("login.proxy.ok"), true) {
				a.applyProxy()
				a.loginProxy = false
			}
			if loginSmallButton(c, i18n.T("login.proxy.cancel"), false) {
				a.loginProxy, a.proxyOpen = false, false
			}
		})
	})
}

// proxyKind is which proxy the panel is set to, as the page names it.
func (a *App) proxyKind() string {
	if a.prefs.Proxy == "" {
		return "none"
	}
	return a.prefs.Proxy
}

// loginRadio is the round tick of the panel: a ring that is filled when it
// is taken, and the words beside it. An empty label draws the tick alone.
func loginRadio(c *ui.Context, on bool, label string, do func()) {
	row := ui.Row(c).Gap(7).AlignItems(ui.Center).Cursor(ui.CursorPointer)
	row.Children(func() {
		ring := ui.Box(c).Size(14, 14).Radius(7).Center()
		if on {
			ring.Border(1, loginAction)
			ring.Children(func() {
				ui.Box(c).Size(6, 6).Radius(3).Background(loginAction)
			})
		} else {
			ring.Border(1, loginRing)
		}
		if label != "" {
			ui.Text(c, i18n.T(label)).FontSize(11.5).TextColor(loginDim)
		}
	})
	if row.Clicked() {
		do()
	}
}

// loginCheck is the square tick of the panel: the same idea as the round
// one, for a setting that is on or off rather than one of a set.
func loginCheck(c *ui.Context, on bool, label string, do func()) {
	row := ui.Row(c).Gap(7).AlignItems(ui.Center).Cursor(ui.CursorPointer)
	row.Children(func() {
		box := ui.Box(c).Size(14, 14).Radius(3).Center()
		if on {
			box.Background(loginAction).Children(func() {
				ui.Icon(c, iconCheck).FontSize(9).TextColor(ui.Hex("#ffffff"))
			})
		} else {
			box.Border(1, loginRing)
		}
		ui.Text(c, i18n.T(label)).FontSize(11.5).TextColor(loginDim)
	})
	if row.Clicked() {
		do()
	}
}

// loginButton is the panel's button: the one that goes in, and the ones
// that answer a page of it.
func loginButton(c *ui.Context, label string, enabled bool, do func()) {
	bg := loginAction
	if !enabled {
		bg = loginOff
	}
	box := ui.Box(c).Width(loginFieldW).Height(loginFieldH).Radius(6).
		Background(bg).Center()
	if enabled {
		box.Cursor(ui.CursorPointer)
	}
	box.Children(func() {
		ink := ui.Hex("#ffffff")
		if !enabled {
			ink = loginOffInk
		}
		ui.Text(c, label).FontSize(16).TextColor(ink)
	})
	if enabled && box.Clicked() {
		do()
	}
}

// loginSmallButton is one of the two answers at the foot of a page of the
// panel. It reports whether it was pressed.
func loginSmallButton(c *ui.Context, label string, primary bool) bool {
	box := ui.Box(c).Width(76).Height(32).Radius(6).Center().
		Cursor(ui.CursorPointer)
	if primary {
		box.Background(loginAction)
	} else {
		box.Background(loginSoft)
	}
	box.Children(func() {
		ui.Text(c, label).FontSize(13).TextColor(loginInk)
	})
	return box.Clicked()
}

// loginLinkText is one of the words at the foot of the panel, which goes
// somewhere when it is pressed.
func loginLinkText(c *ui.Context, key string, do func()) {
	t := ui.Text(c, i18n.T(key)).FontSize(13).TextColor(loginLink).
		Cursor(ui.CursorPointer).SingleLine()
	if t.Hovered() {
		t.Underline()
	}
	if t.Clicked() {
		do()
	}
}

// loginDivider is the line between two of those words.
func loginDivider(c *ui.Context) {
	ui.Text(c, "|").FontSize(12).TextColor(ui.RGBA(255, 255, 255, 0.18))
}

// mark is the app's mark, in the square at the top of a card. It is as
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
