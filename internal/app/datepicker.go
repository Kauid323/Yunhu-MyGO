package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The three pages of the calendar, from the days to the years: the title
// of the calendar walks up them, and a day, a month or a year walks back
// down to the days.
const (
	pickDay = iota
	pickMonth
	pickYear
)

// The calendar is a component of its own: a date is picked from it by
// anything that needs one, and every calendar in the window is the same.
//
// It opens on the month of the day it holds, or of the day it is given
// when it holds none, and its title walks up from the days to the months
// and from the months to the years. A year picked opens its months, a
// month picked opens its days, and a day picked closes it.
const (
	pickerW     = 238
	pickerCell  = 30
	pickerGap   = 2
	pickerPad   = 12
	pickerHeadH = 30
	pickerWeekH = 24
)

// datePicker is one calendar: what it holds, which of its pages is shown,
// and the month or year that page is on.
type datePicker struct {
	open bool
	view int
	// shown is the month the days page is on, or the year the months page
	// is on: the two are the same value, read at a different size.
	shown time.Time
	// at is the day that is picked, if any.
	at time.Time
	// walked is when the page last changed, so that the new one fades in
	// rather than being there all at once.
	walked time.Time
}

// day is the day the calendar holds, or the zero time when it holds none.
func (p *datePicker) day() time.Time { return p.at }

// showDatePicker draws the calendar beside anchor while it is open, and
// reports the day that was picked, if one was in this frame. onPick is
// called with it.
func showDatePicker(c *ui.Context, anchor ui.Element, p *datePicker, onPick func(time.Time)) {
	if !p.open {
		return
	}
	if p.shown.IsZero() {
		p.shown = p.at
		if p.shown.IsZero() {
			p.shown = time.Now()
		}
		p.shown = time.Date(p.shown.Year(), p.shown.Month(), 1, 0, 0, 0, 0, p.shown.Location())
	}
	ui.PopoverBase(c, anchor, &p.open, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorBottomLeft, ui.AnchorTopLeft).Top(6).
			Width(pickerW).Padding(pickerPad).Radius(10).
			Background(ui.Hex("#2b2b2b")).Border(1, ui.Hex("#3a3a3a")).
			Shadow(0, 10, 26, 0, ui.RGBA(0, 0, 0, 0.45)).
			TransformOrigin(0, 0).Transition(menuArrives)
		panel.Children(func() {
			// The page fades in when it changes: a calendar that swapped
			// its grid at once would look like the same one with other
			// numbers.
			k := float32(1)
			if !p.walked.IsZero() {
				k = float32(c.Now().Sub(p.walked)) / float32(120*time.Millisecond)
				if k >= 1 {
					k, p.walked = 1, time.Time{}
				} else {
					c.After(16 * time.Millisecond)
				}
			}
			switch p.view {
			case pickMonth:
				pickerMonths(c, p, k)
			case pickYear:
				pickerYears(c, p, k)
			default:
				pickerDays(c, p, k, onPick)
			}
		})
	})
}

// pickerHead is the line at the top of the calendar: a step back, the
// title that walks up a page, and a step forward.
func pickerHead(c *ui.Context, p *datePicker, title string, up func()) {
	ui.Row(c).Height(pickerHeadH).AlignItems(ui.Center).Children(func() {
		pickerStep(c, iconChevronLeft, func() { p.step(-1) })
		t := ui.Text(c, title).FontSize(14).Bold().TextColor(ui.Hex("#e8e8e8")).
			Grow(1).TextAlign(ui.Center).Cursor(ui.CursorPointer)
		if t.Clicked() && up != nil {
			up()
		}
		pickerStep(c, iconChevronRight, func() { p.step(1) })
	})
}

// pickerStep is one of the two arrows of the calendar's head.
func pickerStep(c *ui.Context, ic *ui.SVG, do func()) {
	box := ui.Box(c).Size(28, 28).Radius(6).Center().Cursor(ui.CursorPointer)
	if box.Hovered() {
		box.Background(ui.RGBA(255, 255, 255, 0.08))
	}
	box.Children(func() {
		ui.Icon(c, ic).FontSize(13).TextColor(ui.Hex("#b8b8b8"))
	})
	if box.Clicked() {
		do()
	}
}

// step moves the page on or back by one of what it shows: a month, a year
// or a page of years.
func (p *datePicker) step(n int) {
	switch p.view {
	case pickYear:
		p.shown = p.shown.AddDate(n*12, 0, 0)
	case pickMonth:
		p.shown = p.shown.AddDate(n, 0, 0)
	default:
		p.shown = p.shown.AddDate(0, n, 0)
	}
	p.walked = time.Now()
}

// walk changes the page the calendar shows, and starts the fade of the one
// that comes.
func (p *datePicker) walk(view int) {
	p.view = view
	p.walked = time.Now()
}

// pickerDays is the month: the days of the week, then the days themselves.
func pickerDays(c *ui.Context, p *datePicker, k float32, onPick func(time.Time)) {
	pickerHead(c, p, p.shown.Format("2006-01"), func() { p.walk(pickMonth) })

	ui.Row(c).Height(pickerWeekH).AlignItems(ui.Center).Children(func() {
		for _, w := range []string{"日", "一", "二", "三", "四", "五", "六"} {
			ui.Text(c, w).FontSize(11.5).TextColor(ui.Hex("#8a8a8a")).
				Width(pickerCell).TextAlign(ui.Center)
		}
	})

	first := p.shown
	lead := int(first.Weekday())
	days := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, first.Location()).Day()
	today := time.Now()
	ui.Column(c).Gap(pickerGap).Opacity(k).Children(func() {
		for row := 0; row < 6; row++ {
			ui.Row(c).Gap(pickerGap).Children(func() {
				for col := 0; col < 7; col++ {
					n := row*7 + col - lead + 1
					if n < 1 || n > days {
						ui.Box(c).Size(pickerCell, pickerCell)
						continue
					}
					day := time.Date(first.Year(), first.Month(), n, 0, 0, 0, 0, first.Location())
					pickerDay(c, p, day, sameDay(day, today), onPick)
				}
			})
		}
	})
}

// pickerDay is one day of the month: blue when it is the one picked, and
// a ring when it is today.
func pickerDay(c *ui.Context, p *datePicker, day time.Time, today bool, onPick func(time.Time)) {
	picked := !p.at.IsZero() && sameDay(p.at, day)
	box := ui.Box(c).Size(pickerCell, pickerCell).Radius(6).Center().
		Cursor(ui.CursorPointer)
	switch {
	case picked:
		box.Background(ui.Hex("#0066cc"))
	case box.Hovered():
		box.Background(ui.RGBA(255, 255, 255, 0.08))
	}
	box.Children(func() {
		col := ui.Hex("#e2e2e2")
		if picked {
			col = ui.Hex("#ffffff")
		} else if today {
			col = ui.Hex("#5aa9f0")
		}
		ui.Text(c, strconv.Itoa(day.Day())).FontSize(12.5).TextColor(col)
	})
	if box.Clicked() {
		p.at = day
		p.open = false
		if onPick != nil {
			onPick(day)
		}
	}
}

// pickerMonths is the year: its twelve months, four to a row.
func pickerMonths(c *ui.Context, p *datePicker, k float32) {
	pickerHead(c, p, strconv.Itoa(p.shown.Year()), func() { p.walk(pickYear) })

	ui.Column(c).Gap(pickerGap).Opacity(k).Children(func() {
		for row := 0; row < 3; row++ {
			ui.Row(c).Gap(pickerGap).Children(func() {
				for col := 0; col < 4; col++ {
					m := time.Month(row*4 + col + 1)
					pickerMonth(c, p, m)
				}
			})
		}
	})
}

// pickerMonth is one month of the year.
func pickerMonth(c *ui.Context, p *datePicker, m time.Month) {
	picked := p.shown.Month() == m
	box := ui.Box(c).Width((pickerW - 2*pickerPad - 3*pickerGap) / 4).
		Height(34).Radius(6).Center().Cursor(ui.CursorPointer)
	switch {
	case picked:
		box.Background(ui.Hex("#0066cc"))
	case box.Hovered():
		box.Background(ui.RGBA(255, 255, 255, 0.08))
	}
	box.Children(func() {
		col := ui.Hex("#e2e2e2")
		if picked {
			col = ui.Hex("#ffffff")
		}
		ui.Text(c, strconv.Itoa(int(m))+"月").FontSize(12.5).TextColor(col)
	})
	if box.Clicked() {
		p.shown = time.Date(p.shown.Year(), m, 1, 0, 0, 0, 0, p.shown.Location())
		p.walk(pickDay)
	}
}

// pickerYears is the page of years: twelve of them, four to a row.
func pickerYears(c *ui.Context, p *datePicker, k float32) {
	start := p.shown.Year() / 12 * 12
	pickerHead(c, p, strconv.Itoa(start)+" - "+strconv.Itoa(start+11), nil)

	ui.Column(c).Gap(pickerGap).Opacity(k).Children(func() {
		for row := 0; row < 3; row++ {
			ui.Row(c).Gap(pickerGap).Children(func() {
				for col := 0; col < 4; col++ {
					pickerYear(c, p, start+row*4+col)
				}
			})
		}
	})
}

// pickerYear is one year of the page of years.
func pickerYear(c *ui.Context, p *datePicker, year int) {
	picked := p.shown.Year() == year
	box := ui.Box(c).Width((pickerW - 2*pickerPad - 3*pickerGap) / 4).
		Height(34).Radius(6).Center().Cursor(ui.CursorPointer)
	switch {
	case picked:
		box.Background(ui.Hex("#0066cc"))
	case box.Hovered():
		box.Background(ui.RGBA(255, 255, 255, 0.08))
	}
	box.Children(func() {
		col := ui.Hex("#e2e2e2")
		if picked {
			col = ui.Hex("#ffffff")
		}
		ui.Text(c, strconv.Itoa(year)).FontSize(12.5).TextColor(col)
	})
	if box.Clicked() {
		p.shown = time.Date(year, p.shown.Month(), 1, 0, 0, 0, 0, p.shown.Location())
		p.walk(pickMonth)
	}
}

// sameDay reports whether two times are the same day of the same year.
func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// monthName is a month as the months page of the calendar writes it: a
// number and the word for month in Chinese, and the month's own name in
// English, whose numbers would say nothing.
func monthName(m time.Month) string {
	if strings.HasPrefix(i18n.Lang(), "zh") {
		return strconv.Itoa(int(m)) + "月"
	}
	return m.String()[:3]
}
