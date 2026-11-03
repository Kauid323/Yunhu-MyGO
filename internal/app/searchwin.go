package app

import (
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The tabs of the search window: everything that was said, the pictures,
// the stickers, the files, and the links.
const (
	searchAll = iota
	searchMedia
	searchSticker
	searchFile
	searchLink
)

// The window the history is searched in, as the client it copies draws it.
const (
	searchW, searchH = 660, 720
	searchFilterW    = 300
	// searchDayH is the band a day of the results opens with, which the
	// day held under the tabs is the same height as: the two are one band
	// handing over to the other, so they have to measure the same.
	searchDayH = 34
)

// searchWin is the window a chat's history is searched in: what is being
// looked for, which tab of it is shown, the filters beside the results,
// and what the search has found so far.
type searchWin struct {
	win      *mygo.Window
	chatID   string
	chatType int

	word string
	tab  int
	// filter says the panel of filters is open beside the results.
	filter bool
	// loading says a page is being asked for; asked and askedTab are what
	// it was asked for, so that one search is asked for once.
	loading  bool
	asked    string
	askedTab int
	// due is when the word being typed is worth asking for: a search is
	// asked for once the typing has stopped.
	due time.Time

	hits []yh.SearchHit
	err  line
	// more says the route may have older messages than the ones here,
	// and loadingMore that a page of them is being asked for.
	more        bool
	loadingMore bool
	// list is the state of the results, and senderList the state of
	// the list of senders: a state shown by two lists is a panic.
	list       ui.ListState
	senderList ui.ListState
	// The line under the tab that is taken slides from the one it was
	// under to the one it is under now: tabX and tabW are where it is,
	// and tabFromX and tabFromW where it comes from, as of tabAt.
	tabX, tabW         float32
	tabFromX, tabFromW float32
	tabAt              time.Time
	// The day at the top of the results is pinned under the tabs: sticky
	// is the day shown, stickyFrom the one it took over from, and
	// stickyAt when it changed, so that one hands over to the other.
	sticky     time.Time
	stickyFrom time.Time
	stickyAt   time.Time

	// The filters: a span of days, and the senders that were picked.
	from, to   time.Time
	senders    map[string]bool
	senderOpen bool
	picker     datePicker
}

// OpenSearch opens the window the chat's history is searched in, or brings
// the one that is open forward: there is one search window, whatever chat
// it was last about.
func (a *App) OpenSearch() {
	if a.current == "" {
		return
	}
	if a.search != nil && a.search.win != nil && !a.search.win.IsDestroyed() {
		s := a.search
		if s.chatID != a.current {
			s.chatID, s.chatType = a.current, a.currentType
			s.hits, s.asked, s.askedTab = nil, "", -1
			a.searchLoad()
		}
		s.win.Show()
		s.win.Focus()
		s.win.Invalidate()
		return
	}
	s := &searchWin{
		chatID:   a.current,
		chatType: a.currentType,
		senders:  map[string]bool{},
	}
	s.win = mygo.NewWindow(mygo.WindowOptions{
		Title:  i18n.T("search.title"),
		Width:  searchW,
		Height: searchH,
		// Narrow enough to sit beside a chat, and wide enough for the
		// filter panel next to the results.
		MinWidth:  520,
		MinHeight: 420,
		// The title and the buttons in its corner are the window's own,
		// as the settings window's are.
		TitleBarStyle: mygo.TitleBarHidden,
		Content:       ui.View(a.searchView),
	})
	a.search = s
	a.searchLoad()
}

// closeSearch closes the search window, if one is open.
func (a *App) closeSearch() {
	if a.search != nil && a.search.win != nil && !a.search.win.IsDestroyed() {
		a.search.win.Close()
	}
	a.search = nil
}

// searchTabKey is the kind of message a tab asks the route for.
func searchTabKey(tab int) string {
	switch tab {
	case searchMedia:
		return "2" // a picture
	case searchSticker:
		return "7"
	case searchFile:
		return "4"
	}
	// Links are said in text, and the route has no kind for them: they
	// are picked out of what comes back.
	return "all"
}

// searchLoad asks for the page the window shows.
func (a *App) searchLoad() {
	s := a.search
	if s == nil || s.loading {
		return
	}
	word, tab := s.word, s.tab
	chatID, chatType := s.chatID, s.chatType
	s.loading, s.asked, s.askedTab = true, word, tab
	// A day picked as the end of the span is the whole of that day: the
	// route answers with what came before the time it is given.
	before := time.Time{}
	if !s.to.IsZero() {
		before = s.to.Add(24 * time.Hour)
	}
	go func() {
		hits, err := a.client.SearchChat(chatID, chatType, word,
			searchTabKey(tab), before, searchPage)
		a.update(func() {
			s.loading = false
			if err != nil {
				s.err = errorLine(err)
				s.hits = nil
				return
			}
			s.err = line{}
			s.hits = hits
			s.more = len(hits) >= searchPage
			s.list.ScrollTo(0, ui.Start)
		})
	}()
}

// searchShown is what the window lists: the hits that the filters leave.
func (a *App) searchShown() []yh.SearchHit {
	s := a.search
	if s == nil {
		return nil
	}
	out := make([]yh.SearchHit, 0, len(s.hits))
	for _, h := range s.hits {
		if s.tab == searchLink && !hasLink(h.Content) {
			continue
		}
		if !s.from.IsZero() && h.Time.Before(startOfDay(s.from)) {
			continue
		}
		if !s.to.IsZero() && !h.Time.Before(startOfDay(s.to).Add(24*time.Hour)) {
			continue
		}
		if len(s.senders) > 0 && !s.senders[h.Name] {
			continue
		}
		out = append(out, h)
	}
	return out
}

// searchSenders is who has been seen saying something in what the search
// found: the route has no list of a chat's members, so the senders the
// search knows of are the ones it can filter by.
func (a *App) searchSenders() []yh.SearchHit {
	s := a.search
	if s == nil {
		return nil
	}
	seen := map[string]bool{}
	out := make([]yh.SearchHit, 0, 8)
	for _, h := range s.hits {
		if h.Name == "" || seen[h.Name] {
			continue
		}
		seen[h.Name] = true
		out = append(out, h)
	}
	return out
}

// searchFiltered reports whether a filter is on, which the button beside
// the tabs says.
func (a *App) searchFiltered() bool {
	s := a.search
	return s != nil && (!s.from.IsZero() || !s.to.IsZero() || len(s.senders) > 0)
}

// searchReset clears every filter.
func (a *App) searchReset() {
	s := a.search
	if s == nil {
		return
	}
	s.from, s.to = time.Time{}, time.Time{}
	s.senders = map[string]bool{}
	s.picker.open = false
	s.senderOpen = false
}

// startOfDay is the first moment of a day.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// hasLink reports whether a message says an address: what the links tab of
// the search window is for.
func hasLink(s string) bool {
	return strings.Contains(s, "http://") || strings.Contains(s, "https://") ||
		strings.Contains(s, "www.")
}

// OpenFirstSearch opens the window the history is searched in on the first
// chat of the list: it is what the capture of the window runs, so that the
// window can be looked at without a pointer to press the button with.
func (a *App) OpenFirstSearch() {
	// The list of chats comes back from the network, so the first of
	// them is waited for rather than assumed to be there.
	for i := 0; i < 30; i++ {
		a.update(func() {
			if a.current == "" && len(a.convs) > 0 {
				a.open(a.convs[0])
			}
		})
		time.Sleep(400 * time.Millisecond)
		if a.current != "" {
			break
		}
	}
	time.Sleep(1200 * time.Millisecond)
	a.update(func() { a.OpenSearch() })
}

// searchPage is how many messages a page of the search holds, and
// searchKeep how many the window holds at once: a search that is scrolled
// for long enough would otherwise keep the whole history of a chat in a
// window that is not even looked at.
const (
	searchPage = 60
	searchKeep = 600
)

// searchLoadMore asks for the page before the ones the window shows: the
// route answers with what came before the time it is given, so the oldest
// message here is where the next page starts.
func (a *App) searchLoadMore() {
	s := a.search
	if s == nil || s.loading || s.loadingMore || !s.more || len(s.hits) == 0 {
		return
	}
	last := s.hits[len(s.hits)-1]
	chatID, chatType, word, tab := s.chatID, s.chatType, s.word, s.tab
	s.loadingMore = true
	go func() {
		hits, err := a.client.SearchChat(chatID, chatType, word,
			searchTabKey(tab), last.Time, searchPage)
		a.update(func() {
			s.loadingMore = false
			if err != nil {
				s.err = errorLine(err)
				return
			}
			// The page starts at the message the last one ended on, which
			// is the same message again: it is dropped.
			seen := make(map[string]bool, len(s.hits))
			for _, h := range s.hits {
				seen[h.ID] = true
			}
			added := 0
			for _, h := range hits {
				if seen[h.ID] {
					continue
				}
				seen[h.ID] = true
				s.hits = append(s.hits, h)
				added++
			}
			// A page that adds nothing is the end of it, whatever its
			// length says.
			s.more = added > 0 && len(hits) >= searchPage
		})
	}()
}

// searchJump goes to the message a result is: the chat it was said in is
// opened at that message, the history is scrolled to it, and the window the
// search is in steps back so that what it found is seen.
//
// Where the message sits in the chat is known from the result, so the page it
// is on is asked for as the chat opens: a result from months ago is one
// request away rather than the pages back to it.
func (a *App) searchJump(s *searchWin, h yh.SearchHit) {
	a.jumpID, a.jumpSeq = h.ID, uint64(h.Sequence)
	a.openConvByID(s.chatID, s.chatType)
	// A chat that was already open loaded nothing for the jump, and the
	// message is looked for in what is here: found, it is only scrolled to.
	a.jumpToMessage(h.ID, uint64(h.Sequence))
	if a.win != nil && !a.win.IsDestroyed() {
		a.win.Show()
		a.win.Focus()
	}
}
