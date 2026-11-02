package app

import (
	"html"
	"strings"
)

// The API sends some messages as markup: rich cards, links, a table of
// text. A native window has no page to show it in, so the words are taken
// out of it and the tags are dropped, and a picture in it is drawn as a
// picture.

// htmlText is what a message of markup says.
func htmlText(s string) string {
	if !strings.ContainsRune(s, '<') {
		return tidy(html.UnescapeString(s))
	}

	var b strings.Builder
	hidden := 0 // inside a script or a style, whose words are not read
	for i := 0; i < len(s); {
		if s[i] != '<' {
			if hidden == 0 {
				b.WriteByte(s[i])
			}
			i++
			continue
		}
		end := tagEnd(s, i)
		if end < 0 { // a '<' that opens nothing is a character
			if hidden == 0 {
				b.WriteByte(s[i])
			}
			i++
			continue
		}
		switch name := tagName(s[i+1 : end]); name {
		case "script", "style":
			hidden++
		case "/script", "/style":
			if hidden > 0 {
				hidden--
			}
		case "br", "/p", "/div", "/tr", "/td", "/li", "/h1", "/h2", "/h3", "/h4", "/h5", "/h6":
			if hidden == 0 {
				b.WriteByte('\n')
			}
		}
		i = end + 1
	}
	return tidy(html.UnescapeString(b.String()))
}

// htmlImage is the first picture a message of markup shows, or "".
func htmlImage(s string) string {
	lower := strings.ToLower(s)
	for i := 0; i < len(s); {
		k := strings.Index(lower[i:], "<img")
		if k < 0 {
			return ""
		}
		k += i
		end := tagEnd(s, k)
		if end < 0 {
			return ""
		}
		if src := attr(s[k+1:end], "src"); src != "" {
			return html.UnescapeString(src)
		}
		i = end + 1
	}
	return ""
}

// tagEnd is where the tag starting at i closes, or -1 when it does not:
// a '>' inside a quoted value does not close it.
func tagEnd(s string, i int) int {
	var quote byte
	for j := i + 1; j < len(s); j++ {
		switch c := s[j]; {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return j
		}
	}
	return -1
}

// tagName is the lowercased name a tag opens with, keeping the slash that
// closes one, as "/p" for "</p>", and dropping the one that ends a tag of
// no content, as "br" for "br/".
func tagName(s string) string {
	s = strings.TrimSpace(s)
	if k := strings.IndexAny(s, " \t\r\n"); k >= 0 {
		s = s[:k]
	}
	if !strings.HasPrefix(s, "/") {
		s = strings.TrimSuffix(s, "/")
	}
	return strings.ToLower(s)
}

// attr is the value of attribute name in a tag's insides, or "".
func attr(tag, name string) string {
	lower := strings.ToLower(tag)
	for i := 0; i < len(tag); {
		k := strings.Index(lower[i:], name)
		if k < 0 {
			return ""
		}
		k += i
		// It is the attribute itself, not the tail of a longer name.
		if k > 0 && isNameByte(tag[k-1]) {
			i = k + len(name)
			continue
		}
		j := k + len(name)
		for j < len(tag) && (tag[j] == ' ' || tag[j] == '\t') {
			j++
		}
		if j >= len(tag) || tag[j] != '=' {
			i = k + len(name)
			continue
		}
		j++
		for j < len(tag) && (tag[j] == ' ' || tag[j] == '\t') {
			j++
		}
		if j >= len(tag) {
			return ""
		}
		if q := tag[j]; q == '"' || q == '\'' {
			if e := strings.IndexByte(tag[j+1:], q); e >= 0 {
				return tag[j+1 : j+1+e]
			}
			return ""
		}
		e := j
		for e < len(tag) && tag[e] != ' ' && tag[e] != '\t' {
			e++
		}
		// An unquoted value ends where the tag does, and a slash there
		// ends the tag rather than the value: "src=a.png/" is "a.png".
		return strings.TrimSuffix(tag[j:e], "/")
	}
	return ""
}

func isNameByte(c byte) bool {
	return c == '-' || c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// tidy makes lines of what was taken out of the markup: runs of spaces
// become one, and the empty lines left by nested tags go.
func tidy(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			// One empty line separates paragraphs; a run of them does not.
			blank = len(out) > 0
			continue
		}
		if blank {
			out = append(out, "")
			blank = false
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
