package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// history fills the open chat with n messages, oldest first.
func history(a *App, n int) {
	base := time.Now().Add(-time.Duration(n) * time.Minute)
	for i := 0; i < n; i++ {
		a.msgs = append(a.msgs, &yh.Message{
			MsgID:       fmt.Sprintf("m%03d", i),
			ChatID:      "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: fmt.Sprintf("第 %d 条", i)},
			TimestampMS: uint64(base.Add(time.Duration(i) * time.Minute).UnixMilli()),
		})
	}
}

// openChat is the app with a chat open and n messages in it.
func openChat(n int) *App {
	a := New()
	a.me = &yh.User{ID: "100", Name: "阿德"}
	a.current, a.currentType = "g1", yh.ChatGroup
	history(a, n)
	return a
}

// TestFarBack checks the point at which more history is asked for: four
// fifths of it read, so that its top fifth is in view.
func TestFarBack(t *testing.T) {
	cases := []struct {
		first, last, rows int
		want              bool
	}{
		{40, 49, 50, false}, // at the end, the newest in view
		{11, 20, 50, false}, // a fifth read, not yet four fifths
		{10, 19, 50, true},  // four fifths read
		{0, 9, 50, true},    // the first message in view
		{0, 2, 3, true},     // all of it in view, more to come
		{0, -1, 50, false},  // the list showed nothing
		{0, 0, 0, false},    // no rows at all
	}
	for _, c := range cases {
		if got := farBack(c.first, c.last, c.rows); got != c.want {
			t.Errorf("farBack(%d, %d, %d) = %v, want %v", c.first, c.last, c.rows, got, c.want)
		}
	}
}

// TestAutoLoadOnScroll checks that scrolling back through the history asks
// the server for more of it, without a button to press.
func TestAutoLoadOnScroll(t *testing.T) {
	asked := make(chan string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case asked <- r.URL.Path:
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	a := openChat(60)
	a.hasMore = true
	a.client.Base = srv.URL

	tt := ui.NewTester(a.view, 900, 700)
	tt.Frame()

	// The chat opens at its newest message, which is far from the top
	// fifth: nothing is asked for.
	select {
	case path := <-asked:
		t.Fatalf("asked for %q while at the end of the history", path)
	case <-time.After(150 * time.Millisecond):
	}

	// Reading back to the top fifth of it asks for the rest.
	tt.Scroll(400, 300, 0, -100000)
	tt.Frame()

	select {
	case path := <-asked:
		if path != "/v1/msg/list-message" {
			t.Errorf("asked for %q, want the message list", path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("scrolling back through the history did not load more of it")
	}
}

// TestTopAsksAgain checks that the top of the history is not a dead end: a
// page asked for at the very top goes in above what is shown, and the reader
// who goes back up to the top of the longer history asks for the next one.
// The row the page was asked at follows the page down, or the top would be a
// row nothing is ever asked at again.
func TestTopAsksAgain(t *testing.T) {
	a := openChat(60)
	a.hasMore = true

	// The reader dragged the scroll bar to the top, and the first page was
	// asked for there.
	a.loadedAt = 0
	a.prepend(olderMessages(30, "p1"))

	// At the top of the longer history there is more to ask for.
	if !a.wantMore(0, 9, len(a.msgs)) {
		t.Errorf("the top of the history is a dead end: loadedAt=%d", a.loadedAt)
	}
}

// TestFailedPageAsksAgain checks that a page that did not come does not leave
// the top of the history a dead end either: the reader sitting there has
// nowhere further back to scroll, so the row it was asked at is forgotten and
// the page is asked for again.
func TestFailedPageAsksAgain(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	a := openChat(60)
	a.hasMore = true
	a.client.Base = srv.URL
	tt := ui.NewTester(a.view, 900, 700)
	tt.Frame()

	// A drag of the scroll bar takes the reader straight to the top, which
	// is where the first page is asked for, and where it fails.
	tt.Scroll(400, 300, 0, -100000)
	tt.Frame()
	waitFor(t, &hits, 1)

	// Asked again, rather than stuck for good.
	time.Sleep(moreWait)
	tt.Scroll(400, 300, 0, -100)
	tt.Frame()
	waitFor(t, &hits, 2)
}

// waitFor waits for the requests counted by hits to reach n.
func waitFor(t *testing.T, hits *int32, n int32) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if atomic.LoadInt32(hits) >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("%d requests were made, want %d", atomic.LoadInt32(hits), n)
}

// olderMessages is a page of messages older than everything on screen.
func olderMessages(n int, tag string) []*yh.Message {
	out := make([]*yh.Message, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &yh.Message{
			MsgID:       fmt.Sprintf("%s%03d", tag, i),
			ChatID:      "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: fmt.Sprintf("更早的 %s%d", tag, i)},
		})
	}
	return out
}

// TestOlderKeepPlace checks that messages arriving above do not move what
// is being read: the message at the top stays at the top.
func TestOlderKeepPlace(t *testing.T) {
	a := openChat(60)
	tt := ui.NewTester(a.view, 900, 700)
	tt.Frame()

	first, _ := a.msgList.Visible()
	if first <= 0 {
		t.Fatalf("the list opened at row %d, not at its end", first)
	}

	// Read back some of it.
	tt.Scroll(400, 300, 0, -900)
	tt.Frame()
	first, _ = a.msgList.Visible()
	if first <= 0 {
		t.Fatalf("scrolling back left the list at row %d", first)
	}
	top := a.msgs[first].MsgID

	// Older messages come in above, as a page of them does.
	older := make([]*yh.Message, 0, 30)
	for i := 0; i < 30; i++ {
		older = append(older, &yh.Message{
			MsgID:       fmt.Sprintf("old%03d", i),
			ChatID:      "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: fmt.Sprintf("更早的 %d", i)},
		})
	}
	a.msgs = append(older, a.msgs...)
	tt.Frame()

	first, _ = a.msgList.Visible()
	if first >= len(a.msgs) {
		t.Fatalf("the list is at row %d of %d", first, len(a.msgs))
	}
	if got := a.msgs[first].MsgID; got != top {
		t.Errorf("the message at the top changed from %s to %s: the view moved", top, got)
	}
}

// TestWantMore checks the guard against paging the whole history in: a
// list that fails to keep its place keeps showing its top, and asking
// again from the same row would fetch page after page with nobody reading
// them.
func TestWantMore(t *testing.T) {
	a := openChat(60)
	a.hasMore = true

	if !a.wantMore(10, 19, 50) {
		t.Error("four fifths read did not ask for more")
	}
	// The page came back, and the list did not move: the same row is at
	// the top, so nothing more is asked for.
	a.loadedAt = 10
	if a.wantMore(10, 19, 50) {
		t.Error("asked again from the row it last asked at")
	}
	// The reader went further back: ask again.
	if !a.wantMore(9, 18, 50) {
		t.Error("a row further back did not ask for more")
	}

	// Nothing is asked for while a page is on its way, or when there is
	// no more of the history, or when the top fifth is not in view.
	a.loadedAt = 1 << 30
	a.loadingMore = true
	if a.wantMore(0, 9, 50) {
		t.Error("asked while a page was on its way")
	}
	a.loadingMore = false
	a.hasMore = false
	if a.wantMore(0, 9, 50) {
		t.Error("asked when the history was all here")
	}
	a.hasMore = true
	if a.wantMore(40, 49, 50) {
		t.Error("asked with the newest messages in view")
	}
}
