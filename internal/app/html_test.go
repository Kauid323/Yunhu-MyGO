package app

import (
	"testing"

	"mygohu/internal/yh"
)

// messageOf is a message of markup, as the API sends some.
func messageOf(body string) *yh.Message {
	return &yh.Message{ContentType: yh.ContentHTML, Content: yh.Content{Text: body}}
}

// TestHTMLText reads markup for its words, as a message of markup shows
// them: the API sends cards and tables this way.
func TestHTMLText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"纯文本", "纯文本"},
		{"<b>加粗</b>和<i>斜体</i>", "加粗和斜体"},
		{"第一行<br>第二行", "第一行\n第二行"},
		{"<p>一段</p><p>另一段</p>", "一段\n另一段"},
		// The words of a script are not read.
		{"<script>var a = 1;</script>看得见", "看得见"},
		{"<style>p{color:red}</style>看得见", "看得见"},
		// Entities come back as the characters they stand for.
		{"a &amp; b &lt;tag&gt;", "a & b <tag>"},
		// A tag broken over lines, and a '>' inside a value.
		{"<a title=\"a>b\" href=\"#\">链接</a>", "链接"},
		// Runs of spaces collapse, and so do the empty lines nested tags
		// leave behind.
		{"  多   个    空格  ", "多 个 空格"},
		{"<div><div><div>深</div></div></div>", "深"},
		{"<table><tr><td>甲</td><td>乙</td></tr></table>", "甲\n乙"},
		// A '<' that opens no tag is a character.
		{"1 < 2", "1 < 2"},
	}
	for _, c := range cases {
		if got := htmlText(c.in); got != c.want {
			t.Errorf("htmlText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestHTMLImage finds the picture a card shows.
func TestHTMLImage(t *testing.T) {
	card := `<a href="#"><table><tr><td><img src="https://chat-img.jwznb.com/9bf377cd.png" ` +
		`width="36" height="36" style="border-radius: 50%"></td>` +
		`<td><div style="font-size: 16px">乐器演奏</div></td></tr></table></a>`
	if got := htmlImage(card); got != "https://chat-img.jwznb.com/9bf377cd.png" {
		t.Errorf("htmlImage(card) = %q", got)
	}

	// Single quotes, no quotes, and an escaped ampersand.
	if got := htmlImage("<img src='https://x/a.png'>"); got != "https://x/a.png" {
		t.Errorf("htmlImage(single quotes) = %q", got)
	}
	if got := htmlImage("<img src=https://x/a.png>"); got != "https://x/a.png" {
		t.Errorf("htmlImage(no quotes) = %q", got)
	}
	if got := htmlImage(`<img src="https://x/a.png?t=1&amp;u=2">`); got != "https://x/a.png?t=1&u=2" {
		t.Errorf("htmlImage(entity) = %q", got)
	}

	// "src" inside another attribute's name is not the picture.
	if got := htmlImage(`<img data-src="https://x/no.png">`); got != "" {
		t.Errorf("htmlImage(data-src) = %q, want empty", got)
	}
	if got := htmlImage("没有图片"); got != "" {
		t.Errorf("htmlImage(no img) = %q, want empty", got)
	}
}

// TestHTMLMessage checks a message of markup as the chat draws it: its
// words, and the picture in it.
func TestHTMLMessage(t *testing.T) {
	m := messageOf(`<div style="font-size:16px">乐器演奏</div><img src="https://x/a.png">`)
	text, detail := messageText(m)
	if text != "乐器演奏" {
		t.Errorf("messageText(html) = %q, want 乐器演奏", text)
	}
	if detail != "" {
		t.Errorf("messageText(html) detail = %q, want empty", detail)
	}
	if got := pictureURL(m); got != "https://x/a.png" {
		t.Errorf("pictureURL(html) = %q", got)
	}

	// Markup of nothing but a picture says nothing.
	bare := messageOf(`<a href="#"><img src="https://x/b.png"></a>`)
	if text, _ := messageText(bare); text != "" {
		t.Errorf("messageText(image only) = %q, want empty", text)
	}
}
