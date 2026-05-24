package pipeline

import (
	"testing"
	"time"
)

func mcEntry(msg string) LogEntry {
	return LogEntry{Message: msg, Level: "info"}
}

func TestMulticast_NoSubscribers(t *testing.T) {
	mc := NewMulticast(MulticastConfig{})
	delivered, dropped := mc.Publish(mcEntry("hello"))
	if delivered != 0 || dropped != 0 {
		t.Fatalf("expected 0/0, got %d/%d", delivered, dropped)
	}
}

func TestMulticast_SingleSubscriberReceives(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 4})
	sub := mc.Subscribe()
	mc.Publish(mcEntry("ping"))
	select {
	case e := <-sub:
		if e.Message != "ping" {
			t.Fatalf("unexpected message: %s", e.Message)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("subscriber did not receive entry")
	}
}

func TestMulticast_MultipleSubscribersAllReceive(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 4})
	sub1 := mc.Subscribe()
	sub2 := mc.Subscribe()
	sub3 := mc.Subscribe()

	mc.Publish(mcEntry("broadcast"))

	for i, sub := range []<-chan LogEntry{sub1, sub2, sub3} {
		select {
		case e := <-sub:
			if e.Message != "broadcast" {
				t.Fatalf("sub%d got wrong message: %s", i+1, e.Message)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("sub%d did not receive entry", i+1)
		}
	}
}

func TestMulticast_FullSubscriberDropsEntry(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 1})
	sub := mc.Subscribe()

	// Fill the buffer
	mc.Publish(mcEntry("first"))
	// This should be dropped
	_, dropped := mc.Publish(mcEntry("second"))
	if dropped != 1 {
		t.Fatalf("expected 1 dropped, got %d", dropped)
	}
	_ = sub
}

func TestMulticast_UnsubscribeRemovesChannel(t *testing.T) {
	mc := NewMulticast(MulticastConfig{})
	sub := mc.Subscribe()
	if mc.Len() != 1 {
		t.Fatalf("expected 1 subscriber, got %d", mc.Len())
	}
	mc.Unsubscribe(sub)
	if mc.Len() != 0 {
		t.Fatalf("expected 0 subscribers after unsubscribe, got %d", mc.Len())
	}
}

func TestMulticast_CloseClosesAllSubscribers(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 4})
	sub1 := mc.Subscribe()
	sub2 := mc.Subscribe()
	mc.Close()

	for i, sub := range []<-chan LogEntry{sub1, sub2} {
		select {
		case _, ok := <-sub:
			if ok {
				t.Fatalf("sub%d channel should be closed", i+1)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("sub%d channel not closed", i+1)
		}
	}
	if mc.Len() != 0 {
		t.Fatalf("expected 0 subscribers after Close, got %d", mc.Len())
	}
}

func TestMulticast_DefaultBufferSize(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 0})
	if mc.bufferSize != 64 {
		t.Fatalf("expected default buffer size 64, got %d", mc.bufferSize)
	}
}
