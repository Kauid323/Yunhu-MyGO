package app

import (
	"testing"

	"mygohu/internal/yh"
)

// TestResized checks the Qiniu transform asked for a picture at the size
// it is drawn: without it the host serves the thousand pixels uploaded.
func TestResized(t *testing.T) {
	cases := []struct {
		raw  string
		w, h int
		want string
	}{
		{"https://chat-img.jwznb.com/a.png", 76, 76,
			"https://chat-img.jwznb.com/a.png?imageView2/1/w/76/h/76"},
		{"https://chat-img.jwznb.com/a.png?t=1", 480, 480,
			"https://chat-img.jwznb.com/a.png?t=1&imageView2/1/w/480/h/480"},
		// A host that cannot resize is asked for the picture itself.
		{"", 76, 76, ""},
		{"https://x/a.png", 0, 0, "https://x/a.png"},
	}
	for _, c := range cases {
		if got := resized(c.raw, c.w, c.h); got != c.want {
			t.Errorf("resized(%q, %d, %d) = %q, want %q", c.raw, c.w, c.h, got, c.want)
		}
	}
}

// TestPictureURL checks which messages are drawn as a picture: an image
// message with a picture in it, and nothing else.
func TestPictureURL(t *testing.T) {
	img := &yh.Message{ContentType: yh.ContentImage,
		Content: yh.Content{ImageURL: "https://chat-img.jwznb.com/a.png"}}
	if got := pictureURL(img); got != "https://chat-img.jwznb.com/a.png" {
		t.Errorf("pictureURL(image) = %q", got)
	}

	// The older field, Image, is the picture too when ImageURL is empty.
	old := &yh.Message{ContentType: yh.ContentImage, Content: yh.Content{Image: "https://x/b.png"}}
	if got := pictureURL(old); got != "https://x/b.png" {
		t.Errorf("pictureURL(image in Image) = %q", got)
	}

	// A text message mentioning a picture is still text.
	text := &yh.Message{ContentType: yh.ContentText,
		Content: yh.Content{Text: "看这个 https://chat-img.jwznb.com/a.png"}}
	if got := pictureURL(text); got != "" {
		t.Errorf("pictureURL(text) = %q, want empty", got)
	}

	// An image message without a picture falls back to its summary.
	bare := &yh.Message{ContentType: yh.ContentImage}
	if got := pictureURL(bare); got != "" {
		t.Errorf("pictureURL(bare image) = %q, want empty", got)
	}
}

// TestPicturePixels checks the size a picture is asked for: a quarter of
// each side of the one sent, and the box's own size when the message does
// not say how big it was.
func TestPicturePixels(t *testing.T) {
	cases := []struct {
		w, h  int
		wantW int
		wantH int
	}{
		{2560, 1530, 640, 382}, // a screenshot
		{1272, 2800, 318, 700}, // a tall one
		{1006, 767, 251, 191},
		{0, 0, pictureBoxPx, pictureBoxPx}, // nothing said
		{3, 3, 1, 1},                       // too small to quarter
	}
	for _, c := range cases {
		m := &yh.Message{ContentType: yh.ContentImage, Content: yh.Content{Width: c.w, Height: c.h}}
		w, h := picturePixels(m)
		if w != c.wantW || h != c.wantH {
			t.Errorf("picturePixels(%d, %d) = %d, %d, want %d, %d",
				c.w, c.h, w, h, c.wantW, c.wantH)
		}
	}
}
