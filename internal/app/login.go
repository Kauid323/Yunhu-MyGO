package app

import (
	"strings"

	"github.com/egoist/mygo"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The pages the panel's words open. This client has no pages of its own,
// so what it can point at is the documentation it is written against: a
// word that goes nowhere would be worse than one that goes there.
const (
	termsURL    = helpURL
	privacyURL  = helpURL
	registerURL = helpURL
	feedbackURL = helpURL
)

// loginName is what the panel calls the account: the one that was here
// last, the one being typed, or nothing at all.
func (a *App) loginName() string {
	switch {
	case a.remembered != nil && a.remembered.Name != "":
		return a.remembered.Name
	case a.email != "":
		return a.email
	case a.phone != "":
		return a.phone
	}
	return i18n.T("login.noAccount")
}

// loginAvatar is the face of that account, or "" when there is none to
// show: the panel draws its own mark then.
func (a *App) loginAvatar() string {
	if a.remembered == nil {
		return ""
	}
	return a.remembered.Avatar
}

// setLoginMode takes another way in. A picture to read out is asked for as
// the phone way is taken, as a code is not texted without one.
func (a *App) setLoginMode(mode int) {
	if a.loginMode == mode {
		return
	}
	a.loginMode = mode
	a.errText, a.hint = line{}, line{}
	a.loginAccounts = false
	if mode == modePassword && a.email == "" && a.remembered != nil {
		a.email = a.remembered.Email
	}
	if mode == modePhone && a.captchaID == "" {
		a.fetchCaptcha()
	}
}

// loginAgain signs in with the account that was here last, which the
// session kept a token for: the panel is a press away from the chat it
// left.
func (a *App) loginAgain() {
	if a.remembered == nil || a.remembered.Token == "" {
		a.setLoginMode(modePassword)
		return
	}
	a.update(func() { a.errText, a.hint = line{}, line{} })
	a.resume()
}

// loginAnother puts the form up for an account that has not been here: the
// circle beside the face is what asks for one.
func (a *App) loginAnother() {
	a.update(func() {
		a.email, a.password = "", ""
		a.loginMode = modePassword
		a.errText, a.hint = line{}, line{}
	})
}

// smsButton is what the button that asks for a code says, and how long
// until it can be pressed again: a code is texted once a minute.
func (a *App) smsButton() (string, int) {
	if left := a.smsLeft(); left > 0 {
		return i18n.T("login.resend", i18n.Num(left)), left
	}
	return i18n.T("login.sendCode"), 0
}

// submit is the button of the form: the number and the code texted to it,
// or the account and its password.
func (a *App) submit() {
	if a.loginMode == modePhone {
		a.loginWithPhone()
		return
	}
	a.login()
}

// formReady reports whether the form has what its button would send.
func (a *App) formReady() bool {
	if !a.prefs.Terms {
		return false
	}
	if a.loginMode == modePhone {
		return strings.TrimSpace(a.phone) != "" && a.phoneCode != "" &&
			a.smsCode != ""
	}
	return strings.TrimSpace(a.email) != "" && a.password != ""
}

// applyProxy sets the way the requests leave this machine, as the panel's
// network page says: no proxy, or one at the address given.
func (a *App) applyProxy() {
	kind, addr := a.prefs.Proxy, strings.TrimSpace(a.prefs.ProxyAt)
	if kind != "" && addr == "" {
		// A proxy with nowhere to go is no proxy: the requests would
		// fail with a message about an address nobody gave.
		kind = ""
	}
	_ = a.client.SetProxy(kind, addr)
}

// openRegister shows where an account is made.
func (a *App) openRegister() { _ = mygo.Shell.OpenExternal(registerURL) }

// openTerms shows what going in is taken to agree to.
func (a *App) openTerms() { _ = mygo.Shell.OpenExternal(termsURL) }

// openPrivacy shows what is done with what the account says.
func (a *App) openPrivacy() { _ = mygo.Shell.OpenExternal(privacyURL) }

// openFeedback shows where a fault in the client is reported.
func (a *App) openFeedback() { _ = mygo.Shell.OpenExternal(feedbackURL) }

// saveSession keeps what the next run needs to sign in again, and what the
// panel shows while it is not signed in.
func (a *App) saveSession(u *yh.User) {
	s := &yh.Session{
		Token:    a.client.Token,
		Email:    u.Email,
		Phone:    u.Phone,
		DeviceID: a.client.DeviceID,
		Base:     a.client.Base,
		Name:     u.Name,
		Avatar:   u.AvatarURL,
	}
	if a.prefs.Remember {
		// The password is kept on this machine so that a token that has
		// gone is a press rather than a form, as the client this one
		// copies keeps it.
		s.Password = a.password
	}
	_ = yh.SaveSession(s)
	a.remembered = s
}
