package hub

import (
	"context"
	"sync/atomic"
)

type SubID uint64

type sub struct {
	counter atomic.Uint64
	id      SubID
	topic   *Topic
	handler Handler
	once    bool
}

func (s *sub) call(ctx context.Context, e *event) error {
	c := s.counter.Add(1)
	if s.once && c > 1 {
		return nil
	}
	if s.handler != nil {
		return s.handler(ctx, e.topic, e.payload)
	}
	return nil
}

func (s *sub) shouldRemove() bool {
	return s.once && s.counter.Load() > 0
}
