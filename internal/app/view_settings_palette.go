package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The palette page: the colour themes the window can be drawn in. A theme is
// the accent the window is lit with; the ground of each appearance is the
// desktop's, as it always is.
func (a *App) palettePage(c *ui.Context, s *settings) {
	gap(c, 30)
	a.palettePreview(c)

	gap(c, 24)
	a.paletteGroup(c, "settings.palette.fresh", "settings.palette.freshNote", freshThemes)
	gap(c, 24)
	a.paletteGroup(c, "settings.palette.ink", "settings.palette.inkNote", inkThemes)
	gap(c, 24)
	a.paletteGroup(c, "settings.palette.mystic", "settings.palette.mysticNote", mysticThemes)
}

// palettePreview is the window as the theme draws it: a card with the accent
// in it, and the four ways of picking one beside it.
func (a *App) palettePreview(c *ui.Context) {
	settingCard(c, 12, 12, func() {
		ui.Row(c).Gap(14).Children(func() {
			shot := ui.Box(c).Grow(1).Height(150).Radius(8).
				Background(qq.setFieldHover).Children(func() {
				// A window in little: a bar, a face, two lines and the
				// accent where a bubble of the account's own would be.
				ui.Box(c).Absolute().Left(12).Top(12).Size(10, 10).
					Radius(3).Background(qq.accent)
				ui.Box(c).Absolute().Left(12).Top(40).Size(22, 22).
					Radius(11).Background(qq.setFieldOpen)
				ui.Box(c).Absolute().Left(42).Top(44).Size(90, 4).
					Radius(2).Background(qq.setFieldOpen)
				ui.Box(c).Absolute().Left(42).Top(54).Size(60, 4).
					Radius(2).Background(qq.setFieldOpen)
				ui.Box(c).Absolute().Right(12).Top(44).Size(96, 12).
					Radius(3).Background(qq.accent)
				ui.Box(c).Absolute().Right(12).Top(62).Size(64, 12).
					Radius(3).Background(qq.accent)
			})
			_ = shot
			ui.Column(c).Gap(8).Children(func() {
				for _, ic := range []*ui.SVG{iconHelp, iconStar, iconPalette, iconImage} {
					box := ui.Box(c).Size(30, 30).Radius(15).Center().
						Cursor(ui.CursorPointer).Background(qq.setFieldHover)
					if box.Hovered() {
						box.Background(qq.setFieldOpen)
					}
					box.Children(func() {
						ui.Icon(c, ic).FontSize(15).TextColor(qq.setTextMuted)
					})
				}
			})
		})
	})
}

// paletteGroup is a row of themes under its own words.
func (a *App) paletteGroup(c *ui.Context, title, note string, themes []swatch) {
	ui.Column(c).Gap(10).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, i18n.T(title)).FontSize(13.5).FontWeight(600).
				TextColor(qq.setText)
			ui.Text(c, i18n.T(note)).FontSize(11).TextColor(qq.setTextMuted)
		})
		ui.Row(c).Gap(10).Children(func() {
			for _, t := range themes {
				a.paletteSwatch(c, t)
			}
		})
	})
}

// swatch is one of the colours the palette page offers: the two ends of its
// gradient, and the accent the window is lit with while it is picked.
type swatch struct {
	name   string // a key of the dictionary
	from   ui.Color
	to     ui.Color
	accent ui.Color
}

var (
	freshThemes = []swatch{
		{"settings.theme.plain", ui.Hex("#ffffff"), ui.Hex("#e8eef6"), ui.Hex("#3d7eff")},
		{"settings.theme.blue", ui.Hex("#dbeafe"), ui.Hex("#bcd4f6"), ui.Hex("#2f6fed")},
		{"settings.theme.sea", ui.Hex("#d7eefb"), ui.Hex("#b6e0f5"), ui.Hex("#0d9bd6")},
		{"settings.theme.sakura", ui.Hex("#ffe4ef"), ui.Hex("#fbd0e6"), ui.Hex("#e0568f")},
		{"settings.theme.spring", ui.Hex("#e6f7d9"), ui.Hex("#f7f0c8"), ui.Hex("#6aa832")},
	}
	inkThemes = []swatch{
		{"settings.theme.gold", ui.Hex("#d8c9a3"), ui.Hex("#c2ac7d"), ui.Hex("#a8813c")},
		{"settings.theme.lilac", ui.Hex("#cfc4dd"), ui.Hex("#b9a8cf"), ui.Hex("#7c5fa8")},
		{"settings.theme.bean", ui.Hex("#c9d8c6"), ui.Hex("#aec4ab"), ui.Hex("#5f8a5c")},
		{"settings.theme.cinnabar", ui.Hex("#dcc3bd"), ui.Hex("#c9a49c"), ui.Hex("#b4543f")},
	}
	mysticThemes = []swatch{
		{"settings.theme.orange", ui.Hex("#f7d9b8"), ui.Hex("#efb98a"), ui.Hex("#d97a25")},
		{"settings.theme.green", ui.Hex("#cfe8cf"), ui.Hex("#a9d4ae"), ui.Hex("#2f8a4a")},
		{"settings.theme.blueMood", ui.Hex("#cfe0f5"), ui.Hex("#a9c7ea"), ui.Hex("#3a6fc4")},
		{"settings.theme.purple", ui.Hex("#e0d0f0"), ui.Hex("#c6aee4"), ui.Hex("#7b4fbb")},
	}
)

// paletteSwatch is one theme: its gradient, its name over it, and a frame when
// it is the one that is picked.
func (a *App) paletteSwatch(c *ui.Context, t swatch) {
	picked := a.prefs.Theme == t.name
	box := ui.Column(c).Width(92).Gap(4).Cursor(ui.CursorPointer).Children(func() {
		tile := ui.Box(c).Width(92).Height(64).Radius(8).
			Gradient(t.from, t.to, 45).Border(2, ui.Transparent)
		if picked {
			tile.Border(2, qq.accent)
		}
		tile.Children(func() {
			// A soft light in the middle, as the swatches of the page this
			// one copies have.
			ui.Box(c).Absolute().Left(30).Top(16).Size(32, 32).Radius(16).
				Background(ui.RGBA(255, 255, 255, 0.35))
			// The name is written dark whatever the window is drawn in: a
			// swatch is a picture of a light theme, and a light word on one
			// is a word nobody reads.
			ui.Text(c, i18n.T(t.name)).FontSize(10.5).TextColor(shotInk).
				Absolute().Left(0).Bottom(5).Width(92).TextAlign(ui.Center)
		})
	})
	if box.Clicked() {
		a.pickPrefs(func(p *prefs) { p.Theme = t.name })
	}
}
