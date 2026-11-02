package app

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/egoist/mygo/ui"

	"mygohu/internal/yh"
)

// TestMemoryBudget holds the window to the memory the project allows it
// under a heavy screen: a thousand messages and a picture cache filled to
// its budget must stay well inside the limit for medium pressure.
//
// It measures the heap, which is where the pictures are kept once decoded
// and what the messages cost. The window's own surfaces are outside it, in
// the driver, and are what the extreme figure is for.
func TestMemoryBudget(t *testing.T) {
	a := openChat(0)
	for i := 0; i < 1000; i++ {
		a.msgs = append(a.msgs, &yh.Message{
			MsgID:       fmt.Sprintf("m%04d", i),
			ChatID:      "g1",
			Sender:      yh.Sender{ChatID: "200", Name: "小明"},
			ContentType: yh.ContentText,
			Content:     yh.Content{Text: fmt.Sprintf("第 %d 条消息，写得长一点，看看换行和布局在长会话里稳不稳。", i)},
			TimestampMS: uint64(1700000000000 + i*1000),
		})
	}

	// Fill the picture cache to its budget, as a scroll through a chat of
	// photos does: each one a picture of its own, not one kept under many
	// names.
	for i := 0; i < 60; i++ {
		a.images.keep(fmt.Sprintf("https://chat-img.jwznb.com/%d.png@480x480", i),
			solidPNG(t, 480, 480))
	}

	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()

	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heap := m.HeapAlloc >> 20
	t.Logf("heap %d MB with %d messages and %d MB of pictures",
		heap, len(a.msgs), a.images.total>>20)

	if heap > 200 {
		t.Errorf("heap is %d MB, over the 200 MB the project allows under medium pressure", heap)
	}
}
