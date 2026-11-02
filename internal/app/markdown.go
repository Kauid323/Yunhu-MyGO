package app

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Some messages are written in markdown. It is parsed rather than read as
// tags and stars: the words keep the weight, the slant and the links the
// author gave them, in a text of spans.

var mdParser = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()

// markdown draws a message written in markdown, block by block.
func (a *App) markdown(c *ui.Context, src string, fg ui.Color) {
	if strings.TrimSpace(src) == "" {
		return
	}
	doc := mdParser.Parse(text.NewReader([]byte(src)))
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		a.mdBlock(c, n, []byte(src), fg)
	}
}

// mdBlock draws one block of markdown: a paragraph, a heading, a list, a
// quotation, a block of code.
func (a *App) mdBlock(c *ui.Context, n ast.Node, src []byte, fg ui.Color) {
	switch b := n.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		// A tight list item holds a text block rather than a paragraph,
		// and reads the same.
		ui.RichText(c, mdSpans(n, src)...).FontSize(13).LineHeight(1.45).
			TextColor(fg).Wrap()

	case *ast.Heading:
		size := float32(13.5)
		if b.Level <= 2 {
			size = 15
		}
		ui.RichText(c, mdSpans(b, src)...).FontSize(size).LineHeight(1.4).
			Bold().TextColor(fg).Wrap()

	case *ast.Blockquote:
		ui.Row(c).Gap(8).Children(func() {
			ui.Box(c).Width(3).FillHeight().Radius(2).
				Background(qq.mdRule)
			ui.Column(c).Grow(1).Gap(4).Children(func() {
				for ch := b.FirstChild(); ch != nil; ch = ch.NextSibling() {
					a.mdBlock(c, ch, src, fg)
				}
			})
		})

	case *ast.List:
		ui.Column(c).Gap(3).Children(func() {
			i := 0
			for item := b.FirstChild(); item != nil; item = item.NextSibling() {
				mark := "•"
				if b.IsOrdered() {
					mark = strconv.Itoa(b.Start+i) + "."
				}
				i++
				ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
					ui.Text(c, mark).FontSize(13).LineHeight(1.45).TextColor(fg).
						Opacity(0.7).MinWidth(14)
					ui.Column(c).Grow(1).Gap(4).Children(func() {
						for ch := item.FirstChild(); ch != nil; ch = ch.NextSibling() {
							a.mdBlock(c, ch, src, fg)
						}
					})
				})
			}
		})

	case *ast.FencedCodeBlock, *ast.CodeBlock:
		ui.Column(c).Padding(6, 9).Radius(4).Gap(0).
			Background(qq.mdCode).Children(func() {
			ui.RichText(c, ui.Span{Text: mdLines(b, src)}).Font("monospace").
				FontSize(11.5).LineHeight(1.4).TextColor(fg).Wrap()
		})

	case *ast.ThematicBreak:
		ui.Box(c).Height(1).Background(qq.mdRule)

	case *ast.HTMLBlock:
		// A block of markup is read for its words, as a message of markup
		// is.
		if t := htmlText(mdLines(n, src)); t != "" {
			ui.Text(c, t).FontSize(13).LineHeight(1.45).TextColor(fg).Wrap()
		}

	default:
		// Anything else is drawn for the blocks it holds.
		for ch := n.FirstChild(); ch != nil; ch = ch.NextSibling() {
			a.mdBlock(c, ch, src, fg)
		}
	}
}

// mdSpans reads the inline part of a node: its words, and the styles they
// are written in.
func mdSpans(n ast.Node, src []byte) []ui.Span {
	var out []ui.Span
	var walk func(ast.Node, mdStyle)
	walk = func(n ast.Node, st mdStyle) {
		for ch := n.FirstChild(); ch != nil; ch = ch.NextSibling() {
			// GFM's strikethrough is a node of its own kind, not a type
			// of its own.
			if ch.Kind() == extast.KindStrikethrough {
				next := st
				next.strike = true
				walk(ch, next)
				continue
			}
			switch t := ch.(type) {
			case *ast.Text:
				out = append(out, mdSpan(string(t.Segment.Value(src)), st))
				// A newline the author typed is a newline here too.
				if t.SoftLineBreak() || t.HardLineBreak() {
					out = append(out, ui.Span{Text: "\n"})
				}
			case *ast.String:
				out = append(out, mdSpan(string(t.Value), st))
			case *ast.Emphasis:
				next := st
				if t.Level >= 2 {
					next.weight = 600
				} else {
					next.italic = true
				}
				walk(t, next)
			case *ast.CodeSpan:
				walk(t, mdStyle{weight: st.weight, italic: st.italic, code: true})
			case *ast.Link:
				next := st
				next.link = true
				walk(t, next)
			case *ast.AutoLink:
				out = append(out, mdSpan(string(t.Label(src)), mdStyle{link: true}))
			case *ast.Image:
				// The picture is not drawn in a line of text: its
				// description stands for it.
				walk(t, st)
			case *ast.RawHTML:
				// Tags are not words.
			default:
				walk(ch, st)
			}
		}
	}
	walk(n, mdStyle{})
	return out
}

// mdStyle is what a run of words is written in.
type mdStyle struct {
	weight int
	italic bool
	strike bool
	code   bool
	link   bool
}

// mdSpan is one run of words in its style.
func mdSpan(s string, st mdStyle) ui.Span {
	span := ui.Span{Text: s, Weight: st.weight, Italic: st.italic, Strikethrough: st.strike}
	switch {
	case st.code:
		span.Font = "monospace"
		span.Size = 12
		span.Background = qq.mdCode
		span.Radius = 3
	case st.link:
		span.Underline = true
		span.Color = qq.accent
	}
	return span
}

// mdLines is the text of a block of code or markup, line by line.
func mdLines(n ast.Node, src []byte) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(src))
	}
	return strings.TrimRight(b.String(), "\n")
}
