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
	imageKeep   = 400
	imageBudget = 48 << 20 // bytes of pixels kept
)

// imageCache keeps the pictures of people, groups and messages:
// downloaded once at the size they are drawn, decoded once, and drawn
// from then on.
type imageCache struct {
	mu      sync.Mutex
	bitmaps map[string]imageEntry
	order   []string // the keys, oldest first
	total   int      // bytes of pixels kept
	loading map[string]bool
	failed  map[string]bool
	hc      *http.Client
}

// imageEntry is one picture and what its pixels cost.
type imageEntry struct {
	bmp  *ui.Bitmap
	size int
}

func newImageCache() *imageCache {
	return &imageCache{
		bitmaps: map[string]imageEntry{},
		loading: map[string]bool{},
		failed:  map[string]bool{},
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

// clear drops every picture the cache holds, as the storage page's way of
// emptying it. What is drawn next fetches what it needs again.
func (c *imageCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bitmaps = map[string]imageEntry{}
	c.order = nil
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
		data, err := c.fetch(resized(url, w, h), url, referer)
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

// keep stores a picture, dropping the oldest while the cache holds more
// than it should.
func (c *imageCache) keep(key string, bmp *ui.Bitmap) {
	w, h := ui.ImageSize(bmp)
	e := imageEntry{bmp: bmp, size: int(w) * int(h) * 4}
	if old, ok := c.bitmaps[key]; ok {
		c.total -= old.size
	} else {
		c.order = append(c.order, key)
	}
	c.bitmaps[key] = e
	c.total += e.size

	for len(c.order) > 1 && (c.total > imageBudget || len(c.order) > imageKeep) {
		old := c.order[0]
		if old == key { // the one just stored goes last, not away
			c.order = append(c.order[1:], old)
			continue
		}
		c.order = c.order[1:]
		if gone, ok := c.bitmaps[old]; ok {
			c.total -= gone.size
			delete(c.bitmaps, old)
		}
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
