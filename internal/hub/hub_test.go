package hub

import (
	"strings"
	"testing"
)

func TestBroadcastAndLimit(t *testing.T) {
	h := New(2)
	a, b := h.Subscribe(), h.Subscribe()
	if a == nil || b == nil || h.Subscribe() != nil {
		t.Fatal("limit not enforced")
	}
	h.Broadcast(Frame{Seq: 1, Data: []byte("x")})
	if f := <-a.C; f.Seq != 1 {
		t.Errorf("%+v", f)
	}
	if f := <-b.C; f.Seq != 1 {
		t.Errorf("%+v", f)
	}
	h.Unsubscribe(a)
	h.Unsubscribe(a) // idempotent
	if h.Count() != 1 || h.Subscribe() == nil {
		t.Errorf("count %d", h.Count())
	}
}

func TestSlowClientDroppedOthersUnaffected(t *testing.T) {
	h := New(0)
	slow, fast := h.Subscribe(), h.Subscribe()
	for i := 1; i <= bufSize+5; i++ {
		h.Broadcast(Frame{Seq: int64(i)})
		<-fast.C // fast client keeps up
	}
	if h.Count() != 1 || h.Dropped() != 1 {
		t.Fatalf("count=%d dropped=%d", h.Count(), h.Dropped())
	}
	n := 0
	for range slow.C { // closed after drop; drains its buffer
		n++
	}
	if n != bufSize {
		t.Errorf("slow got %d buffered frames", n)
	}
}

func TestFormatEvent(t *testing.T) {
	got := string(FormatEvent(7, "caption", []byte(`{"a":1}`)))
	if got != "id: 7\nevent: caption\ndata: {\"a\":1}\n\n" {
		t.Errorf("%q", got)
	}
	if strings.Contains(string(FormatEvent(-1, "preload", []byte("{}"))), "id:") {
		t.Error("id should be omitted")
	}
}
