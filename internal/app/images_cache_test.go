package app

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/egoist/mygo/ui"
)

// solidPNG is a picture of one color, as a bitmap's pixels are what the
// cache counts.
func solidPNG(t *testing.T, w, h int) *ui.Bitmap {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	bmp, err := ui.DecodeBitmap(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return bmp
}

// TestImageBudget checks that the cache holds pictures to a budget of
// pixels, not to a count: a chat of photos must not keep hundreds of
// megabytes because it stayed under a number of them.
func TestImageBudget(t *testing.T) {
	c := newImageCache()
	bmp := solidPNG(t, 400, 400) // 640,000 bytes
	each := 400 * 400 * 4

	for i := 0; i < imageKeep; i++ {
		c.keep(fmt.Sprintf("p%d", i), bmp)
		if c.total > imageBudget+each {
			t.Fatalf("after %d pictures the cache holds %d bytes, over the budget of %d",
				i+1, c.total, imageBudget)
		}
	}
	if len(c.bitmaps) >= imageKeep {
		t.Errorf("kept %d pictures of %d bytes each: the budget did not bite", len(c.bitmaps), each)
	}
	if len(c.order) != len(c.bitmaps) {
		t.Errorf("order has %d keys, bitmaps has %d", len(c.order), len(c.bitmaps))
	}
	// The newest is still here, and the oldest is gone.
	if _, ok := c.bitmaps[fmt.Sprintf("p%d", imageKeep-1)]; !ok {
		t.Error("the picture stored last was dropped")
	}
	if _, ok := c.bitmaps["p0"]; ok {
		t.Error("the oldest picture was kept")
	}

	// Storing the same key again replaces it rather than counting it twice.
	before := c.total
	c.keep(fmt.Sprintf("p%d", imageKeep-1), bmp)
	if c.total != before {
		t.Errorf("storing a picture again changed the total from %d to %d", before, c.total)
	}
}

// TestImageReuse checks that a picture is kept once per size asked for: an
// avatar drawn at 32 and at 38 are two pictures, and the same size twice
// is one.
func TestImageReuse(t *testing.T) {
	c := newImageCache()
	bmp := solidPNG(t, 32, 32)
	c.keep("u@32x32", bmp)
	c.keep("u@32x32", bmp)
	if len(c.bitmaps) != 1 || len(c.order) != 1 {
		t.Errorf("kept %d pictures, %d keys, want 1 and 1", len(c.bitmaps), len(c.order))
	}
}
