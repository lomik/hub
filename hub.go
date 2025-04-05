package hub

import (
	"context"
	"sync"
	"sync/atomic"
)

type Hub struct {
	sync.RWMutex
	seq  atomic.Uint64 // atomic counter for subscription IDs
	subs *sublist
}

func New() *Hub {
	return &Hub{
		subs: &sublist{},
	}
}

func (s *sub) call(ctx context.Context, e *Event) error {
	if s.callbackEvent != nil {
		return s.callbackEvent(ctx, e)
	}
	if s.callbackPayload != nil {
		return s.callbackPayload(ctx, e.Payload())
	}
	return nil
}

// SubscribeEvent registers a new event subscriber with topic matching
func (h *Hub) SubscribeEvent(ctx context.Context, t *Topic, cb func(ctx context.Context, e *Event) error) SubID {
	h.Lock()
	defer h.Unlock()

	id := SubID(h.seq.Add(1))
	newSub := &sub{
		id:            id,
		topic:         t,
		callbackEvent: cb,
	}

	h.subs.add(ctx, newSub)
	return id
}

// SubscribePayload registers a new payload subscriber with topic matching
func (h *Hub) SubscribePayload(ctx context.Context, t *Topic, cb func(ctx context.Context, payload any) error) SubID {
	h.Lock()
	defer h.Unlock()

	id := SubID(h.seq.Add(1))
	newSub := &sub{
		id:              id,
		topic:           t,
		callbackPayload: cb,
	}

	h.subs.add(ctx, newSub)
	return id
}

// PublishEvent delivers an event to all matching subscribers
func (h *Hub) PublishEvent(ctx context.Context, e *Event) {
	h.RLock()
	defer h.RUnlock()

	var wg sync.WaitGroup

	for _, s := range h.subs.lst {
		if s.topic.Match(e.Topic()) {
			wg.Add(1)
			go func(sub *sub) {
				defer wg.Done()

				_ = sub.call(ctx, e)
			}(s)
		}
	}

	if e.wait {
		wg.Wait()
	}
}

// Unsubscribe removes a subscription by ID
func (h *Hub) Unsubscribe(ctx context.Context, id SubID) {
	h.Lock()
	defer h.Unlock()
	h.subs.remove(ctx, id)
}

// Clear removes all active subscriptions
func (h *Hub) Clear(ctx context.Context) {
	h.Lock()
	defer h.Unlock()
	h.subs = &sublist{}
}

// Len returns current number of active subscriptions
func (h *Hub) Len() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.subs.lst)
}
