package i18n

import (
	"sort"
	"strings"
	"testing"
)

// TestEveryKeyIsInEveryLanguage keeps the two files in step: a key added
// to one and forgotten in the other shows as the key itself in the window,
// which is a thing to catch here rather than there.
func TestEveryKeyIsInEveryLanguage(t *testing.T) {
	zh, en := mustLoad("zh-CN"), mustLoad("en-US")
	for _, c := range []struct {
		lang string
		has  map[string]string
		want map[string]string
	}{
		{"en-US", en, zh},
		{"zh-CN", zh, en},
	} {
		var missing []string
		for key := range c.want {
			if _, ok := c.has[key]; !ok {
				missing = append(missing, key)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("%s is missing %d keys: %s", c.lang, len(missing), strings.Join(missing, ", "))
		}
	}
}

// TestPlaceholdersMatch checks that a key takes the same arguments in
// every language: filling one with an argument the other has none of would
// leave a brace in the window.
func TestPlaceholdersMatch(t *testing.T) {
	zh, en := mustLoad("zh-CN"), mustLoad("en-US")
	for key, z := range zh {
		e, ok := en[key]
		if !ok {
			continue
		}
		zhArgs, enArgs := placeholders(z), placeholders(e)
		if zhArgs != enArgs {
			t.Errorf("%s: zh-CN takes %d arguments, en-US takes %d", key, zhArgs, enArgs)
		}
	}
}

// TestFill checks that arguments go in the order the placeholders appear,
// that a number is written the way the language writes it, and that a
// placeholder with no argument is left alone rather than swallowed.
func TestFill(t *testing.T) {
	Set("en-US")
	defer Set(fallback)

	if got := T("chat.failed", "no route"); got != "Failed to send: no route" {
		t.Errorf("chat.failed = %q", got)
	}
	if got := T("chat.headerSub", "Group", "g1"); got != "Group · g1" {
		t.Errorf("chat.headerSub = %q", got)
	}
	if got := T("chat.headerSub", "Group"); got != "Group · {id}" {
		t.Errorf("a placeholder with no argument = %q", got)
	}
	if got := T("msg.fileSize", 1234.5); got != "1,234.5 KB" {
		t.Errorf("msg.fileSize = %q, want the number written the way English writes it", got)
	}
	if got := T("no.such.key"); got != "no.such.key" {
		t.Errorf("an unknown key = %q, want the key itself", got)
	}
}

// TestSetFollowsTheTag checks that the language follows the desktop: any
// Chinese tag reads Chinese, any other reads English, and a tag of no
// language at all leaves the one in use.
func TestSetFollowsTheTag(t *testing.T) {
	defer Set(fallback)
	for _, c := range []struct{ tag, want string }{
		{"zh-CN", "zh-CN"},
		{"zh-Hans", "zh-CN"},
		{"zh-TW", "zh-CN"},
		{"en-US", "en-US"},
		{"en-GB", "en-US"},
		{"de-DE", "en-US"},
		{"", "zh-CN"},
	} {
		Set(fallback)
		Set(c.tag)
		if got := Lang(); got != c.want {
			t.Errorf("Set(%q) is %s, want %s", c.tag, got, c.want)
		}
	}
}

// placeholders counts the {names} of an entry.
func placeholders(s string) int {
	n := 0
	for {
		open := strings.IndexByte(s, '{')
		if open < 0 {
			return n
		}
		end := strings.IndexByte(s[open:], '}')
		if end < 0 {
			return n
		}
		n++
		s = s[open+end+1:]
	}
}
