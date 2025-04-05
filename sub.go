package hub

import "context"

type SubID uint64

type sub struct {
	id              SubID
	topic           *Topic
	callbackEvent   func(ctx context.Context, e *Event) error
	callbackPayload func(ctx context.Context, payload any) error
	once            bool
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
