package app

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// resume signs in again with the saved token, if there is one.
func (a *App) resume() {
	s, err := yh.LoadSession()
	if err != nil || s.Token == "" {
		a.update(func() { a.note = line{} })
		return
	}
	a.client.Token = s.Token
	if s.DeviceID != "" {
		a.client.DeviceID = s.DeviceID
	}
	if s.Base != "" {
		a.client.Base = s.Base
	}
	a.email = s.Email
	a.phone = s.Phone
	a.update(func() { a.busy = true; a.note = say("login.restoring") })

	go func() {
		u, err := a.client.LoginToken(s.Token)
		if err != nil {
			a.update(func() {
				a.busy = false
				a.note = line{}
				a.errText = say("login.expired")
				// The account that was signed in is the one coming
				// back: a number with no email behind it is signed in
				// with again the way it was, with the number already
				// typed in.
				if s.Phone != "" && s.Email == "" {
					a.loginMode = modePhone
				}
			})
			return
		}
		a.afterLogin(u)
	}()
}

// afterLogin shows the chat screen and loads what it needs.
func (a *App) afterLogin(u *yh.User) {
	a.update(func() {
		a.me = u
		a.screen = screenChat
		a.busy = false
		a.errText = line{}
		a.note = say("login.restored")
	})
	_ = yh.SaveSession(&yh.Session{
		Token:    a.client.Token,
		Email:    u.Email,
		Phone:    u.Phone,
		DeviceID: a.client.DeviceID,
		Base:     a.client.Base,
	})
	a.loadProfile()
	a.refreshConversations()
	a.connect(u.ID)
}

// loadProfile asks for what the account says of itself, which the route it
// was read from does not carry: it is on the account's profile, and the
// window shows it under the name. A profile that does not come leaves the
// account as it is rather than failing the sign-in.
func (a *App) loadProfile() {
	go func() {
		intro, err := a.client.Introduction()
		if err != nil || intro == "" {
			return
		}
		a.update(func() {
			if a.me != nil {
				a.me.Introduction = intro
			}
		})
	}()
}

// login signs in with the email and password typed.
func (a *App) login() {
	email := strings.TrimSpace(a.email)
	password := a.password
	if email == "" || password == "" {
		a.update(func() { a.errText = say("login.needCredentials") })
		return
	}
	a.update(func() { a.busy = true; a.errText = line{} })
	go func() {
		token, err := a.client.LoginEmail(email, password)
		if err != nil {
			a.update(func() { a.busy = false; a.errText = errorLine(err) })
			return
		}
		a.client.Token = token
		u, err := a.client.UserInfo()
		if err != nil {
			a.update(func() { a.busy = false; a.errText = errorLine(err) })
			return
		}
		a.update(func() { a.password = "" })
		a.afterLogin(u)
	}()
}

// loginToken signs in with a token pasted in.
func (a *App) loginToken() {
	token := strings.TrimSpace(a.token)
	if token == "" {
		a.update(func() { a.errText = say("login.needToken") })
		return
	}
	a.update(func() { a.busy = true; a.errText = line{} })
	go func() {
		u, err := a.client.LoginToken(token)
		if err != nil {
			a.update(func() { a.busy = false; a.errText = errorLine(err) })
			return
		}
		a.afterLogin(u)
	}()
}

// smsWait is how long after a code is texted another may be asked for.
const smsWait = 60 * time.Second

// fetchCaptcha asks for a picture to read out, which a code is not texted
// without. A picture is read once: another is asked for when the last was
// read wrong, or when the reader presses it.
func (a *App) fetchCaptcha() {
	if a.captchaGo {
		return
	}
	a.update(func() { a.captchaGo = true; a.captcha = nil })
	go func() {
		c, err := a.client.Captcha()
		var bmp *ui.Bitmap
		if err == nil {
			bmp, err = ui.DecodeBitmap(c.Image)
		}
		a.update(func() {
			a.captchaGo = false
			if err != nil {
				a.captchaID, a.captcha = "", nil
				a.errText = errorLine(err)
				return
			}
			a.captchaID, a.captcha = c.ID, bmp
		})
	}()
}

// sendSmsCode asks for a code to be texted to the number typed, having
// read out the picture that came with it.
func (a *App) sendSmsCode() {
	phone := strings.TrimSpace(a.phone)
	code := strings.TrimSpace(a.phoneCode)
	switch {
	case phone == "":
		a.update(func() { a.errText = say("login.needPhone") })
		return
	case a.smsLeft() > 0:
		return
	case a.captchaID == "":
		a.update(func() { a.errText = say("login.needCaptcha") })
		a.fetchCaptcha()
		return
	case code == "":
		a.update(func() { a.errText = say("login.needCaptcha") })
		return
	}
	a.update(func() { a.busy = true; a.errText = line{}; a.hint = line{} })
	go func() {
		err := a.client.SendSmsCode(phone, code, a.captchaID)
		a.update(func() {
			a.busy = false
			if err != nil {
				a.errText = errorLine(err)
				// A picture is read out once, and one the server turned
				// down has been read: another is put in its place.
				var api *yh.APIError
				if errors.As(err, &api) {
					a.fetchCaptcha()
				}
				return
			}
			a.smsAt = time.Now()
			a.hint = say("login.codeSent")
			// The picture was read out for this code alone: another is
			// asked for, so that the next code is not texted on a
			// picture that has already been used.
			a.phoneCode = ""
			a.fetchCaptcha()
		})
	}()
}

// smsLeft is how long until another code may be asked for, in seconds: it
// counts down to what the button over the field says. It counts up, so
// that a second is not shown as none while the code is still not on its
// way again.
func (a *App) smsLeft() int {
	if a.smsAt.IsZero() {
		return 0
	}
	left := smsWait - time.Since(a.smsAt)
	if left <= 0 {
		return 0
	}
	return int(math.Ceil(left.Seconds()))
}

// loginWithPhone signs in with the number typed and the code texted to it.
func (a *App) loginWithPhone() {
	phone := strings.TrimSpace(a.phone)
	code := strings.TrimSpace(a.smsCode)
	switch {
	case phone == "":
		a.update(func() { a.errText = say("login.needPhone") })
		return
	case code == "":
		a.update(func() { a.errText = say("login.needCode") })
		return
	}
	a.update(func() { a.busy = true; a.errText = line{}; a.hint = line{} })
	go func() {
		token, err := a.client.LoginPhone(phone, code)
		if err != nil {
			a.update(func() { a.busy = false; a.errText = errorLine(err) })
			return
		}
		a.client.Token = token
		u, err := a.client.UserInfo()
		if err != nil {
			a.update(func() { a.busy = false; a.errText = errorLine(err) })
			return
		}
		a.update(func() { a.smsCode = ""; a.phoneCode = "" })
		a.afterLogin(u)
	}()
}

// logout forgets the session and shows the sign-in screen.
func (a *App) logout() {
	a.update(func() {
		a.screen = screenLogin
		a.me = nil
		a.convs = nil
		a.msgs = nil
		a.current = ""
		a.note = line{}
		a.errText = line{}
		a.hint = line{}
		// A code was texted for the account that is being left, and the
		// picture was read out for it: neither is worth keeping.
		a.smsCode, a.phoneCode = "", ""
		a.captchaID, a.captcha = "", nil
		a.smsAt = time.Time{}
	})
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
	token := a.client.Token
	a.client.Token = ""
	go func() {
		if token != "" {
			_ = a.client.Logout()
		}
		_ = yh.ClearSession()
	}()
}
