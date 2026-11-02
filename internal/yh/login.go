package yh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Signing in with a phone number takes three steps: the API hands out a
// picture of some letters, a code is texted to the number once that
// picture has been read out, and the code signs in. The picture is fetched
// once and read once, and the id it came with says which picture was read.

// Captcha is a picture to read out before a code is texted to a phone.
type Captcha struct {
	// ID says which picture this is, and goes with what it says.
	ID string
	// Image is the picture itself: a PNG or a JPEG.
	Image []byte
}

// Captcha asks for a picture to read out.
func (c *Client) Captcha() (*Captcha, error) {
	data, err := c.do("POST", "/v1/user/captcha", []byte("{}"), false)
	if err != nil {
		return nil, err
	}
	var res struct {
		jsonStatus
		Data struct {
			B64S string `json:"b64s"`
			ID   string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if err := res.err(); err != nil {
		return nil, err
	}
	img, err := decodeDataURL(res.Data.B64S)
	if err != nil {
		return nil, err
	}
	if len(img) == 0 || res.Data.ID == "" {
		return nil, ErrNoCaptcha
	}
	return &Captcha{ID: res.Data.ID, Image: img}, nil
}

// SendSmsCode asks for a code to be texted to a phone number, having read
// out the picture it came with.
func (c *Client) SendSmsCode(mobile, code, captchaID string) error {
	body, _ := json.Marshal(map[string]string{
		"mobile":   mobile,
		"code":     code,
		"id":       captchaID,
		"platform": c.Platform,
	})
	data, err := c.do("POST", "/v1/verification/get-verification-code", body, false)
	if err != nil {
		return err
	}
	var res jsonStatus
	if err := json.Unmarshal(data, &res); err != nil {
		return fmt.Errorf("%w: %v", ErrResponse, err)
	}
	return res.err()
}

// LoginPhone signs in with a phone number and the code it was texted, and
// returns the token.
func (c *Client) LoginPhone(mobile, code string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"mobile":   mobile,
		"captcha":  code,
		"deviceId": c.DeviceID,
		"platform": c.Platform,
	})
	return c.login("/v1/user/verification-login", body)
}

// login posts a sign-in and returns the token it answers with.
func (c *Client) login(path string, body []byte) (string, error) {
	data, err := c.do("POST", path, body, false)
	if err != nil {
		return "", err
	}
	var res struct {
		jsonStatus
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return "", fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if err := res.err(); err != nil {
		return "", err
	}
	if res.Data.Token == "" {
		return "", ErrNoToken
	}
	return res.Data.Token, nil
}

// decodeDataURL takes the picture out of the data URL the API writes it in,
// which it spells its own way: the header says base64 with the data: in
// front of it, and the payload is base64 with its padding or without.
func decodeDataURL(s string) ([]byte, error) {
	if i := strings.Index(s, "base64,"); i >= 0 {
		s = s[i+len("base64,"):]
	}
	s = strings.Map(dropSpace, s)
	if s == "" {
		return nil, ErrNoCaptcha
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	b, err = base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResponse, err)
	}
	return b, nil
}

// dropSpace drops what is not part of the base64 itself, as a data URL
// that was wrapped over several lines carries.
func dropSpace(r rune) rune {
	switch r {
	case '\n', '\r', '\t', ' ':
		return -1
	}
	return r
}
