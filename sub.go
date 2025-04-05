package hub

import "context"

type SubID uint64

type sub struct {
	id              SubID
	topic           *Topic
	callbackEvent   func(ctx context.Context, e *Event) error
	callbackPayload func(ctx context.Context, payload any) error
}
