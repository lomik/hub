package hub

import (
	"context"
	"sync"
	"sync/atomic"
)

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
func (h *Hub) SubscribeEvent(ctx context.Context, t *Topic, cb func(ctx context.Context, e *Event) error) SubID {
	h.Lock()
	defer h.Unlock()

	id := SubID(h.seq.Add(1))

	h.add(ctx, &sub{
		id:            id,
		topic:         t,
		callbackEvent: cb,
	})
	return id
}

// SubscribePayload registers a new payload subscriber with topic matching
func (h *Hub) SubscribePayload(ctx context.Context, t *Topic, cb func(ctx context.Context, payload any) error) SubID {
	h.Lock()
	defer h.Unlock()

	id := SubID(h.seq.Add(1))

	h.add(ctx, &sub{
		id:              id,
		topic:           t,
		callbackPayload: cb,
	})
	return id
}

// add adds a subscription to all relevant indexes
func (h *Hub) add(ctx context.Context, s *sub) {
	h.all.add(ctx, s)

	// Process each key-value pair in the topic
	s.topic.Each(func(k, v string) {
		// Initialize nested maps if needed
		if _, exists := h.indexKeyValue[k]; !exists {
			h.indexKeyValue[k] = make(map[string]*sublist)
		}
		if _, exists := h.indexKeyValue[k][v]; !exists {
			h.indexKeyValue[k][v] = &sublist{}
		}
		h.indexKeyValue[k][v].add(context.Background(), s)

		// Add to wildcard index for this key
		if _, exists := h.indexKey[k]; !exists {
			h.indexKey[k] = &sublist{}
		}
		h.indexKey[k].add(context.Background(), s)
	})

	// Handle empty topics
	if s.topic.Len() == 0 {
		h.indexEmpty.add(context.Background(), s)
	}
}

// PublishEvent delivers an event to all matching subscribers
func (h *Hub) PublishEvent(ctx context.Context, e *Event) {
	h.RLock()
	defer h.RUnlock()

	// Collect potential candidate subscriptions lists
	candidates := make([]*sublist, 0)

	// Query indexes for each event attribute
	e.Topic().Each(func(k, v string) {
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
	candidates = append(candidates, h.indexEmpty)

	// Process matching subscriptions in parallel
	var wg sync.WaitGroup
	for s := range mergeSubLists(candidates...) {
		if s.topic.Match(e.Topic()) {
			wg.Add(1)
			go func(s *sub) {
				defer wg.Done()
				_ = s.call(ctx, e)
			}(s)
		}
	}

	// Wait if event requires synchronous processing
	if e.wait {
		wg.Wait()
	}
}

// Unsubscribe removes a subscription by ID
func (h *Hub) Unsubscribe(ctx context.Context, id SubID) {
	panic("not implemented")
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
	return len(h.all.lst)
}
