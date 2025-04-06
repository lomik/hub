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

func TestHubSubscribe(t *testing.T) {
	h := New()
	ctx := context.Background()

	t.Run("SubscribeEvent", func(t *testing.T) {
		id := h.SubscribeEvent(ctx, T("type=test"), func(ctx context.Context, e *Event) error {
			return nil
		})
		if id == 0 {
			t.Error("Expected non-zero subscription ID")
		}
		if h.Len() != 1 {
			t.Error("Expected 1 subscription after SubscribeEvent")
		}
	})
}

func TestHubPublish(t *testing.T) {
	h := New()
	ctx := context.Background()

	t.Run("simple publish", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		h.SubscribeEvent(ctx, T("type=test"), func(ctx context.Context, e *Event) error {
			defer wg.Done()
			if e.Topic().String() != "type=test" {
				t.Error("Unexpected topic in event")
			}
			return nil
		})

		h.PublishEvent(ctx, E("type=test"))
		wg.Wait()
	})

	t.Run("wildcard matching", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		h.SubscribeEvent(ctx, T("type=*"), func(ctx context.Context, e *Event) error {
			defer wg.Done()
			return nil
		})

		h.PublishEvent(ctx, E("type=wildcard"))
		wg.Wait()
	})

	t.Run("wait mode", func(t *testing.T) {
		start := time.Now()
		h.SubscribeEvent(ctx, T("wait=test"), func(ctx context.Context, e *Event) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		h.PublishEvent(ctx, E("wait=test").WithWait(true))
		if time.Since(start) < 100*time.Millisecond {
			t.Error("Wait mode didn't wait for handlers")
		}
	})
}

func TestHubUnsubscribe(t *testing.T) {
	h := New()
	ctx := context.Background()

	t.Run("unsubscribe existing", func(t *testing.T) {
		id := h.SubscribeEvent(ctx, T("type=test"), nil)
		h.Unsubscribe(ctx, id)
		if h.Len() != 0 {
			t.Error("Expected 0 subscriptions after unsubscribe")
		}
	})

	t.Run("unsubscribe non-existent", func(t *testing.T) {
		h.Unsubscribe(ctx, 999) // Should not panic
	})

	t.Run("unsubscribe from indexes", func(t *testing.T) {
		id := h.SubscribeEvent(ctx, T("type=alert"), nil)
		h.Unsubscribe(ctx, id)

		h.RLock()
		defer h.RUnlock()
		if h.indexKey["type"].len() != 0 {
			t.Error("Subscription not removed from key index")
		}
		if h.indexKeyValue["type"]["alert"].len() != 0 {
			t.Error("Subscription not removed from key-value index")
		}
	})
}

func TestHubClear(t *testing.T) {
	h := New()
	ctx := context.Background()

	h.SubscribeEvent(ctx, T("type=a"), nil)
	h.SubscribeEvent(ctx, T("type=b"), nil)
	h.Clear(ctx)

	if h.Len() != 0 {
		t.Error("Expected 0 subscriptions after clear")
	}

	h.RLock()
	defer h.RUnlock()
	if len(h.indexKey) != 0 {
		t.Error("Expected empty key index after clear")
	}
	if len(h.indexKeyValue) != 0 {
		t.Error("Expected empty key-value index after clear")
	}
}

func TestHubConcurrency(t *testing.T) {
	h := New()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Test concurrent subscriptions
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.SubscribeEvent(ctx, T("concurrent=test"), nil)
		}()
	}

	// Test concurrent publish
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.PublishEvent(ctx, E("concurrent=test"))
		}()
	}

	// Test concurrent unsubscribe
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id SubID) {
			defer wg.Done()
			h.Unsubscribe(ctx, id)
		}(SubID(i + 1))
	}

	wg.Wait()
}

func TestHubIndexes(t *testing.T) {
	h := New()
	ctx := context.Background()

	// Test key-value index
	h.SubscribeEvent(ctx, T("type=alert"), nil)
	h.RLock()
	if h.indexKeyValue["type"]["alert"].len() != 1 {
		t.Error("Subscription not added to key-value index")
	}
	h.RUnlock()

	// Test wildcard index
	h.SubscribeEvent(ctx, T("type=*"), nil)
	h.RLock()
	if h.indexKey["type"].len() != 2 {
		t.Error("Subscription not added to key index")
	}
	h.RUnlock()

	// Test empty topic
	h.SubscribeEvent(ctx, T(""), nil)
	h.RLock()
	if h.indexEmpty.len() != 1 {
		t.Error("Subscription not added to empty index")
	}
	h.RUnlock()
}

func TestHubLen(t *testing.T) {
	h := New()
	ctx := context.Background()

	if h.Len() != 0 {
		t.Error("New hub should have length 0")
	}

	h.SubscribeEvent(ctx, T("type=test"), nil)
	if h.Len() != 1 {
		t.Error("Expected length 1 after subscribe")
	}

	h.Unsubscribe(ctx, 1)
	if h.Len() != 0 {
		t.Error("Expected length 0 after unsubscribe")
	}
}
