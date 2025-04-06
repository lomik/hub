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

func BenchmarkWrappedCallbacks(b *testing.B) {
	ctx := context.Background()
	event := &Event{
		topic:   T("type=test", "priority=high"),
		payload: "test payload",
	}

	// Тестовые callback-функции разных типов
	noopMinimal := func(ctx context.Context) error { return nil }
	noopPayload := func(ctx context.Context, payload any) error { return nil }
	noopTopicPayload := func(ctx context.Context, topic *Topic, payload any) error { return nil }
	noopEvent := func(ctx context.Context, e *Event) error { return nil }

	// Обернутые версии
	minimalProxy, _ := WrapSubscribeCallback(ctx, noopMinimal)
	payloadProxy, _ := WrapSubscribeCallback(ctx, noopPayload)
	topicPayloadProxy, _ := WrapSubscribeCallback(ctx, noopTopicPayload)
	eventProxy, _ := WrapSubscribeCallback(ctx, noopEvent)

	b.Run("Minimal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = minimalProxy(ctx, event)
		}
	})

	b.Run("Payload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = payloadProxy(ctx, event)
		}
	})

	b.Run("Topic+Payload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = topicPayloadProxy(ctx, event)
		}
	})

	b.Run("Event", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = eventProxy(ctx, event)
		}
	})

	// Бенчмарк с реальной работой в callback
	b.Run("PayloadWithWork", func(b *testing.B) {
		cb := func(ctx context.Context, payload any) error {
			// Имитация полезной нагрузки
			s := payload.(string)
			if len(s) > 0 {
				return nil
			}
			return errors.New("empty payload")
		}
		proxy, _ := WrapSubscribeCallback(ctx, cb)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = proxy(ctx, event)
		}
	})
}
