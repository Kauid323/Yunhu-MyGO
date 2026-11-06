package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// Pictures are kept while they cost little: a list is scrolled through
// once, and what was scrolled past is not worth the room. Both bounds
// matter: avatars are a few KB each, and a message's picture is up to a
// megabyte, so a count alone would let a chat of photos take hundreds.
const (
	imageKeep   = 320
	imageBudget = 24 << 20 // bytes of pixels kept
)

// imageCache keeps the pictures of people, groups and messages:
// downloaded once at the size they are drawn, decoded once, and drawn
// from then on.
type imageCache struct {
	mu      sync.Mutex
	bitmaps map[string]imageEntry
	total   int // bytes of pixels kept
	// frame counts the frames drawn, and used says which one a picture was
	// last drawn in: what is drawn now is not taken away to make room for
	// what is drawn next.
	frame   uint64
	loading map[string]bool
	failed  map[string]bool
	// fetchSem bounds how many pictures are being fetched at once. A panel
	// that asks for a whole collection of them in one frame would otherwise
	// have every answer in memory at once, downloaded but not yet decoded.
	fetchSem chan struct{}
	hc       *http.Client
}

// imageEntry is one picture and what its pixels cost.
type imageEntry struct {
	bmp  *ui.Bitmap
	size int
	used uint64
}

func newImageCache() *imageCache {
	return &imageCache{
		bitmaps:  map[string]imageEntry{},
		loading:  map[string]bool{},
		failed:   map[string]bool{},
		fetchSem: make(chan struct{}, 8),
		hc: &http.Client{
			Timeout: 20 * time.Second,
			// One picture is asked for at a time per URL, but several
			// at once across the rows of a list.
			Transport: &http.Transport{MaxIdleConnsPerHost: 8},
		},
	}
}

// bytes is how many bytes of pixels the cache holds: what the storage page
// reports as what the pictures this run has fetched take up.
func (c *imageCache) bytes() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return int64(c.total)
}

// busy reports whether pictures are still on their way: a card drawn in a
// window of its own waits for them before it is taken, as nothing draws
// again in a window that is never shown.
func (c *imageCache) busy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.loading) > 0
}

// beginFrame starts a frame: from here until the next one, every picture
// asked for is one the frame is drawing, and is kept while it is.
func (c *imageCache) beginFrame() {
	c.mu.Lock()
	c.frame++
	c.mu.Unlock()
}

// clear drops every picture the cache holds, as the storage page's way of
// emptying it. What is drawn next fetches what it needs again.
func (c *imageCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bitmaps = map[string]imageEntry{}
	c.total = 0
	c.failed = map[string]bool{}
}

// get returns the square picture of url at px pixels, and starts
// downloading it when it is not here yet: the next frame shows it.
func (c *imageCache) get(url string, px int, app *App) *ui.Bitmap {
	return c.getBox(url, px, px, app)
}

// getBox returns url as a picture fitted in a w×h box, and starts
// downloading it when it is not here yet.
func (c *imageCache) getBox(url string, w, h int, app *App) *ui.Bitmap {
	key := fmt.Sprintf("%s@%dx%d", url, w, h)

	c.mu.Lock()
	if e, ok := c.bitmaps[key]; ok {
		// Asking for it is drawing it: it is the last thing that may be
		// dropped now.
		e.used = c.frame
		c.bitmaps[key] = e
		c.mu.Unlock()
		return e.bmp
	}
	if c.loading[key] || c.failed[key] {
		c.mu.Unlock()
		return nil
	}
	c.loading[key] = true
	c.mu.Unlock()

	referer := app.imageReferer()
	go func() {
		c.fetchSem <- struct{}{}
		data, err := c.fetch(resized(url, w, h), url, referer)
		<-c.fetchSem
		if err != nil {
			c.mu.Lock()
			delete(c.loading, key)
			c.failed[key] = true
			c.mu.Unlock()
			return
		}
		// Bitmaps belong to the main thread, as drawing does.
		app.update(func() {
			bmp, err := ui.DecodeBitmap(data)
			c.mu.Lock()
			delete(c.loading, key)
			if err != nil {
				c.failed[key] = true
			} else {
				c.keep(key, bmp)
			}
			c.mu.Unlock()
		})
	}()
	return nil
}

// keep stores a picture, dropping the ones no frame is drawing while the
// cache holds more than it should.
func (c *imageCache) keep(key string, bmp *ui.Bitmap) {
	w, h := ui.ImageSize(bmp)
	e := imageEntry{bmp: bmp, size: int(w) * int(h) * 4, used: c.frame}
	if old, ok := c.bitmaps[key]; ok {
		c.total -= old.size
	}
	c.bitmaps[key] = e
	c.total += e.size
	c.trim()
}

// trim drops the pictures the frames before this one drew, the least
// recently drawn first, while the cache holds more than it should.
//
// A picture the frame being built asked for is never dropped, whatever the
// cache costs: dropping it would only have it fetched and decoded again for
// the next frame, and a frame that draws a whole panel of pictures would
// spend every frame fetching them all over again — which is what a panel of
// stickers did, once the pictures in it cost more than the cache may keep.
func (c *imageCache) trim() {
	for c.total > imageBudget || len(c.bitmaps) > imageKeep {
		key, oldest := "", c.frame
		for k, e := range c.bitmaps {
			if e.used >= c.frame {
				continue
			}
			if key == "" || e.used < oldest {
				key, oldest = k, e.used
			}
		}
		if key == "" {
			// Everything here is on screen: it is kept, and the budget
			// is what gives way.
			return
		}
		c.total -= c.bitmaps[key].size
		delete(c.bitmaps, key)
	}
}

// resized asks the image host for a picture at the size it is drawn.
// Qiniu's imageView2 fits it in a box, so a 38 DIP avatar comes as 76
// pixels rather than as the thousand the account uploaded.
func resized(rawURL string, w, h int) string {
	if rawURL == "" || w <= 0 || h <= 0 {
		return rawURL
	}
	sep := "?"
	if strings.Contains(rawURL, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%simageView2/1/w/%d/h/%d", rawURL, sep, w, h)
}

// fetch asks for the picture at the size drawn, and takes it whole when
// the host will not process it: not every host is Qiniu's, and a host
// that cannot resize answers with an error rather than the original.
func (c *imageCache) fetch(sized, raw, referer string) ([]byte, error) {
	if sized != raw {
		if data, err := c.download(sized, referer); err == nil {
			return data, nil
		}
	}
	return c.download(raw, referer)
}

// download fetches a picture. The image host refuses one asked for
// without a Referer it knows, answering 403, so the request carries the
// chat API's own host: it is on the host's list.
func (c *imageCache) download(url, referer string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errStatus(resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

type errStatus int

func (e errStatus) Error() string { return http.StatusText(int(e)) }

// DebugStats reports what the caches hold, for the memory log.
func (a *App) DebugStats() string {
	a.images.mu.Lock()
	n, total := len(a.images.bitmaps), a.images.total
	a.images.mu.Unlock()
	return fmt.Sprintf("convs=%d msgs=%d imgs=%d imgBytes=%dMB", len(a.convs), len(a.msgs), n, total>>20)
}
