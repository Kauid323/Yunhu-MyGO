package app

import (
	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// The sign-in page: whether the window starts with the computer, whether it
// signs in by itself, and whether it asks for the account again once it has
// been signed out.
func (a *App) signInPage(c *ui.Context, s *settings) {
	gap(c, 30)
	sectionTitle(c, "settings.signIn.start")
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.signIn.autoStart", "", &a.prefs.AutoStart)
		a.switchRow(c, true, "settings.signIn.autoLogin", "", &a.prefs.AutoLogin)
	})

	gap(c, 24)
	sectionTitle(c, "settings.signIn.way")
	settingCard(c, 16, 0, func() {
		a.switchRow(c, false, "settings.signIn.quick", "", &a.prefs.QuickLogin)
	})
}

// aboutPage is what this window is: its mark, the version it is, and what it
// is built on. The version is the app's own — the client this one is written
// for has one of its own, which says nothing about this build.
func (a *App) aboutPage(c *ui.Context, s *settings) {
	gap(c, 30)
	settingCard(c, 16, 14, func() {
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(40, 40).Radius(10).Center().
				Background(qq.accent).Children(func() { a.mark(c) })
			ui.Text(c, i18n.T("about.version", appVersion)).
				FontSize(13).TextColor(qq.text)
		})
	})

	ui.Spacer(c)
	ui.Column(c).AlignItems(ui.Center).Gap(5).Padding(0, 0, 12, 0).Children(func() {
		ui.Text(c, i18n.T("about.built")).FontSize(10.5).
			TextColor(qq.setTextMuted)
		ui.Text(c, i18n.T("about.rights")).FontSize(10.5).
			TextColor(qq.setTextMuted)
	})
}
