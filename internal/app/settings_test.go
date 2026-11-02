package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/i18n"
)

// settingsTester builds the settings window without a window, and the app it
// is the settings of. It is given the room the window's own buttons take, so
// that the menu starts where it does in the window the desktop draws.
func settingsTester(t *testing.T) (*App, *settings, *ui.Tester) {
	t.Helper()
	// The settings are kept in the user's own directory, which a test must
	// neither read nor write.
	t.Setenv("APPDATA", t.TempDir())
	a := New()
	s := &settings{}
	tt := ui.NewTester(func(c *ui.Context) { a.settingsFrame(c, s) }, settingsW, settingsH)
	tt.SetTitleBar(ui.TitleBar{Height: 32, Right: 138})
	return a, s, tt
}

// TestSettingsPicks checks that the window is what it offers: a picture is
// taken by pressing its word as well as its frame, a drop-down opens over the
// row it belongs to and takes what is chosen in it, and the menu shows the
// page of the section that is picked.
func TestSettingsPicks(t *testing.T) {
	i18n.Set("zh-CN")
	defer i18n.Set("zh-CN")

	a, s, tt := settingsTester(t)
	if !tt.HasText(i18n.T("settings.general.mode")) || !tt.HasText(i18n.T("settings.look.system")) {
		t.Fatalf("the general page is not drawn; texts: %q", tt.Texts())
	}

	// The choices are the app's: picking one is what the window does, not a
	// value the page keeps.
	if err := tt.Click(i18n.T("settings.mode.classic")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.prefs.Mode != modeClassic {
		t.Errorf("the classic layout was not taken")
	}
	if err := tt.Click(i18n.T("settings.look.night")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.prefs.Look != lookNight {
		t.Errorf("the night appearance was not taken")
	}

	// The drop-down of the first row, and the other value in it.
	if err := tt.Click(i18n.T("settings.value.enter")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !s.openSend {
		t.Fatalf("pressing the drop-down did not open it")
	}
	if err := tt.Click(i18n.T("settings.value.ctrlEnter")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.prefs.SendKey != "ctrlEnter" || s.openSend {
		t.Errorf("the choice left sendKey=%q open=%v", a.prefs.SendKey, s.openSend)
	}

	// And another section: its word goes over the page, and the page that
	// was shown goes with it.
	if err := tt.Click(i18n.T("settings.nav.shortcuts")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.section != 3 {
		t.Errorf("the menu opened section %d, want 3", s.section)
	}
	if tt.HasText(i18n.T("settings.general.mode")) {
		t.Errorf("the page of the section that was left is still shown")
	}
	// The keys are the page that is shown, with the way to more of them.
	if !tt.HasText(i18n.T("settings.hotkeys.more")) {
		t.Errorf("the hotkeys page is not drawn; texts: %q", tt.Texts())
	}
}

// TestSwitchAnimates checks that a switch does not jump from one end to the
// other: the frame it is pressed in has the knob where it was, the frames
// after it carry it across, and it settles at the far end.
func TestSwitchAnimates(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		qq = darkSkin
		ui.Row(c).Fill().Background(qq.setBg).Padding(20).Children(func() {
			switchBox(c, &on, nil)
		})
	}, 120, 60)

	midY := 20 + switchH/2
	tt.ClickAt(20+switchW/2, float32(midY))
	tt.Frame()
	if !on {
		t.Fatalf("the switch did not turn on")
	}
	off := knobCentre(tt, midY)
	tt.Advance(60 * time.Millisecond)
	moving := knobCentre(tt, midY)
	tt.Advance(400 * time.Millisecond)
	settled := knobCentre(tt, midY)

	// The knob starts at the near end, is part of the way across the frame
	// after that, and ends at the far end, inset from it: the scan of its
	// white averages its first and last pixel, half a point short of its
	// middle.
	want := float32(20+switchW-switchInset-switchKnob) + float32(switchKnob-1)/2
	if off > 20+switchW/3 {
		t.Errorf("the knob starts at %v: the switch jumped", off)
	}
	if moving <= off || moving >= want {
		t.Errorf("the knob is at %v after 60ms, between %v and %v", moving, off, want)
	}
	if settled != want {
		t.Errorf("the knob settled at %v, want %v", settled, want)
	}
}

// knobCentre is where the white of a switch's knob is across the row y, in
// the window's points, or -1 when the row holds none.
func knobCentre(tt *ui.Tester, y int) float32 {
	img := tt.Image()
	first, last := -1, -1
	for x := 18; x < 50; x++ {
		r, g, b, _ := img.At(x, y).RGBA()
		if r > 0xf000 && g > 0xf000 && b > 0xf000 {
			if first < 0 {
				first = x
			}
			last = x
		}
	}
	if first < 0 {
		return -1
	}
	return float32(first+last) / 2
}
