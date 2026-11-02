package yh

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// answers is a server that answers the three steps of signing in with a
// phone number, and hands back what was asked of it.
func answers(t *testing.T) (*httptest.Server, chan string) {
	t.Helper()
	asked := make(chan string, 8)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/v1/user/captcha":
			asked <- "captcha"
			b64 := base64.StdEncoding.EncodeToString([]byte("letters"))
			io.WriteString(w, `{"code":1,"data":{"b64s":"data:image/png;base64,`+b64+`","id":"c1"}}`)
		case "/v1/verification/get-verification-code":
			asked <- string(body)
			io.WriteString(w, `{"code":1,"msg":"success"}`)
		case "/v1/user/verification-login":
			asked <- string(body)
			io.WriteString(w, `{"code":1,"data":{"token":"tok-1"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, asked
}

// TestCaptchaIsReadOut checks the picture a code is not texted without: its
// bytes are taken out of the data URL the API wraps them in, and the id
// they came with is kept.
func TestCaptchaIsReadOut(t *testing.T) {
	srv, _ := answers(t)
	c := New()
	c.Base = srv.URL

	got, err := c.Captcha()
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "c1" {
		t.Errorf("the picture is %q, want c1", got.ID)
	}
	if string(got.Image) != "letters" {
		t.Errorf("the picture reads %q, want the bytes of it", got.Image)
	}
}

// TestDecodeDataURL checks what the API wraps a picture in: the header it
// writes, with or without the data: in front of it, and padding or none.
func TestDecodeDataURL(t *testing.T) {
	raw := []byte("a picture")
	std := base64.StdEncoding.EncodeToString(raw)
	raw64 := base64.RawStdEncoding.EncodeToString(raw)

	for _, in := range []string{
		"data:image/png;base64," + std,
		"image/png;base64," + std,
		"image//png;base64," + std,
		std,
		"data:image/jpeg;base64," + raw64,
		"data:image/png;base64," + std[:4] + "\n" + std[4:],
	} {
		got, err := decodeDataURL(in)
		if err != nil {
			t.Errorf("decodeDataURL(%q): %v", in, err)
			continue
		}
		if string(got) != string(raw) {
			t.Errorf("decodeDataURL(%q) reads %q", in, got)
		}
	}

	if _, err := decodeDataURL(""); err == nil {
		t.Error("an empty picture was read as one")
	}
	if _, err := decodeDataURL("data:image/png;base64,!!!"); err == nil {
		t.Error("a picture that is not base64 was read")
	}
}

// TestSendSmsCodeAsksForACode checks what is sent to have a code texted:
// the number, what the picture said, and which picture it was.
func TestSendSmsCodeAsksForACode(t *testing.T) {
	srv, asked := answers(t)
	c := New()
	c.Base = srv.URL

	if err := c.SendSmsCode("13800000000", "ab12", "c1"); err != nil {
		t.Fatal(err)
	}
	var sent map[string]string
	if err := json.Unmarshal([]byte(<-asked), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["mobile"] != "13800000000" || sent["code"] != "ab12" || sent["id"] != "c1" {
		t.Errorf("the request was %v", sent)
	}
}

// TestSendSmsCodeSaysNo checks that a server that refuses to text a code
// is a failure the window can say.
func TestSendSmsCodeSaysNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"code":0,"msg":"手机号不对"}`)
	}))
	defer srv.Close()
	c := New()
	c.Base = srv.URL

	err := c.SendSmsCode("1", "ab12", "c1")
	if err == nil {
		t.Fatal("a refused code was taken as sent")
	}
	var api *APIError
	if !errors.As(err, &api) {
		t.Fatalf("the failure is %v, not the server's own words", err)
	}
	if api.Msg != "手机号不对" {
		t.Errorf("the failure reads %q, want the server's words", api.Msg)
	}
}

// TestLoginPhone checks the sign-in itself: the number and the code go,
// and the token comes back.
func TestLoginPhone(t *testing.T) {
	srv, asked := answers(t)
	c := New()
	c.Base = srv.URL

	token, err := c.LoginPhone("13800000000", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if token != "tok-1" {
		t.Errorf("the sign-in returned %q, want the token", token)
	}
	var sent map[string]string
	if err := json.Unmarshal([]byte(<-asked), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["mobile"] != "13800000000" || sent["captcha"] != "123456" {
		t.Errorf("the request was %v", sent)
	}
	if sent["deviceId"] == "" || sent["platform"] == "" {
		t.Errorf("the request did not say what signed in: %v", sent)
	}
}

// TestLoginPhoneWithoutAToken checks that an answer with no token in it is
// a failure rather than a sign-in with nothing.
func TestLoginPhoneWithoutAToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"code":1,"data":{}}`)
	}))
	defer srv.Close()
	c := New()
	c.Base = srv.URL

	if _, err := c.LoginPhone("13800000000", "123456"); err == nil {
		t.Error("a sign-in that returned no token was taken as one")
	} else if !strings.Contains(err.Error(), "token") {
		t.Errorf("the failure reads %q", err)
	}
}
