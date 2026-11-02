// Package i18n holds the words the window shows, in the languages it
// speaks: one JSON file per language, keyed by where the words are used
// and what for, "chat.send", so that a key says the same thing wherever it
// is read.
//
// A language is never chosen here: the window asks the desktop which one
// it is in, every frame, and Set makes it the one in use. Nothing in the
// package draws, and nothing in it knows about mygo.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// zh-CN is the language every entry must be in: it is what an entry
// missing from another language falls back to.
const (
	fallback = "zh-CN"
	other    = "en-US"
)

//go:embed zh-CN.json en-US.json
var files embed.FS

var (
	mu sync.RWMutex
	// dict is the entries of the language in use, and pr writes the
	// numbers that fill them the way that language writes them.
	dict = mustLoad(fallback)
	pr   = message.NewPrinter(language.Make(fallback))
	tag  = fallback
)

// base is the fallback's entries, which every language falls back to.
var base = mustLoad(fallback)

// matcher picks the language closest to the one the desktop is in: a
// Chinese desktop reads Chinese, and a language the window has no words in
// reads English.
var matcher = language.NewMatcher([]language.Tag{
	language.SimplifiedChinese, language.AmericanEnglish,
})

// mustLoad reads the entries of a language, which are part of the binary:
// a file that does not parse is a mistake to catch at build time, not
// something to carry on without.
func mustLoad(name string) map[string]string {
	raw, err := files.ReadFile(name + ".json")
	if err != nil {
		panic(fmt.Sprintf("i18n: %s: %v", name, err))
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		panic(fmt.Sprintf("i18n: %s: %v", name, err))
	}
	return m
}

// Set makes tag the language in use. A tag of no language, or one the
// window has no words in, keeps the one in use or reads English; the call
// costs nothing when the language does not change, which is every frame
// but the one the desktop's language changes on.
func Set(t string) {
	l := match(t)
	mu.Lock()
	defer mu.Unlock()
	if l == tag {
		return
	}
	tag = l
	dict = mustLoad(l)
	pr = message.NewPrinter(language.Make(l))
}

// Lang is the language in use, as a BCP 47 tag.
func Lang() string {
	mu.RLock()
	defer mu.RUnlock()
	return tag
}

// match is the language of a tag: Chinese for any Chinese tag, English for
// the rest, and the fallback for a tag that names no language at all.
func match(t string) string {
	if strings.TrimSpace(t) == "" {
		return fallback
	}
	got, _, _ := matcher.Match(language.Make(t))
	if base, _ := got.Base(); base.String() == "zh" {
		return fallback
	}
	return other
}

// T is the words for key, with the placeholders of the entry filled from
// args in the order they appear: T("chat.failed", reason). A number is
// written the way the language writes numbers, so that 1234 is 1,234 where
// that is how it is written.
//
// A key with no entry is returned as it is, so that a mistake shows in the
// window rather than leaving a blank, and TestEveryKeyIsInEveryLanguage
// keeps it from getting there.
func T(key string, args ...any) string {
	mu.RLock()
	s, ok := dict[key]
	p := pr
	mu.RUnlock()
	if !ok {
		if s, ok = base[key]; !ok {
			return key
		}
	}
	if len(args) == 0 {
		return s
	}
	return fill(s, p, args)
}

// fill writes args into the {placeholders} of s, in order. A placeholder
// with no argument left is left as it is written, which is a mistake in
// the entry and reads as one.
func fill(s string, p *message.Printer, args []any) string {
	var b strings.Builder
	b.Grow(len(s) + 8*len(args))
	rest, n := s, 0
	for {
		open := strings.IndexByte(rest, '{')
		if open < 0 {
			b.WriteString(rest)
			return b.String()
		}
		end := strings.IndexByte(rest[open:], '}')
		if end < 0 {
			b.WriteString(rest)
			return b.String()
		}
		end += open
		b.WriteString(rest[:open])
		if n < len(args) {
			b.WriteString(word(p, args[n]))
			n++
		} else {
			b.WriteString(rest[open : end+1])
		}
		rest = rest[end+1:]
	}
}

// word is one argument as the language writes it.
func word(p *message.Printer, v any) string {
	switch x := v.(type) {
	case string:
		return x
	case error:
		return x.Error()
	case int:
		return p.Sprint(x)
	case int64:
		return p.Sprint(x)
	case float64:
		return p.Sprintf("%.1f", x)
	case nil:
		return ""
	}
	return p.Sprint(v)
}

// Num is a number as the language in use writes it, for the places that
// build a line of their own around one.
func Num(v any) string {
	mu.RLock()
	p := pr
	mu.RUnlock()
	return word(p, v)
}
