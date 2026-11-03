package app

import "github.com/egoist/mygo/ui"

// A palette is every color the window is drawn in. There are two of them,
// one per appearance of the desktop: the window takes the one the system
// is in, so that it changes with the desktop rather than at a setting of
// its own.
//
// Both are built around the background the user asked for: #222222 for the
// dark one and #f5f5f5 for the light one, with the rail, the list and the
// panels a step or two off it in the same gray.
type palette struct {
	dark bool

	rail, railHover, railActive      ui.Color
	railIcon, railIconActive         ui.Color
	listBg, listHover, listSelected  ui.Color
	chatBg, panelBg, inputBg, cardBg ui.Color
	border, borderSoft               ui.Color
	searchBg                         ui.Color

	surfaceHover, surfacePressed ui.Color

	bubbleOther ui.Color

	text, textMuted, textFaint ui.Color

	accent, accentHover, accentPressed, accentText ui.Color
	selection, focus, scrollbar                    ui.Color

	badge, badgeMuted  ui.Color
	tagOwner, tagAdmin ui.Color

	// error, warning and success say how a thing went, and mdCode and
	// mdRule are what markdown sets its code and its rules in: a tint of
	// the text, so that they read on either background.
	error, warning, success ui.Color
	mdCode, mdRule          ui.Color

	// timePill is behind the time of a stretch of talking, which sits on the
	// history as a mark rather than as a word. quoteBar is the line down the
	// left of what a message answers, which has no ground of its own.
	timePill, quoteBar ui.Color

	// The popup menu is a step off the panels, with its own line.
	menuBg, menuHover, menuLine ui.Color
	menuText, menuTextMuted     ui.Color

	// The settings window is a window of its own, and a darker one than
	// the chat: setBg is the ground its cards sit on, setSidebar the menu
	// down its left side, and setCard a step up from the ground.
	setBg, setSidebar, setSidebarHover, setSidebarActive ui.Color
	setCard, setLine, setText, setTextMuted              ui.Color
	// setField is the line around a drop-down, setFieldHover what it is
	// filled with while it is pointed at or open, and setFieldOpen the
	// line it takes then. setPopup is the list it opens, setPopupHover the
	// row of that list under the pointer.
	setField, setFieldHover, setFieldOpen ui.Color
	setPopup, setPopupHover, setLink      ui.Color
	// setAccent lights the picture that is picked, and fills a switch that
	// is on; setSwitchOff is the track of one that is not, and
	// setSwitchKnob the knob on it, which is white in both appearances.
	setAccent, setSwitchOff, setSwitchOn, setSwitchKnob ui.Color

	// viewerBg is behind a picture opened in a window of its own. That
	// window is dark in both appearances, as picture viewers are: a
	// photograph is looked at against the dark, not against the desktop.
	viewerBg ui.Color
}

// viewerCSS is viewerBg as a window's background color, which is CSS: the
// window is filled with it before its view paints, so that it does not
// flash white on the way up.
const viewerCSS = "#141414"

// darkSkin is the window in the dark appearance.
var darkSkin = palette{
	dark: true,

	rail:           ui.Hex("#191919"),
	railHover:      ui.Hex("#292929"),
	railActive:     ui.Hex("#343434"),
	railIcon:       ui.Hex("#8a8a8a"),
	railIconActive: ui.Hex("#ffffff"),

	listBg:       ui.Hex("#1e1e1e"),
	listHover:    ui.Hex("#282828"),
	listSelected: ui.Hex("#343434"),

	chatBg:  ui.Hex("#222222"),
	panelBg: ui.Hex("#2a2a2a"),
	inputBg: ui.Hex("#2c2c2c"),
	cardBg:  ui.Hex("#2c2c2c"),

	border:     ui.Hex("#333333"),
	borderSoft: ui.Hex("#2a2a2a"),
	searchBg:   ui.Hex("#2e2e2e"),

	surfaceHover:   ui.Hex("#333333"),
	surfacePressed: ui.Hex("#3c3c3c"),

	bubbleOther: ui.Hex("#2e2e2e"),

	text:      ui.Hex("#e6e6e6"),
	textMuted: ui.Hex("#8f8f8f"),
	textFaint: ui.Hex("#6b6b6b"),

	accent:        ui.Hex("#0099ff"),
	accentHover:   ui.Hex("#33adff"),
	accentPressed: ui.Hex("#0088e6"),
	accentText:    ui.Hex("#ffffff"),
	selection:     ui.RGBA(0, 153, 255, 0.30),
	focus:         ui.RGBA(0, 153, 255, 0.55),
	scrollbar:     ui.RGBA(255, 255, 255, 0.16),

	badge:      ui.Hex("#ff4d4f"),
	badgeMuted: ui.Hex("#4a4a4a"),

	tagOwner: ui.Hex("#ff9f43"),
	tagAdmin: ui.Hex("#4a9eff"),

	error:   ui.Hex("#ff6b6b"),
	warning: ui.Hex("#ffb020"),
	success: ui.Hex("#22c55e"),
	mdCode:  ui.RGBA(255, 255, 255, 0.10),
	mdRule:  ui.RGBA(255, 255, 255, 0.16),

	timePill: ui.Hex("#2c2c2c"),
	quoteBar: ui.RGBA(255, 255, 255, 0.22),

	menuBg:        ui.Hex("#303030"),
	menuHover:     ui.Hex("#404040"),
	menuLine:      ui.Hex("#3d3d3d"),
	menuText:      ui.Hex("#e6e6e6"),
	menuTextMuted: ui.Hex("#a0a0a0"),

	setBg:            ui.Hex("#111111"),
	setSidebar:       ui.Hex("#1b1b1b"),
	setSidebarHover:  ui.Hex("#262626"),
	setSidebarActive: ui.Hex("#373737"),
	setCard:          ui.Hex("#262626"),
	setLine:          ui.Hex("#333333"),
	setText:          ui.Hex("#e8e8e8"),
	setTextMuted:     ui.Hex("#9a9a9a"),
	setField:         ui.Hex("#373737"),
	setFieldHover:    ui.Hex("#303030"),
	setFieldOpen:     ui.Hex("#4e4e4e"),
	setPopup:         ui.Hex("#303030"),
	setPopupHover:    ui.Hex("#494949"),
	setLink:          ui.Hex("#2d77e5"),
	setAccent:        ui.Hex("#0066cc"),
	setSwitchOff:     ui.Hex("#515151"),
	setSwitchOn:      ui.Hex("#0066cc"),
	setSwitchKnob:    ui.Hex("#ffffff"),

	viewerBg: ui.Hex("#141414"),
}

// lightSkin is the window in the light appearance.
var lightSkin = palette{
	dark: false,

	rail:           ui.Hex("#e9e9e9"),
	railHover:      ui.Hex("#dedede"),
	railActive:     ui.Hex("#d4d4d4"),
	railIcon:       ui.Hex("#7a7a7a"),
	railIconActive: ui.Hex("#1a1a1a"),

	listBg:       ui.Hex("#f0f0f0"),
	listHover:    ui.Hex("#e8e8e8"),
	listSelected: ui.Hex("#dcdcdc"),

	chatBg:  ui.Hex("#f5f5f5"),
	panelBg: ui.Hex("#ffffff"),
	inputBg: ui.Hex("#ffffff"),
	cardBg:  ui.Hex("#ffffff"),

	border:     ui.Hex("#e2e2e2"),
	borderSoft: ui.Hex("#ededed"),
	searchBg:   ui.Hex("#e6e6e6"),

	surfaceHover:   ui.Hex("#f0f0f0"),
	surfacePressed: ui.Hex("#e6e6e6"),

	bubbleOther: ui.Hex("#ffffff"),

	text:      ui.Hex("#1a1a1a"),
	textMuted: ui.Hex("#767676"),
	textFaint: ui.Hex("#9c9c9c"),

	accent:        ui.Hex("#0099ff"),
	accentHover:   ui.Hex("#33adff"),
	accentPressed: ui.Hex("#0088e6"),
	accentText:    ui.Hex("#ffffff"),
	selection:     ui.RGBA(0, 153, 255, 0.22),
	focus:         ui.RGBA(0, 153, 255, 0.45),
	scrollbar:     ui.RGBA(0, 0, 0, 0.28),

	badge:      ui.Hex("#ff4d4f"),
	badgeMuted: ui.Hex("#c4c4c4"),

	tagOwner: ui.Hex("#e8890c"),
	tagAdmin: ui.Hex("#1a73e8"),

	error:   ui.Hex("#d93f3f"),
	warning: ui.Hex("#b45309"),
	success: ui.Hex("#15803d"),
	mdCode:  ui.RGBA(0, 0, 0, 0.06),
	mdRule:  ui.RGBA(0, 0, 0, 0.14),

	timePill: ui.Hex("#e6e6e6"),
	quoteBar: ui.RGBA(0, 0, 0, 0.20),

	menuBg:        ui.Hex("#ffffff"),
	menuHover:     ui.Hex("#f0f0f0"),
	menuLine:      ui.Hex("#e6e6e6"),
	menuText:      ui.Hex("#1a1a1a"),
	menuTextMuted: ui.Hex("#8a8a8a"),

	setBg:            ui.Hex("#ffffff"),
	setSidebar:       ui.Hex("#f2f2f2"),
	setSidebarHover:  ui.Hex("#e8e8e8"),
	setSidebarActive: ui.Hex("#e0e0e0"),
	setCard:          ui.Hex("#f5f5f5"),
	setLine:          ui.Hex("#e6e6e6"),
	setText:          ui.Hex("#1a1a1a"),
	setTextMuted:     ui.Hex("#8a8a8a"),
	setField:         ui.Hex("#d9d9d9"),
	setFieldHover:    ui.Hex("#efefef"),
	setFieldOpen:     ui.Hex("#b3b3b3"),
	setPopup:         ui.Hex("#ffffff"),
	setPopupHover:    ui.Hex("#f0f0f0"),
	setLink:          ui.Hex("#0066cc"),
	setAccent:        ui.Hex("#0066cc"),
	setSwitchOff:     ui.Hex("#c4c4c4"),
	setSwitchOn:      ui.Hex("#0066cc"),
	setSwitchKnob:    ui.Hex("#ffffff"),

	viewerBg: ui.Hex("#141414"),
}

// A shotSkin is the few colors a thumbnail of the window is drawn in. A
// thumbnail shows the window in one appearance, not in the one this window
// is in — the settings window offers both, so both are drawn whatever the
// desktop is set to — so its colors are here rather than taken from qq.
type shotSkin struct {
	ground, window, bar, mark, bubble ui.Color
}

// shotInk is the word written on a swatch of the palette page: a swatch is a
// picture of a light theme whatever the window is drawn in, so its name is
// dark in both appearances.
var shotInk = ui.Hex("#3b3b3b")

var (
	dayShot = shotSkin{
		ground: ui.Hex("#e0e0e0"),
		window: ui.Hex("#ffffff"),
		bar:    ui.Hex("#f0f0f0"),
		mark:   ui.Hex("#dcdcdc"),
		bubble: ui.Hex("#cbedfe"),
	}
	nightShot = shotSkin{
		ground: ui.Hex("#171717"),
		window: ui.Hex("#1f1f1f"),
		bar:    ui.Hex("#2a2a2a"),
		mark:   ui.Hex("#3d3d3d"),
		bubble: ui.Hex("#00334d"),
	}
)

// qq is the palette of the frame being built. view sets it from the
// appearance the system is in, and every drawing function reads it: the
// alternative is a palette threaded through every call of every view for a
// value that is the same for all of them.
var qq = darkSkin

// skin is the palette of an appearance.
func skin(dark bool) palette {
	if dark {
		return darkSkin
	}
	return lightSkin
}

var (
	darkTheme  = themeOf(darkSkin)
	lightTheme = themeOf(lightSkin)
)

// themeOf is the mygo theme the widgets see, so that buttons, inputs and
// scroll bars come out in the same skin as the panels drawn by hand.
func themeOf(p palette) *ui.Theme {
	return &ui.Theme{
		Dark:           p.dark,
		Background:     p.chatBg,
		Surface:        p.inputBg,
		SurfaceHover:   p.surfaceHover,
		SurfacePressed: p.surfacePressed,
		Border:         p.border,
		Text:           p.text,
		TextMuted:      p.textMuted,
		Accent:         p.accent,
		AccentHover:    p.accentHover,
		AccentPressed:  p.accentPressed,
		AccentText:     p.accentText,
		Danger:         p.error,
		Warning:        p.warning,
		Success:        p.success,
		Selection:      p.selection,
		Focus:          p.focus,
		Scrollbar:      p.scrollbar,
		Radius:         8,
		Spacing:        4,
		FontSize:       13,
	}
}

// theme is the mygo theme of an appearance.
func theme(dark bool) *ui.Theme {
	if dark {
		return darkTheme
	}
	return lightTheme
}
