// Command mygohu is mygo湖, a native chat client for Yunhu (云湖) written
// in pure Go: the window is drawn by MyGo's native UI, and the messages
// come from the Yunhu API documented at https://yh-api.yyyyt.top/api/v1.
package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/egoist/mygo"

	// MyGo's own words — its menus, its scroll bars, what its controls say
	// to assistive technology — are a catalog of their own, and one is
	// imported per language the app speaks beyond English, which is built
	// in.
	_ "github.com/egoist/mygo/locales/zh-Hans"

	"mygohu/internal/app"
)

func main() {
	// The languages the window is written in, its own first: the desktop's
	// language is matched against them, and the first is what a desktop
	// speaking neither shows. Without this MyGo matches the desktop
	// against its own catalogs, which is how a Chinese desktop came to be
	// shown an English window.
	mygo.App.SetLanguages("zh-CN", "en-US")

	a := app.New()
	// MYGOHU_CAPTURE=shot.png draws the window to that file once it has
	// loaded, which is how the interface is looked at over a remote
	// session, where a screen grab comes out black.
	if path := os.Getenv("MYGOHU_CAPTURE"); path != "" {
		go func() {
			time.Sleep(6 * time.Second)
			if err := a.Capture(path); err != nil {
				log.Println("capture:", err)
			}
		}()
	}
	dbg := func() string { return "" }
	if os.Getenv("MYGOHU_MEM") != "" {
		dbg = func() string { return a.DebugStats() }
		go func() {
			for {
				time.Sleep(5 * time.Second)
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				log.Printf("heap=%dMB sys=%dMB stack=%dMB objs=%d gc=%d %s",
					m.HeapAlloc>>20, m.Sys>>20, m.StackSys>>20, m.HeapObjects, m.NumGC, dbg())
			}
		}()
	}
	// MYGOHU_SETTINGS=1 opens the settings window as the app starts, which
	// is how it is looked at without picking it out of the menu, and
	// MYGOHU_CAPTURE_SETTINGS=shot.png draws that window to a file, as
	// MYGOHU_CAPTURE draws the window of the chat.
	openSettings := os.Getenv("MYGOHU_SETTINGS") != ""
	settingsShot := os.Getenv("MYGOHU_CAPTURE_SETTINGS")
	openSearch := os.Getenv("MYGOHU_SEARCH") != ""
	searchShot := os.Getenv("MYGOHU_CAPTURE_SEARCH")
	mygo.App.WhenReady(func() {
		a.Start()
		if openSearch {
			// The window the history is searched in, opened on the first
			// chat once the list of them has come back.
			go func() {
				time.Sleep(5 * time.Second)
				a.OpenFirstSearch()
				if searchShot == "" {
					return
				}
				time.Sleep(3 * time.Second)
				if err := a.CaptureSearch(searchShot); err != nil {
					log.Println("capture search:", err)
				}
			}()
		}
		if !openSettings {
			return
		}
		a.OpenSettings()
		if settingsShot == "" {
			return
		}
		go func() {
			time.Sleep(4 * time.Second)
			if err := a.CaptureSettings(settingsShot); err != nil {
				log.Println("capture settings:", err)
			}
		}()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
