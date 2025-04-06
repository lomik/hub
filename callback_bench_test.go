package hub

import (
	"context"
	"errors"
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
			name: "Payload String",
			fn: func(ctx context.Context, payload string) error {
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

func BenchmarkWrappedCallbacks(b *testing.B) {
	ctx := context.Background()
	event := &Event{
		topic:   T("type=test", "priority=high"),
		payload: "test payload",
	}

	// direct
	noopMinimal := func(ctx context.Context) error { return nil }
	noopPayload := func(ctx context.Context, payload any) error { return nil }
	noopPayloadString := func(ctx context.Context, payload string) error { return nil }
	noopEvent := func(ctx context.Context, e *Event) error { return nil }

	// wrapped
	minimalProxy, _ := WrapSubscribeCallback(ctx, noopMinimal)
	payloadProxy, _ := WrapSubscribeCallback(ctx, noopPayload)
	payloadStringProxy, _ := WrapSubscribeCallback(ctx, noopPayloadString)
	eventProxy, _ := WrapSubscribeCallback(ctx, noopEvent)

	// basic
	b.Run("Direct/Minimal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopMinimal(ctx)
		}
	})

	b.Run("Wrapped/Minimal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = minimalProxy(ctx, event)
		}
	})

	b.Run("Direct/Event", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopEvent(ctx, event)
		}
	})

	b.Run("Wrapped/Event", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = eventProxy(ctx, event)
		}
	})

	b.Run("Wrapped/Payload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = payloadProxy(ctx, event)
		}
	})

	b.Run("Direct/PayloadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = noopPayloadString(ctx, "")
		}
	})

	b.Run("Wrapped/PayloadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = payloadStringProxy(ctx, event)
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

	processPayloadProxy, _ := WrapSubscribeCallback(ctx, processPayload)

	b.Run("Wrapped/RealPayload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = processPayloadProxy(ctx, event)
		}
	})
}
