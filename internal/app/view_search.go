package app

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// searchItem is one row of the results: a day that opens what came on it,
// or what was said. The pictures of a day are one row of up to three, as
// the window this one copies shows them.
type searchItem struct {
	head bool
	day  time.Time
	hits []yh.SearchHit
}

// searchView is the window the chat's history is searched in: the field,
// the tabs, the results, and the filters beside them.
func (a *App) searchView(c *ui.Context) {
	qq = skin(a.dark(c))
	c.SetTheme(theme(qq.dark))
	i18n.Set(c.Locale())

	s := a.search
	if s == nil {
		return
	}
	a.searchLoadIfDue(c, s)

	ui.Column(c).Fill().Background(qq.chatBg).Children(func() {
		a.searchTitle(c, s)
		a.searchField(c, s)
		a.searchTabs(c, s)
		ui.Box(c).Height(1).Background(qq.border)
		ui.Row(c).Grow(1).MinHeight(0).Children(func() {
			a.searchResults(c, s)
			if s.filter {
				a.searchFilterPanel(c, s)
			}
		})
	})
}

// searchTitle is the strip across the top of the window: what the search
// is of, and the buttons of the window in its corner.
func (a *App) searchTitle(c *ui.Context, s *searchWin) {
	ui.Row(c).Height(36).AlignItems(ui.Center).DragWindow().Children(func() {
		title := a.currentTitle()
		ui.Text(c, title).FontSize(13).TextColor(qq.text).
			Grow(1).TextAlign(ui.Center).SingleLine()
	})
}

// currentTitle is what the chat that is open is called.
func (a *App) currentTitle() string {
	for _, conv := range a.convs {
		if conv.ChatID == a.current {
			return conv.Title()
		}
	}
	return i18n.T("search.title")
}

// searchField is the box the words are typed in.
func (a *App) searchField(c *ui.Context, s *searchWin) {
	box := ui.Box(c).Height(30).Margin(14, 14, 0, 14).Radius(15).
		Background(qq.inputBg)
	box.Children(func() {
		ui.Icon(c, iconSearch).FontSize(14).TextColor(qq.textFaint).
			Absolute().Left(10).Top(8)
		in := ui.TextInput(c, &s.word).Fill().
			Placeholder(i18n.T("search.placeholder")).
			Background(ui.Transparent).Border(0, ui.Transparent).
			FontSize(12.5).TextColor(qq.text)
		_ = in
	})
	// The field is typed in, not entered: the search is asked for once the
	// typing has stopped.
	if s.word != s.asked && !s.loading {
		s.due = c.Now().Add(280 * time.Millisecond)
		c.After(280 * time.Millisecond)
	}
}

// searchLoadIfDue asks for the page the window shows once the typing has
// stopped, or at once when it is a tab that was picked.
func (a *App) searchLoadIfDue(c *ui.Context, s *searchWin) {
	if s.loading || s.due.IsZero() || c.Now().Before(s.due) {
		return
	}
	s.due = time.Time{}
	a.searchLoad()
}

// searchTabs is the row of what the search can be narrowed to, and the
// button that opens the filters.
func (a *App) searchTabs(c *ui.Context, s *searchWin) {
	ui.Row(c).Height(34).Padding(0, 14).AlignItems(ui.Center).Gap(22).Children(func() {
		keys := [...]string{"search.tab.all", "search.tab.media", "search.tab.sticker",
			"search.tab.file", "search.tab.link"}
		for i, key := range keys {
			a.searchTab(c, s, i, key)
		}
		ui.Spacer(c)
		on := a.searchFiltered()
		box := ui.Row(c).Height(22).Padding(0, 9).Gap(5).Radius(11).
			AlignItems(ui.Center).Cursor(ui.CursorPointer)
		if s.filter || on {
			box.Background(qq.inputBg)
		}
		box.Children(func() {
			col := qq.textMuted
			if on {
				col = qq.accent
			}
			ui.Text(c, i18n.T("search.filter")).FontSize(11.5).TextColor(col)
			ui.Icon(c, iconFilter).FontSize(12).TextColor(col)
		})
		if box.Clicked() {
			s.filter = !s.filter
		}
	})
}

// searchTab is one of the tabs: the word, and the line under the one that
// is taken.
func (a *App) searchTab(c *ui.Context, s *searchWin, tab int, key string) {
	on := s.tab == tab
	box := ui.Column(c).Gap(6).AlignItems(ui.Center).Cursor(ui.CursorPointer)
	box.Children(func() {
		ui.Text(c, i18n.T(key)).FontSize(12.5).
			TextColor(map[bool]ui.Color{true: qq.accent, false: qq.textMuted}[on])
		line := ui.Box(c).Height(2).Radius(1).Width(18)
		if on {
			line.Background(qq.accent)
		}
	})
	if box.Clicked() && !on {
		s.tab = tab
		s.list = ui.ListState{}
		a.searchLoad()
	}
}

// searchResults is what the search found, a day at a time.
func (a *App) searchResults(c *ui.Context, s *searchWin) {
	items := a.searchItems()
	switch {
	case s.loading && len(items) == 0:
		ui.Box(c).Grow(1).FillHeight().Center().Children(func() {
			ui.Spinner(c)
		})
		return
	case !s.err.empty():
		ui.Box(c).Grow(1).FillHeight().Center().Padding(30).Children(func() {
			ui.Text(c, s.err.String()).FontSize(12).TextColor(qq.error).Wrap()
		})
		return
	case len(items) == 0:
		a.searchEmpty(c, s)
		return
	}
	ui.List(c, &s.list, len(items), func(i int) {
		it := items[i]
		if it.head {
			a.searchDay(c, it.day)
			return
		}
		a.searchRow(c, s, it.hits)
	}).Grow(1).FillHeight().Padding(0, 16).Gap(0)
}

// searchEmpty is the window with nothing in it: a mark and the word for
// what was not found.
func (a *App) searchEmpty(c *ui.Context, s *searchWin) {
	ui.Box(c).Grow(1).FillHeight().Center().Children(func() {
		ui.Column(c).AlignItems(ui.Center).Gap(16).Children(func() {
			ui.Box(c).Size(120, 120).Radius(60).Border(1, qq.border).Center().
				Children(func() {
					ui.Icon(c, iconChat).FontSize(46).TextColor(qq.textMuted)
				})
			key := "search.empty"
			if s.tab != searchAll {
				key = "search.emptyKind"
			}
			ui.Text(c, i18n.T(key)).FontSize(13).TextColor(qq.textMuted)
		})
	})
}

// searchDay is the line that opens a day of the results.
func (a *App) searchDay(c *ui.Context, day time.Time) {
	ui.Column(c).Gap(7).Padding(0, 0, 0, 0).Children(func() {
		ui.Text(c, day.Format("2006/01/02")).FontSize(12.5).TextColor(qq.textMuted).
			Margin(12, 0, 0, 0)
		ui.Box(c).Height(1).Background(qq.border)
	})
}

// searchRow is one message the search found, or the pictures of one day
// beside each other.
func (a *App) searchRow(c *ui.Context, s *searchWin, hits []yh.SearchHit) {
	ui.Row(c).Gap(10).Padding(12, 0, 14, 0).AlignItems(ui.Start).Children(func() {
		a.avatar(c, hits[0].AvatarURL, hits[0].Name, 40)
		ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, hits[0].Name).FontSize(11).TextColor(qq.textMuted).SingleLine()
				ui.Text(c, hits[0].Time.Format("15:04")).FontSize(11).
					TextColor(qq.textFaint)
			})
			if s.tab == searchMedia || s.tab == searchSticker {
				ui.Row(c).Gap(8).Children(func() {
					for _, h := range hits {
						a.searchThumb(c, h)
					}
				})
				return
			}
			for _, h := range hits {
				a.searchText(c, s, h)
			}
		})
	})
}

// searchThumb is one picture of the results, as large as the window this
// one copies draws it.
func (a *App) searchThumb(c *ui.Context, h yh.SearchHit) {
	box := ui.Box(c).Size(88, 88).Radius(6).Background(qq.inputBg).
		Center().Cursor(ui.CursorPointer)
	box.Children(func() {
		url := h.URL
		if url == "" {
			return
		}
		bmp := a.images.get(url, 176, a)
		if bmp == nil {
			ui.Icon(c, iconImage).FontSize(22).TextColor(qq.textFaint)
			return
		}
		bw, bh := ui.ImageSize(bmp)
		k := min(88/bw, 88/bh)
		ui.Image(c, bmp).Width(bw * k).Height(bh * k).Radius(6)
	})
}

// searchText is what a message says: a link, a quote, or the words
// themselves, with the pictures of a file said as their names.
func (a *App) searchText(c *ui.Context, s *searchWin, h yh.SearchHit) {
	switch h.Type {
	case yh.ContentImage:
		ui.Text(c, i18n.T("search.image")).FontSize(13).TextColor(qq.textMuted)
	case yh.ContentSticker:
		ui.Text(c, i18n.T("search.sticker")).FontSize(13).TextColor(qq.textMuted)
	case yh.ContentFile:
		ui.Text(c, h.Content).FontSize(13).TextColor(qq.text).Wrap()
	default:
		ui.Text(c, h.Content).FontSize(13).TextColor(qq.text).LineHeight(1.5).Wrap()
	}
}

// searchItems groups what was found into days, and the pictures of a day
// into rows of three.
func (a *App) searchItems() []searchItem {
	hits := a.searchShown()
	grid := a.search != nil && (a.search.tab == searchMedia || a.search.tab == searchSticker)
	out := make([]searchItem, 0, len(hits)+4)
	var day time.Time
	for i := 0; i < len(hits); i++ {
		h := hits[i]
		d := startOfDay(h.Time)
		if day.IsZero() || !sameDay(day, d) {
			day = d
			out = append(out, searchItem{head: true, day: d})
		}
		if !grid {
			out = append(out, searchItem{hits: hits[i : i+1]})
			continue
		}
		row := []yh.SearchHit{h}
		for len(row) < 3 && i+1 < len(hits) && sameDay(startOfDay(hits[i+1].Time), d) {
			i++
			row = append(row, hits[i])
		}
		out = append(out, searchItem{hits: row})
	}
	return out
}

// searchFilterPanel is the panel beside the results: the day the search is
// narrowed to, and who said it.
func (a *App) searchFilterPanel(c *ui.Context, s *searchWin) {
	ui.Column(c).Width(searchFilterW).FillHeight().Background(qq.panelBg).
		BorderWidth(0, 0, 0, 1).BorderColor(qq.border).
		Padding(14, 16, 16, 16).Children(func() {
		ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("search.filterTitle")).FontSize(13).
				TextColor(qq.text).Grow(1)
			if a.searchFiltered() {
				reset := ui.Text(c, i18n.T("search.reset")).FontSize(11.5).
					TextColor(qq.accent).Cursor(ui.CursorPointer).Margin(0, 10, 0, 10)
				if reset.Clicked() {
					a.searchReset()
				}
			}
			a.closeBox(c, func() { s.filter = false })
		})

		ui.Text(c, i18n.T("search.date")).FontSize(11.5).
			TextColor(qq.textMuted).Margin(14, 0, 0, 0)
		dateBtn := a.searchFilterField(c, 8, a.dateLabel(s), iconCalendar, s.to.IsZero())
		if dateBtn.Clicked() {
			s.picker.open = !s.picker.open
			s.senderOpen = false
		}
		showDatePicker(c, dateBtn, &s.picker, func(d time.Time) {
			s.from, s.to = d, d
		})

		ui.Text(c, i18n.T("search.sender")).FontSize(11.5).
			TextColor(qq.textMuted).Margin(16, 0, 0, 0)
		senderBtn := a.searchFilterField(c, 8, a.senderLabel(s), iconChevron, len(s.senders) == 0)
		if senderBtn.Clicked() {
			s.senderOpen = !s.senderOpen
			s.picker.open = false
		}
		if s.senderOpen {
			a.searchSenderList(c, s, senderBtn)
		}
	})
}

// searchFilterField is one of the two boxes of the filter panel: what it
// holds, and the mark at its right.
func (a *App) searchFilterField(c *ui.Context, top float32, label string, ic *ui.SVG, dim bool) ui.Element {
	box := ui.Box(c).FillWidth().Height(30).Margin(top, 0, 0, 0).Radius(6).
		Border(1, qq.border).Background(qq.inputBg).Cursor(ui.CursorPointer)
	box.Children(func() {
		col := qq.text
		if dim {
			col = qq.textMuted
		}
		ui.Text(c, label).FontSize(12).TextColor(col).
			Absolute().Left(9).Top(7).SingleLine()
		ui.Icon(c, ic).FontSize(13).TextColor(qq.textMuted).
			Absolute().Right(9).Top(8)
	})
	return box
}

// dateLabel is what the date box says: the day that was picked, or that
// none was.
func (a *App) dateLabel(s *searchWin) string {
	if s.to.IsZero() {
		return i18n.T("search.pickDate")
	}
	if !s.from.IsZero() && !sameDay(s.from, s.to) {
		return s.from.Format("2006/01/02") + " - " + s.to.Format("2006/01/02")
	}
	return s.to.Format("2006/01/02")
}

// senderLabel is what the sender box says.
func (a *App) senderLabel(s *searchWin) string {
	switch len(s.senders) {
	case 0:
		return i18n.T("search.pickSender")
	case 1:
		for name := range s.senders {
			return name
		}
	}
	return i18n.T("search.selected", i18n.Num(len(s.senders)))
}

// searchSenderList is the list of who said something in what was found,
// hung under the box it belongs to, with a tick on the ones picked.
func (a *App) searchSenderList(c *ui.Context, s *searchWin, anchor ui.Element) {
	ui.PopoverBase(c, anchor, &s.senderOpen, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorBottomLeft, ui.AnchorTopLeft).Top(6).
			Width(searchFilterW-32).Padding(4).Radius(8).
			Background(qq.menuBg).Border(1, qq.menuLine).
			Shadow(0, 10, 26, 0, ui.RGBA(0, 0, 0, 0.40)).
			Transition(menuArrives)
		panel.Children(func() {
			rows := a.searchSenders()
			if len(rows) == 0 {
				ui.Text(c, i18n.T("search.noSender")).FontSize(12).
					TextColor(qq.textMuted).Padding(10, 8)
				return
			}
			ui.Column(c).MaxHeight(320).Children(func() {
				ui.List(c, &s.senderList, len(rows), func(i int) {
					h := rows[i]
					row := ui.Row(c).Height(34).Padding(0, 8).Gap(8).
						AlignItems(ui.Center).Radius(5).Cursor(ui.CursorPointer)
					if row.Hovered() {
						row.Background(qq.menuHover)
					}
					row.Children(func() {
						a.avatar(c, h.AvatarURL, h.Name, 22)
						ui.Text(c, h.Name).FontSize(12).TextColor(qq.menuText).
							Grow(1).MinWidth(0).SingleLine()
						if s.senders[h.Name] {
							ui.Icon(c, iconCheck).FontSize(12).TextColor(qq.accent)
						}
					})
					if row.Clicked() {
						if s.senders[h.Name] {
							delete(s.senders, h.Name)
						} else {
							s.senders[h.Name] = true
						}
					}
				}).FillHeight()
			})
		})
	})
}

// closeBox is the small cross that closes a panel.
func (a *App) closeBox(c *ui.Context, do func()) {
	box := ui.Box(c).Size(22, 22).Radius(5).Center().Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(qq.inputBg)
	}
	box.Children(func() {
		ui.Icon(c, iconClose).FontSize(11).TextColor(qq.textMuted)
	})
	if box.Clicked() {
		do()
	}
}

// searchCount is how many of the results are shown, which the window says
// nowhere else.
func searchCount(hits []yh.SearchHit) string { return strconv.Itoa(len(hits)) }
