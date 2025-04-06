package hub

import (
	"context"
	"sync"
	"sync/atomic"
)

// Handler defines a function signature for processing events in the hub.
// It receives a context for cancellation/timeout and the event to process.
// Return an error to indicate processing failure.
//
// Usage:
//
//	var myHandler Handler = func(ctx context.Context, e *Event) error {
//	    log.Printf("Processing event: %s", e.Topic())
//	    return nil // Return nil on success
//	}
//
//	h.SubscribeEvent(ctx, topic, myHandler)
type Handler func(ctx context.Context, e *Event) error

// Hub implements a pub/sub system with optimized subscription matching
// using multi-level indexes for efficient event distribution
type Hub struct {
	sync.RWMutex
	seq atomic.Uint64 // Atomic counter for generating subscription IDs

	all *sublist
	// Index structures:
	indexKeyValue map[string]map[string]*sublist // Exact key-value pair index
	indexKey      map[string]*sublist            // Wildcard value index (key=*)
	indexEmpty    *sublist                       // Subscriptions without topic attributes

	// customize
	onCallback [](func(cb any) (Handler, error))
}

// New creates and initializes a new Hub instance
func New() *Hub {
	return &Hub{
		all:           &sublist{},
		indexKeyValue: make(map[string]map[string]*sublist),
		indexKey:      make(map[string]*sublist),
		indexEmpty:    &sublist{},
	}
}

// SubscribeEvent registers a new event subscriber with topic matching
func (h *Hub) SubscribeEvent(ctx context.Context, t *Topic, cb Handler, opts ...SubscribeOption) SubID {
	h.Lock()
	defer h.Unlock()

	id := SubID(h.seq.Add(1))
	s := &sub{
		id:            id,
		topic:         t,
		callbackEvent: cb,
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.modifySub(ctx, s)
	}

	h.add(ctx, s)
	return id
}

// Subscribe registers an event handler with flexible callback signature options.
// It provides a more convenient interface than SubscribeEvent by automatically
// wrapping different callback signatures while maintaining type safety.
//
// Supported callback formats:
//  1. Minimal:      func(ctx context.Context) error
//  2. Event style:  func(ctx context.Context, e *Event) error
//  3. Typed payload: func(ctx context.Context, payload Type) error
//  4. Generic payload: func(ctx context.Context, payload any) error
//
// Supported payload types (Type):
//   - All integer types (int8-int64, uint8-uint64)
//   - Floating point (float32, float64)
//   - String and boolean
//   - Time types (time.Time, time.Duration)
//   - Common collections ([]string, map[string]any)
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - t: Topic to subscribe to (use hub.T() for all events)
//   - cb: Callback function in one of supported formats
//   - opts: Optional subscription settings (e.g., Once for single delivery)
//
// Returns:
//   - Subscription ID that can be used for unsubscribing
//   - Error if:
//   - Callback signature is invalid
//   - Topic is nil
//   - Unsupported parameter type in callback
//
// Behavior:
//   - For typed callbacks, attempts direct type assertion first
//   - Falls back to automatic conversion using spf13/cast
//   - Returns conversion errors during event delivery
//   - Supports all standard SubscribeOption configurations
//
// Example usage:
//
//	// Minimal callback
//	id, err := hub.Subscribe(ctx, topic, func(ctx context.Context) error {
//	    return nil
//	})
//
//	// Typed payload
//	id, err := hub.Subscribe(ctx, topic, func(ctx context.Context, id int) error {
//	    log.Printf("Processing ID: %d", id)
//	    return nil
//	})
//
//	// With options
//	id, err := hub.Subscribe(ctx, topic,
//	    func(ctx context.Context, msg string) error {
//	        return nil
//	    },
//	    hub.Once(true), // Auto-unsubscribe after first event
//	)
//
// Notes:
// - Prefer specific typed callbacks when possible for better performance
// - The generic 'any' signature provides flexibility at small performance cost
// - All type validation occurs during subscription, not event delivery
func (h *Hub) Subscribe(ctx context.Context, t *Topic, cb interface{}, opts ...SubscribeOption) (SubID, error) {
	eventCb, err := h.ToHandler(ctx, cb)
	if err != nil {
		return 0, err
	}

	return h.SubscribeEvent(ctx, t, eventCb, opts...), nil
}

// add adds a subscription to all relevant indexes
func (h *Hub) add(_ context.Context, s *sub) {
	h.all.add(s)

	// Process each key-value pair in the topic
	s.topic.Each(func(k, v string) {
		// Initialize nested maps if needed
		if _, exists := h.indexKeyValue[k]; !exists {
			h.indexKeyValue[k] = make(map[string]*sublist)
		}
		if _, exists := h.indexKeyValue[k][v]; !exists {
			h.indexKeyValue[k][v] = &sublist{}
		}
		h.indexKeyValue[k][v].add(s)

		// Add to wildcard index for this key
		if _, exists := h.indexKey[k]; !exists {
			h.indexKey[k] = &sublist{}
		}
		h.indexKey[k].add(s)
	})

	// Handle empty topics
	if s.topic.Len() == 0 {
		h.indexEmpty.add(s)
	}
}

// Publish sends an event to all subscribers of the specified topic with the given payload.
// It provides a simplified interface compared to PublishEvent by automatically creating
// the Event structure for common use cases.
//
// Parameters:
//   - ctx:       Context for cancellation and timeouts
//   - topic:     Destination topic for the event (required)
//   - payload:   Event data (can be any type)
//   - opts:      Optional publishing settings:
//   - hub.Wait(true) - wait for all handlers to complete
//   - hub.Sync(true) - process handlers synchronously
//   - hub.OnFinish() - add completion callback
//
// Behavior:
//   - Creates a new Event with the provided topic and payload
//   - Applies all specified PublishOptions
//   - Delivers to all matching subscribers
//   - Handles payload conversion automatically when subscribers use typed callbacks
//
// Example usage:
//
//	// Simple publish
//	hub.Publish(ctx,
//	    hub.T("type=alert", "priority=high"),
//	    "server is down",
//	)
//
//	// With options
//	hub.Publish(ctx,
//	    hub.T("type=metrics"),
//	    map[string]any{"cpu": 85, "mem": 45},
//	    hub.Wait(true),          // Wait for processing
//	    hub.OnFinish(func(ctx context.Context, e *hub.Event) {
//	        log.Println("Event processed")
//	    }),
//	)
//
// Notes:
// - For advanced event configuration, use PublishEvent directly
// - The payload will be automatically converted when subscribers use typed callbacks
// - Topic is required (use hub.T() to create topics)
// - Safe for concurrent use
func (h *Hub) Publish(ctx context.Context, topic *Topic, payload any, opts ...PublishOption) {
	h.PublishEvent(ctx, E().WithTopic(topic).WithPayload(payload), opts...)
}

// match finds subscriptions that match the event.
// Must be called while holding the Hub's read lock (h.RLock()).
func (h *Hub) match(t *Topic, cb func(s *sub)) int {
	// Collect potential candidate subscriptions lists
	candidates := make([]*sublist, 0)

	// Query indexes for each event attribute
	t.Each(func(k, v string) {
		// For any values add only list by key
		if v == Any {
			if sl, exists := h.indexKey[k]; exists {
				candidates = append(candidates, sl)
			}
			return
		}

		// Check exact value matches
		if vals, exists := h.indexKeyValue[k]; exists {
			if sl, exists := vals[v]; exists {
				candidates = append(candidates, sl)
			}
			// Check wildcard matches for this key
			if sl, exists := vals[Any]; exists {
				candidates = append(candidates, sl)
			}
		}
	})

	// Include subscriptions without topic attributes
	if h.indexEmpty.len() > 0 {
		candidates = append(candidates, h.indexEmpty)
	}

	var matched int
	for s := range mergeSubLists(candidates...) {
		if s.topic.Match(t) {
			matched++
			cb(s)
		}
	}
	return matched
}

// sync = true
func (h *Hub) publishEventSync(ctx context.Context, e *Event) {
	var unsub []SubID

	h.RLock()
	h.match(e.Topic(), func(s *sub) {
		_ = s.call(ctx, e)
		// handle limited subscription
		if s.shouldRemove() {
			unsub = append(unsub, s.id)
		}
	})
	h.RUnlock()

	e.finish(ctx)

	for _, sid := range unsub {
		h.Unsubscribe(ctx, sid)
	}

	return
}

// sync = false, wait = true
func (h *Hub) publishEventAsyncWait(ctx context.Context, e *Event) {
	var wg sync.WaitGroup

	h.RLock()
	h.match(e.Topic(), func(s *sub) {
		wg.Add(1)
		go func(s *sub) {
			_ = s.call(ctx, e)
			wg.Done()
			// handle limited subscription
			if s.shouldRemove() {
				// will remove after unlock
				h.Unsubscribe(ctx, s.id)
			}
		}(s)
	})
	h.RUnlock()

	wg.Wait()
	e.finish(ctx)
}

// sync = false, wait = false, hasOnFinish = true
func (h *Hub) publishEventAsyncNoWaitFinish(ctx context.Context, e *Event) {
	var wg sync.WaitGroup
	var once sync.Once

	h.RLock()
	n := h.match(e.Topic(), func(s *sub) {
		go func(s *sub) {
			_ = s.call(ctx, e)
			wg.Done()

			once.Do(func() {
				wg.Wait()
				e.finish(ctx)
			})

			// handle limited subscription
			if s.shouldRemove() {
				// will remove after unlock
				h.Unsubscribe(ctx, s.id)
			}
		}(s)
	})
	h.RUnlock()

	if n == 0 {
		go e.finish(ctx)
	}
}

// sync = false, wait = false, hasOnFinish = false
func (h *Hub) publishEventAsyncNoWaitNoFinish(ctx context.Context, e *Event) {
	// run all async and don't wait anything
	h.RLock()
	h.match(e.Topic(), func(s *sub) {
		go func(s *sub) {
			_ = s.call(ctx, e)
			// handle limited subscription
			if s.shouldRemove() {
				// will remove after unlock
				h.Unsubscribe(ctx, s.id)
			}
		}(s)
	})
	h.RUnlock()
}

// PublishEvent delivers an event to all matching subscribers
func (h *Hub) PublishEvent(ctx context.Context, e *Event, opts ...PublishOption) {
	for _, o := range opts {
		if o == nil {
			continue
		}
		e = o.modifyEvent(ctx, e)
	}

	if e.sync {
		h.publishEventSync(ctx, e)
		return
	}

	if e.wait {
		h.publishEventAsyncWait(ctx, e)
		return
	}

	if e.hasOnFinish() {
		h.publishEventAsyncNoWaitFinish(ctx, e)
		return
	}

	h.publishEventAsyncNoWaitNoFinish(ctx, e)
}

// Unsubscribe removes a subscription by ID
func (h *Hub) Unsubscribe(ctx context.Context, id SubID) {
	h.Lock()
	defer h.Unlock()

	// Find the subscription in the main list
	idx := h.all.find(id)
	if idx == -1 {
		return // Subscription not found
	}

	s := h.all.lst[idx] // Get the subscription

	// Remove from the main list first
	h.all.remove(id)

	// Remove from all key-value indexes
	s.topic.Each(func(k, v string) {
		// Remove from exact value index
		if vals, exists := h.indexKeyValue[k]; exists {
			if sl, exists := vals[v]; exists {
				sl.remove(id)

				// Cleanup empty sublists
				if sl.len() == 0 {
					delete(h.indexKeyValue[k], v)
				}
			}
		}

		// Remove from wildcard index
		if sl, exists := h.indexKey[k]; exists {
			sl.remove(id)

			// Cleanup empty sublists
			if sl.len() == 0 {
				delete(h.indexKey, k)
			}
		}
	})

	// Remove from empty topic index if needed
	if s.topic.Len() == 0 {
		h.indexEmpty.remove(id)
	}
}

// Clear removes all active subscriptions
func (h *Hub) Clear(ctx context.Context) {
	h.Lock()
	defer h.Unlock()

	h.all = &sublist{}
	h.indexKeyValue = make(map[string]map[string]*sublist)
	h.indexKey = make(map[string]*sublist)
	h.indexEmpty = &sublist{}

}

// Len returns current number of active subscriptions
func (h *Hub) Len() int {
	h.RLock()
	defer h.RUnlock()
	return h.all.len()
}
