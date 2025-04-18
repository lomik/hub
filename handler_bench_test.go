package hub

import (
	"context"
	"errors"
	"testing"
)

func BenchmarkToHandler(b *testing.B) {
	ctx := context.Background()
	h := New()

	testEvent := &event{
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
			name: "Payload String",
			fn: func(ctx context.Context, payload string) error {
				return nil
			},
		},
		{
			name: "Event",
			fn: func(ctx context.Context, topic *Topic, p any) error {
				return nil
			},
		},
	}

	for _, cb := range callbacks {
		b.Run("Wrap_"+cb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = h.ToHandler(ctx, cb.fn)
			}
		})

		proxy, err := h.ToHandler(ctx, cb.fn)
		if err != nil {
			b.Fatalf("Wrap failed: %v", err)
		}

		b.Run("Exec_"+cb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = proxy(ctx, testEvent.topic, testEvent.payload)
			}
		})
	}
}

func BenchmarkWrappedCallbacks(b *testing.B) {
	ctx := context.Background()
	h := New()
	event := &event{
		topic:   T("type=test", "priority=high"),
		payload: "test payload",
	}

	// direct
	noopMinimal := func(ctx context.Context) error { return nil }
	noopPayload := func(ctx context.Context, payload any) error { return nil }
	noopPayloadString := func(ctx context.Context, payload string) error { return nil }
	noopEvent := func(ctx context.Context, topic *Topic, p any) error { return nil }

	// wrapped
	minimalProxy, _ := h.ToHandler(ctx, noopMinimal)
	payloadProxy, _ := h.ToHandler(ctx, noopPayload)
	payloadStringProxy, _ := h.ToHandler(ctx, noopPayloadString)
	eventProxy, _ := h.ToHandler(ctx, noopEvent)

	// basic
	b.Run("Direct/Minimal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopMinimal(ctx)
		}
	})

	b.Run("Wrapped/Minimal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = minimalProxy(ctx, event.topic, event.payload)
		}
	})

	b.Run("Direct/Event", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopEvent(ctx, event.topic, event.payload)
		}
	})

	b.Run("Wrapped/Event", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = eventProxy(ctx, event.topic, event.payload)
		}
	})

	b.Run("Wrapped/Payload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = payloadProxy(ctx, event.topic, event.payload)
		}
	})

	b.Run("Direct/PayloadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopPayloadString(ctx, "")
		}
	})

	b.Run("Wrapped/PayloadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = payloadStringProxy(ctx, event.topic, event.payload)
		}
	})

	// with non-noop
	processPayload := func(ctx context.Context, payload any) error {
		s, ok := payload.(string)
		if !ok {
			return errors.New("invalid payload")
		}
		if len(s) < 10 {
			return nil
		}
		return errors.New("payload too long")
	}

	processPayloadProxy, _ := h.ToHandler(ctx, processPayload)

	b.Run("Wrapped/RealPayload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = processPayloadProxy(ctx, event.topic, event.payload)
		}
	})
}
