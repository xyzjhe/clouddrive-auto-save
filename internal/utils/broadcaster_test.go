package utils

import (
	"strings"
	"testing"
	"time"
)

func TestBroadcaster(t *testing.T) {
	ch := GlobalBroadcaster.Subscribe()
	defer GlobalBroadcaster.Unsubscribe(ch)

	msg := "test message"
	GlobalBroadcaster.Broadcast(msg)

	select {
	case got := <-ch:
		if got != msg {
			t.Errorf("got %v, want %v", got, msg)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

func TestRingBuffer_PushAndSnapshot(t *testing.T) {
	rb := newRingBuffer(5)

	// 写入 3 条，应全部读取
	rb.Push("a")
	rb.Push("b")
	rb.Push("c")

	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 items, got %d", len(snap))
	}
	if snap[0] != "a" || snap[1] != "b" || snap[2] != "c" {
		t.Errorf("unexpected snapshot order: %v", snap)
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := newRingBuffer(3)

	// 写入 5 条，应只保留最后 3 条
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		rb.Push(s)
	}

	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 items after overflow, got %d", len(snap))
	}
	if snap[0] != "c" || snap[1] != "d" || snap[2] != "e" {
		t.Errorf("expected [c d e], got %v", snap)
	}
}

func TestRingBuffer_Reset(t *testing.T) {
	rb := newRingBuffer(5)
	rb.Push("a")
	rb.Push("b")

	rb.Reset()

	if rb.len != 0 {
		t.Errorf("expected len 0 after reset, got %d", rb.len)
	}
	snap := rb.Snapshot()
	if snap != nil {
		t.Errorf("expected nil snapshot after reset, got %v", snap)
	}
}

func TestRingBuffer_ExactCapacity(t *testing.T) {
	rb := newRingBuffer(3)
	rb.Push("x")
	rb.Push("y")
	rb.Push("z")

	// 恰好填满，不应丢失
	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3, got %d", len(snap))
	}
	if snap[0] != "x" || snap[1] != "y" || snap[2] != "z" {
		t.Errorf("expected [x y z], got %v", snap)
	}
}

func TestRingBuffer_WrapTwice(t *testing.T) {
	rb := newRingBuffer(3)
	// 写入 7 条，环形绕了两圈多
	for _, s := range []string{"1", "2", "3", "4", "5", "6", "7"} {
		rb.Push(s)
	}

	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3, got %d", len(snap))
	}
	if snap[0] != "5" || snap[1] != "6" || snap[2] != "7" {
		t.Errorf("expected [5 6 7], got %v", snap)
	}
}

func TestBroadcaster_HistoryFiltering(t *testing.T) {
	b := NewBroadcaster()
	go b.run()

	// 广播普通日志和事件
	b.Broadcast("normal log line")
	b.Broadcast("[EVENT:task_update|{\"id\":1}]")
	b.Broadcast("another log")

	// 等待消息处理
	time.Sleep(50 * time.Millisecond)

	recent := b.GetRecent()
	// 事件消息不应出现在历史中
	for _, r := range recent {
		if strings.HasPrefix(r, "[EVENT:") {
			t.Errorf("EVENT message should not appear in history: %s", r)
		}
	}
	if len(recent) != 2 {
		t.Errorf("expected 2 history items, got %d: %v", len(recent), recent)
	}
}
