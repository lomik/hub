package hub

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/lomik/hub/pkg/cmap"
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

	t.Run("Subscribe", func(t *testing.T) {
		id, _ := h.Subscribe(ctx, T("type=test"), func(ctx context.Context, topic *Topic, p any) error {
			return nil
		})
		if id == 0 {
			t.Error("Expected non-zero subscription ID")
		}
		if h.Len() != 1 {
			t.Error("Expected 1 subscription after Subscribe")
		}
	})
}

func TestHubPublish(t *testing.T) {
	h := New()
	ctx := context.Background()

	t.Run("simple publish", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		h.Subscribe(ctx, T("type=test"), func(ctx context.Context, topic *Topic, p any) error {
			defer wg.Done()
			if topic.String() != "type=test" {
				t.Error("Unexpected topic in event")
			}
			return nil
		})

		h.Publish(ctx, T("type=test"), nil)
		wg.Wait()
	})

	t.Run("wildcard matching", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		h.Subscribe(ctx, T("type=*"), func(ctx context.Context, topic *Topic, p any) error {
			defer wg.Done()
			return nil
		})

		h.Publish(ctx, T("type=wildcard"), nil)
		wg.Wait()
	})

	t.Run("wait mode", func(t *testing.T) {
		start := time.Now()
		h.Subscribe(ctx, T("wait=test"), func(ctx context.Context, topic *Topic, p any) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		h.Publish(ctx, T("wait=test"), nil, Wait(true))
		if time.Since(start) < 100*time.Millisecond {
			t.Error("Wait mode didn't wait for handlers")
		}
	})
}

func TestHubUnsubscribe(t *testing.T) {
	h := New()
	ctx := context.Background()

	t.Run("unsubscribe existing", func(t *testing.T) {
		id, _ := h.Subscribe(ctx, T("type=test"), nil)
		h.Unsubscribe(ctx, id)
		if h.Len() != 0 {
			t.Error("Expected 0 subscriptions after unsubscribe")
		}
	})

	t.Run("unsubscribe non-existent", func(t *testing.T) {
		h.Unsubscribe(ctx, 999) // Should not panic
	})

	t.Run("unsubscribe from indexes", func(t *testing.T) {
		id, _ := h.Subscribe(ctx, T("type=alert"), nil)
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

	h.Subscribe(ctx, T("type=a"), nil)
	h.Subscribe(ctx, T("type=b"), nil)
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
			h.Subscribe(ctx, T("concurrent=test"), nil)
		}()
	}

	// Test concurrent publish
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Publish(ctx, T("concurrent=test"), nil)
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
	h.Subscribe(ctx, T("type=alert"), Handler(nil))
	h.RLock()
	if h.indexKeyValue["type"]["alert"].len() != 1 {
		t.Error("Subscription not added to key-value index")
	}
	h.RUnlock()

	// Test wildcard index
	h.Subscribe(ctx, T("type=*"), Handler(nil))
	h.RLock()
	if h.indexKey["type"].len() != 2 {
		t.Error("Subscription not added to key index")
	}
	h.RUnlock()

	// Test empty topic
	h.Subscribe(ctx, T(""), Handler(nil))
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

	h.Subscribe(ctx, T("type=test"), Handler(nil))
	if h.Len() != 1 {
		t.Error("Expected length 1 after subscribe")
	}

	h.Unsubscribe(ctx, 1)
	if h.Len() != 0 {
		t.Error("Expected length 0 after unsubscribe")
	}
}

func makeTestHub(subTimeout time.Duration) (*Hub, *cmap.CMap) {
	ctx := context.Background()
	h := New()
	c := cmap.New()

	h.Subscribe(ctx, T("a=10", "b=20"), func(ctx context.Context) {
		if subTimeout > 0 {
			time.Sleep(subTimeout)
		}
		c.Add("a=10, b=20", 1)
	})

	h.Subscribe(ctx, T("a=10", "b=21"), func(ctx context.Context) {
		if subTimeout > 0 {
			time.Sleep(subTimeout)
		}
		c.Add("a=10, b=21", 1)
	})

	h.Subscribe(ctx, T("a=11", "b=20"), func(ctx context.Context) {
		if subTimeout > 0 {
			time.Sleep(subTimeout)
		}
		c.Add("a=11, b=20", 1)
	})

	h.Subscribe(ctx, T("a=11", "b=21"), func(ctx context.Context) {
		if subTimeout > 0 {
			time.Sleep(subTimeout)
		}
		c.Add("a=11, b=21", 1)
	})

	h.Subscribe(ctx, T("a=*", "b=21"), func(ctx context.Context) {
		if subTimeout > 0 {
			time.Sleep(subTimeout)
		}
		c.Add("a=*, b=21", 1)
	})

	return h, c
}

func TestHubPubSubMatch(t *testing.T) {
	ctx := context.Background()
	h, c := makeTestHub(0)

	checkC := func(t *testing.T, mp map[string]int) {
		if !c.Eq(mp) {
			t.Error("Result mismatch")
		}
	}

	t.Run("a=10, b=20, sync", func(t *testing.T) {
		c.Clear()
		h.Publish(ctx, T("a=10", "b=20"), nil, Sync(true))
		checkC(t, map[string]int{"a=10, b=20": 1})
	})

	t.Run("a=10, b=21, sync", func(t *testing.T) {
		c.Clear()
		h.Publish(ctx, T("a=10", "b=21"), nil, Sync(true))
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1})
	})

	t.Run("a=10, b=*, sync", func(t *testing.T) {
		c.Clear()
		h.Publish(ctx, T("a=10", "b=*"), nil, Sync(true))
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

}

func TestHubPubSubSyncAsync(t *testing.T) {
	ctx := context.Background()
	h, c := makeTestHub(time.Second)

	checkT := func(t *testing.T, s time.Time, v int) {
		if int(math.Round(time.Since(s).Seconds())) != v {
			t.Error("Result duration ", time.Since(s))
		}
	}

	checkC := func(t *testing.T, mp map[string]int) {
		if !c.Eq(mp) {
			t.Error("Result mismatch")
		}
	}

	t.Run("sync", func(t *testing.T) {
		c.Clear()
		s := time.Now()
		h.Publish(ctx, T("a=10", "b=*"), nil, Sync(true))
		checkT(t, s, 3)
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

	t.Run("async", func(t *testing.T) {
		c.Clear()
		s := time.Now()
		h.Publish(ctx, T("a=10", "b=*"), nil)
		checkT(t, s, 0)
		time.Sleep(2 * time.Second)
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

	t.Run("wait", func(t *testing.T) {
		c.Clear()
		s := time.Now()
		h.Publish(ctx, T("a=10", "b=*"), nil, Wait(true))
		checkT(t, s, 1)
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

}

func TestHubPubSubOnFinish(t *testing.T) {
	ctx := context.Background()
	h, c := makeTestHub(0)

	checkC := func(t *testing.T, mp map[string]int) {
		if !c.Eq(mp) {
			t.Error("Result mismatch")
		}
	}

	t.Run("sync", func(t *testing.T) {
		c.Clear()
		ch := make(chan struct{})
		h.Publish(ctx, T("a=10", "b=*"), nil, Sync(true), OnFinish(func(ctx context.Context) { close(ch) }))
		<-ch
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

	t.Run("async", func(t *testing.T) {
		c.Clear()
		ch := make(chan struct{})
		h.Publish(ctx, T("a=10", "b=*"), nil, OnFinish(func(ctx context.Context) { close(ch) }))
		<-ch
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

	t.Run("wait", func(t *testing.T) {
		c.Clear()
		ch := make(chan struct{})
		h.Publish(ctx, T("a=10", "b=*"), nil, Wait(true), OnFinish(func(ctx context.Context) { close(ch) }))
		<-ch
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1})
	})

	t.Run("async, not matched", func(t *testing.T) {
		c.Clear()
		ch := make(chan struct{})
		h.Publish(ctx, T("a=31", "b=32"), nil, OnFinish(func(ctx context.Context) { close(ch) }), nil)
		<-ch
		checkC(t, map[string]int{})
	})
}

func TestHubPubSubOnce(t *testing.T) {
	ctx := context.Background()
	h, c := makeTestHub(0)

	checkC := func(t *testing.T, mp map[string]int) {
		if !c.Eq(mp) {
			t.Error("Result mismatch")
		}
	}
	t.Run("sync", func(t *testing.T) {
		c.Clear()
		h.Subscribe(ctx, T(), func(ctx context.Context) {
			c.Add("once", 1)
		}, Once(true), nil)
		h.Publish(ctx, T("a=10", "b=*"), nil, Sync(true))
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1, "once": 1})
		h.Publish(ctx, T("a=10", "b=*"), nil, Sync(true))
		checkC(t, map[string]int{"a=10, b=21": 2, "a=*, b=21": 2, "a=10, b=20": 2, "once": 1})
	})

	t.Run("async", func(t *testing.T) {
		c.Clear()
		h.Subscribe(ctx, T(), func(ctx context.Context) {
			c.Add("once", 1)
		}, Once(true), nil)
		h.Publish(ctx, T("a=10", "b=*"), nil)
		time.Sleep(100 * time.Millisecond)
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1, "once": 1})
		h.Publish(ctx, T("a=10", "b=*"), nil)
		time.Sleep(100 * time.Millisecond)
		checkC(t, map[string]int{"a=10, b=21": 2, "a=*, b=21": 2, "a=10, b=20": 2, "once": 1})
	})

	t.Run("async, with on finish", func(t *testing.T) {
		c.Clear()
		h.Subscribe(ctx, T(), func(ctx context.Context) {
			c.Add("once", 1)
		}, Once(true), nil)
		h.Publish(ctx, T("a=10", "b=*"), nil, OnFinish(func(ctx context.Context) {}))
		time.Sleep(100 * time.Millisecond)
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1, "once": 1})
		h.Publish(ctx, T("a=10", "b=*"), nil, OnFinish(func(ctx context.Context) {}))
		time.Sleep(100 * time.Millisecond)
		checkC(t, map[string]int{"a=10, b=21": 2, "a=*, b=21": 2, "a=10, b=20": 2, "once": 1})
	})

	t.Run("wait", func(t *testing.T) {
		c.Clear()
		h.Subscribe(ctx, T(), func(ctx context.Context) {
			c.Add("once", 1)
		}, Once(true), nil)
		h.Publish(ctx, T("a=10", "b=*"), nil, Wait(true))
		checkC(t, map[string]int{"a=10, b=21": 1, "a=*, b=21": 1, "a=10, b=20": 1, "once": 1})
		h.Publish(ctx, T("a=10", "b=*"), nil, Wait(true))
		checkC(t, map[string]int{"a=10, b=21": 2, "a=*, b=21": 2, "a=10, b=20": 2, "once": 1})
	})

}
