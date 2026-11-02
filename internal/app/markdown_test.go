package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/yuin/goldmark/text"

	"mygohu/internal/yh"
)

// spansOf reads the inline part of the first block of src.
func spansOf(src string) []ui.Span {
	b := []byte(src)
	doc := mdParser.Parse(text.NewReader(b))
	return mdSpans(doc.FirstChild(), b)
}

// find returns the span whose text is s, if there is one.
func find(spans []ui.Span, s string) (ui.Span, bool) {
	for _, sp := range spans {
		if sp.Text == s {
			return sp, true
		}
	}
	return ui.Span{}, false
}

// TestMarkdownSpans checks that markdown's stars and hashes become the
// styles they stand for, rather than staying in the text.
func TestMarkdownSpans(t *testing.T) {
	spans := spansOf("**粗**和*斜*还有`代码`~~删掉~~[链接](https://x)")

	bold, ok := find(spans, "粗")
	if !ok || bold.Weight < 600 {
		t.Errorf("粗 = %+v, want a bold span", bold)
	}
	italic, ok := find(spans, "斜")
	if !ok || !italic.Italic {
		t.Errorf("斜 = %+v, want an italic span", italic)
	}
	code, ok := find(spans, "代码")
	if !ok || code.Font != "monospace" {
		t.Errorf("代码 = %+v, want a monospaced span", code)
	}
	struck, ok := find(spans, "删掉")
	if !ok || !struck.Strikethrough {
		t.Errorf("删掉 = %+v, want a struck span", struck)
	}
	link, ok := find(spans, "链接")
	if !ok || !link.Underline || link.Color != qq.accent {
		t.Errorf("链接 = %+v, want a link", link)
	}

	// None of the markup is left in the words.
	for _, sp := range spans {
		switch sp.Text {
		case "**", "*", "`", "~~", "](", "https://x":
			t.Errorf("markup %q stayed in the text", sp.Text)
		}
	}
}

// TestMarkdownLines checks that the newlines the author typed are kept,
// and that a hard break is one too.
func TestMarkdownLines(t *testing.T) {
	spans := spansOf("第一行\n第二行")
	var joined string
	for _, sp := range spans {
		joined += sp.Text
	}
	if joined != "第一行\n第二行" {
		t.Errorf("spans read %q, want the line break kept", joined)
	}
}

// TestMarkdownDraws checks that a message written in markdown shows its
// words: a list item holds a text block rather than a paragraph, and a
// text block left unread draws bullets with nothing beside them.
func TestMarkdownDraws(t *testing.T) {
	a := openChat(0)
	a.msgs = []*yh.Message{{
		MsgID: "md", ChatID: "g1", Sender: yh.Sender{ChatID: "200", Name: "小明"},
		ContentType: yh.ContentMarkdown,
		Content: yh.Content{Text: "### 今日更新\n" +
			"- 滚动到 80% 自动加载\n" +
			"> 位置不变才是关键\n"},
	}}
	tt := ui.NewTester(a.view, 900, 700)
	tt.Frame()

	for _, want := range []string{"今日更新", "滚动到 80% 自动加载", "位置不变才是关键"} {
		if !tt.HasText(want) {
			t.Errorf("the message does not show %q: %v", want, tt.Texts())
		}
	}
}
