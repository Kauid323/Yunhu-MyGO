package app

import "github.com/egoist/mygo/ui"

// icon parses the shapes of a 24×24 stroked icon, drawn in currentColor as
// icon sets draw them.
func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

// The rail's sections, the header's actions and the composer's tools.
var (
	iconChat         = icon(`<path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 9.9 9.9 0 0 1-4-.9L3 21l1.9-4.6A8.4 8.4 0 0 1 12 3.1a8.4 8.4 0 0 1 9 8.4z"/>`)
	iconContacts     = icon(`<path d="M16 20v-1.5a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4V20"/><circle cx="9.5" cy="7.5" r="3.5"/><path d="M21 20v-1.5a4 4 0 0 0-3-3.8"/><path d="M16.5 4.2a3.5 3.5 0 0 1 0 6.6"/>`)
	iconStar         = icon(`<path d="M12 3.5l2.6 5.4 5.9.8-4.3 4.2 1 5.9L12 17l-5.2 2.8 1-5.9-4.3-4.2 5.9-.8z"/>`)
	iconFolder       = icon(`<path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4l2 2.5h9A1.5 1.5 0 0 1 21 10v8a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18z"/>`)
	iconGrid         = icon(`<rect x="3" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="3" y="13.5" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="13.5" width="7.5" height="7.5" rx="1.6"/>`)
	iconSettings     = icon(`<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1A1.7 1.7 0 0 0 8.9 19a1.7 1.7 0 0 0-1.9.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 5 8.9a1.7 1.7 0 0 0-.3-1.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.9.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.9-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.9V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/>`)
	iconMenu         = icon(`<path d="M4 7h16M4 12h16M4 17h16"/>`)
	iconClose        = icon(`<path d="M5.5 5.5l13 13M18.5 5.5l-13 13"/>`)
	iconFilter       = icon(`<path d="M3 5h18l-7 8v6l-4 2v-8z"/>`)
	iconPin          = icon(`<path d="M9.5 3h5l-.8 5.5 3.3 3.3H7l3.3-3.3z"/><path d="M12 11.8V21"/>`)
	iconCopy         = icon(`<rect x="8.5" y="8.5" width="11.5" height="11.5" rx="2"/><path d="M15.5 8.5V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7.5a2 2 0 0 0 2 2h2.5"/>`)
	iconTrash        = icon(`<path d="M4 7h16"/><path d="M9.5 7V4.5h5V7"/><path d="M6.5 7l1 13h9l1-13"/>`)
	iconCheckCircle  = icon(`<circle cx="12" cy="12" r="8.5"/><path d="M8.2 12.3l2.6 2.6 5-5.2"/>`)
	iconWindow       = icon(`<rect x="3" y="4.5" width="18" height="15" rx="2"/><path d="M3 9.5h18"/>`)
	iconForward      = icon(`<path d="M13.5 5.5L20 12l-6.5 6.5"/><path d="M20 12H9a4.5 4.5 0 0 0-4.5 4.5V20"/>`)
	iconQuote        = icon(`<path d="M9.5 6.5H6A2.5 2.5 0 0 0 3.5 9v2A2.5 2.5 0 0 0 6 13.5h2.5"/><path d="M9.5 6.5v4.5c0 3.3-1.6 5.3-4 6.5"/><path d="M20 6.5h-3.5A2.5 2.5 0 0 0 14 9v2a2.5 2.5 0 0 0 2.5 2.5H19"/><path d="M20 6.5v4.5c0 3.3-1.6 5.3-4 6.5"/>`)
	iconGroupPlus    = icon(`<path d="M3 20v-1.5A4.5 4.5 0 0 1 7.5 14h4a4.5 4.5 0 0 1 4.5 4.5V20"/><circle cx="9.5" cy="7.5" r="3.5"/><path d="M18.5 8.5v5M16 11h5"/>`)
	iconUserPlus     = icon(`<circle cx="9.5" cy="8" r="3.6"/><path d="M3 20v-1.4A4.6 4.6 0 0 1 7.6 14h3.8a4.6 4.6 0 0 1 4.6 4.6V20"/><path d="M18.5 8.5v5M16 11h5"/>`)
	iconBot          = icon(`<rect x="4" y="8" width="16" height="11" rx="3"/><path d="M12 4v4"/><circle cx="9" cy="13" r="1.2"/><circle cx="15" cy="13" r="1.2"/><path d="M9.5 16.5h5"/>`)
	iconCalendar     = icon(`<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M8 3v4M16 3v4M3 10h18"/>`)
	iconChevronRight = icon(`<path d="M9 5l7 7-7 7"/>`)
	iconChevronLeft  = icon(`<path d="M15 5l-7 7 7 7"/>`)

	// The menu's own: the quick actions across its top, then its rows.
	iconBookmark   = icon(`<path d="M6 3.5h12v17l-6-4.5-6 4.5z"/><path d="M12 8l.9 1.9 2.1.3-1.5 1.5.4 2.1-1.9-1-1.9 1 .4-2.1-1.5-1.5 2.1-.3z"/>`)
	iconPalette    = icon(`<path d="M12 3a9 9 0 1 0 0 18c1 0 1.8-.8 1.8-1.7 0-.5-.2-.9-.5-1.2-.3-.3-.5-.7-.5-1.2 0-.9.8-1.7 1.7-1.7H16a5 5 0 0 0 5-5c0-4-4-7.2-9-7.2z"/><circle cx="7.5" cy="12" r="1.1" fill="currentColor" stroke="none"/><circle cx="9.8" cy="7.8" r="1.1" fill="currentColor" stroke="none"/><circle cx="14.6" cy="7.2" r="1.1" fill="currentColor" stroke="none"/>`)
	iconHistoryLog = icon(`<rect x="3" y="4.5" width="11" height="8" rx="2"/><rect x="10" y="11" width="11" height="8" rx="2"/>`)
	iconUpdate     = icon(`<circle cx="12" cy="12" r="9"/><path d="M12 16.5v-9"/><path d="M8.5 11L12 7.5 15.5 11"/>`)
	iconLock       = icon(`<rect x="4.5" y="10.5" width="15" height="10" rx="2.2"/><path d="M8 10.5V7.5a4 4 0 0 1 8 0v3"/>`)
	iconPower      = icon(`<path d="M12 3.5v8"/><path d="M7.2 6.4a8 8 0 1 0 9.6 0"/>`)

	iconSearch = icon(`<circle cx="10.5" cy="10.5" r="6.5"/><path d="M20 20l-4.8-4.8"/>`)
	iconPlus   = icon(`<path d="M12 5v14M5 12h14"/>`)

	// The settings window's sections, and the mark of the value a
	// drop-down holds.
	iconSliders  = icon(`<path d="M4 7.5h4.5M13.5 7.5h6.5M4 16.5h6.5M15.5 16.5h4.5"/><circle cx="11" cy="7.5" r="2.2"/><circle cx="13" cy="16.5" r="2.2"/>`)
	iconDisk     = icon(`<rect x="3.5" y="7" width="17" height="13.5" rx="2.5"/><path d="M8.5 7V5.5A1.5 1.5 0 0 1 10 4h4a1.5 1.5 0 0 1 1.5 1.5V7"/>`)
	iconKeyboard = icon(`<rect x="2.5" y="6" width="19" height="12" rx="2.4"/><path d="M6.5 10h.01M10 10h.01M13.5 10h.01M17 10h.01M7.5 14h9"/>`)
	iconPeople   = icon(`<path d="M15 20v-1.4a3.6 3.6 0 0 0-3.6-3.6H6.6A3.6 3.6 0 0 0 3 18.6V20"/><circle cx="9" cy="7.5" r="3.5"/><path d="M21 20v-1.4a3.6 3.6 0 0 0-2.7-3.5"/><path d="M16.5 4.2a3.5 3.5 0 0 1 0 6.6"/>`)
	iconLink     = icon(`<path d="M9.5 14.5l5-5"/><path d="M11 6.6l1.6-1.6a3.9 3.9 0 0 1 5.5 5.5L16.5 12"/><path d="M13 17.4l-1.6 1.6a3.9 3.9 0 0 1-5.5-5.5L7.5 12"/>`)
	iconCheck    = icon(`<path d="M4.5 12.5l5 5 10-11"/>`)

	iconPhone      = icon(`<path d="M21 16.9v2.6a1.7 1.7 0 0 1-1.9 1.7 17 17 0 0 1-7.4-2.6 16.7 16.7 0 0 1-5.1-5.1A17 17 0 0 1 4 6.1 1.7 1.7 0 0 1 5.7 4.2h2.6a1.7 1.7 0 0 1 1.7 1.5c.1 1 .3 1.9.7 2.8a1.7 1.7 0 0 1-.4 1.8l-1.1 1.1a13.6 13.6 0 0 0 5.1 5.1l1.1-1.1a1.7 1.7 0 0 1 1.8-.4c.9.4 1.8.6 2.8.7a1.7 1.7 0 0 1 1.5 1.7z"/>`)
	iconVideo      = icon(`<rect x="2.5" y="6" width="13" height="12" rx="2.4"/><path d="M15.5 10.5l5.2-3v9l-5.2-3z"/>`)
	iconShare      = icon(`<rect x="2.5" y="4" width="19" height="13" rx="2"/><path d="M8 21h8M12 17v4"/>`)
	iconFolderOpen = icon(`<path d="M3 8V6.5A1.5 1.5 0 0 1 4.5 5h4l2 2.5h7A1.5 1.5 0 0 1 19 9v1"/><path d="M3 8h16.5a1.5 1.5 0 0 1 1.4 2l-2 7a1.5 1.5 0 0 1-1.4 1H4.5A1.5 1.5 0 0 1 3 16.5z"/>`)
	iconHelp       = icon(`<circle cx="12" cy="12" r="9"/><path d="M9.6 9.2a2.5 2.5 0 0 1 4.8.8c0 1.7-2.4 2.5-2.4 2.5"/><path d="M12 17h.01"/>`)
	iconMore       = icon(`<circle cx="5" cy="12" r="1.4" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1.4" fill="currentColor" stroke="none"/><circle cx="19" cy="12" r="1.4" fill="currentColor" stroke="none"/>`)
	iconInfo       = icon(`<circle cx="12" cy="12" r="9"/><path d="M12 11v6M12 7.6h.01"/>`)
	iconBell       = icon(`<path d="M18 8.5a6 6 0 1 0-12 0c0 5-2 6.5-2 6.5h16s-2-1.5-2-6.5z"/><path d="M13.7 19a2 2 0 0 1-3.4 0"/>`)

	iconEmoji    = icon(`<circle cx="12" cy="12" r="9"/><path d="M8.5 14.5a4.5 4.5 0 0 0 7 0"/><path d="M9 9.5h.01M15 9.5h.01"/>`)
	iconScissors = icon(`<circle cx="6.5" cy="6.5" r="2.5"/><circle cx="6.5" cy="17.5" r="2.5"/><path d="M20 4L8.5 15.5M14.5 14.5L20 20M8.5 8.5L12 12"/>`)
	iconImage    = icon(`<rect x="3" y="4.5" width="18" height="15" rx="2.4"/><circle cx="8.5" cy="9.5" r="1.6"/><path d="M21 15.5l-4.6-4.6L6 21"/>`)
	iconMic      = icon(`<rect x="9" y="2.5" width="6" height="11" rx="3"/><path d="M5.5 11a6.5 6.5 0 0 0 13 0"/><path d="M12 17.5V21M8.5 21h7"/>`)
	iconAt       = icon(`<circle cx="12" cy="12" r="4"/><path d="M16 8v5a3 3 0 0 0 5.2 2A9 9 0 1 0 18 19.7"/>`)
	iconHistory  = icon(`<path d="M3.5 12a8.5 8.5 0 1 0 2.6-6.1"/><path d="M3 4v4.5h4.5"/><path d="M12 8v4.5l3 1.8"/>`)
	iconChevron  = icon(`<path d="M6 9l6 6 6-6"/>`)
	iconSend     = icon(`<path d="M21 3L10.5 13.5M21 3l-6.8 18-3.7-7.5L3 9.8z"/>`)
	iconLogout   = icon(`<path d="M9 21H5.5A1.5 1.5 0 0 1 4 19.5v-15A1.5 1.5 0 0 1 5.5 3H9"/><path d="M16 16.5l4.5-4.5L16 7.5"/><path d="M20.5 12H9"/>`)
)

// iconFill parses the same shapes as icon, filled rather than stroked: what
// a section of the rail becomes once it is the one that is open. The stroke
// stays, thin, so that the edge of a filled shape is as clean as a drawn
// one's.
func iconFill(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

// The filled rail: the same sections as above, solid.
var (
	iconChatFill     = iconFill(`<path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 9.9 9.9 0 0 1-4-.9L3 21l1.9-4.6A8.4 8.4 0 0 1 12 3.1a8.4 8.4 0 0 1 9 8.4z"/>`)
	iconContactsFill = iconFill(`<circle cx="9.5" cy="7.5" r="4"/><path d="M16 20v-1.5a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4V20"/><path d="M21 20v-1.5a4 4 0 0 0-3-3.8" fill="none"/><path d="M16.5 4.2a3.5 3.5 0 0 1 0 6.6" fill="none"/>`)
	iconStarFill     = iconFill(`<path d="M12 3.5l2.6 5.4 5.9.8-4.3 4.2 1 5.9L12 17l-5.2 2.8 1-5.9-4.3-4.2 5.9-.8z"/>`)
	iconFolderFill   = iconFill(`<path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4l2 2.5h9A1.5 1.5 0 0 1 21 10v8a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18z"/>`)
	iconGridFill     = iconFill(`<rect x="3" y="3" width="7.5" height="7.5" rx="1.8"/><rect x="13.5" y="3" width="7.5" height="7.5" rx="1.8"/><rect x="3" y="13.5" width="7.5" height="7.5" rx="1.8"/><rect x="13.5" y="13.5" width="7.5" height="7.5" rx="1.8"/>`)
	iconBellFill     = iconFill(`<path d="M18 8.5a6 6 0 1 0-12 0c0 5-2 6.5-2 6.5h16s-2-1.5-2-6.5z"/><path d="M13.7 19a2 2 0 0 1-3.4 0" fill="none"/>`)
)
