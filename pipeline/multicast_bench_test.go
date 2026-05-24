package pipeline

import (
	"testing"
)

func BenchmarkMulticast_Publish_1Sub(b *testing.B) {
	mc := NewMulticast(MulticastConfig{BufferSize: 1024})
	sub := mc.Subscribe()
	entry := mcEntry("bench")

	// Drain in background
	go func() {
		for range sub {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Publish(entry)
	}
	b.StopTimer()
	mc.Close()
}

func BenchmarkMulticast_Publish_10Subs(b *testing.B) {
	mc := NewMulticast(MulticastConfig{BufferSize: 1024})
	for i := 0; i < 10; i++ {
		sub := mc.Subscribe()
		go func() {
			for range sub {
			}
		}()
	}
	entry := mcEntry("bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Publish(entry)
	}
	b.StopTimer()
	mc.Close()
}

func BenchmarkMulticast_SubscribeUnsubscribe(b *testing.B) {
	mc := NewMulticast(MulticastConfig{BufferSize: 4})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub := mc.Subscribe()
		mc.Unsubscribe(sub)
	}
}
