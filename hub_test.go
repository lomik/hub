package hub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	h := New()
	if h == nil {
		t.Fatal("Expected new Hub instance, got nil")
	}
	if h.Len() != 0 {
		t.Error("New hub should have 0 subscriptions")
	}
}

func TestHubSubscribeEvent(t *testing.T) {
	h := New()
	cb := func(ctx context.Context, e *Event) error { return nil }

	t.Run("basic subscription", func(t *testing.T) {
		id := h.SubscribeEvent(context.Background(), T("type=test"), cb)
		if id == 0 {
			t.Error("Expected non-zero subscription ID")
		}
		if h.Len() != 1 {
			t.Error("Expected 1 subscription")
		}
	})

	t.Run("multiple subscriptions", func(t *testing.T) {
		h.Clear(context.Background())
		h.SubscribeEvent(context.Background(), T("type=a"), cb)
		h.SubscribeEvent(context.Background(), T("type=b"), cb)
		if h.Len() != 2 {
			t.Error("Expected 2 subscriptions")
		}
	})
}

func TestHubSubscribePayload(t *testing.T) {
	h := New()
	cb := func(ctx context.Context, p any) error { return nil }

	id := h.SubscribePayload(context.Background(), T("type=test"), cb)
	if id == 0 {
		t.Error("Expected non-zero subscription ID")
	}
	if h.Len() != 1 {
		t.Error("Expected 1 subscription")
	}
}

func TestHubPublishEvent(t *testing.T) {
	h := New()
	var wg sync.WaitGroup

	t.Run("event delivery", func(t *testing.T) {
		wg.Add(1)
		h.SubscribeEvent(context.Background(), T("type=test"), func(ctx context.Context, e *Event) error {
			defer wg.Done()
			if e.Topic().String() != "type=test" {
				t.Error("Wrong topic in event")
			}
			return nil
		})

		h.PublishEvent(context.Background(), E("type=test"))
		wg.Wait()
	})

	t.Run("payload delivery", func(t *testing.T) {
		wg.Add(1)
		testPayload := "test payload"
		h.SubscribePayload(context.Background(), T("type=payload"), func(ctx context.Context, p any) error {
			defer wg.Done()
			if p != testPayload {
				t.Error("Wrong payload received")
			}
			return nil
		})

		h.PublishEvent(context.Background(), E("type=payload").WithPayload(testPayload))
		wg.Wait()
	})

	t.Run("topic matching", func(t *testing.T) {
		wg.Add(1)
		h.SubscribeEvent(context.Background(), T("type=*"), func(ctx context.Context, e *Event) error {
			defer wg.Done()
			return nil
		})

		h.PublishEvent(context.Background(), E("type=any"))
		wg.Wait()
	})

	t.Run("wait mode", func(t *testing.T) {
		start := time.Now()
		h.SubscribeEvent(context.Background(), T("wait=test"), func(ctx context.Context, e *Event) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		h.PublishEvent(context.Background(), E("wait=test").WithWait(true))
		if time.Since(start) < 100*time.Millisecond {
			t.Error("Wait mode didn't wait for handlers")
		}
	})
}

func TestHubUnsubscribe(t *testing.T) {
	h := New()
	cb := func(ctx context.Context, e *Event) error { return nil }

	t.Run("basic unsubscribe", func(t *testing.T) {
		id := h.SubscribeEvent(context.Background(), T("type=test"), cb)
		h.Unsubscribe(context.Background(), id)
		if h.Len() != 0 {
			t.Error("Expected 0 subscriptions after unsubscribe")
		}
	})

	t.Run("unsubscribe non-existent", func(t *testing.T) {
		h.Unsubscribe(context.Background(), 999) // should not panic
	})
}

func TestHubClear(t *testing.T) {
	h := New()
	cb := func(ctx context.Context, e *Event) error { return nil }

	h.SubscribeEvent(context.Background(), T("type=a"), cb)
	h.SubscribeEvent(context.Background(), T("type=b"), cb)
	h.Clear(context.Background())

	if h.Len() != 0 {
		t.Error("Expected 0 subscriptions after clear")
	}
}

func TestHubConcurrency(t *testing.T) {
	h := New()
	var wg sync.WaitGroup
	const workers = 100

	// Concurrent subscriptions
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			h.SubscribeEvent(context.Background(), T("type=concurrent"), func(ctx context.Context, e *Event) error {
				return nil
			})
		}()
	}

	// Concurrent publish
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			h.PublishEvent(context.Background(), E("type=concurrent"))
		}()
	}

	wg.Wait()
	if h.Len() != workers {
		t.Errorf("Expected %d subscriptions, got %d", workers, h.Len())
	}
}
