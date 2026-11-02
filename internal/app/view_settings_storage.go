package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"

	"mygohu/internal/i18n"
	"mygohu/internal/yh"
)

// The storage page: the two folders the app writes to, and what it takes up.
// The space it reports is what is on this machine — the folder of the session
// and the settings, and the pictures this run has fetched — rather than a
// figure that was made up for the page.
func (a *App) storagePage(c *ui.Context, s *settings) {
	gap(c, 30)
	sectionTitle(c, "settings.storage.where")
	settingCard(c, 16, 0, func() {
		a.folderRow(c, false, "settings.storage.files", a.filesDir(), a.chooseFilesDir)
		a.folderRow(c, true, "settings.storage.messages", a.msgDir(), a.chooseMsgDir)
	})

	gap(c, 24)
	settingCard(c, 16, 0, func() {
		ui.Row(c).Height(settingRowH).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T("settings.storage.space")).FontSize(13.5).
				TextColor(qq.setText).Grow(1).SingleLine()
			ui.Button(c, i18n.T("settings.storage.manage")).Radius(6).
				Padding(3, 12).FontSize(12).OnClick(func() {
				s.sub = storageSpacePage
				a.measureSpace()
			})
		})
	})
}

// folderRow is one of the two folders: what it holds, where it is, and the way
// to another one.
func (a *App) folderRow(c *ui.Context, divider bool, key, dir string, choose func()) {
	if divider {
		ui.Box(c).Height(1).Background(qq.setLine)
	}
	ui.Row(c).Height(settingRowNoteH).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, i18n.T(key)).FontSize(13.5).
				TextColor(qq.setText).SingleLine()
			ui.Text(c, dir).FontSize(11.5).TextColor(qq.setTextMuted).SingleLine()
		})
		ui.Button(c, i18n.T("settings.storage.change")).Radius(6).
			Padding(3, 12).FontSize(12).OnClick(choose)
	})
}

// The pages a section holds under its own.
const (
	storageSpacePage = iota + 1
	hotkeysMorePage
)

// space is what the app takes up on this machine, as the storage space page
// reports it: what the app's own folder holds, and what the pictures it has
// fetched hold.
type space struct {
	known bool
	files int64
	cache int64
	other int64
	total int64
	disk  int64 // the whole disk, for the share the page shows
	free  int64
}

// measureSpace adds up what the app holds, off the drawing thread: walking a
// folder is not something to do between two frames.
func (a *App) measureSpace() {
	if a.space.known || a.spaceGo {
		return
	}
	a.spaceGo = true
	go func() {
		var got space
		dir, err := yh.ConfigPath()
		if err == nil {
			got.other = dirSize(filepath.Dir(dir))
		}
		got.cache = a.images.bytes()
		got.files = dirSize(a.filesDir()) + dirSize(a.msgDir())
		got.total = got.files + got.cache + got.other
		got.disk, got.free = diskSpace(a.filesDir())
		a.update(func() {
			a.spaceGo = false
			got.known = true
			a.space = got
		})
	}()
}

// dirSize is what a folder holds, or zero when it is not there.
func dirSize(dir string) int64 {
	if dir == "" {
		return 0
	}
	var n int64
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		n += info.Size()
		return nil
	})
	return n
}

// storageSpacePage is what the app takes up, as a ring of what is used and
// what is not, and the three kinds of it below.
func (a *App) storageSpacePage(c *ui.Context, s *settings) {
	gap(c, 24)
	if !a.space.known {
		if !a.spaceGo {
			a.measureSpace()
		}
		ui.Spinner(c).Margin(ui.Auto)
		return
	}
	sp := a.space
	ui.Column(c).AlignItems(ui.Center).Gap(10).Children(func() {
		// The ring: the app's share of what is used, what the rest of the
		// computer uses, and what is left, as the page this one copies
		// draws it.
		used := max(sp.disk-sp.free, sp.total)
		otherUsed := max(used-sp.total, 0)
		ring := ui.Box(c).Size(180, 180).Draw(func(p *ui.Painter, r ui.Rect) {
			cx, cy := r.X+r.W/2, r.Y+r.H/2
			values := []float64{float64(sp.total), float64(otherUsed), float64(sp.free)}
			colors := []ui.Color{qq.accent, qq.setAccent, qq.setLine}
			for i, slice := range chart.Pie(values, 0.05) {
				p.FillPath(chart.Sector(cx, cy, 62, 80, slice), colors[i])
			}
		})
		ring.Children(func() {
			ui.Column(c).Absolute().Left(0).Top(60).Width(180).
				AlignItems(ui.Center).Gap(2).Children(func() {
				ui.Text(c, sizeText(sp.total)).FontSize(19).Bold().
					TextColor(qq.text)
				ui.Text(c, i18n.T("settings.storage.share", share(sp.total, sp.disk))).
					FontSize(10).TextColor(qq.setTextMuted)
			})
		})
		ui.Row(c).Gap(16).Children(func() {
			for i, key := range []string{"settings.storage.app", "settings.storage.pc", "settings.storage.free"} {
				col := qq.accent
				switch i {
				case 1:
					col = qq.setAccent
				case 2:
					col = qq.setLine
				}
				ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(7, 7).Radius(4).Background(col)
					ui.Text(c, i18n.T(key)).FontSize(10.5).TextColor(qq.setTextMuted)
				})
			}
		})
	})

	gap(c, 20)
	a.spaceCard(c, "settings.storage.chatFiles", "settings.storage.chatFilesNote", sp.files,
		i18n.T("settings.storage.manage"), a.openFilesDir)
	gap(c, 12)
	a.spaceCard(c, "settings.storage.cache", "settings.storage.cacheNote", sp.cache,
		i18n.T("settings.storage.clear"), func() {
			a.images.clear()
			a.space = space{}
			a.measureSpace()
		})
	gap(c, 12)
	a.spaceCard(c, "settings.storage.other", "settings.storage.otherNote", sp.other, "", nil)
}

// spaceCard is one kind of what the app takes up: what it is, how much, and
// what can be done about it.
func (a *App) spaceCard(c *ui.Context, key, note string, n int64, action string, do func()) {
	settingCard(c, 16, 14, func() {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
				ui.Text(c, i18n.T(key)).FontSize(13).TextColor(qq.setText).SingleLine()
				ui.Text(c, sizeText(n)).FontSize(17).Bold().TextColor(qq.text)
				ui.Text(c, i18n.T(note)).FontSize(11).TextColor(qq.setTextMuted).Wrap()
			})
			if action == "" {
				return
			}
			ui.Button(c, action).Radius(6).Padding(3, 12).FontSize(12).OnClick(do)
		})
	})
}

// sizeText is a number of bytes as a person reads it: in MB once it is worth
// one, and in KB below that.
func sizeText(n int64) string {
	mb := float64(n) / (1 << 20)
	if mb >= 1 {
		return fmt.Sprintf("%.2f MB", mb)
	}
	return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
}

// share is n of whole as a percentage, as the page writes it.
func share(n, whole int64) string {
	if whole <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(n)/float64(whole))
}

// msgDir is where the messages are kept, or the folder the desktop gives the
// app.
func (a *App) msgDir() string {
	if a.prefs.MsgDir != "" {
		return a.prefs.MsgDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Documents", "Tencent Files")
}

// chooseMsgDir asks the desktop for the folder the messages are kept in.
func (a *App) chooseMsgDir() {
	a.chooseDir(a.msgDir(), func(p *prefs, dir string) { p.MsgDir = dir })
}

// chooseDir asks the desktop for a folder and keeps it: the dialog blocks
// until it is answered, so it is asked from a goroutine, and the page that
// asked is told after. What the app holds is measured again, as a folder that
// moved is not the one that was measured.
func (a *App) chooseDir(start string, set func(p *prefs, dir string)) {
	parent := a.win
	title := i18n.T("settings.files.choose")
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      parent,
			Title:       title,
			DefaultPath: start,
			Directory:   true,
		})
		if err != nil || len(paths) == 0 {
			return
		}
		dir := paths[0]
		a.update(func() {
			a.pickPrefs(func(p *prefs) { set(p, dir) })
			a.space = space{}
		})
	}()
}
