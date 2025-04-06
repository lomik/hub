package hub

import (
	"context"
	"sync/atomic"
)

type SubID uint64

type sub struct {
	counter       atomic.Uint64
	id            SubID
	topic         *Topic
	callbackEvent func(ctx context.Context, e *Event) error
	once          bool
}

func (s *sub) call(ctx context.Context, e *Event) error {
	c := s.counter.Add(1)
	if s.once && c > 1 {
		return nil
	}
	if s.callbackEvent != nil {
		return s.callbackEvent(ctx, e)
	}
	return nil
}

func (s *sub) shouldRemove() bool {
	return s.once && s.counter.Load() > 0
}
