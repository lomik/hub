package hub

import (
	"context"
	"testing"
)

func BenchmarkWrapSubscribeCallback(b *testing.B) {
	ctx := context.Background()
	testEvent := &Event{
		topic:   T("type=benchmark"),
		payload: "benchmark payload",
	}

	callbacks := []struct {
		name string
		fn   interface{}
	}{
		{
			name: "Minimal",
			fn: func(ctx context.Context) error {
				return nil
			},
		},
		{
			name: "Payload",
			fn: func(ctx context.Context, payload any) error {
				return nil
			},
		},
		{
			name: "TopicPayload",
			fn: func(ctx context.Context, topic *Topic, payload any) error {
				return nil
			},
		},
		{
			name: "Event",
			fn: func(ctx context.Context, e *Event) error {
				return nil
			},
		},
	}

	for _, cb := range callbacks {
		b.Run("Wrap_"+cb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = WrapSubscribeCallback(ctx, cb.fn)
			}
		})

		proxy, err := WrapSubscribeCallback(ctx, cb.fn)
		if err != nil {
			b.Fatalf("Wrap failed: %v", err)
		}

		b.Run("Exec_"+cb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = proxy(ctx, testEvent)
			}
		})
	}
}
