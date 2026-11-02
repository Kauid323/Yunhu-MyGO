package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// pbString is a protobuf field holding a string, which is how the API
// answers the account's information.
func pbString(field int, s string) []byte {
	return append([]byte{byte(field<<3 | 2), byte(len(s))}, s...)
}

// captchaPicture is a picture to read out, as the API sends one: a data
// URL of a PNG.
func captchaPicture(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// phoneAPI is a server that answers the three steps of signing in with a
// phone number, and hands back the bodies that were sent to it.
func phoneAPI(t *testing.T) (*httptest.Server, chan string) {
	t.Helper()
	pic := captchaPicture(t)
	// The account the sign-in is for, in protobuf: field 1 of the answer is
	// the status, whose own field 2 is the code — 1, a success — and field
	// 2 is the account, which holds its id and its phone number.
	inner := append(pbString(1, "100"), pbString(6, "13800000000")...)
	user := append(pbString(1, string([]byte{0x10, 0x01})), pbString(2, string(inner))...)

	asked := make(chan string, 8)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/v1/user/captcha":
			asked <- "captcha"
			io.WriteString(w, `{"code":1,"data":{"b64s":"`+pic+`","id":"c1"}}`)
		case "/v1/verification/get-verification-code":
			asked <- string(body)
			io.WriteString(w, `{"code":1,"msg":"success"}`)
		case "/v1/user/verification-login":
			asked <- string(body)
			io.WriteString(w, `{"code":1,"data":{"token":"tok-1"}}`)
		case "/v1/user/info":
			w.Write(user)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, asked
}

// until waits for what a goroutine is doing to be done, as the window
// would draw it once it is.
func until(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("%s did not happen", what)
}

// TestPhoneSignIn walks the whole way in with a phone number: a picture to
// read out, a code texted to the number, and the code signing in.
func TestPhoneSignIn(t *testing.T) {
	defer i18n.Set("zh-CN")
	// A sign-in writes the session it is for: it is written to a
	// directory of the test's own rather than the account's.
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	a := New()
	srv, asked := phoneAPI(t)
	a.client.Base = srv.URL
	// The live connection is not what is being tested, and it is not
	// asked of the network.
	a.client.WSBase = "ws://127.0.0.1:1"

	// Taking the phone way asks for a picture to read out.
	a.setLoginMode(modePhone)
	select {
	case got := <-asked:
		if got != "captcha" {
			t.Fatalf("taking the phone way asked for %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("taking the phone way asked for no picture")
	}
	until(t, "the picture to arrive", func() bool { return a.captchaID != "" })
	if a.captcha == nil {
		t.Fatal("the picture did not come back as a picture")
	}

	// A code is not texted without a number and a picture read out.
	a.sendSmsCode()
	if got := a.errText.String(); got != i18n.T("login.needPhone") {
		t.Errorf("asking for a code without a number said %q", got)
	}
	a.phone = "13800000000"
	a.sendSmsCode()
	if got := a.errText.String(); got != i18n.T("login.needCaptcha") {
		t.Errorf("asking for a code without reading the picture said %q", got)
	}

	// With both, the code is asked for, with what the picture said.
	a.phoneCode = "ab12"
	a.sendSmsCode()
	var sent map[string]string
	select {
	case body := <-asked:
		if err := json.Unmarshal([]byte(body), &sent); err != nil {
			t.Fatalf("%q is not what was sent: %v", body, err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no code was asked for")
	}
	if sent["mobile"] != "13800000000" || sent["code"] != "ab12" || sent["id"] != "c1" {
		t.Errorf("the code was asked for with %v", sent)
	}
	until(t, "the code to be marked sent", func() bool { return a.smsLeft() > 0 })
	if a.hint.String() != i18n.T("login.codeSent") {
		t.Errorf("sending a code said %q", a.hint.String())
	}
	// What the picture said is used up with the code: another picture is
	// asked for, and the field is cleared for what it says.
	if a.phoneCode != "" {
		t.Errorf("the picture's code was kept: %q", a.phoneCode)
	}
	select {
	case got := <-asked:
		if got != "captcha" {
			t.Errorf("after the code was sent, %q was asked for", got)
		}
	case <-time.After(3 * time.Second):
		t.Error("no other picture was asked for after the code was sent")
	}

	// The code that was texted signs in.
	a.smsCode = ""
	a.loginWithPhone()
	if got := a.errText.String(); got != i18n.T("login.needCode") {
		t.Errorf("signing in without the code said %q", got)
	}
	a.smsCode = "123456"
	a.loginWithPhone()
	select {
	case body := <-asked:
		if err := json.Unmarshal([]byte(body), &sent); err != nil {
			t.Fatalf("%q is not what was sent: %v", body, err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the sign-in was not asked for")
	}
	if sent["mobile"] != "13800000000" || sent["captcha"] != "123456" {
		t.Errorf("the sign-in was asked for with %v", sent)
	}
	until(t, "the chat screen to open", func() bool { return a.screen == screenChat })
	if a.me == nil || a.me.ID != "100" {
		t.Fatalf("the account that signed in is %+v", a.me)
	}

	// And the number is kept for the next run.
	s, err := yh.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	if s.Token != "tok-1" || s.Phone != "13800000000" {
		t.Errorf("the session kept %+v", s)
	}
}

// TestCodeCountsDown checks the wait between two codes: a code is texted
// once a minute, and the button says how long is left.
func TestCodeCountsDown(t *testing.T) {
	defer i18n.Set("zh-CN")

	a := New()
	if got := a.smsLeft(); got != 0 {
		t.Errorf("a screen that never texted a code counts %d down", got)
	}
	a.smsAt = time.Now()
	if got := a.smsLeft(); got != int(smsWait/time.Second) {
		t.Errorf("a code just texted counts %d down, want %d", got, int(smsWait/time.Second))
	}
	if label, left := a.smsButton(); left <= 0 || label != i18n.T("login.resend", i18n.Num(left)) {
		t.Errorf("the button reads %q, %d", label, left)
	}

	a.smsAt = time.Now().Add(-smsWait)
	if got := a.smsLeft(); got != 0 {
		t.Errorf("a minute after the code it still counts %d down", got)
	}
	if label, left := a.smsButton(); left != 0 || label != i18n.T("login.sendCode") {
		t.Errorf("the button reads %q, %d, want the way to ask again", label, left)
	}
}

// TestPhoneModeAsksForAPictureOnce checks that the picture to read out is
// asked for once as the phone way is taken, and not again on every frame.
func TestPhoneModeAsksForAPictureOnce(t *testing.T) {
	a := New()
	srv, asked := phoneAPI(t)
	a.client.Base = srv.URL

	a.setLoginMode(modePhone)
	select {
	case <-asked:
	case <-time.After(3 * time.Second):
		t.Fatal("taking the phone way asked for no picture")
	}
	until(t, "the picture to arrive", func() bool { return a.captchaID != "" })

	// Taking it again, and drawing the screen, asks for nothing more.
	a.setLoginMode(modePassword)
	a.setLoginMode(modePhone)
	tt := ui.NewTester(a.view, 900, 620)
	tt.Frame()
	tt.Frame()
	select {
	case got := <-asked:
		t.Errorf("another picture was asked for: %q", got)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestAWrongReadingAsksForAnotherPicture checks that a picture read out
// wrong is replaced: it has been read, and the server will not take the
// same one twice.
func TestAWrongReadingAsksForAnotherPicture(t *testing.T) {
	defer i18n.Set("zh-CN")

	pic := captchaPicture(t)
	asked := make(chan string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/user/captcha" {
			asked <- "captcha"
			io.WriteString(w, `{"code":1,"data":{"b64s":"`+pic+`","id":"c2"}}`)
			return
		}
		asked <- "refused"
		io.WriteString(w, `{"code":0,"msg":"验证码错误"}`)
	}))
	defer srv.Close()

	a := New()
	a.client.Base = srv.URL
	a.captchaID = "c1"
	a.phone = "13800000000"
	a.phoneCode = "wrong"
	a.sendSmsCode()

	select {
	case got := <-asked:
		if got != "refused" {
			t.Fatalf("the first thing asked for was %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no code was asked for")
	}
	until(t, "the refusal to be said", func() bool { return !a.errText.empty() })
	if got := a.errText.String(); !strings.Contains(got, "验证码错误") {
		t.Errorf("a wrong reading is said as %q, want the server's words", got)
	}
	if a.smsLeft() != 0 {
		t.Error("a code that was refused was counted as sent")
	}
	select {
	case got := <-asked:
		if got != "captcha" {
			t.Errorf("after a wrong reading, %q was asked for", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no other picture was asked for after a wrong reading")
	}
}
