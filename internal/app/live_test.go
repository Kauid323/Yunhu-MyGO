package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// TestLivePicture asks the real host for a real image message at the size
// it is drawn, and checks that it is a picture, not an error page.
// Skipped unless MYGOHU_LIVE=1, as it needs the saved session and the
// network.
func TestLivePicture(t *testing.T) {
	if os.Getenv("MYGOHU_LIVE") == "" {
		t.Skip("set MYGOHU_LIVE=1 to ask the real API")
	}
	s, err := yh.LoadSession()
	if err != nil || s == nil || s.Token == "" {
		t.Skip("no saved session")
	}
	c := yh.New()
	c.Base = s.Base
	c.Token = s.Token
	c.DeviceID = s.DeviceID

	convs, _, err := c.Conversations("")
	if err != nil {
		t.Skipf("conversations: %v", err)
	}

	hc := &http.Client{Timeout: 20 * time.Second}
	for _, conv := range convs {
		msgs, err := c.Messages(conv.ChatID, conv.ChatType, 30)
		if err != nil {
			continue
		}
		for _, m := range msgs {
			raw := pictureURL(m)
			if raw == "" {
				continue
			}
			for _, px := range []int{pictureBoxPx, 76} {
				u := resized(raw, px, px)
				req, _ := http.NewRequest("GET", u, nil)
				req.Header.Set("Referer", c.Base)
				resp, err := hc.Do(req)
				if err != nil {
					t.Fatalf("%s: %v", u, err)
				}
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
				resp.Body.Close()
				kind := http.DetectContentType(body)
				if resp.StatusCode != 200 || !strings.HasPrefix(kind, "image/") {
					t.Fatalf("%s: %d %s %.80q", u, resp.StatusCode, kind, body)
				}
				t.Logf("chat %q px=%d -> %d %s %d bytes", conv.Name, px, resp.StatusCode, kind, len(body))
			}
			return
		}
	}
	t.Log("no image message in the first page of any conversation")
}

// TestLiveLoadOlder pages a real conversation back and checks that each
// page brings messages not already on screen.
func TestLiveLoadOlder(t *testing.T) {
	if os.Getenv("MYGOHU_LIVE") == "" {
		t.Skip("set MYGOHU_LIVE=1 to ask the real API")
	}
	s, err := yh.LoadSession()
	if err != nil || s == nil || s.Token == "" {
		t.Skip("no saved session")
	}
	a := New()
	a.client.Base, a.client.Token, a.client.DeviceID = s.Base, s.Token, s.DeviceID

	convs, _, err := a.client.Conversations("")
	if err != nil {
		t.Skipf("conversations: %v", err)
	}
	if len(convs) == 0 {
		t.Skip("no conversations")
	}

	// The first conversation that has more than one page of history.
	for _, conv := range convs {
		msgs, err := a.client.Messages(conv.ChatID, conv.ChatType, msgFirst)
		if err != nil || len(msgs) < msgFirst {
			continue
		}
		a.current, a.currentType = conv.ChatID, conv.ChatType
		a.msgs, a.hasMore = msgs, true

		before := len(a.msgs)
		oldest := a.msgs[0].MsgID
		a.loadOlder()
		// loadOlder runs on its own goroutine; wait for it to come back.
		for i := 0; i < 100 && a.loadingMore; i++ {
			time.Sleep(20 * time.Millisecond)
		}
		if a.loadingMore {
			t.Fatal("loading older messages did not finish")
		}
		if len(a.msgs) <= before {
			t.Fatalf("no messages came in: still %d", len(a.msgs))
		}
		// What came in sits above what was here: the message that was
		// oldest is now just past the ones that came in.
		if got := a.msgs[len(a.msgs)-before].MsgID; got != oldest {
			t.Errorf("the page did not go before the oldest: %s then %s", oldest, got)
		}
		seen := map[string]bool{}
		for _, m := range a.msgs {
			if seen[m.MsgID] {
				t.Fatalf("message %s is on screen twice", m.MsgID)
			}
			seen[m.MsgID] = true
		}
		t.Logf("chat %q: %d -> %d messages", conv.Name, before, len(a.msgs))
		return
	}
	t.Skip("no conversation with more than one page of history")
}

// TestLiveMessages reports the shape of a page of messages, to see that
// the picture fields are the ones read.
func TestLiveMessages(t *testing.T) {
	if os.Getenv("MYGOHU_LIVE") == "" {
		t.Skip("set MYGOHU_LIVE=1 to ask the real API")
	}
	s, err := yh.LoadSession()
	if err != nil || s == nil || s.Token == "" {
		t.Skip("no saved session")
	}
	c := yh.New()
	c.Base = s.Base
	c.Token = s.Token
	c.DeviceID = s.DeviceID
	convs, _, err := c.Conversations("")
	if err != nil {
		t.Skipf("conversations: %v", err)
	}
	kinds := map[int]int{}
	for _, conv := range convs {
		msgs, err := c.Messages(conv.ChatID, conv.ChatType, 20)
		if err != nil {
			continue
		}
		for _, m := range msgs {
			kinds[m.ContentType]++
		}
	}
	var b strings.Builder
	for k, n := range kinds {
		fmt.Fprintf(&b, "%d:%d ", k, n)
	}
	t.Logf("content types across %d conversations: %s", len(convs), b.String())
}

// TestLiveMarkdownShot draws real markdown messages, as the bots send
// them, to a picture to look at.
func TestLiveMarkdownShot(t *testing.T) {
	if os.Getenv("MYGOHU_LIVE") == "" {
		t.Skip("set MYGOHU_LIVE=1 to ask the real API")
	}
	s, err := yh.LoadSession()
	if err != nil || s == nil || s.Token == "" {
		t.Skip("no saved session")
	}
	a := New()
	a.client.Base, a.client.Token, a.client.DeviceID = s.Base, s.Token, s.DeviceID
	a.me = &yh.User{ID: "100", Name: "阿德"}
	a.live = yh.StateOnline

	convs, _, err := a.client.Conversations("")
	if err != nil {
		t.Skipf("conversations: %v", err)
	}
	for _, conv := range convs {
		msgs, err := a.client.Messages(conv.ChatID, conv.ChatType, 40)
		if err != nil {
			continue
		}
		var md []*yh.Message
		for _, m := range msgs {
			if m.ContentType == yh.ContentMarkdown {
				md = append(md, m)
			}
		}
		if len(md) < 2 {
			continue
		}
		if len(md) > 6 {
			md = md[len(md)-6:]
		}
		a.current, a.currentType = conv.ChatID, conv.ChatType
		a.convs = []yh.Conversation{{ChatID: conv.ChatID, ChatType: conv.ChatType, Name: conv.Name}}
		a.msgs = md
		writePNG(t, filepath.Join("..", "..", "shots", "md-live.png"),
			ui.NewTester(a.view, 1000, 700))
		t.Logf("chat %q: %d markdown messages drawn", conv.Name, len(md))
		return
	}
	t.Skip("no conversation with markdown messages")
}
